package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic/platform/compliance/internal/config"
	"github.com/csic/platform/compliance/internal/handler"
	"github.com/csic/platform/compliance/internal/messaging"
	"github.com/csic/platform/compliance/internal/repository"
	"github.com/csic/platform/compliance/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Update logger level based on config
	if cfg.App.Debug {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	logger.Info("Starting Compliance Service",
		zap.String("name", cfg.App.Name),
		zap.String("environment", cfg.App.Environment))

	// Initialize database connection
	dbConfig := repository.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	repo, err := repository.NewPostgresRepository(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer repo.Close()

	// Initialize Redis connection
	redisConfig := repository.RedisConfig{
		Host:      cfg.Redis.Host,
		Port:      cfg.Redis.Port,
		Password:  cfg.Redis.Password,
		DB:        cfg.Redis.DB,
		KeyPrefix: cfg.Redis.KeyPrefix,
		PoolSize:  cfg.Redis.PoolSize,
	}

	redisClient, err := repository.NewRedisClient(redisConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka consumer
	kafkaConfig := messaging.KafkaConfig{
		Brokers:       cfg.Kafka.Brokers,
		ConsumerGroup: cfg.Kafka.ConsumerGroup,
		Topics:        cfg.Kafka.Topics,
	}

	kafkaConsumer, err := messaging.NewKafkaConsumer(kafkaConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Kafka consumer", zap.Error(err))
	}
	defer kafkaConsumer.Close()

	// Initialize Kafka producer for violations
	violationProducer, err := messaging.NewKafkaProducer(messaging.KafkaProducerConfig{
		Brokers:       cfg.Kafka.Brokers,
		RequiredAcks:  cfg.Kafka.Producer.RequiredAcks,
		RetryMax:      cfg.Kafka.Producer.RetryMax,
	}, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Kafka producer", zap.Error(err))
	}
	defer violationProducer.Close()

	// Initialize services
	complianceService := service.NewComplianceService(
		repo,
		redisClient,
		violationProducer,
		cfg,
		logger,
	)

	// Initialize HTTP handlers
	httpHandler := handler.NewHTTPHandler(complianceService, logger)

	// Initialize Kafka event handlers
	eventHandler := messaging.NewEventHandler(complianceService, logger)

	// Set up Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(handler.CORSMiddleware())

	// Register routes
	httpHandler.RegisterRoutes(router)

	// Start Kafka consumer in background
	go func() {
		if err := kafkaConsumer.Consume(context.Background(), eventHandler.HandleTransaction); err != nil {
			logger.Error("Kafka consumer error", zap.Error(err))
		}
	}()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Compliance Service listening",
			zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Compliance Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Compliance Service stopped")
}
