package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"audit-log/internal/adapter/messaging"
	"audit-log/internal/adapter/repository"
	"audit-log/internal/config"
	"audit-log/internal/core/ports"
	"audit-log/internal/core/service"
	"audit-log/internal/handler"
)

func main() {
	// Load configuration
	configPath := "config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize logger
	logger := initLogger(cfg.App.LogLevel)
	logger.Info("Starting Audit Log Service", "name", cfg.App.Name, "env", cfg.App.Environment)

	// Initialize database connection
	db, err := initDatabase(cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewPostgresAuditRepository(db, logger)

	// Initialize Kafka producer
	var producer ports.KafkaProducer
	if cfg.Messaging.Type == "kafka" {
		producer = messaging.NewKafkaAuditProducer(cfg.Messaging.Brokers, cfg.Messaging.TopicPrefix, logger)

		// Ensure Kafka topics exist
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := messaging.EnsureTopics(ctx, cfg.Messaging.Brokers, []string{
			"audit.entries",
			"audit.verifications",
		}, cfg.Messaging.TopicPrefix); err != nil {
			logger.Warn("Failed to ensure Kafka topics", "error", err)
		}

		logger.Info("Kafka producer initialized", "brokers", cfg.Messaging.Brokers)
	}

	// Initialize service
	auditService := service.NewAuditLogService(repo, producer, logger)

	// Initialize handler
	auditHandler := handler.NewAuditHandler(auditService, logger)

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS middleware
	router.Use(corsMiddleware())

	// Security middleware
	if cfg.Security.AuthEnabled {
		router.Use(authMiddleware(cfg.Security.JWTSecret, logger))
	}

	// Rate limiting middleware
	router.Use(rateLimitMiddleware(cfg.Security.RateLimit, logger))

	// Register routes
	auditHandler.RegisterRoutes(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		auditHandler.HealthCheck(c)
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close producer
	if producer != nil {
		producer.Close()
	}

	logger.Info("Server stopped")
}

func initDatabase(cfg config.DatabaseConfig, logger ports.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.GetConnMaxLifeDuration())

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected", "host", cfg.Host, "database", cfg.Name)
	return db, nil
}

func initLogger(level string) ports.Logger {
	return &stdLogger{level: level}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func authMiddleware(jwtSecret string, logger ports.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header", "path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// In a real implementation, validate JWT token here
		c.Set("user_id", "demo-user")
		c.Next()
	}
}

func rateLimitMiddleware(requestsPerMinute int, logger ports.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// stdLogger implements ports.Logger using standard log package
type stdLogger struct {
	level string
}

func (l *stdLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Printf("[INFO] "+msg, keysAndValues...)
}

func (l *stdLogger) Error(msg string, keysAndValues ...interface{}) {
	log.Printf("[ERROR] "+msg, keysAndValues...)
}

func (l *stdLogger) Warn(msg string, keysAndValues ...interface{}) {
	log.Printf("[WARN] "+msg, keysAndValues...)
}

func (l *stdLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.level == "debug" {
		log.Printf("[DEBUG] "+msg, keysAndValues...)
	}
}
