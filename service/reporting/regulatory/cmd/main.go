package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/csic/platform/service/reporting/regulatory/internal/generator"
	"github.com/csic/platform/service/reporting/regulatory/internal/handler"
	"github.com/csic/platform/service/reporting/regulatory/internal/messaging"
	"github.com/csic/platform/service/reporting/regulatory/internal/repository"
	"github.com/csic/platform/service/reporting/regulatory/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize database connection
	db, err := repository.NewDatabase(cfg.Database)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := repository.RunMigrations(db); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		os.Exit(1)
	}

	// Initialize Kafka producer
	kafkaProducer, err := messaging.NewKafkaProducer(cfg.Kafka)
	if err != nil {
		fmt.Printf("Failed to create Kafka producer: %v\n", err)
		os.Exit(1)
	}
	defer kafkaProducer.Close()

	// Initialize repositories
	reportRepo := repository.NewReportRepository(db, cfg.Redis.KeyPrefix)
	templateRepo := repository.NewTemplateRepository(db)

	// Initialize report generators
	reportGenerator := generator.NewReportGenerator(cfg, templateRepo)

	// Initialize services
	reportService := service.NewReportService(cfg, reportRepo, kafkaProducer, reportGenerator)
	scheduleService := service.NewScheduleService(cfg, reportService, kafkaProducer)
	templateService := service.NewTemplateService(cfg, templateRepo)

	// Initialize handlers
	reportHandler := handler.NewReportHandler(reportService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	templateHandler := handler.NewTemplateHandler(templateService)
	healthHandler := handler.NewHealthHandler(cfg)

	// Start scheduled report generation
	go scheduleService.StartScheduler(ctx)

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup routes
	setupRoutes(router, reportHandler, scheduleHandler, templateHandler, healthHandler, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("Starting Regulatory Reports Service on %s:%d\n", cfg.App.Host, cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited")
}

func setupRoutes(
	router *gin.Engine,
	reportHandler *handler.ReportHandler,
	scheduleHandler *handler.ScheduleHandler,
	templateHandler *handler.TemplateHandler,
	healthHandler *handler.HealthHandler,
	cfg *config.Config,
) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Report management endpoints
		reports := v1.Group("/reports")
		{
			reports.POST("", reportHandler.CreateReport)
			reports.GET("", reportHandler.ListReports)
			reports.GET("/:id", reportHandler.GetReport)
			reports.DELETE("/:id", reportHandler.DeleteReport)
			reports.POST("/:id/generate", reportHandler.GenerateReport)
			reports.GET("/:id/download", reportHandler.DownloadReport)
		}

		// Report types endpoints
		types := v1.Group("/report-types")
		{
			types.GET("", reportHandler.ListReportTypes)
			types.GET("/:type", reportHandler.GetReportType)
		}

		// Scheduled reports endpoints
		schedules := v1.Group("/schedules")
		{
			schedules.GET("", scheduleHandler.ListSchedules)
			schedules.POST("", scheduleHandler.CreateSchedule)
			schedules.GET("/:id", scheduleHandler.GetSchedule)
			schedules.PUT("/:id", scheduleHandler.UpdateSchedule)
			schedules.DELETE("/:id", scheduleHandler.DeleteSchedule)
			schedules.POST("/:id/trigger", scheduleHandler.TriggerSchedule)
		}

		// Template management endpoints
		templates := v1.Group("/templates")
		{
			templates.GET("", templateHandler.ListTemplates)
			templates.GET("/:id", templateHandler.GetTemplate)
			templates.POST("", templateHandler.CreateTemplate)
			templates.PUT("/:id", templateHandler.UpdateTemplate)
			templates.DELETE("/:id", templateHandler.DeleteTemplate)
		}
	}

	// Health endpoints
	router.GET("/health", healthHandler.GetHealth)
	router.GET("/health/live", healthHandler.LivenessCheck)
	router.GET("/health/ready", healthHandler.ReadinessCheck)
}
