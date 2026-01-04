package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// BaseModel 包含所有模型的公共字段
type BaseModel struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Version   int       `json:"version" db:"version"`
}

// SystemStatus 表示整体系统状态
type SystemStatus struct {
	BaseModel
	State            string    `json:"state"` // ONLINE, DEGRADED, OFFLINE, EMERGENCY
	LastHeartbeat    time.Time `json:"last_heartbeat"`
	ActiveExchanges  int       `json:"active_exchanges"`
	MonitoredWallets int       `json:"monitored_wallets"`
	PendingAlerts    int       `json:"pending_alerts"`
	HSMStatus        string    `json:"hsm_status"`
	DatabaseStatus   string    `json:"database_status"`
	Uptime           Duration  `json:"uptime"`
}

// Exchange 表示持牌加密货币交易所
type Exchange struct {
	BaseModel
	Name           string         `json:"name" db:"name"`
	LicenseNumber  string         `json:"license_number" db:"license_number"`
	LicenseType    string         `json:"license_type" db:"license_type"` // SPOT, DERIVATIVES, CUSTODY
	Status         string         `json:"status" db:"status"` // ACTIVE, SUSPENDED, REVOKED, PENDING
	Jurisdiction   string         `json:"jurisdiction" db:"jurisdiction"`
	RegistrationID string         `json:"registration_id" db:"registration_id"`
	ContactEmail   string         `json:"contact_email" db:"contact_email"`
	Website        string         `json:"website" db:"website"`
	KYCPolicy      json.RawMessage `json:"kyc_policy" db:"kyc_policy"`
	AMLPolicy      json.RawMessage `json:"aml_policy" db:"aml_policy"`
	FeeSchedule    json.RawMessage `json:"fee_schedule" db:"fee_schedule"`
	TradingPairs   []string       `json:"trading_pairs"`
	Volume24H      float64        `json:"volume_24h"`
	Fees24H        float64        `json:"fees_24h"`
	LastAuditAt    sql.NullTime   `json:"last_audit_at"`
	NextAuditDue   sql.NullTime   `json:"next_audit_due"`
}

// ExchangeMetrics 表示实时交易所指标
type ExchangeMetrics struct {
	BaseModel
	ExchangeID   string    `json:"exchange_id" db:"exchange_id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	VolumeBTC    float64   `json:"volume_btc" db:"volume_btc"`
	VolumeUSD    float64   `json:"volume_usd" db:"volume_usd"`
	TradeCount   int64     `json:"trade_count" db:"trade_count"`
	OrderBookDepth float64 `json:"order_book_depth" db:"order_book_depth"`
	SpreadBPS    float64   `json:"spread_bps" db:"spread_bps"`
	LatencyMs    float64   `json:"latency_ms" db:"latency_ms"`
}

// ExchangeHealth 表示交易所健康评分
type ExchangeHealth struct {
	ExchangeID     string  `json:"exchange_id"`
	OverallScore   float64 `json:"overall_score"` // 0-100
	Availability   float64 `json:"availability"`
	LatencyScore   float64 `json:"latency_score"`
	VolumeScore    float64 `json:"volume_score"`
	ComplianceScore float64 `json:"compliance_score"`
	LastChecked    time.Time `json:"last_checked"`
}

// Wallet 表示托管钱包
type Wallet struct {
	BaseModel
	Address        string         `json:"address" db:"address"`
	AddressType    string         `json:"address_type" db:"address_type"` // BTC, ETH, MULTISIG
	ExchangeID     sql.NullString `json:"exchange_id" db:"exchange_id"`
	WalletType     string         `json:"wallet_type" db:"wallet_type"` // HOT, COLD, WARM
	Status         string         `json:"status" db:"status"` // ACTIVE, FROZEN, CLOSED
	Balance        float64        `json:"balance" db:"balance"`
	BalanceCurrency string        `json:"balance_currency" db:"balance_currency"`
	RiskScore      int            `json:"risk_score" db:"risk_score"`
	Blacklisted    bool           `json:"blacklisted" db:"blacklisted"`
	FreezeReason   sql.NullString `json:"freeze_reason" db:"freeze_reason"`
	FreezeOrderID  sql.NullString `json:"freeze_order_id" db:"freeze_order_id"`
	LastActivityAt sql.NullTime   `json:"last_activity_at"`
}

// Transaction 表示区块链交易
type Transaction struct {
	BaseModel
	TxID        string         `json:"tx_id" db:"tx_id"`
	BlockHash   sql.NullString `json:"block_hash" db:"block_hash"`
	BlockNumber sql.NullInt64  `json:"block_number" db:"block_number"`
	Timestamp   time.Time      `json:"timestamp" db:"timestamp"`
	FromAddress string         `json:"from_address" db:"from_address"`
	ToAddress   string         `json:"to_address" db:"to_address"`
	Amount      float64        `json:"amount" db:"amount"`
	Currency    string         `json:"currency" db:"currency"`
	GasUsed     sql.NullInt64  `json:"gas_used" db:"gas_used"`
	GasPrice    sql.NullFloat64 `json:"gas_price" db:"gas_price"`
	Fee         sql.NullFloat64 `json:"fee" db:"fee"`
	Status      string         `json:"status" db:"status"` // PENDING, CONFIRMED, FAILED
	ExchangeID  sql.NullString `json:"exchange_id" db:"exchange_id"`
	RiskScore   int            `json:"risk_score" db:"risk_score"`
	Flagged     bool           `json:"flagged" db:"flagged"`
	FlagReason  sql.NullString `json:"flag_reason" db:"flag_reason"`
	Metadata    json.RawMessage `json:"metadata" db:"metadata"`
}

// TransactionFilter 用于交易搜索
type TransactionFilter struct {
	FromAddress   string    `json:"from_address,omitempty"`
	ToAddress     string    `json:"to_address,omitempty"`
	Currency      string    `json:"currency,omitempty"`
	Status        string    `json:"status,omitempty"`
	Flagged       *bool     `json:"flagged,omitempty"`
	MinAmount     float64   `json:"min_amount,omitempty"`
	MaxAmount     float64   `json:"max_amount,omitempty"`
	StartTime     time.Time `json:"start_time,omitempty"`
	EndTime       time.Time `json:"end_time,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
}

// License 表示监管许可证
type License struct {
	BaseModel
	LicenseNumber    string         `json:"license_number" db:"license_number"`
	EntityName       string         `json:"entity_name" db:"entity_name"`
	EntityType       string         `json:"entity_type" db:"entity_type"` // EXCHANGE, CUSTODIAN, MINER, PAYMENT
	LicenseType      string         `json:"license_type" db:"license_type"`
	Status           string         `json:"status" db:"status"` // ACTIVE, SUSPENDED, EXPIRED, REVOKED
	IssueDate        time.Time      `json:"issue_date" db:"issue_date"`
	ExpiryDate       time.Time      `json:"expiry_date" db:"expiry_date"`
	Jurisdiction     string         `json:"jurisdiction" db:"jurisdiction"`
	LicenseDocument  json.RawMessage `json:"license_document" db:"license_document"`
	Conditions       json.RawMessage `json:"conditions" db:"conditions"`
	ApprovedBy       string         `json:"approved_by" db:"approved_by"`
	ApproverTitle    string         `json:"approver_title" db:"approver_title"`
}

// Miner 表示挖矿作业
type Miner struct {
	BaseModel
	Name             string         `json:"name" db:"name"`
	OperatorID       string         `json:"operator_id" db:"operator_id"`
	LicenseID        sql.NullString `json:"license_id" db:"license_id"`
	Location         string         `json:"location" db:"location"`
	Coordinates      sql.NullString `json:"coordinates" db:"coordinates"`
	HashRate         float64        `json:"hash_rate" db:"hash_rate"`
	HashRateUnit     string         `json:"hash_rate_unit" db:"hash_rate_unit"` // TH/s, PH/s, EH/s
	Status           string         `json:"status" db:"status"` // ONLINE, OFFLINE, THROTTLED, SHUTDOWN
	PowerConsumption float64        `json:"power_consumption" db:"power_consumption"`
	PowerUnit        string         `json:"power_unit" db:"power_unit"` // MW, GW
	EnergySource     string         `json:"energy_source" db:"energy_source"` // GRID, RENEWABLE, MIXED
	ASICCount        int            `json:"asic_count" db:"asic_count"`
	UptimePercent    float64        `json:"uptime_percent" db:"uptime_percent"`
	RemoteShutdown   bool           `json:"remote_shutdown" db:"remote_shutdown"`
}

// MiningMetrics 表示挖矿作业指标
type MiningMetrics struct {
	BaseModel
	MinerID          string    `json:"miner_id" db:"miner_id"`
	Timestamp        time.Time `json:"timestamp" db:"timestamp"`
	HashRate         float64   `json:"hash_rate" db:"hash_rate"`
	BlocksFound      int       `json:"blocks_found" db:"blocks_found"`
	Revenue          float64   `json:"revenue" db:"revenue"`
	RevenueCurrency  string    `json:"revenue_currency" db:"revenue_currency"`
	PowerDraw        float64   `json:"power_draw" db:"power_draw"`
	Temperature      float64   `json:"temperature" db:"temperature"`
	FanSpeed         int       `json:"fan_speed" db:"fan_speed"`
}

// EnergyData 表示能源网格数据
type EnergyData struct {
	BaseModel
	Timestamp          time.Time `json:"timestamp" db:"timestamp"`
	RegionID           string    `json:"region_id" db:"region_id"`
	TotalLoad          float64   `json:"total_load" db:"total_load"`
	CryptoLoad         float64   `json:"crypto_load" db:"crypto_load"`
	AvailableCapacity  float64   `json:"available_capacity" db:"available_capacity"`
	GridFrequency      float64   `json:"grid_frequency" db:"grid_frequency"`
	Voltage            float64   `json:"voltage" db:"voltage"`
	Status             string    `json:"status" db:"status"` // NORMAL, WARNING, CRITICAL
}

// Alert 表示系统警报
type Alert struct {
	BaseModel
	Severity      string         `json:"severity" db:"severity"` // INFO, WARNING, CRITICAL, EMERGENCY
	Category      string         `json:"category" db:"category"` // SYSTEM, EXCHANGE, TRANSACTION, MINING, SECURITY
	Title         string         `json:"title" db:"title"`
	Description   string         `json:"description" db:"description"`
	Source        string         `json:"source" db:"source"`
	EntityID      sql.NullString `json:"entity_id" db:"entity_id"`
	EntityType    sql.NullString `json:"entity_type" db:"entity_type"`
	Status        string         `json:"status" db:"status"` // ACTIVE, ACKNOWLEDGED, RESOLVED
	AssignedTo    sql.NullString `json:"assigned_to" db:"assigned_to"`
	ResolvedAt    sql.NullTime   `json:"resolved_at" db:"resolved_at"`
	Resolution    sql.NullString `json:"resolution" db:"resolution"`
	Metadata      json.RawMessage `json:"metadata" db:"metadata"`
}

// AuditLog 表示不可变审计日志条目
type AuditLog struct {
	BaseModel
	UserID       string         `json:"user_id" db:"user_id"`
	UserRole     string         `json:"user_role" db:"user_role"`
	Action       string         `json:"action" db:"action"`
	ResourceType string         `json:"resource_type" db:"resource_type"`
	ResourceID   sql.NullString `json:"resource_id" db:"resource_id"`
	IPAddress    string         `json:"ip_address" db:"ip_address"`
	UserAgent    sql.NullString `json:"user_agent" db:"user_agent"`
	RequestBody  sql.NullString `json:"request_body" db:"request_body"`
	ResponseCode int            `json:"response_code" db:"response_code"`
	PreviousHash string         `json:"previous_hash" db:"previous_hash"`
	CurrentHash  string         `json:"current_hash" db:"current_hash"`
	Nonce        int64          `json:"nonce" db:"nonce"`
	Timestamp    time.Time      `json:"timestamp" db:"timestamp"`
}

// Report 表示生成的监管报告
type Report struct {
	BaseModel
	ReportType    string         `json:"report_type" db:"report_type"` // DAILY, MONTHLY, QUARTERLY, ANNUAL, CUSTOM
	Title         string         `json:"title" db:"title"`
	Description   sql.NullString `json:"description" db:"description"`
	PeriodStart   time.Time      `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time      `json:"period_end" db:"period_end"`
	Status        string         `json:"status" db:"status"` // GENERATING, COMPLETED, FAILED
	GeneratedBy   string         `json:"generated_by" db:"generated_by"`
	FilePath      sql.NullString `json:"file_path" db:"file_path"`
	FileSize      sql.NullInt64  `json:"file_size" db:"file_size"`
	Format        string         `json:"format" db:"format"` // PDF, CSV, JSON
	Checksum      sql.NullString `json:"checksum" db:"checksum"`
	Parameters    json.RawMessage `json:"parameters" db:"parameters"`
}

// RiskScore 表示钱包风险评分评估
type RiskScore struct {
	WalletAddress  string       `json:"wallet_address"`
	Score          int          `json:"score"` // 0-100
	RiskLevel      string       `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	Factors        []RiskFactor `json:"factors"`
	LastAssessed   time.Time    `json:"last_assessed"`
	PreviousScore  int          `json:"previous_score"`
	Trend          string       `json:"trend"` // IMPROVING, STABLE, DEGRADING
}

// RiskFactor 表示风险评分的组成部分
type RiskFactor struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Score       int     `json:"score"`
	Description string  `json:"description"`
}

// User 表示系统用户
type User struct {
	BaseModel
	Username     string         `json:"username" db:"username"`
	Email        string         `json:"email" db:"email"`
	PasswordHash string         `json:"-" db:"password_hash"`
	Role         string         `json:"role" db:"role"` // ADMIN, OPERATOR, AUDITOR, VIEWER
	Department   string         `json:"department" db:"department"`
	Status       string         `json:"status" db:"status"` // ACTIVE, INACTIVE, LOCKED
	LastLogin    sql.NullTime   `json:"last_login" db:"last_login"`
	MFAEnabled   bool           `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret    sql.NullString `json:"-" db:"mfa_secret"`
}

// UserRole 表示用户角色及其权限
type UserRole struct {
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description"`
}

// EntityCluster 表示相关实体的集群
type EntityCluster struct {
	BaseModel
	ClusterID   string    `json:"cluster_id" db:"cluster_id"`
	EntityType  string    `json:"entity_type" db:"entity_type"`
	EntityIDs   []string  `json:"entity_ids"`
	RiskScore   int       `json:"risk_score" db:"risk_score"`
	Labels      []string  `json:"labels"`
	FirstSeen   time.Time `json:"first_seen" db:"first_seen"`
	LastSeen    time.Time `json:"last_seen" db:"last_seen"`
}

// FreezeOrder 表示资产冻结命令
type FreezeOrder struct {
	BaseModel
	OrderType     string         `json:"order_type"` // FREEZE, UNFREEZE
	EntityType    string         `json:"entity_type"` // WALLET, EXCHANGE, ACCOUNT
	EntityID      string         `json:"entity_id"`
	Reason        string         `json:"reason"`
	LegalBasis    string         `json:"legal_basis"`
	IssuedBy      string         `json:"issued_by"`
	IssuerTitle   string         `json:"issuer_title"`
	EffectiveFrom time.Time      `json:"effective_from"`
	EffectiveTo   sql.NullTime   `json:"effective_to"`
	Status        string         `json:"status"` // ACTIVE, EXPIRED, REVOKED
	Metadata      json.RawMessage `json:"metadata"`
}

// EmergencyStop 表示紧急停止事件
type EmergencyStop struct {
	BaseModel
	StopType     string         `json:"stop_type"` // GLOBAL, EXCHANGE, TRANSACTION
	EntityID     sql.NullString `json:"entity_id"`
	Reason       string         `json:"reason"`
	IssuedBy     string         `json:"issued_by"`
	IssuedAt     time.Time      `json:"issued_at"`
	ResolvedAt   sql.NullTime   `json:"resolved_at"`
	Resolution   sql.NullString `json:"resolution"`
	Status       string         `json:"status"` // ACTIVE, RESOLVED
}

// PolicyRule 表示监管策略规则
type PolicyRule struct {
	BaseModel
	RuleID       string         `json:"rule_id"`
	Category     string         `json:"category"` // TRANSACTION, TRADING, REPORTING
	RuleType     string         `json:"rule_type"` // LIMIT, PROHIBITION, REQUIREMENT
	Description  string         `json:"description"`
	Conditions   json.RawMessage `json:"conditions"`
	Actions      json.RawMessage `json:"actions"`
	EffectiveFrom time.Time     `json:"effective_from"`
	EffectiveTo   sql.NullTime  `json:"effective_to"`
	Status       string         `json:"status"` // ACTIVE, INACTIVE
	Priority     int            `json:"priority"`
}

// Duration 是用于在JSON中处理time.Duration的自定义类型
type Duration struct {
	time.Duration
}

// MarshalJSON 实现json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// UnmarshalJSON 实现json.Unmarshaler
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Metadata  *APIMeta    `json:"metadata,omitempty"`
}

// APIError API错误详情
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// APIMeta API元信息
type APIMeta struct {
	RequestID  string    `json:"request_id"`
	Timestamp  time.Time `json:"timestamp"`
	Version    string    `json:"version"`
	Processing int64     `json:"processing_ms"`
}
