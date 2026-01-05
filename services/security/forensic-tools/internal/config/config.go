package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
	Messaging MessagingConfig `yaml:"messaging"`
	Security SecurityConfig `yaml:"security"`
}

// AppConfig represents application settings
type AppConfig struct {
	Name         string `yaml:"name"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Environment  string `yaml:"environment"`
	Debug        bool   `yaml:"debug"`
	LogLevel     string `yaml:"log_level"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// DatabaseConfig represents database connection settings
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Name         string `yaml:"name"`
	SSLMode      string `yaml:"ssl_mode"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	ConnMaxLife  int    `yaml:"conn_max_life"` // in seconds
}

// DSN returns the PostgreSQL connection string
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Name, d.SSLMode,
	)
}

// GetConnMaxLifeDuration returns connection max lifetime as duration
func (d *DatabaseConfig) GetConnMaxLifeDuration() time.Duration {
	return time.Duration(d.ConnMaxLife) * time.Second
}

// StorageConfig represents storage settings
type StorageConfig struct {
	Type     string `yaml:"type"`
	BasePath string `yaml:"base_path"` // For local storage
	Bucket   string `yaml:"bucket"`    // For S3 storage
	Prefix   string `yaml:"prefix"`    // For S3 storage
	MaxSize  int64  `yaml:"max_size"`  // Max evidence size in bytes
}

// MessagingConfig represents messaging settings
type MessagingConfig struct {
	Type        string   `yaml:"type"`
	Brokers     []string `yaml:"brokers"`
	TopicPrefix string   `yaml:"topic_prefix"`
	Consumer    string   `yaml:"consumer"`
}

// SecurityConfig represents security settings
type SecurityConfig struct {
	JWTSecret    string `yaml:"jwt_secret"`
	AuthEnabled  bool   `yaml:"auth_enabled"`
	RequireMFA   bool   `yaml:"require_mfa"`
	AllowedIPs   []string `yaml:"allowed_ips"`
	RateLimit    int    `yaml:"rate_limit"` // requests per minute
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	config.applyDefaults()

	// Apply environment overrides
	config.applyEnvOverrides()

	return config, nil
}

// applyDefaults sets default values for configuration
func (c *Config) applyDefaults() {
	// App defaults
	if c.App.Name == "" {
		c.App.Name = "forensic-tools"
	}
	if c.App.Host == "" {
		c.App.Host = "0.0.0.0"
	}
	if c.App.Port == 0 {
		c.App.Port = 8080
	}
	if c.App.Environment == "" {
		c.App.Environment = "development"
	}
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.App.ReadTimeout == 0 {
		c.App.ReadTimeout = 30
	}
	if c.App.WriteTimeout == 0 {
		c.App.WriteTimeout = 30
	}

	// Database defaults
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLife == 0 {
		c.Database.ConnMaxLife = 300
	}

	// Storage defaults
	if c.Storage.Type == "" {
		c.Storage.Type = "local"
	}
	if c.Storage.BasePath == "" {
		c.Storage.BasePath = "/data/forensic/evidence"
	}
	if c.Storage.MaxSize == 0 {
		c.Storage.MaxSize = 10 * 1024 * 1024 * 1024 // 10 GB
	}

	// Messaging defaults
	if c.Messaging.Type == "" {
		c.Messaging.Type = "kafka"
	}
	if c.Messaging.TopicPrefix == "" {
		c.Messaging.TopicPrefix = "forensic"
	}

	// Security defaults
	if c.Security.RateLimit == 0 {
		c.Security.RateLimit = 100
	}
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	// Database overrides
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Database.Port)
	}
	if user := os.Getenv("DB_USERNAME"); user != "" {
		c.Database.Username = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		c.Database.Password = pass
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		c.Database.Name = name
	}

	// Storage overrides
	if basePath := os.Getenv("STORAGE_PATH"); basePath != "" {
		c.Storage.BasePath = basePath
	}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		c.Storage.Bucket = bucket
	}

	// Messaging overrides
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		c.Messaging.Brokers = []string{brokers}
	}

	// Security overrides
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.Security.JWTSecret = secret
	}

	// App overrides
	if port := os.Getenv("APP_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.App.Port)
	}
	if host := os.Getenv("APP_HOST"); host != "" {
		c.App.Host = host
	}
	if debug := os.Getenv("APP_DEBUG"); debug == "true" {
		c.App.Debug = true
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Storage.BasePath == "" && c.Storage.Type == "local" {
		return fmt.Errorf("storage base path is required for local storage")
	}
	return nil
}
