package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/csic-platform/services/api-gateway/router"
    "github.com/csic-platform/services/api-gateway/middleware"
    "github.com/csic-platform/services/api-gateway/auth"
    "github.com/gin-gonic/gin"
)

func main() {
    // Initialize configuration
    cfg := loadConfiguration()
    
    // Initialize services
    authService := auth.NewAuthService(cfg.JWT)
    authMiddleware := middleware.NewAuthMiddleware(authService)
    rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)
    logger := middleware.NewLogger(cfg.Logging)
    
    // Create router
    apiRouter := router.NewAPIRouter(authMiddleware, rateLimiter, logger)
    
    // Setup Gin router
    gin.SetMode(gin.ReleaseMode)
    router := gin.New()
    
    // Apply global middleware
    router.Use(gin.Recovery())
    router.Use(middleware.SecurityHeaders())
    router.Use(middleware.CORS())
    router.Use(middleware.RequestID())
    
    // Setup API routes
    apiRouter.SetupRoutes(router)
    
    // Create HTTP server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
        Handler:      router,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server in goroutine
    go func() {
        log.Printf("Starting API Gateway on port %d", cfg.Server.HTTPPort)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited gracefully")
}

func loadConfiguration() *Config {
    // Load configuration from environment or config file
    return &Config{
        Server: ServerConfig{
            HTTPPort: getEnvInt("API_GATEWAY_PORT", 8080),
        },
        JWT: JWTConfig{
            Secret: getEnv("JWT_SECRET", "default-secret-change-in-production"),
            Expiry: getEnvInt("JWT_EXPIRY_HOURS", 8),
        },
        RateLimit: RateLimitConfig{
            RequestsPerMinute: getEnvInt("RATE_LIMIT_RPM", 1000),
            Burst: getEnvInt("RATE_LIMIT_BURST", 100),
        },
        Logging: LoggingConfig{
            Level: getEnv("LOG_LEVEL", "INFO"),
            Format: "json",
        },
    }
}

type Config struct {
    Server    ServerConfig
    JWT       JWTConfig
    RateLimit RateLimitConfig
    Logging   LoggingConfig
}

type ServerConfig struct {
    HTTPPort int
}

type JWTConfig struct {
    Secret string
    Expiry int
}

type RateLimitConfig struct {
    RequestsPerMinute int
    Burst             int
}

type LoggingConfig struct {
    Level  string
    Format string
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        var result int
        if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
            return result
        }
    }
    return defaultValue
}
