package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Version   int       `json:"version" db:"version"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	BaseModel
	State          string    `json:"state"` // ONLINE, DEGRADED, OFFLINE, EMERGENCY
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	ActiveExchanges int      `json:"active_exchanges"`
	MonitoredWallets int     `json:"monitored_wallets"`
	PendingAlerts   int      `json:"pending_alerts"`
	HSMStatus       string   `json:"hsm_status"`
	DatabaseStatus  string   `json:"database_status"`
	Uptime          duration `json:"uptime"`
}

// Exchange represents a licensed cryptocurrency exchange
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

// ExchangeMetrics represents real-time exchange metrics
type ExchangeMetrics struct {
	BaseModel
	ExchangeID     string    `json:"exchange_id" db:"exchange_id"`
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
	VolumeBTC      float64   `json:"volume_btc" db:"volume_btc"`
	VolumeUSD      float64   `json:"volume_usd" db:"volume_usd"`
	TradeCount     int64     `json:"trade_count" db:"trade_count"`
	OrderBookDepth float64   `json:"order_book_depth" db:"order_book_depth"`
	SpreadBPS      float64   `json:"spread_bps" db:"spread_bps"`
	LatencyMs      float64   `json:"latency_ms" db:"latency_ms"`
}

// Wallet represents a custodial wallet
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

// Transaction represents a blockchain transaction
type Transaction struct {
	BaseModel
	TxID          string         `json:"tx_id" db:"tx_id"`
	BlockHash     sql.NullString `json:"block_hash" db:"block_hash"`
	BlockNumber   sql.NullInt64  `json:"block_number" db:"block_number"`
	Timestamp     time.Time      `json:"timestamp" db:"timestamp"`
	FromAddress   string         `json:"from_address" db:"from_address"`
	ToAddress     string         `json:"to_address" db:"to_address"`
	Amount        float64        `json:"amount" db:"amount"`
	Currency      string         `json:"currency" db:"currency"`
	GasUsed       sql.NullInt64  `json:"gas_used" db:"gas_used"`
	GasPrice      sql.NullFloat64 `json:"gas_price" db:"gas_price"`
	Fee           sql.NullFloat64 `json:"fee" db:"fee"`
	Status        string         `json:"status" db:"status"` // PENDING, CONFIRMED, FAILED
	ExchangeID    sql.NullString `json:"exchange_id" db:"exchange_id"`
	RiskScore     int            `json:"risk_score" db:"risk_score"`
	Flagged       bool           `json:"flagged" db:"flagged"`
	FlagReason    sql.NullString `json:"flag_reason" db:"flag_reason"`
	Metadata      json.RawMessage `json:"metadata" db:"metadata"`
}

// License represents a regulatory license
type License struct {
	BaseModel
	LicenseNumber   string         `json:"license_number" db:"license_number"`
	EntityName      string         `json:"entity_name" db:"entity_name"`
	EntityType      string         `json:"entity_type" db:"entity_type"` // EXCHANGE, CUSTODIAN, MINER, PAYMENT
	LicenseType     string         `json:"license_type" db:"license_type"`
	Status          string         `json:"status" db:"status"` // ACTIVE, SUSPENDED, EXPIRED, REVOKED
	IssueDate       time.Time      `json:"issue_date" db:"issue_date"`
	ExpiryDate      time.Time      `json:"expiry_date" db:"expiry_date"`
	Jurisdiction    string         `json:"jurisdiction" db:"jurisdiction"`
	LicenseDocument json.RawMessage `json:"license_document" db:"license_document"`
	Conditions      json.RawMessage `json:"conditions" db:"conditions"`
	ApprovedBy      string         `json:"approved_by" db:"approved_by"`
	ApproverTitle   string         `json:"approver_title" db:"approver_title"`
}

// Miner represents a mining operation
type Miner struct {
	BaseModel
	Name           string         `json:"name" db:"name"`
	OperatorID     string         `json:"operator_id" db:"operator_id"`
	LicenseID      sql.NullString `json:"license_id" db:"license_id"`
	Location       string         `json:"location" db:"location"`
	Coordinates    sql.NullString `json:"coordinates" db:"coordinates"`
	HashRate       float64        `json:"hash_rate" db:"hash_rate"`
	HashRateUnit   string         `json:"hash_rate_unit" db:"hash_rate_unit"` // TH/s, PH/s, EH/s
	Status         string         `json:"status" db:"status"` // ONLINE, OFFLINE, THROTTLED, SHUTDOWN
	PowerConsumption float64      `json:"power_consumption" db:"power_consumption"`
	PowerUnit      string         `json:"power_unit" db:"power_unit"` // MW, GW
	EnergySource   string         `json:"energy_source" db:"energy_source"` // GRID, RENEWABLE, MIXED
	ASICCount      int            `json:"asic_count" db:"asic_count"`
	UptimePercent  float64        `json:"uptime_percent" db:"uptime_percent"`
	RemoteShutdown bool           `json:"remote_shutdown" db:"remote_shutdown"`
}

// MiningMetrics represents mining operation metrics
type MiningMetrics struct {
	BaseModel
	MinerID       string    `json:"miner_id" db:"miner_id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	HashRate      float64   `json:"hash_rate" db:"hash_rate"`
	BlocksFound   int       `json:"blocks_found" db:"blocks_found"`
	Revenue       float64   `json:"revenue" db:"revenue"`
	RevenueCurrency string  `json:"revenue_currency" db:"revenue_currency"`
	PowerDraw     float64   `json:"power_draw" db:"power_draw"`
	Temperature   float64   `json:"temperature" db:"temperature"`
	FanSpeed      int       `json:"fan_speed" db:"fan_speed"`
}

// EnergyData represents energy grid data
type EnergyData struct {
	BaseModel
	Timestamp        time.Time `json:"timestamp" db:"timestamp"`
	RegionID         string    `json:"region_id" db:"region_id"`
	TotalLoad        float64   `json:"total_load" db:"total_load"`
	CryptoLoad       float64   `json:"crypto_load" db:"crypto_load"`
	AvailableCapacity float64  `json:"available_capacity" db:"available_capacity"`
	GridFrequency    float64   `json:"grid_frequency" db:"grid_frequency"`
	Voltage          float64   `json:"voltage" db:"voltage"`
	Status           string    `json:"status" db:"status"` // NORMAL, WARNING, CRITICAL
}

// Alert represents a system alert
type Alert struct {
	BaseModel
	Severity     string         `json:"severity" db:"severity"` // INFO, WARNING, CRITICAL, EMERGENCY
	Category     string         `json:"category" db:"category"` // SYSTEM, EXCHANGE, TRANSACTION, MINING, SECURITY
	Title        string         `json:"title" db:"title"`
	Description  string         `json:"description" db:"description"`
	Source       string         `json:"source" db:"source"`
	EntityID     sql.NullString `json:"entity_id" db:"entity_id"`
	EntityType   sql.NullString `json:"entity_type" db:"entity_type"`
	Status       string         `json:"status" db:"status"` // ACTIVE, ACKNOWLEDGED, RESOLVED
	AssignedTo   sql.NullString `json:"assigned_to" db:"assigned_to"`
	ResolvedAt   sql.NullTime   `json:"resolved_at" db:"resolved_at"`
	Resolution   sql.NullString `json:"resolution" db:"resolution"`
	Metadata     json.RawMessage `json:"metadata" db:"metadata"`
}

// AuditLog represents an immutable audit log entry
type AuditLog struct {
	BaseModel
	UserID        string         `json:"user_id" db:"user_id"`
	UserRole      string         `json:"user_role" db:"user_role"`
	Action        string         `json:"action" db:"action"`
	ResourceType  string         `json:"resource_type" db:"resource_type"`
	ResourceID    sql.NullString `json:"resource_id" db:"resource_id"`
	IPAddress     string         `json:"ip_address" db:"ip_address"`
	UserAgent     sql.NullString `json:"user_agent" db:"user_agent"`
	RequestBody   sql.NullString `json:"request_body" db:"request_body"`
	ResponseCode  int            `json:"response_code" db:"response_code"`
	PreviousHash  string         `json:"previous_hash" db:"previous_hash"`
	CurrentHash   string         `json:"current_hash" db:"current_hash"`
	Nonce         int64          `json:"nonce" db:"nonce"`
	Timestamp     time.Time      `json:"timestamp" db:"timestamp"`
}

// Report represents a generated regulatory report
type Report struct {
	BaseModel
	ReportType   string         `json:"report_type" db:"report_type"` // DAILY, MONTHLY, QUARTERLY, ANNUAL, CUSTOM
	Title        string         `json:"title" db:"title"`
	Description  sql.NullString `json:"description" db:"description"`
	PeriodStart  time.Time      `json:"period_start" db:"period_start"`
	PeriodEnd    time.Time      `json:"period_end" db:"period_end"`
	Status       string         `json:"status" db:"status"` // GENERATING, COMPLETED, FAILED
	GeneratedBy  string         `json:"generated_by" db:"generated_by"`
	FilePath     sql.NullString `json:"file_path" db:"file_path"`
	FileSize     sql.NullInt64  `json:"file_size" db:"file_size"`
	Format       string         `json:"format" db:"format"` // PDF, CSV, JSON
	Checksum     sql.NullString `json:"checksum" db:"checksum"`
	Parameters   json.RawMessage `json:"parameters" db:"parameters"`
}

// RiskScore represents a wallet risk score assessment
type RiskScore struct {
	WalletAddress  string    `json:"wallet_address"`
	Score          int       `json:"score"` // 0-100
	RiskLevel      string    `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	Factors        []RiskFactor `json:"factors"`
	LastAssessed   time.Time `json:"last_assessed"`
	PreviousScore  int       `json:"previous_score"`
	Trend          string    `json:"trend"` // IMPROVING, STABLE, DEGRADING
}

// RiskFactor represents a component of the risk score
type RiskFactor struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Score       int     `json:"score"`
	Description string  `json:"description"`
}

// User represents a system user
type User struct {
	BaseModel
	Username    string         `json:"username" db:"username"`
	Email       string         `json:"email" db:"email"`
	PasswordHash string        `json:"-" db:"password_hash"`
	Role        string         `json:"role" db:"role"` // ADMIN, OPERATOR, AUDITOR, VIEWER
	Department  string         `json:"department" db:"department"`
	Status      string         `json:"status" db:"status"` // ACTIVE, INACTIVE, LOCKED
	LastLogin   sql.NullTime   `json:"last_login" db:"last_login"`
	MFAEnabled  bool           `json:"mfa_enabled" db:"mfa_enabled"`
	MFA secret  sql.NullString `json:"-" db:"mfa_secret"`
}

// EntityCluster represents a cluster of related entities
type EntityCluster struct {
	BaseModel
	ClusterID    string   `json:"cluster_id" db:"cluster_id"`
	EntityType   string   `json:"entity_type" db:"entity_type"`
	EntityIDs    []string `json:"entity_ids"`
	RiskScore    int      `json:"risk_score" db:"risk_score"`
	Labels       []string `json:"labels"`
	FirstSeen    time.Time `json:"first_seen" db:"first_seen"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
}

// duration is a custom type for handling time.Duration in JSON
type duration struct {
	time.Duration
}

// MarshalJSON implements json.Marshaler
func (d duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (d *duration) UnmarshalJSON(b []byte) error {
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
