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

	"github.com/csic-platform/services/internal/config"
	"github.com/csic-platform/services/internal/grpc/interceptors"
	"github.com/csic-platform/services/internal/handlers"
	"github.com/csic-platform/services/internal/middleware"
	"github.com/csic-platform/services/internal/repository"
	"github.com/csic-platform/services/internal/services"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// 全局变量用于优雅关闭
var (
	server      *http.Server
	grpcServer  *grpc.Server
	db          *sql.DB
	wormStorage *services.WORMStorage
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库连接
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 初始化WORM存储用于不可变日志
	wormStorage, err = services.NewWORMStorage(cfg.WORMStorage.Path)
	if err != nil {
		log.Fatalf("初始化WORM存储失败: %v", err)
	}

	// 初始化仓储层
	repos := repository.NewRepositories(db)

	// 初始化服务层
	svc := services.NewServices(cfg, repos, wormStorage)

	// 初始化处理器
	h := handlers.NewHandlers(svc)

	// 设置中间件
	authMiddleware := middleware.NewAuthMiddleware(svc)
	loggingMiddleware := middleware.NewLoggingMiddleware(wormStorage)
	auditMiddleware := middleware.NewAuditMiddleware(wormStorage)

	// 启动后台服务
	go startBackgroundServices(svc)

	// 启动HTTP服务器
	go startHTTPServer(cfg, h, authMiddleware, loggingMiddleware, auditMiddleware)

	// 启动gRPC服务器
	go startGRPCServer(cfg, h, authMiddleware)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP服务器关闭错误: %v", err)
	}

	grpcServer.GracefulStop()
	wormStorage.Close()

	log.Println("服务器已优雅停止")
}

func initDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库Ping失败: %w", err)
	}

	log.Println("数据库连接成功")
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

	// 健康检查端点
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API v1路由
	v1 := router.Group("/api/v1")
	{
		// 核心控制层
		v1.GET("/status", h.GetSystemStatus)
		v1.POST("/emergency/stop", h.EmergencyStop)
		v1.POST("/emergency/resume", h.EmergencyResume)

		// 交易所监管
		v1.GET("/exchanges", h.GetExchanges)
		v1.GET("/exchanges/:id", h.GetExchangeDetails)
		v1.POST("/exchanges/:id/freeze", h.FreezeExchange)
		v1.POST("/exchanges/:id/thaw", h.ThawExchange)
		v1.GET("/exchanges/:id/metrics", h.GetExchangeMetrics)

		// 钱包治理
		v1.GET("/wallets", h.GetWallets)
		v1.POST("/wallets", h.CreateWallet)
		v1.POST("/wallets/:id/freeze", h.FreezeWallet)
		v1.POST("/wallets/:id/unfreeze", h.UnfreezeWallet)
		v1.POST("/wallets/:id/transfer", h.TransferFromWallet)

		// 交易监控
		v1.GET("/transactions", h.GetTransactions)
		v1.GET("/transactions/:id", h.GetTransactionDetails)
		v1.GET("/transactions/search", h.SearchTransactions)
		v1.POST("/transactions/flag", h.FlagTransaction)
		v1.GET("/risk/score/:address", h.GetWalletRiskScore)

		// 许可合规
		v1.GET("/licenses", h.GetLicenses)
		v1.POST("/licenses", h.CreateLicense)
		v1.GET("/licenses/:id", h.GetLicenseDetails)
		v1.PUT("/licenses/:id", h.UpdateLicense)
		v1.POST("/licenses/:id/revoke", h.RevokeLicense)

		// 挖矿控制
		v1.GET("/miners", h.GetMiners)
		v1.GET("/miners/:id", h.GetMinerDetails)
		v1.POST("/miners/:id/shutdown", h.ShutdownMiner)
		v1.POST("/miners/:id/start", h.StartMiner)
		v1.GET("/mining/metrics", h.GetMiningMetrics)

		// 能源集成
		v1.GET("/energy/grid", h.GetGridStatus)
		v1.GET("/energy/consumption", h.GetEnergyConsumption)
		v1.POST("/energy/load-shedding", h.TriggerLoadShedding)

		// 报告生成
		v1.GET("/reports", h.GetReports)
		v1.POST("/reports/generate", h.GenerateReport)
		v1.GET("/reports/:id/download", h.DownloadReport)

		// 安全审计
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

	log.Printf("HTTP服务器启动于端口 %d", cfg.Server.HTTPPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP服务器失败: %v", err)
	}
}

func startGRPCServer(cfg *config.Config, h *handlers.Handlers, authMiddleware *middleware.AuthMiddleware) {
	// 加载TLS证书
	cert, err := tls.LoadX509KeyPair(cfg.TLSCert.CertFile, cfg.TLSCert.KeyFile)
	if err != nil {
		log.Fatalf("加载TLS证书失败: %v", err)
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    loadCACertPool(cfg.TLSCert.CAFile),
	}

	// 创建带凭证的gRPC服务器
	grpcServer = grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor,
			interceptors.AuthInterceptor(authMiddleware),
			interceptors.AuditInterceptor,
		),
	)

	// 注册服务（示例）
	// pb.RegisterCoreControlServer(grpcServer, coreService)

	// 开始监听
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("监听gRPC端口失败: %v", err)
	}

	log.Printf("gRPC服务器启动于端口 %d", cfg.Server.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC服务器失败: %v", err)
	}
}

func loadCACertPool(caFile string) *x509.CertPool {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		log.Printf("警告: 无法加载CA证书: %v", err)
		return nil
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		log.Printf("警告: 无法附加CA证书")
		return nil
	}

	return certPool
}

func startBackgroundServices(svc *services.Services) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 更新系统指标
		if err := svc.UpdateSystemMetrics(); err != nil {
			log.Printf("更新系统指标错误: %v", err)
		}

		// 处理待处理警报
		if err := svc.ProcessAlerts(); err != nil {
			log.Printf("处理警报错误: %v", err)
		}

		// 同步区块链节点
		if err := svc.SyncBlockchainNodes(); err != nil {
			log.Printf("同步区块链节点错误: %v", err)
		}
	}
}
