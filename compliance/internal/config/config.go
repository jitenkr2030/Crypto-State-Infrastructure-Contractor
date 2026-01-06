package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the compliance service
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	RulesEngine   RulesEngineConfig   `mapstructure:"rules_engine"`
	ControlLayer  ControlLayerConfig  `mapstructure:"control_layer"`
	AuditLog      AuditLogConfig      `mapstructure:"audit_log"`
	Health        HealthConfig        `mapstructure:"health"`
	Metrics       MetricsConfig       `mapstructure:"metrics"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	LogLevel    string `mapstructure:"log_level"`
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db"`
	KeyPrefix string `mapstructure:"key_prefix"`
	PoolSize  int    `mapstructure:"pool_size"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers       []string           `mapstructure:"brokers"`
	ConsumerGroup string             `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig  `mapstructure:"topics"`
	Producer      KafkaProducerConfig `mapstructure:"producer"`
}

// KafkaTopicsConfig holds Kafka topic names
type KafkaTopicsConfig struct {
	Transactions string `mapstructure:"transactions"`
	Violations   string `mapstructure:"violations"`
	Audit        string `mapstructure:"audit"`
}

// KafkaProducerConfig holds Kafka producer configuration
type KafkaProducerConfig struct {
	RequiredAcks string `mapstructure:"required_acks"`
	RetryMax     int    `mapstructure:"retry_max"`
}

// RulesEngineConfig holds rules engine configuration
type RulesEngineConfig struct {
	CacheTTL           int  `mapstructure:"cache_ttl"`
	MaxRulesPerCheck   int  `mapstructure:"max_rules_per_check"`
	EvaluationTimeout  int  `mapstructure:"evaluation_timeout"`
	ParallelEvaluation bool `mapstructure:"parallel_evaluation"`
}

// ControlLayerConfig holds control layer integration configuration
type ControlLayerConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Endpoint   string `mapstructure:"endpoint"`
	Timeout    int    `mapstructure:"timeout"`
	RetryCount int    `mapstructure:"retry_count"`
}

// AuditLogConfig holds audit log integration configuration
type AuditLogConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	Timeout  int    `mapstructure:"timeout"`
	Async    bool   `mapstructure:"async"`
}

// HealthConfig holds health check configuration
type HealthConfig struct {
	Interval int `mapstructure:"interval"`
	Timeout  int `mapstructure:"timeout"`
	Retries  int `mapstructure:"retries"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	Port     int    `mapstructure:"port"`
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/csic/compliance/")

	// Read environment variables
	v.SetEnvPrefix("CSIC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for unset values
	cfg.applyDefaults()

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "compliance-service")
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.port", 8082)
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.log_level", "info")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.username", "csic_user")
	v.SetDefault("database.password", "csic_password")
	v.SetDefault("database.name", "csic_compliance")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 300)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.key_prefix", "csic:compliance:")
	v.SetDefault("redis.pool_size", 10)

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.consumer_group", "compliance-consumer")
	v.SetDefault("kafka.topics.transactions", "csic.transactions")
	v.SetDefault("kafka.topics.violations", "csic.compliance.violations")
	v.SetDefault("kafka.topics.audit", "csic.audit.events")
	v.SetDefault("kafka.producer.required_acks", "all")
	v.SetDefault("kafka.producer.retry_max", 3)

	// Rules engine defaults
	v.SetDefault("rules_engine.cache_ttl", 3600)
	v.SetDefault("rules_engine.max_rules_per_check", 100)
	v.SetDefault("rules_engine.evaluation_timeout", 5000)
	v.SetDefault("rules_engine.parallel_evaluation", true)

	// Control layer defaults
	v.SetDefault("control_layer.enabled", true)
	v.SetDefault("control_layer.endpoint", "http://localhost:8081")
	v.SetDefault("control_layer.timeout", 5000)
	v.SetDefault("control_layer.retry_count", 3)

	// Audit log defaults
	v.SetDefault("audit_log.enabled", true)
	v.SetDefault("audit_log.endpoint", "http://localhost:8080")
	v.SetDefault("audit_log.timeout", 5000)
	v.SetDefault("audit_log.async", true)

	// Health defaults
	v.SetDefault("health.interval", 30)
	v.SetDefault("health.timeout", 10)
	v.SetDefault("health.retries", 3)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.endpoint", "/metrics")
	v.SetDefault("metrics.port", 9090)
}

// applyDefaults applies default values for unset configuration
func (c *Config) applyDefaults() {
	if c.App.Name == "" {
		c.App.Name = "compliance-service"
	}
	if c.App.Host == "" {
		c.App.Host = "0.0.0.0"
	}
	if c.App.Port == 0 {
		c.App.Port = 8082
	}
	if c.App.Environment == "" {
		c.App.Environment = "development"
	}
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
	if c.Kafka.Producer.RetryMax == 0 {
		c.Kafka.Producer.RetryMax = 3
	}
	if c.RulesEngine.MaxRulesPerCheck == 0 {
		c.RulesEngine.MaxRulesPerCheck = 100
	}
	if c.ControlLayer.RetryCount == 0 {
		c.ControlLayer.RetryCount = 3
	}
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

// GetAddress returns the Redis address
func (c *RedisConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddress returns the server address
func (c *AppConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
