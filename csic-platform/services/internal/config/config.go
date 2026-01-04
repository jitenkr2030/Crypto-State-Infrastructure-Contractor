package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 表示完整的应用程序配置
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	TimescaleDB TimescaleDBConfig `yaml:"timescaledb"`
	Redis       RedisConfig       `yaml:"redis"`
	OpenSearch  OpenSearchConfig  `yaml:"opensearch"`
	WORMStorage WORMStorageConfig `yaml:"worm_storage"`
	Kafka       KafkaConfig       `yaml:"kafka"`
	Blockchain  BlockchainConfig  `yaml:"blockchain"`
	HSM         HSMConfig         `yaml:"hsm"`
	TLSCert     TLSCertConfig     `yaml:"tls"`
	Security    SecurityConfig    `yaml:"security"`
	Regulatory  RegulatoryConfig  `yaml:"regulatory"`
	Mining      MiningConfig      `yaml:"mining"`
	Exchange    ExchangeConfig    `yaml:"exchange"`
	Logging     LoggingConfig     `yaml:"logging"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	DR          DRConfig          `yaml:"disaster_recovery"`
}

// ServerConfig 包含HTTP和gRPC服务器设置
type ServerConfig struct {
	HTTPPort         int    `yaml:"http_port"`
	GRPCPort         int    `yaml:"grpc_port"`
	ReadTimeout      int    `yaml:"read_timeout"`
	WriteTimeout     int    `yaml:"write_timeout"`
	Env              string `yaml:"env"`
	MaxRequestSize   int64  `yaml:"max_request_size"`
	RateLimit        RateLimitConfig `yaml:"rate_limit"`
}

// RateLimitConfig 包含速率限制设置
type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

// DatabaseConfig 包含PostgreSQL连接设置
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	PoolMode        string `yaml:"pool_mode"`
}

// TimescaleDBConfig 包含TimescaleDB连接设置
type TimescaleDBConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Name         string `yaml:"name"`
	SSLMode      string `yaml:"ssl_mode"`
	Retention    string `yaml:"retention"`
	Compression  bool   `yaml:"compression"`
}

// RedisConfig 包含Redis连接设置
type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MaxRetries   int    `yaml:"max_retries"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// OpenSearchConfig 包含OpenSearch设置
type OpenSearchConfig struct {
	Hosts          []string    `yaml:"hosts"`
	IndexPrefix    string      `yaml:"index_prefix"`
	RetentionDays  int         `yaml:"retention_days"`
	SSLEnabled     bool        `yaml:"ssl_enabled"`
	Auth           AuthConfig  `yaml:"auth"`
}

// AuthConfig 包含认证配置
type AuthConfig struct {
	Username     string `yaml:"username"`
	UsernameFile string `yaml:"username_file"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`
}

// WORMStorageConfig 包含WORM存储设置
type WORMStorageConfig struct {
	Path          string `yaml:"primary_path"`
	ReplicaPath   string `yaml:"replica_path"`
	RetentionDays int    `yaml:"retention_days"`
	Encryption    string `yaml:"encryption"`
	WriteOnce     bool   `yaml:"write_once"`
}

// KafkaConfig 包含Kafka代理设置
type KafkaConfig struct {
	Brokers       []string        `yaml:"brokers"`
	ConsumerGroup string          `yaml:"consumer_group"`
	Topics        KafkaTopics     `yaml:"topics"`
	Security      KafkaSecurity   `yaml:"security"`
}

// KafkaTopics 包含Kafka主题名称
type KafkaTopics struct {
	Transactions  string `yaml:"transactions"`
	Alerts        string `yaml:"alerts"`
	AuditLogs     string `yaml:"audit_logs"`
	ExchangeData  string `yaml:"exchange_data"`
	MiningMetrics string `yaml:"mining_metrics"`
}

// KafkaSecurity 包含Kafka安全设置
type KafkaSecurity struct {
	SASLMechanism string `yaml:"sasl_mechanism"`
	TLSEnabled    bool   `yaml:"tls_enabled"`
}

// BlockchainConfig 包含区块链节点设置
type BlockchainConfig struct {
	Bitcoin     BitcoinConfig  `yaml:"bitcoin"`
	Ethereum    EthereumConfig `yaml:"ethereum"`
	SyncInterval int           `yaml:"sync_interval"`
}

// BitcoinConfig 包含比特币节点设置
type BitcoinConfig struct {
	RPCURL        string `yaml:"rpc_url"`
	RPCUser       string `yaml:"rpc_user"`
	RPCUserFile   string `yaml:"rpc_user_file"`
	RPCPassword   string `yaml:"rpc_password"`
	RPCPasswordFile string `yaml:"rpc_password_file"`
	ZMQURL        string `yaml:"zmq_url"`
	WalletName    string `yaml:"wallet_name"`
	SyncInterval  int    `yaml:"sync_interval"`
}

// EthereumConfig 包含以太坊节点设置
type EthereumConfig struct {
	RPCURL          string `yaml:"rpc_url"`
	RPCUser         string `yaml:"rpc_user"`
	RPCUserFile     string `yaml:"rpc_user_file"`
	RPCPassword     string `yaml:"rpc_password"`
	RPCPasswordFile string `yaml:"rpc_password_file"`
	WSURL           string `yaml:"ws_url"`
	ChainID         int    `yaml:"chain_id"`
	SyncInterval    int    `yaml:"sync_interval"`
}

// HSMConfig 包含硬件安全模块设置
type HSMConfig struct {
	Provider      string `yaml:"provider"`
	LibraryPath   string `yaml:"library_path"`
	Slot          int    `yaml:"slot"`
	PinFile       string `yaml:"pin_file"`
	KeyLabel      string `yaml:"key_label"`
	KeyType       string `yaml:"key_type"`
	AutoRotate    bool   `yaml:"auto_rotate"`
	RotationPeriod string `yaml:"rotation_period"`
}

// TLSCertConfig 包含TLS证书路径
type TLSCertConfig struct {
	CertFile    string   `yaml:"cert_file"`
	KeyFile     string   `yaml:"key_file"`
	CAFile      string   `yaml:"ca_file"`
	MinVersion  string   `yaml:"min_version"`
	CipherSuites []string `yaml:"cipher_suites"`
}

// SecurityConfig 包含安全相关设置
type SecurityConfig struct {
	JWTSecret       string         `yaml:"jwt_secret"`
	JWTSecretFile   string         `yaml:"jwt_secret_file"`
	TokenExpiryHours int           `yaml:"token_expiry_hours"`
	RefreshExpiryHours int         `yaml:"refresh_expiry_hours"`
	Algorithm       string         `yaml:"algorithm"`
	PasswordPolicy  PasswordPolicy `yaml:"password_policy"`
	MFA             MFAConfig      `yaml:"mfa"`
	Session         SessionConfig  `yaml:"session"`
}

// PasswordPolicy 包含密码策略
type PasswordPolicy struct {
	MinLength        int  `yaml:"min_length"`
	RequireUppercase bool `yaml:"require_uppercase"`
	RequireLowercase bool `yaml:"require_lowercase"`
	RequireNumbers   bool `yaml:"require_numbers"`
	RequireSpecial   bool `yaml:"require_special_chars"`
	PasswordHistory  int  `yaml:"password_history"`
	MaxAgeDays       int  `yaml:"max_age_days"`
}

// MFAConfig 包含多因素认证设置
type MFAConfig struct {
	Enabled      bool `yaml:"enabled"`
	Issuer       string `yaml:"issuer"`
	Window       int    `yaml:"window"`
	BackupCodes  int    `yaml:"backup_codes"`
}

// SessionConfig 包含会话设置
type SessionConfig struct {
	IdleTimeout     int `yaml:"idle_timeout"`
	AbsoluteTimeout int `yaml:"absolute_timeout"`
	ConcurrentLimit int `yaml:"concurrent_limit"`
}

// RegulatoryConfig 包含监管规则设置
type RegulatoryConfig struct {
	ComplianceFramework string              `yaml:"compliance_framework"`
	RiskWeights         map[string]float64  `yaml:"risk_weights"`
	Thresholds          ThresholdConfig     `yaml:"thresholds"`
	Reporting           ReportingConfig     `yaml:"reporting"`
}

// ThresholdConfig 包含阈值设置
type ThresholdConfig struct {
	LargeTransaction float64 `yaml:"large_transaction"`
	SuspiciousPattern int    `yaml:"suspicious_pattern"`
	VolumeAnomaly    float64 `yaml:"volume_anomaly"`
}

// ReportingConfig 包含报告设置
type ReportingConfig struct {
	DailyDeadline    string `yaml:"daily_deadline"`
	MonthlyDeadline  string `yaml:"monthly_deadline"`
	QuarterlyDeadline string `yaml:"quarterly_deadline"`
}

// MiningConfig 包含挖矿控制设置
type MiningConfig struct {
	HashRateUnit         string  `yaml:"hash_rate_unit"`
	PowerUnit            string  `yaml:"power_unit"`
	MonitoringInterval   int     `yaml:"monitoring_interval"`
	RemoteShutdownEnabled bool   `yaml:"remote_shutdown_enabled"`
	EnergyThresholdWarning float64 `yaml:"energy_threshold_warning"`
	EnergyThresholdCritical float64 `yaml:"energy_threshold_critical"`
}

// ExchangeConfig 包含交易所监控设置
type ExchangeConfig struct {
	DataFeedInterval      int               `yaml:"data_feed_interval"`
	ManipulationDetection ManipulationConfig `yaml:"manipulation_detection"`
	HealthCheckInterval   int               `yaml:"health_check_interval"`
	AutoSuspendThreshold  float64           `yaml:"auto_suspend_threshold"`
}

// ManipulationConfig 包含操纵检测设置
type ManipulationConfig struct {
	WashTradingThreshold float64 `yaml:"wash_trading_threshold"`
	SpoofingDetection    bool    `yaml:"spoofing_detection"`
	LayeringDetection    bool    `yaml:"layering_detection"`
}

// LoggingConfig 包含日志设置
type LoggingConfig struct {
	Level           string       `yaml:"level"`
	Format          string       `yaml:"format"`
	Output          []LogOutput  `yaml:"output"`
	AuditLogEnabled bool         `yaml:"audit_log_enabled"`
}

// LogOutput 包含日志输出配置
type LogOutput struct {
	Type      string `yaml:"type"`
	Path      string `yaml:"path"`
	MaxSize   string `yaml:"max_size"`
	MaxBackups int   `yaml:"max_backups"`
	Host      string `yaml:"host"`
	Protocol  string `yaml:"protocol"`
}

// MonitoringConfig 包含监控设置
type MonitoringConfig struct {
	MetricsEnabled   bool `yaml:"metrics_enabled"`
	TracingEnabled   bool `yaml:"tracing_enabled"`
	HealthCheckIntv  int  `yaml:"health_check_interval"`
	AlertThresholdCPU int `yaml:"alert_threshold_cpu"`
	AlertThresholdMemory int `yaml:"alert_threshold_memory"`
	AlertThresholdDisk int `yaml:"alert_threshold_disk"`
}

// DRConfig 包含灾难恢复设置
type DRConfig struct {
	BackupInterval    int    `yaml:"backup_interval"`
	RetentionDays     int    `yaml:"retention_days"`
	ReplicationTarget string `yaml:"replication_target"`
	RPO              int    `yaml:"rpo"`
	RTO              int    `yaml:"rto"`
	FailoverMode      string `yaml:"failover_mode"`
}

// LoadConfig 从指定的YAML文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 应用环境变量覆盖
	cfg.applyEnvOverrides()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides 应用环境变量覆盖到配置
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

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("无效的HTTP端口: %d", c.Server.HTTPPort)
	}
	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("无效的gRPC端口: %d", c.Server.GRPCPort)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("数据库主机必填")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("数据库名称必填")
	}
	if c.WORMStorage.Path == "" {
		return fmt.Errorf("WORM存储路径必填")
	}
	if c.HSM.LibraryPath == "" {
		return fmt.Errorf("HSM库路径必填")
	}
	return nil
}

// GetDSN 返回PostgreSQL连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetRedisAddr 返回Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
