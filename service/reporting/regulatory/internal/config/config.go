package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App         AppConfig              `mapstructure:"app"`
	Database    DatabaseConfig         `mapstructure:"database"`
	Redis       RedisConfig            `mapstructure:"redis"`
	Kafka       KafkaConfig            `mapstructure:"kafka"`
	Reporting   ReportingConfig        `mapstructure:"reporting"`
	Schedules   map[string]ScheduleConfig `mapstructure:"schedules"`
	Regulatory  RegulatoryConfig       `mapstructure:"regulatory"`
	Templates   TemplateConfig         `mapstructure:"templates"`
	Logging     LoggingConfig          `mapstructure:"logging"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name        string   `mapstructure:"name"`
	Host        string   `mapstructure:"host"`
	Port        int      `mapstructure:"port"`
	Environment string   `mapstructure:"environment"`
	LogLevel    string   `mapstructure:"log_level"`
	APIKeys     []string `mapstructure:"api_keys"`
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db"`
	KeyPrefix string `mapstructure:"key_prefix"`
	PoolSize  int    `mapstructure:"pool_size"`
}

// KafkaConfig contains Kafka settings
type KafkaConfig struct {
	Brokers       []string             `mapstructure:"brokers"`
	ConsumerGroup string               `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig    `mapstructure:"topics"`
	Producer      KafkaProducerConfig  `mapstructure:"producer"`
}

// KafkaTopicsConfig contains Kafka topic names
type KafkaTopicsConfig struct {
	ComplianceEvents string `mapstructure:"compliance_events"`
	Transactions     string `mapstructure:"transactions"`
}

// KafkaProducerConfig contains Kafka producer settings
type KafkaProducerConfig struct {
	Acks    string `mapstructure:"acks"`
	Retries int    `mapstructure:"retries"`
}

// ReportingConfig contains report generation settings
type ReportingConfig struct {
	RetentionDays       int                 `mapstructure:"retention_days"`
	MaxRecordsPerReport int                 `mapstructure:"max_records_per_report"`
	GenerationTimeout   int                 `mapstructure:"generation_timeout"`
	Formats             []string            `mapstructure:"formats"`
	Storage             StorageConfig       `mapstructure:"storage"`
}

// StorageConfig contains report storage settings
type StorageConfig struct {
	Type     string      `mapstructure:"type"`
	BasePath string      `mapstructure:"base_path"`
	S3       S3Config    `mapstructure:"s3"`
}

// S3Config contains S3 storage settings
type S3Config struct {
	Enabled    bool   `mapstructure:"enabled"`
	Bucket     string `mapstructure:"bucket"`
	Region     string `mapstructure:"region"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
}

// ScheduleConfig contains scheduled report settings
type ScheduleConfig struct {
	Enabled    bool     `mapstructure:"enabled"`
	Cron       string   `mapstructure:"cron"`
	Recipients []string `mapstructure:"recipients"`
}

// RegulatoryConfig contains regulatory framework settings
type RegulatoryConfig struct {
	Frameworks []FrameworkConfig `mapstructure:"frameworks"`
}

// FrameworkConfig contains a regulatory framework definition
type FrameworkConfig struct {
	Name        string   `mapstructure:"name"`
	Version     string   `mapstructure:"version"`
	Requirements []string `mapstructure:"requirements"`
}

// TemplateConfig contains template settings
type TemplateConfig struct {
	Path         string `mapstructure:"path"`
	CacheEnabled bool   `mapstructure:"cache_enabled"`
	CacheTTL     int    `mapstructure:"cache_ttl"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Format        string `mapstructure:"format"`
	Output        string `mapstructure:"output"`
	Level         string `mapstructure:"level"`
	IncludeCaller bool   `mapstructure:"include_caller"`
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/csic/reports/")

	// Read from environment variables
	v.SetEnvPrefix("CSIC")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "regulatory-reports")
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.port", 8082)
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.log_level", "info")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 300)

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.key_prefix", "csic:reports:")
	v.SetDefault("redis.pool_size", 10)

	v.SetDefault("reporting.retention_days", 2555)
	v.SetDefault("reporting.max_records_per_report", 100000)
	v.SetDefault("reporting.generation_timeout", 300)

	v.SetDefault("templates.cache_enabled", true)
	v.SetDefault("templates.cache_ttl", 3600)

	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.level", "info")
}

// GetGenerationTimeout returns the report generation timeout as a duration
func (c *Config) GetGenerationTimeout() time.Duration {
	return time.Duration(c.Reporting.GenerationTimeout) * time.Second
}

// GetRetentionPeriod returns the report retention period as a duration
func (c *Config) GetRetentionPeriod() time.Duration {
	return time.Duration(c.Reporting.RetentionDays) * 24 * time.Hour
}

// GetTemplateCacheTTL returns the template cache TTL as a duration
func (c *Config) GetTemplateCacheTTL() time.Duration {
	return time.Duration(c.Templates.CacheTTL) * time.Second
}
