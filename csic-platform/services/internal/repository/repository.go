package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pql"
)

// Repositories 包含所有数据仓储
type Repositories struct {
	DB                *sql.DB
	Exchanges         *ExchangeRepository
	Wallets           *WalletRepository
	Transactions      *TransactionRepository
	Licenses          *LicenseRepository
	Miners            *MinerRepository
	Alerts            *AlertRepository
	AuditLogs         *AuditLogRepository
	Reports           *ReportRepository
	Users             *UserRepository
	FreezeOrders      *FreezeOrderRepository
	EmergencyStops    *EmergencyStopRepository
	PolicyRules       *PolicyRuleRepository
	EntityClusters    *EntityClusterRepository
}

// NewRepositories 创建所有仓储实例
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		DB:             db,
		Exchanges:      NewExchangeRepository(db),
		Wallets:        NewWalletRepository(db),
		Transactions:   NewTransactionRepository(db),
		Licenses:       NewLicenseRepository(db),
		Miners:         NewMinerRepository(db),
		Alerts:         NewAlertRepository(db),
		AuditLogs:      NewAuditLogRepository(db),
		Reports:        NewReportRepository(db),
		Users:          NewUserRepository(db),
		FreezeOrders:   NewFreezeOrderRepository(db),
		EmergencyStops: NewEmergencyStopRepository(db),
		PolicyRules:    NewPolicyRuleRepository(db),
		EntityClusters: NewEntityClusterRepository(db),
	}
}

// ExchangeRepository 交易所数据仓储
type ExchangeRepository struct {
	db *sql.DB
}

// NewExchangeRepository 创建交易所仓储
func NewExchangeRepository(db *sql.DB) *ExchangeRepository {
	return &ExchangeRepository{db: db}
}

// Create 创建新交易所
func (r *ExchangeRepository) Create(exchange *models.Exchange) error {
	if exchange.ID == "" {
		exchange.ID = uuid.New().String()
	}
	exchange.CreatedAt = time.Now()
	exchange.UpdatedAt = time.Now()

	query := `
		INSERT INTO exchanges (
			id, name, license_number, license_type, status, jurisdiction,
			registration_id, contact_email, website, kyc_policy, aml_policy,
			fee_schedule, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 1)
	`

	_, err := r.db.Exec(query,
		exchange.ID, exchange.Name, exchange.LicenseNumber, exchange.LicenseType,
		exchange.Status, exchange.Jurisdiction, exchange.RegistrationID,
		exchange.ContactEmail, exchange.Website, exchange.KYCPolicy,
		exchange.AMLPolicy, exchange.FeeSchedule, exchange.CreatedAt, exchange.UpdatedAt,
	)
	return err
}

// GetByID 通过ID获取交易所
func (r *ExchangeRepository) GetByID(id string) (*models.Exchange, error) {
	query := `
		SELECT id, name, license_number, license_type, status, jurisdiction,
			registration_id, contact_email, website, kyc_policy, aml_policy,
			fee_schedule, created_at, updated_at, version
		FROM exchanges WHERE id = $1
	`

	var exchange models.Exchange
	err := r.db.QueryRow(query, id).Scan(
		&exchange.ID, &exchange.Name, &exchange.LicenseNumber, &exchange.LicenseType,
		&exchange.Status, &exchange.Jurisdiction, &exchange.RegistrationID,
		&exchange.ContactEmail, &exchange.Website, &exchange.KYCPolicy,
		&exchange.AMLPolicy, &exchange.FeeSchedule, &exchange.CreatedAt,
		&exchange.UpdatedAt, &exchange.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &exchange, err
}

// GetAll 获取所有交易所
func (r *ExchangeRepository) GetAll(status string, limit, offset int) ([]models.Exchange, error) {
	query := `
		SELECT id, name, license_number, license_type, status, jurisdiction,
			registration_id, contact_email, website, kyc_policy, aml_policy,
			fee_schedule, created_at, updated_at, version
		FROM exchanges
	`
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exchanges []models.Exchange
	for rows.Next() {
		var exchange models.Exchange
		err := rows.Scan(
			&exchange.ID, &exchange.Name, &exchange.LicenseNumber, &exchange.LicenseType,
			&exchange.Status, &exchange.Jurisdiction, &exchange.RegistrationID,
			&exchange.ContactEmail, &exchange.Website, &exchange.KYCPolicy,
			&exchange.AMLPolicy, &exchange.FeeSchedule, &exchange.CreatedAt,
			&exchange.UpdatedAt, &exchange.Version,
		)
		if err != nil {
			return nil, err
		}
		exchanges = append(exchanges, exchange)
	}
	return exchanges, rows.Err()
}

// Update 更新交易所信息
func (r *ExchangeRepository) Update(exchange *models.Exchange) error {
	exchange.UpdatedAt = time.Now()
	exchange.Version++

	query := `
		UPDATE exchanges SET
			name = $1, license_type = $2, status = $3, jurisdiction = $4,
			contact_email = $5, website = $6, kyc_policy = $7, aml_policy = $8,
			fee_schedule = $9, updated_at = $10, version = $11
		WHERE id = $12 AND version = $13
	`

	result, err := r.db.Exec(query,
		exchange.Name, exchange.LicenseType, exchange.Status, exchange.Jurisdiction,
		exchange.ContactEmail, exchange.Website, exchange.KYCPolicy, exchange.AMLPolicy,
		exchange.FeeSchedule, exchange.UpdatedAt, exchange.Version, exchange.ID,
		exchange.Version-1,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("exchange not found or version mismatch")
	}
	return nil
}

// CountByStatus 按状态统计交易所数量
func (r *ExchangeRepository) CountByStatus() (map[string]int, error) {
	query := "SELECT status, COUNT(*) FROM exchanges GROUP BY status"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}
	return result, rows.Err()
}

// WalletRepository 钱包数据仓储
type WalletRepository struct {
	db *sql.DB
}

// NewWalletRepository 创建钱包仓储
func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// Create 创建新钱包
func (r *WalletRepository) Create(wallet *models.Wallet) error {
	if wallet.ID == "" {
		wallet.ID = uuid.New().String()
	}
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	query := `
		INSERT INTO wallets (
			id, address, address_type, exchange_id, wallet_type, status,
			balance, balance_currency, risk_score, blacklisted, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 1)
	`

	_, err := r.db.Exec(query,
		wallet.ID, wallet.Address, wallet.AddressType, wallet.ExchangeID,
		wallet.WalletType, wallet.Status, wallet.Balance, wallet.BalanceCurrency,
		wallet.RiskScore, wallet.Blacklisted, wallet.CreatedAt, wallet.UpdatedAt,
	)
	return err
}

// GetByID 通过ID获取钱包
func (r *WalletRepository) GetByID(id string) (*models.Wallet, error) {
	query := `
		SELECT id, address, address_type, exchange_id, wallet_type, status,
			balance, balance_currency, risk_score, blacklisted, freeze_reason,
			freeze_order_id, last_activity_at, created_at, updated_at, version
		FROM wallets WHERE id = $1
	`

	var wallet models.Wallet
	err := r.db.QueryRow(query, id).Scan(
		&wallet.ID, &wallet.Address, &wallet.AddressType, &wallet.ExchangeID,
		&wallet.WalletType, &wallet.Status, &wallet.Balance, &wallet.BalanceCurrency,
		&wallet.RiskScore, &wallet.Blacklisted, &wallet.FreezeReason,
		&wallet.FreezeOrderID, &wallet.LastActivityAt, &wallet.CreatedAt,
		&wallet.UpdatedAt, &wallet.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &wallet, err
}

// GetByAddress 通过地址获取钱包
func (r *WalletRepository) GetByAddress(address string) (*models.Wallet, error) {
	query := `
		SELECT id, address, address_type, exchange_id, wallet_type, status,
			balance, balance_currency, risk_score, blacklisted, freeze_reason,
			freeze_order_id, last_activity_at, created_at, updated_at, version
		FROM wallets WHERE address = $1
	`

	var wallet models.Wallet
	err := r.db.QueryRow(query, address).Scan(
		&wallet.ID, &wallet.Address, &wallet.AddressType, &wallet.ExchangeID,
		&wallet.WalletType, &wallet.Status, &wallet.Balance, &wallet.BalanceCurrency,
		&wallet.RiskScore, &wallet.Blacklisted, &wallet.FreezeReason,
		&wallet.FreezeOrderID, &wallet.LastActivityAt, &wallet.CreatedAt,
		&wallet.UpdatedAt, &wallet.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &wallet, err
}

// GetFrozen 获取所有冻结钱包
func (r *WalletRepository) GetFrozen(limit, offset int) ([]models.Wallet, error) {
	query := `
		SELECT id, address, address_type, exchange_id, wallet_type, status,
			balance, balance_currency, risk_score, blacklisted, freeze_reason,
			freeze_order_id, last_activity_at, created_at, updated_at, version
		FROM wallets WHERE status = 'FROZEN'
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []models.Wallet
	for rows.Next() {
		var wallet models.Wallet
		err := rows.Scan(
			&wallet.ID, &wallet.Address, &wallet.AddressType, &wallet.ExchangeID,
			&wallet.WalletType, &wallet.Status, &wallet.Balance, &wallet.BalanceCurrency,
			&wallet.RiskScore, &wallet.Blacklisted, &wallet.FreezeReason,
			&wallet.FreezeOrderID, &wallet.LastActivityAt, &wallet.CreatedAt,
			&wallet.UpdatedAt, &wallet.Version,
		)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	return wallets, rows.Err()
}

// Freeze 冻结钱包
func (r *WalletRepository) Freeze(id, reason, orderID string) error {
	query := `
		UPDATE wallets SET
			status = 'FROZEN', freeze_reason = $1, freeze_order_id = $2,
			updated_at = $3, version = version + 1
		WHERE id = $4
	`

	_, err := r.db.Exec(query, reason, orderID, time.Now(), id)
	return err
}

// Unfreeze 解冻钱包
func (r *WalletRepository) Unfreeze(id string) error {
	query := `
		UPDATE wallets SET
			status = 'ACTIVE', freeze_reason = NULL, freeze_order_id = NULL,
			updated_at = $1, version = version + 1
		WHERE id = $2
	`

	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// TransactionRepository 交易数据仓储
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository 创建交易仓储
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create 创建新交易
func (r *TransactionRepository) Create(tx *models.Transaction) error {
	if tx.ID == "" {
		tx.ID = uuid.New().String()
	}
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()

	query := `
		INSERT INTO transactions (
			id, tx_id, block_hash, block_number, timestamp, from_address,
			to_address, amount, currency, gas_used, gas_price, fee, status,
			exchange_id, risk_score, flagged, flag_reason, metadata,
			created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, 1)
	`

	_, err := r.db.Exec(query,
		tx.ID, tx.TxID, tx.BlockHash, tx.BlockNumber, tx.Timestamp,
		tx.FromAddress, tx.ToAddress, tx.Amount, tx.Currency, tx.GasUsed,
		tx.GasPrice, tx.Fee, tx.Status, tx.ExchangeID, tx.RiskScore,
		tx.Flagged, tx.FlagReason, tx.Metadata, tx.CreatedAt, tx.UpdatedAt,
	)
	return err
}

// GetByID 通过ID获取交易
func (r *TransactionRepository) GetByID(id string) (*models.Transaction, error) {
	query := `
		SELECT id, tx_id, block_hash, block_number, timestamp, from_address,
			to_address, amount, currency, gas_used, gas_price, fee, status,
			exchange_id, risk_score, flagged, flag_reason, metadata,
			created_at, updated_at, version
		FROM transactions WHERE id = $1
	`

	var tx models.Transaction
	err := r.db.QueryRow(query, id).Scan(
		&tx.ID, &tx.TxID, &tx.BlockHash, &tx.BlockNumber, &tx.Timestamp,
		&tx.FromAddress, &tx.ToAddress, &tx.Amount, &tx.Currency, &tx.GasUsed,
		&tx.GasPrice, &tx.Fee, &tx.Status, &tx.ExchangeID, &tx.RiskScore,
		&tx.Flagged, &tx.FlagReason, &tx.Metadata, &tx.CreatedAt, &tx.UpdatedAt,
		&tx.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tx, err
}

// Search 搜索交易
func (r *TransactionRepository) Search(filter models.TransactionFilter) ([]models.Transaction, int64, error) {
	baseQuery := "FROM transactions WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.FromAddress != "" {
		baseQuery += fmt.Sprintf(" AND from_address = $%d", argIndex)
		args = append(args, filter.FromAddress)
		argIndex++
	}
	if filter.ToAddress != "" {
		baseQuery += fmt.Sprintf(" AND to_address = $%d", argIndex)
		args = append(args, filter.ToAddress)
		argIndex++
	}
	if filter.Currency != "" {
		baseQuery += fmt.Sprintf(" AND currency = $%d", argIndex)
		args = append(args, filter.Currency)
		argIndex++
	}
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}
	if filter.Flagged != nil {
		baseQuery += fmt.Sprintf(" AND flagged = $%d", argIndex)
		args = append(args, *filter.Flagged)
		argIndex++
	}
	if filter.MinAmount > 0 {
		baseQuery += fmt.Sprintf(" AND amount >= $%d", argIndex)
		args = append(args, filter.MinAmount)
		argIndex++
	}
	if filter.MaxAmount > 0 {
		baseQuery += fmt.Sprintf(" AND amount <= $%d", argIndex)
		args = append(args, filter.MaxAmount)
		argIndex++
	}
	if !filter.StartTime.IsZero() {
		baseQuery += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, filter.StartTime)
		argIndex++
	}
	if !filter.EndTime.IsZero() {
		baseQuery += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, filter.EndTime)
		argIndex++
	}

	// 计算总数
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 获取数据
	selectQuery := `
		SELECT id, tx_id, block_hash, block_number, timestamp, from_address,
			to_address, amount, currency, gas_used, gas_price, fee, status,
			exchange_id, risk_score, flagged, flag_reason, metadata,
			created_at, updated_at, version
	` + baseQuery + " ORDER BY timestamp DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	selectQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID, &tx.TxID, &tx.BlockHash, &tx.BlockNumber, &tx.Timestamp,
			&tx.FromAddress, &tx.ToAddress, &tx.Amount, &tx.Currency, &tx.GasUsed,
			&tx.GasPrice, &tx.Fee, &tx.Status, &tx.ExchangeID, &tx.RiskScore,
			&tx.Flagged, &tx.FlagReason, &tx.Metadata, &tx.CreatedAt, &tx.UpdatedAt,
			&tx.Version,
		)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, total, rows.Err()
}

// Flag 标记交易
func (r *TransactionRepository) Flag(id, reason string) error {
	query := `
		UPDATE transactions SET
			flagged = true, flag_reason = $1, updated_at = $2, version = version + 1
		WHERE id = $3
	`

	_, err := r.db.Exec(query, reason, time.Now(), id)
	return err
}

// GetFlagged 获取所有标记交易
func (r *TransactionRepository) GetFlagged(limit, offset int) ([]models.Transaction, error) {
	query := `
		SELECT id, tx_id, block_hash, block_number, timestamp, from_address,
			to_address, amount, currency, gas_used, gas_price, fee, status,
			exchange_id, risk_score, flagged, flag_reason, metadata,
			created_at, updated_at, version
		FROM transactions WHERE flagged = true
		ORDER BY timestamp DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID, &tx.TxID, &tx.BlockHash, &tx.BlockNumber, &tx.Timestamp,
			&tx.FromAddress, &tx.ToAddress, &tx.Amount, &tx.Currency, &tx.GasUsed,
			&tx.GasPrice, &tx.Fee, &tx.Status, &tx.ExchangeID, &tx.RiskScore,
			&tx.Flagged, &tx.FlagReason, &tx.Metadata, &tx.CreatedAt, &tx.UpdatedAt,
			&tx.Version,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, rows.Err()
}

// LicenseRepository 许可证数据仓储
type LicenseRepository struct {
	db *sql.DB
}

// NewLicenseRepository 创建许可证仓储
func NewLicenseRepository(db *sql.DB) *LicenseRepository {
	return &LicenseRepository{db: db}
}

// Create 创建新许可证
func (r *LicenseRepository) Create(license *models.License) error {
	if license.ID == "" {
		license.ID = uuid.New().String()
	}
	license.CreatedAt = time.Now()
	license.UpdatedAt = time.Now()

	query := `
		INSERT INTO licenses (
			id, license_number, entity_name, entity_type, license_type, status,
			issue_date, expiry_date, jurisdiction, license_document, conditions,
			approved_by, approver_title, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 1)
	`

	_, err := r.db.Exec(query,
		license.ID, license.LicenseNumber, license.EntityName, license.EntityType,
		license.LicenseType, license.Status, license.IssueDate, license.ExpiryDate,
		license.Jurisdiction, license.LicenseDocument, license.Conditions,
		license.ApprovedBy, license.ApproverTitle, license.CreatedAt, license.UpdatedAt,
	)
	return err
}

// GetByID 通过ID获取许可证
func (r *LicenseRepository) GetByID(id string) (*models.License, error) {
	query := `
		SELECT id, license_number, entity_name, entity_type, license_type, status,
			issue_date, expiry_date, jurisdiction, license_document, conditions,
			approved_by, approver_title, created_at, updated_at, version
		FROM licenses WHERE id = $1
	`

	var license models.License
	err := r.db.QueryRow(query, id).Scan(
		&license.ID, &license.LicenseNumber, &license.EntityName, &license.EntityType,
		&license.LicenseType, &license.Status, &license.IssueDate, &license.ExpiryDate,
		&license.Jurisdiction, &license.LicenseDocument, &license.Conditions,
		&license.ApprovedBy, &license.ApproverTitle, &license.CreatedAt, &license.UpdatedAt,
		&license.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &license, err
}

// GetByNumber 通过许可证号获取
func (r *LicenseRepository) GetByNumber(licenseNumber string) (*models.License, error) {
	query := `
		SELECT id, license_number, entity_name, entity_type, license_type, status,
			issue_date, expiry_date, jurisdiction, license_document, conditions,
			approved_by, approver_title, created_at, updated_at, version
		FROM licenses WHERE license_number = $1
	`

	var license models.License
	err := r.db.QueryRow(query, licenseNumber).Scan(
		&license.ID, &license.LicenseNumber, &license.EntityName, &license.EntityType,
		&license.LicenseType, &license.Status, &license.IssueDate, &license.ExpiryDate,
		&license.Jurisdiction, &license.LicenseDocument, &license.Conditions,
		&license.ApprovedBy, &license.ApproverTitle, &license.CreatedAt, &license.UpdatedAt,
		&license.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &license, err
}

// GetAll 获取所有许可证
func (r *LicenseRepository) GetAll(status string, entityType string, limit, offset int) ([]models.License, error) {
	query := `
		SELECT id, license_number, entity_name, entity_type, license_type, status,
			issue_date, expiry_date, jurisdiction, license_document, conditions,
			approved_by, approver_title, created_at, updated_at, version
		FROM licenses
	`
	args := []interface{}{}
	argIndex := 1
	conditions := ""

	if status != "" {
		conditions += fmt.Sprintf(" status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	if entityType != "" {
		if conditions != "" {
			conditions += " AND"
		}
		conditions += fmt.Sprintf(" entity_type = $%d", argIndex)
		args = append(args, entityType)
		argIndex++
	}
	if conditions != "" {
		query += " WHERE" + conditions
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licenses []models.License
	for rows.Next() {
		var license models.License
		err := rows.Scan(
			&license.ID, &license.LicenseNumber, &license.EntityName, &license.EntityType,
			&license.LicenseType, &license.Status, &license.IssueDate, &license.ExpiryDate,
			&license.Jurisdiction, &license.LicenseDocument, &license.Conditions,
			&license.ApprovedBy, &license.ApproverTitle, &license.CreatedAt, &license.UpdatedAt,
			&license.Version,
		)
		if err != nil {
			return nil, err
		}
		licenses = append(licenses, license)
	}
	return licenses, rows.Err()
}

// Update 更新许可证
func (r *LicenseRepository) Update(license *models.License) error {
	license.UpdatedAt = time.Now()
	license.Version++

	query := `
		UPDATE licenses SET
			status = $1, license_document = $2, conditions = $3,
			updated_at = $4, version = $5
		WHERE id = $6 AND version = $7
	`

	result, err := r.db.Exec(query,
		license.Status, license.LicenseDocument, license.Conditions,
		license.UpdatedAt, license.Version, license.ID, license.Version-1,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("license not found or version mismatch")
	}
	return nil
}

// Revoke 撤销许可证
func (r *LicenseRepository) Revoke(id string) error {
	query := `
		UPDATE licenses SET
			status = 'REVOKED', updated_at = $1, version = version + 1
		WHERE id = $2
	`

	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// GetExpiring 获取即将过期的许可证
func (r *LicenseRepository) GetExpiring(days int) ([]models.License, error) {
	query := `
		SELECT id, license_number, entity_name, entity_type, license_type, status,
			issue_date, expiry_date, jurisdiction, license_document, conditions,
			approved_by, approver_title, created_at, updated_at, version
		FROM licenses
		WHERE status = 'ACTIVE' AND expiry_date <= NOW() + INTERVAL '%d days'
		ORDER BY expiry_date ASC
	`

	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licenses []models.License
	for rows.Next() {
		var license models.License
		err := rows.Scan(
			&license.ID, &license.LicenseNumber, &license.EntityName, &license.EntityType,
			&license.LicenseType, &license.Status, &license.IssueDate, &license.ExpiryDate,
			&license.Jurisdiction, &license.LicenseDocument, &license.Conditions,
			&license.ApprovedBy, &license.ApproverTitle, &license.CreatedAt, &license.UpdatedAt,
			&license.Version,
		)
		if err != nil {
			return nil, err
		}
		licenses = append(licenses, license)
	}
	return licenses, rows.Err()
}

// MinerRepository 矿工数据仓储
type MinerRepository struct {
	db *sql.DB
}

// NewMinerRepository 创建矿工仓储
func NewMinerRepository(db *sql.DB) *MinerRepository {
	return &MinerRepository{db: db}
}

// Create 创建新矿工记录
func (r *MinerRepository) Create(miner *models.Miner) error {
	if miner.ID == "" {
		miner.ID = uuid.New().String()
	}
	miner.CreatedAt = time.Now()
	miner.UpdatedAt = time.Now()

	query := `
		INSERT INTO miners (
			id, name, operator_id, license_id, location, coordinates,
			hash_rate, hash_rate_unit, status, power_consumption, power_unit,
			energy_source, asic_count, uptime_percent, remote_shutdown,
			created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 1)
	`

	_, err := r.db.Exec(query,
		miner.ID, miner.Name, miner.OperatorID, miner.LicenseID, miner.Location,
		miner.Coordinates, miner.HashRate, miner.HashRateUnit, miner.Status,
		miner.PowerConsumption, miner.PowerUnit, miner.EnergySource, miner.ASICCount,
		miner.UptimePercent, miner.RemoteShutdown, miner.CreatedAt, miner.UpdatedAt,
	)
	return err
}

// GetByID 通过ID获取矿工
func (r *MinerRepository) GetByID(id string) (*models.Miner, error) {
	query := `
		SELECT id, name, operator_id, license_id, location, coordinates,
			hash_rate, hash_rate_unit, status, power_consumption, power_unit,
			energy_source, asic_count, uptime_percent, remote_shutdown,
			created_at, updated_at, version
		FROM miners WHERE id = $1
	`

	var miner models.Miner
	err := r.db.QueryRow(query, id).Scan(
		&miner.ID, &miner.Name, &miner.OperatorID, &miner.LicenseID, &miner.Location,
		&miner.Coordinates, &miner.HashRate, &miner.HashRateUnit, &miner.Status,
		&miner.PowerConsumption, &miner.PowerUnit, &miner.EnergySource, &miner.ASICCount,
		&miner.UptimePercent, &miner.RemoteShutdown, &miner.CreatedAt, &miner.UpdatedAt,
		&miner.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &miner, err
}

// GetAll 获取所有矿工
func (r *MinerRepository) GetAll(status string, limit, offset int) ([]models.Miner, error) {
	query := `
		SELECT id, name, operator_id, license_id, location, coordinates,
			hash_rate, hash_rate_unit, status, power_consumption, power_unit,
			energy_source, asic_count, uptime_percent, remote_shutdown,
			created_at, updated_at, version
		FROM miners
	`
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var miners []models.Miner
	for rows.Next() {
		var miner models.Miner
		err := rows.Scan(
			&miner.ID, &miner.Name, &miner.OperatorID, &miner.LicenseID, &miner.Location,
			&miner.Coordinates, &miner.HashRate, &miner.HashRateUnit, &miner.Status,
			&miner.PowerConsumption, &miner.PowerUnit, &miner.EnergySource, &miner.ASICCount,
			&miner.UptimePercent, &miner.RemoteShutdown, &miner.CreatedAt, &miner.UpdatedAt,
			&miner.Version,
		)
		if err != nil {
			return nil, err
		}
		miners = append(miners, miner)
	}
	return miners, rows.Err()
}

// UpdateStatus 更新矿工状态
func (r *MinerRepository) UpdateStatus(id, status string) error {
	query := `
		UPDATE miners SET
			status = $1, updated_at = $2, version = version + 1
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// UpdateMetrics 更新矿工指标
func (r *MinerRepository) UpdateMetrics(id string, hashRate, powerDraw float64, temperature float64, fanSpeed int) error {
	query := `
		UPDATE miners SET
			hash_rate = $1, power_consumption = $2, updated_at = $3,
			version = version + 1
		WHERE id = $4
	`

	_, err := r.db.Exec(query, hashRate, powerDraw, time.Now(), id)
	// 在实际实现中，还应该插入到mining_metrics表中
	_ = temperature
	_ = fanSpeed
	return err
}

// GetOnlineCount 获取在线矿工数量
func (r *MinerRepository) GetOnlineCount() (int, error) {
	query := "SELECT COUNT(*) FROM miners WHERE status IN ('ONLINE', 'THROTTLED')"
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// AlertRepository 警报数据仓储
type AlertRepository struct {
	db *sql.DB
}

// NewAlertRepository 创建警报仓储
func NewAlertRepository(db *sql.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Create 创建新警报
func (r *AlertRepository) Create(alert *models.Alert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	query := `
		INSERT INTO alerts (
			id, severity, category, title, description, source,
			entity_id, entity_type, status, metadata, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 1)
	`

	_, err := r.db.Exec(query,
		alert.ID, alert.Severity, alert.Category, alert.Title, alert.Description,
		alert.Source, alert.EntityID, alert.EntityType, alert.Status, alert.Metadata,
		alert.CreatedAt, alert.UpdatedAt,
	)
	return err
}

// GetActive 获取所有活动警报
func (r *AlertRepository) GetActive(severity string, limit, offset int) ([]models.Alert, error) {
	query := `
		SELECT id, severity, category, title, description, source,
			entity_id, entity_type, status, assigned_to, resolved_at, resolution,
			metadata, created_at, updated_at, version
		FROM alerts WHERE status = 'ACTIVE'
	`
	args := []interface{}{}
	argIndex := 1

	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(
			&alert.ID, &alert.Severity, &alert.Category, &alert.Title, &alert.Description,
			&alert.Source, &alert.EntityID, &alert.EntityType, &alert.Status,
			&alert.AssignedTo, &alert.ResolvedAt, &alert.Resolution, &alert.Metadata,
			&alert.CreatedAt, &alert.UpdatedAt, &alert.Version,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

// Acknowledge 确认警报
func (r *AlertRepository) Acknowledge(id, assignedTo string) error {
	query := `
		UPDATE alerts SET
			status = 'ACKNOWLEDGED', assigned_to = $1, updated_at = $2, version = version + 1
		WHERE id = $3
	`

	_, err := r.db.Exec(query, assignedTo, time.Now(), id)
	return err
}

// Resolve 解决警报
func (r *AlertRepository) Resolve(id, resolution string) error {
	query := `
		UPDATE alerts SET
			status = 'RESOLVED', resolution = $1, resolved_at = $2,
			updated_at = $3, version = version + 1
		WHERE id = $4
	`

	_, err := r.db.Exec(query, resolution, time.Now(), time.Now(), id)
	return err
}

// CountBySeverity 按严重程度统计警报数量
func (r *AlertRepository) CountBySeverity() (map[string]int, error) {
	query := "SELECT severity, COUNT(*) FROM alerts WHERE status = 'ACTIVE' GROUP BY severity"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, err
		}
		result[severity] = count
	}
	return result, rows.Err()
}

// AuditLogRepository 审计日志仓储
type AuditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository 创建审计日志仓储
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create 创建审计日志（追加操作）
func (r *AuditLogRepository) Create(log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	log.Timestamp = time.Now()

	query := `
		INSERT INTO audit_logs (
			id, user_id, user_role, action, resource_type, resource_id,
			ip_address, user_agent, request_body, response_code,
			previous_hash, current_hash, nonce, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Exec(query,
		log.ID, log.UserID, log.UserRole, log.Action, log.ResourceType, log.ResourceID,
		log.IPAddress, log.UserAgent, log.RequestBody, log.ResponseCode,
		log.PreviousHash, log.CurrentHash, log.Nonce, log.Timestamp,
	)
	return err
}

// GetByID 通过ID获取审计日志
func (r *AuditLogRepository) GetByID(id string) (*models.AuditLog, error) {
	query := `
		SELECT id, user_id, user_role, action, resource_type, resource_id,
			ip_address, user_agent, request_body, response_code,
			previous_hash, current_hash, nonce, timestamp
		FROM audit_logs WHERE id = $1
	`

	var log models.AuditLog
	err := r.db.QueryRow(query, id).Scan(
		&log.ID, &log.UserID, &log.UserRole, &log.Action, &log.ResourceType, &log.ResourceID,
		&log.IPAddress, &log.UserAgent, &log.RequestBody, &log.ResponseCode,
		&log.PreviousHash, &log.CurrentHash, &log.Nonce, &log.Timestamp,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &log, err
}

// GetByUser 获取用户的审计日志
func (r *AuditLogRepository) GetByUser(userID string, limit, offset int) ([]models.AuditLog, error) {
	query := `
		SELECT id, user_id, user_role, action, resource_type, resource_id,
			ip_address, user_agent, request_body, response_code,
			previous_hash, current_hash, nonce, timestamp
		FROM audit_logs WHERE user_id = $1
		ORDER BY timestamp DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.UserRole, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.IPAddress, &log.UserAgent, &log.RequestBody, &log.ResponseCode,
			&log.PreviousHash, &log.CurrentHash, &log.Nonce, &log.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// GetByResource 获取指定资源的审计日志
func (r *AuditLogRepository) GetByResource(resourceType, resourceID string, limit, offset int) ([]models.AuditLog, error) {
	query := `
		SELECT id, user_id, user_role, action, resource_type, resource_id,
			ip_address, user_agent, request_body, response_code,
			previous_hash, current_hash, nonce, timestamp
		FROM audit_logs WHERE resource_type = $1 AND resource_id = $2
		ORDER BY timestamp DESC LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(query, resourceType, resourceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.UserRole, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.IPAddress, &log.UserAgent, &log.RequestBody, &log.ResponseCode,
			&log.PreviousHash, &log.CurrentHash, &log.Nonce, &log.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// GetChainHead 获取链头（最新哈希）
func (r *AuditLogRepository) GetChainHead() (string, int64, error) {
	query := "SELECT current_hash, nonce, timestamp FROM audit_logs ORDER BY nonce DESC LIMIT 1"
	var hash string
	var nonce int64
	var timestamp time.Time
	err := r.db.QueryRow(query).Scan(&hash, &nonce, &timestamp)
	if err == sql.ErrNoRows {
		return "GENESIS", 0, nil
	}
	return hash, nonce, err
}

// GetCount 获取审计日志总数
func (r *AuditLogRepository) GetCount() (int64, error) {
	query := "SELECT COUNT(*) FROM audit_logs"
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// ReportRepository 报告数据仓储
type ReportRepository struct {
	db *sql.DB
}

// NewReportRepository 创建报告仓储
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create 创建新报告
func (r *ReportRepository) Create(report *models.Report) error {
	if report.ID == "" {
		report.ID = uuid.New().String()
	}
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()

	query := `
		INSERT INTO reports (
			id, report_type, title, description, period_start, period_end,
			status, generated_by, format, parameters, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 1)
	`

	_, err := r.db.Exec(query,
		report.ID, report.ReportType, report.Title, report.Description,
		report.PeriodStart, report.PeriodEnd, report.Status, report.GeneratedBy,
		report.Format, report.Parameters, report.CreatedAt, report.UpdatedAt,
	)
	return err
}

// GetByID 通过ID获取报告
func (r *ReportRepository) GetByID(id string) (*models.Report, error) {
	query := `
		SELECT id, report_type, title, description, period_start, period_end,
			status, generated_by, file_path, file_size, format, checksum,
			parameters, created_at, updated_at, version
		FROM reports WHERE id = $1
	`

	var report models.Report
	err := r.db.QueryRow(query, id).Scan(
		&report.ID, &report.ReportType, &report.Title, &report.Description,
		&report.PeriodStart, &report.PeriodEnd, &report.Status, &report.GeneratedBy,
		&report.FilePath, &report.FileSize, &report.Format, &report.Checksum,
		&report.Parameters, &report.CreatedAt, &report.UpdatedAt, &report.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &report, err
}

// GetAll 获取所有报告
func (r *ReportRepository) GetAll(status, reportType string, limit, offset int) ([]models.Report, error) {
	query := `
		SELECT id, report_type, title, description, period_start, period_end,
			status, generated_by, file_path, file_size, format, checksum,
			parameters, created_at, updated_at, version
		FROM reports
	`
	args := []interface{}{}
	argIndex := 1
	conditions := ""

	if status != "" {
		conditions += fmt.Sprintf(" status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	if reportType != "" {
		if conditions != "" {
			conditions += " AND"
		}
		conditions += fmt.Sprintf(" report_type = $%d", argIndex)
		args = append(args, reportType)
		argIndex++
	}
	if conditions != "" {
		query += " WHERE" + conditions
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var report models.Report
		err := rows.Scan(
			&report.ID, &report.ReportType, &report.Title, &report.Description,
			&report.PeriodStart, &report.PeriodEnd, &report.Status, &report.GeneratedBy,
			&report.FilePath, &report.FileSize, &report.Format, &report.Checksum,
			&report.Parameters, &report.CreatedAt, &report.UpdatedAt, &report.Version,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

// UpdateStatus 更新报告状态
func (r *ReportRepository) UpdateStatus(id, status, filePath, checksum string, fileSize int64) error {
	query := `
		UPDATE reports SET
			status = $1, file_path = $2, file_size = $3, checksum = $4,
			updated_at = $5, version = version + 1
		WHERE id = $6
	`

	_, err := r.db.Exec(query, status, filePath, fileSize, checksum, time.Now(), id)
	return err
}

// UserRepository 用户数据仓储
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建新用户
func (r *UserRepository) Create(user *models.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (
			id, username, email, password_hash, role, department,
			status, mfa_enabled, mfa_secret, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 1)
	`

	_, err := r.db.Exec(query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role,
		user.Department, user.Status, user.MFAEnabled, user.MFASecret,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// GetByID 通过ID获取用户
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, department,
			status, last_login, mfa_enabled, mfa_secret, created_at, updated_at, version
		FROM users WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.Department, &user.Status, &user.LastLogin, &user.MFAEnabled,
		&user.MFASecret, &user.CreatedAt, &user.UpdatedAt, &user.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

// GetByUsername 通过用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, department,
			status, last_login, mfa_enabled, mfa_secret, created_at, updated_at, version
		FROM users WHERE username = $1
	`

	var user models.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.Department, &user.Status, &user.LastLogin, &user.MFAEnabled,
		&user.MFASecret, &user.CreatedAt, &user.UpdatedAt, &user.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(id string) error {
	query := `
		UPDATE users SET last_login = $1, updated_at = $2, version = version + 1
		WHERE id = $3
	`

	_, err := r.db.Exec(query, time.Now(), time.Now(), id)
	return err
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(id, status string) error {
	query := `
		UPDATE users SET status = $1, updated_at = $2, version = version + 1
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// GetByRole 按角色获取用户
func (r *UserRepository) GetByRole(role string) ([]models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, department,
			status, last_login, mfa_enabled, mfa_secret, created_at, updated_at, version
		FROM users WHERE role = $1
	`

	rows, err := r.db.Query(query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
			&user.Department, &user.Status, &user.LastLogin, &user.MFAEnabled,
			&user.MFASecret, &user.CreatedAt, &user.UpdatedAt, &user.Version,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// FreezeOrderRepository 冻结订单仓储
type FreezeOrderRepository struct {
	db *sql.DB
}

// NewFreezeOrderRepository 创建冻结订单仓储
func NewFreezeOrderRepository(db *sql.DB) *FreezeOrderRepository {
	return &FreezeOrderRepository{db: db}
}

// Create 创建冻结订单
func (r *FreezeOrderRepository) Create(order *models.FreezeOrder) error {
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	query := `
		INSERT INTO freeze_orders (
			id, order_type, entity_type, entity_id, reason, legal_basis,
			issued_by, issuer_title, effective_from, effective_to, status,
			metadata, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 1)
	`

	_, err := r.db.Exec(query,
		order.ID, order.OrderType, order.EntityType, order.EntityID, order.Reason,
		order.LegalBasis, order.IssuedBy, order.IssuerTitle, order.EffectiveFrom,
		order.EffectiveTo, order.Status, order.Metadata, order.CreatedAt, order.UpdatedAt,
	)
	return err
}

// GetActive 获取活动冻结订单
func (r *FreezeOrderRepository) GetActive(entityType string) ([]models.FreezeOrder, error) {
	query := `
		SELECT id, order_type, entity_type, entity_id, reason, legal_basis,
			issued_by, issuer_title, effective_from, effective_to, status,
			metadata, created_at, updated_at, version
		FROM freeze_orders
		WHERE status = 'ACTIVE' AND effective_from <= NOW()
		AND (effective_to IS NULL OR effective_to > NOW())
	`
	args := []interface{}{}

	if entityType != "" {
		query += " AND entity_type = $1"
		args = append(args, entityType)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.FreezeOrder
	for rows.Next() {
		var order models.FreezeOrder
		err := rows.Scan(
			&order.ID, &order.OrderType, &order.EntityType, &order.EntityID, &order.Reason,
			&order.LegalBasis, &order.IssuedBy, &order.IssuerTitle, &order.EffectiveFrom,
			&order.EffectiveTo, &order.Status, &order.Metadata, &order.CreatedAt, &order.UpdatedAt,
			&order.Version,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

// Revoke 撤销冻结订单
func (r *FreezeOrderRepository) Revoke(id string) error {
	query := `
		UPDATE freeze_orders SET
			status = 'REVOKED', updated_at = $1, version = version + 1
		WHERE id = $2
	`

	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// EmergencyStopRepository 紧急停止仓储
type EmergencyStopRepository struct {
	db *sql.DB
}

// NewEmergencyStopRepository 创建紧急停止仓储
func NewEmergencyStopRepository(db *sql.DB) *EmergencyStopRepository {
	return &EmergencyStopRepository{db: db}
}

// Create 创建紧急停止记录
func (r *EmergencyStopRepository) Create(stop *models.EmergencyStop) error {
	if stop.ID == "" {
		stop.ID = uuid.New().String()
	}
	stop.IssuedAt = time.Now()

	query := `
		INSERT INTO emergency_stops (
			id, stop_type, entity_id, reason, issued_by, issued_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query,
		stop.ID, stop.StopType, stop.EntityID, stop.Reason, stop.IssuedBy,
		stop.IssuedAt, stop.Status,
	)
	return err
}

// GetActive 获取活动紧急停止
func (r *EmergencyStopRepository) GetActive() ([]models.EmergencyStop, error) {
	query := `
		SELECT id, stop_type, entity_id, reason, issued_by, issued_at,
			resolved_at, resolution, status
		FROM emergency_stops WHERE status = 'ACTIVE'
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stops []models.EmergencyStop
	for rows.Next() {
		var stop models.EmergencyStop
		err := rows.Scan(
			&stop.ID, &stop.StopType, &stop.EntityID, &stop.Reason, &stop.IssuedBy,
			&stop.IssuedAt, &stop.ResolvedAt, &stop.Resolution, &stop.Status,
		)
		if err != nil {
			return nil, err
		}
		stops = append(stops, stop)
	}
	return stops, rows.Err()
}

// Resolve 解决紧急停止
func (r *EmergencyStopRepository) Resolve(id, resolution string) error {
	query := `
		UPDATE emergency_stops SET
			status = 'RESOLVED', resolution = $1, resolved_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, resolution, time.Now(), id)
	return err
}

// PolicyRuleRepository 策略规则仓储
type PolicyRuleRepository struct {
	db *sql.DB
}

// NewPolicyRuleRepository 创建策略规则仓储
func NewPolicyRuleRepository(db *sql.DB) *PolicyRuleRepository {
	return &PolicyRuleRepository{db: db}
}

// Create 创建策略规则
func (r *PolicyRuleRepository) Create(rule *models.PolicyRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	query := `
		INSERT INTO policy_rules (
			id, rule_id, category, rule_type, description, conditions,
			actions, effective_from, effective_to, status, priority,
			created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 1)
	`

	_, err := r.db.Exec(query,
		rule.ID, rule.RuleID, rule.Category, rule.RuleType, rule.Description,
		rule.Conditions, rule.Actions, rule.EffectiveFrom, rule.EffectiveTo,
		rule.Status, rule.Priority, rule.CreatedAt, rule.UpdatedAt,
	)
	return err
}

// GetActive 获取活动策略规则
func (r *PolicyRuleRepository) GetActive(category string) ([]models.PolicyRule, error) {
	query := `
		SELECT id, rule_id, category, rule_type, description, conditions,
			actions, effective_from, effective_to, status, priority,
			created_at, updated_at, version
		FROM policy_rules
		WHERE status = 'ACTIVE' AND effective_from <= NOW()
		AND (effective_to IS NULL OR effective_to > NOW())
	`
	args := []interface{}{}

	if category != "" {
		query += " AND category = $1"
		args = append(args, category)
	}
	query += " ORDER BY priority DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.PolicyRule
	for rows.Next() {
		var rule models.PolicyRule
		err := rows.Scan(
			&rule.ID, &rule.RuleID, &rule.Category, &rule.RuleType, &rule.Description,
			&rule.Conditions, &rule.Actions, &rule.EffectiveFrom, &rule.EffectiveTo,
			&rule.Status, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
			&rule.Version,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// UpdateStatus 更新策略规则状态
func (r *PolicyRuleRepository) UpdateStatus(id, status string) error {
	query := `
		UPDATE policy_rules SET
			status = $1, updated_at = $2, version = version + 1
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// EntityClusterRepository 实体集群仓储
type EntityClusterRepository struct {
	db *sql.DB
}

// NewEntityClusterRepository 创建实体集群仓储
func NewEntityClusterRepository(db *sql.DB) *EntityClusterRepository {
	return &EntityClusterRepository{db: db}
}

// Create 创建实体集群
func (r *EntityClusterRepository) Create(cluster *models.EntityCluster) error {
	if cluster.ID == "" {
		cluster.ID = uuid.New().String()
	}
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()

	entityIDsJSON, _ := json.Marshal(cluster.EntityIDs)

	query := `
		INSERT INTO entity_clusters (
			id, cluster_id, entity_type, entity_ids, risk_score, labels,
			first_seen, last_seen, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1)
	`

	_, err := r.db.Exec(query,
		cluster.ID, cluster.ClusterID, cluster.EntityType, entityIDsJSON,
		cluster.RiskScore, cluster.Labels, cluster.FirstSeen, cluster.LastSeen,
		cluster.CreatedAt, cluster.UpdatedAt,
	)
	return err
}

// GetHighRisk 获取高风险集群
func (r *EntityClusterRepository) GetHighRisk(minScore int, limit int) ([]models.EntityCluster, error) {
	query := `
		SELECT id, cluster_id, entity_type, entity_ids, risk_score, labels,
			first_seen, last_seen, created_at, updated_at, version
		FROM entity_clusters
		WHERE risk_score >= $1
		ORDER BY risk_score DESC LIMIT $2
	`

	rows, err := r.db.Query(query, minScore, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []models.EntityCluster
	for rows.Next() {
		var cluster models.EntityCluster
		var entityIDsJSON []byte
		err := rows.Scan(
			&cluster.ID, &cluster.ClusterID, &cluster.EntityType, &entityIDsJSON,
			&cluster.RiskScore, &cluster.Labels, &cluster.FirstSeen, &cluster.LastSeen,
			&cluster.CreatedAt, &cluster.UpdatedAt, &cluster.Version,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(entityIDsJSON, &cluster.EntityIDs)
		clusters = append(clusters, cluster)
	}
	return clusters, rows.Err()
}

// UpdateRiskScore 更新风险评分
func (r *EntityClusterRepository) UpdateRiskScore(id string, score int) error {
	query := `
		UPDATE entity_clusters SET
			risk_score = $1, last_seen = $2, updated_at = $3, version = version + 1
		WHERE id = $4
	`

	_, err := r.db.Exec(query, score, time.Now(), time.Now(), id)
	return err
}
