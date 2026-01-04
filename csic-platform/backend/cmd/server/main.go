package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/backend/internal/config"
	"github.com/csic-platform/backend/internal/grpc/interceptors"
	"github.com/csic-platform/backend/internal/handlers"
	"github.com/csic-platform/backend/internal/middleware"
	"github.com/csic-platform/backend/internal/repository"
	"github.com/csic-platform/backend/internal/services"
	"github.com/csic-platform/backend/api/proto"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Global variables for graceful shutdown
var (
	server      *http.Server
	grpcServer  *grpc.Server
	db          *sql.DB
	wormStorage *services.WORMStorage
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize WORM storage for immutable logs
	wormStorage, err = services.NewWORMStorage(cfg.WORMStorage.Path)
	if err != nil {
		log.Fatalf("Failed to initialize WORM storage: %v", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	svc := services.NewServices(cfg, repos, wormStorage)

	// Initialize handlers
	h := handlers.NewHandlers(svc)

	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(svc)
	loggingMiddleware := middleware.NewLoggingMiddleware(wormStorage)
	auditMiddleware := middleware.NewAuditMiddleware(wormStorage)

	// Start background services
	go startBackgroundServices(svc)

	// Start HTTP server
	go startHTTPServer(cfg, h, authMiddleware, loggingMiddleware, auditMiddleware)

	// Start gRPC server
	go startGRPCServer(cfg, h, authMiddleware)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
	wormStorage.Close()

	log.Println("Servers stopped gracefully")
}

func initDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return db, nil
}

func startHTTPServer(
	cfg *config.Config,
	h *handlers.Handlers,
	authMiddleware *middleware.AuthMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	auditMiddleware *middleware.AuditMiddleware,
) {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware.Logger())
	router.Use(authMiddleware.Authenticate())

	// Health check endpoint
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Core Control Layer
		v1.GET("/status", h.GetSystemStatus)
		v1.POST("/emergency/stop", h.EmergencyStop)
		v1.POST("/emergency/resume", h.EmergencyResume)

		// Exchange Oversight
		v1.GET("/exchanges", h.GetExchanges)
		v1.GET("/exchanges/:id", h.GetExchangeDetails)
		v1.POST("/exchanges/:id/freeze", h.FreezeExchange)
		v1.POST("/exchanges/:id/thaw", h.ThawExchange)
		v1.GET("/exchanges/:id/metrics", h.GetExchangeMetrics)

		// Wallet Governance
		v1.GET("/wallets", h.GetWallets)
		v1.POST("/wallets", h.CreateWallet)
		v1.POST("/wallets/:id/freeze", h.FreezeWallet)
		v1.POST("/wallets/:id/unfreeze", h.UnfreezeWallet)
		v1.POST("/wallets/:id/transfer", h.TransferFromWallet)

		// Transaction Monitoring
		v1.GET("/transactions", h.GetTransactions)
		v1.GET("/transactions/:id", h.GetTransactionDetails)
		v1.GET("/transactions/search", h.SearchTransactions)
		v1.POST("/transactions/flag", h.FlagTransaction)
		v1.GET("/risk/score/:address", h.GetWalletRiskScore)

		// Licensing & Compliance
		v1.GET("/licenses", h.GetLicenses)
		v1.POST("/licenses", h.CreateLicense)
		v1.GET("/licenses/:id", h.GetLicenseDetails)
		v1.PUT("/licenses/:id", h.UpdateLicense)
		v1.POST("/licenses/:id/revoke", h.RevokeLicense)

		// Mining Control
		v1.GET("/miners", h.GetMiners)
		v1.GET("/miners/:id", h.GetMinerDetails)
		v1.POST("/miners/:id/shutdown", h.ShutdownMiner)
		v1.POST("/miners/:id/start", h.StartMiner)
		v1.GET("/mining/metrics", h.GetMiningMetrics)

		// Energy Integration
		v1.GET("/energy/grid", h.GetGridStatus)
		v1.GET("/energy/consumption", h.GetEnergyConsumption)
		v1.POST("/energy/load-shedding", h.TriggerLoadShedding)

		// Reporting
		v1.GET("/reports", h.GetReports)
		v1.POST("/reports/generate", h.GenerateReport)
		v1.GET("/reports/:id/download", h.DownloadReport)

		// Security & Audit
		v1.GET("/audit/logs", h.GetAuditLogs)
		v1.GET("/audit/logs/:id", h.GetAuditLogDetails)
		v1.POST("/audit/export", h.ExportAuditLogs)
		v1.GET("/security/hsm/status", h.GetHSMStatus)
		v1.POST("/security/keys/rotate", h.RotateKeys)
	}

	server = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("HTTP server starting on port %d", cfg.Server.HTTPPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func startGRPCServer(cfg *config.Config, h *handlers.Handlers, authMiddleware *middleware.AuthMiddleware) {
	// Load TLS certificates
	cert, err := tls.LoadX509KeyPair(cfg.TLSCert.CertFile, cfg.TLSCert.KeyFile)
	if err != nil {
		log.Fatalf("Failed to load TLS certificates: %v", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    loadCACertPool(cfg.TLSCert.CAFile),
	}

	// Create gRPC server with credentials
	grpcServer = grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor,
			interceptors.AuthInterceptor(authMiddleware),
			interceptors.AuditInterceptor,
		),
	)

	// Register services
	coreService := services.NewCoreControlService()
	proto.RegisterCoreControlServer(grpcServer, coreService)

	exchangeService := services.NewExchangeOversightService()
	proto.RegisterExchangeOversightServer(grpcServer, exchangeService)

	walletService := services.NewWalletGovernanceService()
	proto.RegisterWalletGovernanceServer(grpcServer, walletService)

	monitoringService := services.NewTransactionMonitoringService()
	proto.RegisterTransactionMonitoringServer(grpcServer, monitoringService)

	// Start listening
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	log.Printf("gRPC server starting on port %d", cfg.Server.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}

func loadCACertPool(caFile string) *x509.CertPool {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		log.Printf("Warning: Could not load CA cert: %v", err)
		return nil
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		log.Printf("Warning: Could not append CA cert")
		return nil
	}

	return certPool
}

func startBackgroundServices(svc *services.Services) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Update system metrics
		if err := svc.UpdateSystemMetrics(); err != nil {
			log.Printf("Error updating system metrics: %v", err)
		}

		// Process pending alerts
		if err := svc.ProcessAlerts(); err != nil {
			log.Printf("Error processing alerts: %v", err)
		}

		// Sync with blockchain nodes
		if err := svc.SyncBlockchainNodes(); err != nil {
			log.Printf("Error syncing blockchain nodes: %v", err)
		}
	}
}
