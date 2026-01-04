package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Redis       RedisConfig       `yaml:"redis"`
	Kafka       KafkaConfig       `yaml:"kafka"`
	WORMStorage WORMStorageConfig `yaml:"worm_storage"`
	HSM         HSMConfig         `yaml:"hsm"`
	TLSCert     TLSCertConfig     `yaml:"tls_cert"`
	Blockchain  BlockchainConfig  `yaml:"blockchain"`
	Security    SecurityConfig    `yaml:"security"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
}

// ServerConfig contains HTTP and gRPC server settings
type ServerConfig struct {
	HTTPPort     int    `yaml:"http_port"`
	GRPCPort     int    `yaml:"grpc_port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	Env          string `yaml:"env"`
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Name           string `yaml:"name"`
	SSLMode        string `yaml:"ssl_mode"`
	MaxOpenConns   int    `yaml:"max_open_conns"`
	MaxIdleConns   int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int   `yaml:"conn_max_lifetime"`
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// KafkaConfig contains Kafka broker settings
type KafkaConfig {
	Brokers        []string `yaml:"brokers"`
	ConsumerGroup  string   `yaml:"consumer_group"`
	Topics         TopicsConfig `yaml:"topics"`
	SecurityConfig SecurityConfig `yaml:"security_config"`
}

// TopicsConfig contains Kafka topic names
type TopicsConfig struct {
	Transactions  string `yaml:"transactions"`
	Alerts        string `yaml:"alerts"`
	AuditLogs     string `yaml:"audit_logs"`
	ExchangeData  string `yaml:"exchange_data"`
	MiningMetrics string `yaml:"mining_metrics"`
}

// WORMStorageConfig contains WORM storage settings
type WORMStorageConfig struct {
	Path          string `yaml:"path"`
	RetentionDays int    `yaml:"retention_days"`
	Compression   bool   `yaml:"compression"`
}

// HSMConfig contains Hardware Security Module settings
type HSMConfig struct {
	Provider      string `yaml:"provider"`
	Slot          int    `yaml:"slot"`
	Pin           string `yaml:"pin"`
	KeyLabel      string `yaml:"key_label"`
	LibraryPath   string `yaml:"library_path"`
}

// TLSCertConfig contains TLS certificate paths
type TLSCertConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

// BlockchainConfig contains blockchain node settings
type BlockchainConfig struct {
	Bitcoin struct {
		RPCURL      string `yaml:"rpc_url"`
		RPCUser     string `yaml:"rpc_user"`
		RPCPassword string `yaml:"rpc_password"`
		WalletName  string `yaml:"wallet_name"`
	} `yaml:"bitcoin"`
	Ethereum struct {
		RPCURL      string `yaml:"rpc_url"`
		RPCUser     string `yaml:"rpc_user"`
		RPCPassword string `yaml:"rpc_password"`
	} `yaml:"ethereum"`
	SyncInterval int `yaml:"sync_interval"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	JWTSecret        string `yaml:"jwt_secret"`
	TokenExpiryHours int    `yaml:"token_expiry_hours"`
	RateLimit        int    `yaml:"rate_limit"`
	MaxFailedLogins  int    `yaml:"max_failed_logins"`
	LockoutDuration  int    `yaml:"lockout_duration"`
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	MetricsEnabled  bool `yaml:"metrics_enabled"`
	TracingEnabled  bool `yaml:"tracing_enabled"`
	HealthCheckIntv int  `yaml:"health_check_interval"`
	AlertThreshold  int  `yaml:"alert_threshold"`
}

// LoadConfig loads configuration from the specified YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to configuration
func (c *Config) applyEnvOverrides() {
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Database.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Database.Password = password
	}
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		c.Redis.Host = redisHost
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		c.Security.JWTSecret = jwtSecret
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.Server.HTTPPort)
	}
	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", c.Server.GRPCPort)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.WORMStorage.Path == "" {
		return fmt.Errorf("WORM storage path is required")
	}
	return nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
