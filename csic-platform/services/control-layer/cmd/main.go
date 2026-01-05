package main

import (
	"log"
	"os"

	"csic-platform/control-layer/internal/adapters/handlers"
	"csic-platform/control-layer/internal/adapters/messaging"
	"csic-platform/control-layer/internal/adapters/storage"
	"csic-platform/control-layer/internal/config"
	"csic-platform/control-layer/internal/core/ports"
	"csic-platform/control-layer/internal/core/services"
	"csic-platform/control-layer/pkg/logger"
	"csic-platform/control-layer/pkg/metrics"
)

func main() {
	// Initialize logger
	logLogger := logger.NewLogger()
	
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger with config
	zapLogger, err := logger.NewZapLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting Control Layer Service",
		logger.String("service", "control-layer"),
		logger.String("environment", cfg.Environment),
	)

	// Initialize metrics
	metricsCollector := metrics.NewPrometheusCollector("control_layer")
	
	// Initialize PostgreSQL repository
	policyRepo, err := storage.NewPostgresPolicyRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to PostgreSQL", logger.Error(err))
	}
	defer policyRepo.Close()
	
	stateRepo, err := storage.NewPostgresStateRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to state repository", logger.Error(err))
	}
	defer stateRepo.Close()
	
	enforcementRepo, err := storage.NewPostgresEnforcementRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to enforcement repository", logger.Error(err))
	}
	defer enforcementRepo.Close()

	// Initialize Redis client
	redisClient, err := storage.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		zapLogger.Fatal("Failed to connect to Redis", logger.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka producer
	kafkaProducer, err := messaging.NewKafkaProducer(cfg.KafkaBrokers)
	if err != nil {
		zapLogger.Fatal("Failed to create Kafka producer", logger.Error(err))
	}
	defer kafkaProducer.Close()

	// Initialize Kafka consumer for policy updates
	kafkaConsumer, err := messaging.NewKafkaConsumer(cfg.KafkaBrokers, "control-layer-policy-updates")
	if err != nil {
		zapLogger.Fatal("Failed to create Kafka consumer", logger.Error(err))
	}
	defer kafkaConsumer.Close()

	// Initialize repositories ports
	repositories := ports.Repositories{
		PolicyRepository:     policyRepo,
		StateRepository:      stateRepo,
		EnforcementRepository: enforcementRepo,
	}

	// Initialize cache port
	cachePort := ports.CachePort{
		Client: redisClient,
	}

	// Initialize messaging port
	messagingPort := ports.MessagingPort{
		Producer: kafkaProducer,
		Consumer: kafkaConsumer,
	}

	// Initialize services
	policyEngine := services.NewPolicyEngine(repositories, cachePort, zapLogger)
	enforcementHandler := services.NewEnforcementHandler(repositories, messagingPort, zapLogger)
	stateRegistry := services.NewStateRegistry(repositories, cachePort, zapLogger)
	interventionService := services.NewInterventionService(repositories, messagingPort, zapLogger, policyEngine)

	// Initialize HTTP handler
	httpHandler := handlers.NewHTTPHandler(
		policyEngine,
		enforcementHandler,
		stateRegistry,
		interventionService,
		metricsCollector,
		zapLogger,
	)

	// Initialize gRPC handler
	grpcHandler := handlers.NewGRPCHandler(
		policyEngine,
		enforcementHandler,
		stateRegistry,
		interventionService,
		metricsCollector,
		zapLogger,
	)

	// Start policy update consumer
	go policyEngine.StartPolicyUpdateConsumer(zapLogger)

	// Start intervention monitor
	go interventionService.StartInterventionMonitor(zapLogger)

	// Start HTTP server
	go httpHandler.Start(cfg.HTTPPort, zapLogger)

	// Start gRPC server
	grpcHandler.Start(cfg.GRPCPort, zapLogger)

	zapLogger.Info("Control Layer Service started successfully",
		logger.Int("http_port", cfg.HTTPPort),
		logger.Int("grpc_port", cfg.GRPCPort),
	)

	// Wait for shutdown signal
	handlers.WaitForShutdown(zapLogger)
}
