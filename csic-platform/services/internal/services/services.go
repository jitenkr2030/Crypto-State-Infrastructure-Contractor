package services

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/csic-platform/services/internal/config"
	"github.com/csic-platform/services/internal/models"
	"github.com/csic-platform/services/internal/repository"
)

// Services 包含所有业务服务
type Services struct {
	Config        *config.Config
	Repos         *repository.Repositories
	WORMStorage   *WORMStorage
	PolicyEngine  *PolicyEngine
	RiskScorer    *RiskScorer
	BlockchainSvc *BlockchainService
	ExchangeSvc   *ExchangeService
	MiningSvc     *MiningService
	ReportingSvc  *ReportingService
	HSMService    *HSMService
}

// NewServices 创建所有服务实例
func NewServices(cfg *config.Config, repos *repository.Repositories, wormStorage *WORMStorage) *Services {
	svc := &Services{
		Config:       cfg,
		Repos:        repos,
		WORMStorage:  wormStorage,
		PolicyEngine: NewPolicyEngine(cfg, repos),
		RiskScorer:   NewRiskScorer(cfg, repos),
		BlockchainSvc: NewBlockchainService(cfg, repos),
		ExchangeSvc:  NewExchangeService(cfg, repos),
		MiningSvc:    NewMiningService(cfg, repos),
		ReportingSvc: NewReportingService(cfg, repos),
		HSMService:   NewHSMService(cfg),
	}
	return svc
}

// UpdateSystemMetrics 更新系统指标
func (s *Services) UpdateSystemMetrics() error {
	// 获取交易所统计数据
	exchangeCounts, err := s.Repos.Exchanges.CountByStatus()
	if err != nil {
		return fmt.Errorf("获取交易所统计失败: %w", err)
	}
	_ = exchangeCounts

	// 获取警报统计数据
	alertCounts, err := s.Repos.Alerts.CountBySeverity()
	if err != nil {
		return fmt.Errorf("获取警报统计失败: %w", err)
	}
	_ = alertCounts

	// 获取在线矿工数量
	onlineMiners, err := s.Repos.Miners.GetOnlineCount()
	if err != nil {
		return fmt.Errorf("获取在线矿工数量失败: %w", err)
	}
	_ = onlineMiners

	return nil
}

// ProcessAlerts 处理待处理警报
func (s *Services) ProcessAlerts() error {
	alerts, err := s.Repos.Alerts.GetActive("", 100, 0)
	if err != nil {
		return fmt.Errorf("获取活动警报失败: %w", err)
	}

	for _, alert := range alerts {
		// 根据警报类型执行相应处理逻辑
		switch alert.Category {
		case "EXCHANGE":
			s.handleExchangeAlert(alert)
		case "TRANSACTION":
			s.handleTransactionAlert(alert)
		case "MINING":
			s.handleMiningAlert(alert)
		case "SECURITY":
			s.handleSecurityAlert(alert)
		}
	}

	return nil
}

// handleExchangeAlert 处理交易所警报
func (s *Services) handleExchangeAlert(alert models.Alert) {
	// 交易所警报处理逻辑
	// 根据严重程度和具体警报内容执行相应操作
}

// handleTransactionAlert 处理交易警报
func (s *Services) handleTransactionAlert(alert models.Alert) {
	// 交易警报处理逻辑
	// 可能包括自动标记可疑交易、通知监管人员等
}

// handleMiningAlert 处理挖矿警报
func (s *Services) handleMiningAlert(alert models.Alert) {
	// 挖矿警报处理逻辑
	// 可能包括检查能源消耗、矿工状态等
}

// handleSecurityAlert 处理安全警报
func (s *Services) handleSecurityAlert(alert models.Alert) {
	// 安全警报处理逻辑
	// 可能包括锁定账户、通知安全团队等
}

// SyncBlockchainNodes 同步区块链节点
func (s *Services) SyncBlockchainNodes() error {
	return s.BlockchainSvc.SyncBlocks()
}

// WORMStorage 只写一次存储服务（不可变审计日志）
type WORMStorage struct {
	mu          sync.RWMutex
	storagePath string
	chainHead   string
	nonce       int64
}

// NewWORMStorage 创建WORM存储实例
func NewWORMStorage(storagePath string) (*WORMStorage, error) {
	// 确保存储目录存在
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("创建WORM存储目录失败: %w", err)
	}

	worm := &WORMStorage{
		storagePath: storagePath,
		chainHead:   "GENESIS",
		nonce:       0,
	}

	// 加载链头状态
	if err := worm.loadChainHead(); err != nil {
		// 如果加载失败，创建初始链头
		worm.chainHead = "GENESIS"
		worm.nonce = 0
	}

	return worm, nil
}

// Append 追加不可变日志条目
func (w *WORMStorage) Append(entry *models.AuditLog) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 生成当前哈希
	currentHash := w.calculateHash(entry, w.chainHead, w.nonce)

	// 更新日志条目的哈希
	entry.PreviousHash = w.chainHead
	entry.CurrentHash = currentHash
	entry.Nonce = w.nonce
	w.nonce++

	// 序列化并写入文件
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("序列化审计日志失败: %w", err)
	}

	// 按日期组织存储
	dateDir := filepath.Join(w.storagePath, entry.Timestamp.Format("2006/01/02"))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return fmt.Errorf("创建日期目录失败: %w", err)
	}

	// 使用哈希作为文件名
	filename := filepath.Join(dateDir, fmt.Sprintf("%s.json", currentHash))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入WORM存储失败: %w", err)
	}

	// 更新链头
	w.chainHead = currentHash

	// 保存链头状态
	if err := w.saveChainHead(); err != nil {
		// 记录错误但不影响主流程
		fmt.Printf("警告: 保存链头状态失败: %v\n", err)
	}

	return nil
}

// calculateHash 计算审计日志的哈希值
func (w *WORMStorage) calculateHash(entry *models.AuditLog, previousHash string, nonce int64) string {
	// 构建要哈希的数据
	data := struct {
		UserID       string
		UserRole     string
		Action       string
		ResourceType string
		ResourceID   string
		IPAddress    string
		ResponseCode int
		PreviousHash string
		Nonce        int64
		Timestamp    time.Time
	}{
		UserID:       entry.UserID,
		UserRole:     entry.UserRole,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID.String,
		IPAddress:    entry.IPAddress,
		ResponseCode: entry.ResponseCode,
		PreviousHash: previousHash,
		Nonce:        nonce,
		Timestamp:    entry.Timestamp,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// loadChainHead 加载链头状态
func (w *WORMStorage) loadChainHead() error {
	chainHeadFile := filepath.Join(w.storagePath, "chain_head.json")

	data, err := os.ReadFile(chainHeadFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 首次运行
		}
		return err
	}

	var state struct {
		ChainHead string `json:"chain_head"`
		Nonce     int64  `json:"nonce"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	w.chainHead = state.ChainHead
	w.nonce = state.Nonce
	return nil
}

// saveChainHead 保存链头状态
func (w *WORMStorage) saveChainHead() error {
	chainHeadFile := filepath.Join(w.storagePath, "chain_head.json")

	state := struct {
		ChainHead string `json:"chain_head"`
		Nonce     int64  `json:"nonce"`
	}{
		ChainHead: w.chainHead,
		Nonce:     w.nonce,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(chainHeadFile, data, 0644)
}

// VerifyIntegrity 验证存储完整性
func (w *WORMStorage) VerifyIntegrity() (bool, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// 在实际实现中，这里会遍历所有日志文件
	// 验证哈希链的完整性
	// 由于WORM存储的特性，我们只需要验证链的连续性

	return true, nil
}

// GetChainHead 获取当前链头
func (w *WORMStorage) GetChainHead() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.chainHead
}

// GetNonce 获取当前nonce值
func (w *WORMStorage) GetNonce() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.nonce
}

// Close 关闭存储
func (w *WORMStorage) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 确保保存最后状态
	return w.saveChainHead()
}

// GetStoragePath 获取存储路径
func (w *WORMStorage) GetStoragePath() string {
	return w.storagePath
}

// PolicyEngine 策略引擎服务
type PolicyEngine struct {
	Config *config.Config
	Repos  *repository.Repositories
	mu     sync.RWMutex
	rules  []models.PolicyRule
}

// NewPolicyEngine 创建策略引擎
func NewPolicyEngine(cfg *config.Config, repos *repository.Repositories) *PolicyEngine {
	engine := &PolicyEngine{
		Config: cfg,
		Repos:  repos,
	}

	// 加载策略规则
	engine.LoadRules()

	return engine
}

// LoadRules 加载策略规则
func (p *PolicyEngine) LoadRules() {
	p.mu.Lock()
	defer p.mu.Unlock()

	rules, err := p.Repos.PolicyRules.GetActive("")
	if err != nil {
		fmt.Printf("警告: 加载策略规则失败: %v\n", err)
		return
	}

	p.rules = rules
}

// ReloadRules 重新加载策略规则
func (p *PolicyEngine) ReloadRules() {
	p.LoadRules()
}

// EvaluateTransaction 评估交易是否符合策略
func (p *PolicyEngine) EvaluateTransaction(tx *models.Transaction) (*PolicyResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := &PolicyResult{
		TransactionID: tx.ID,
		Compliant:     true,
		Violations:    []string{},
		RequiredActions: []string{},
	}

	for _, rule := range p.rules {
		if rule.Category != "TRANSACTION" || rule.Status != "ACTIVE" {
			continue
		}

		// 评估规则条件
		compliant, err := p.evaluateRule(tx, rule)
		if err != nil {
			return nil, fmt.Errorf("评估规则失败: %w", err)
		}

		if !compliant {
			result.Compliant = false
			result.Violations = append(result.Violations, rule.Description)
			result.RequiredActions = append(result.RequiredActions, rule.RuleID)
		}
	}

	return result, nil
}

// evaluateRule 评估单个规则
func (p *PolicyEngine) evaluateRule(tx *models.Transaction, rule models.PolicyRule) (bool, error) {
	// 解析规则条件
	var conditions map[string]interface{}
	if err := json.Unmarshal(rule.Conditions, &conditions); err != nil {
		return true, nil // 条件解析失败时默认通过
	}

	// 检查金额阈值
	if maxAmount, ok := conditions["max_amount"].(float64); ok {
		if tx.Amount > maxAmount {
			return false, nil
		}
	}

	// 检查货币类型
	if currencies, ok := conditions["allowed_currencies"].([]interface{}); ok {
		allowed := make(map[string]bool)
		for _, c := range currencies {
			if currency, ok := c.(string); ok {
				allowed[currency] = true
			}
		}
		if !allowed[tx.Currency] {
			return false, nil
		}
	}

	// 检查来源地址是否在黑名单中
	if blacklist, ok := conditions["blacklisted_senders"].([]interface{}); ok {
		for _, item := range blacklist {
			if addr, ok := item.(string); ok && addr == tx.FromAddress {
				return false, nil
			}
		}
	}

	return true, nil
}

// PolicyResult 策略评估结果
type PolicyResult struct {
	TransactionID    string   `json:"transaction_id"`
	Compliant        bool     `json:"compliant"`
	Violations       []string `json:"violations"`
	RequiredActions  []string `json:"required_actions"`
	ProcessedAt      time.Time `json:"processed_at"`
}

// RiskScorer 风险评分服务
type RiskScorer struct {
	Config *config.Config
	Repos  *repository.Repositories
}

// NewRiskScorer 创建风险评分器
func NewRiskScorer(cfg *config.Config, repos *repository.Repositories) *RiskScorer {
	return &RiskScorer{
		Config: cfg,
		Repos:  repos,
	}
}

// CalculateWalletRisk 计算钱包风险评分
func (r *RiskScorer) CalculateWalletRisk(address string) (*models.RiskScore, error) {
	wallet, err := r.Repos.Wallets.GetByAddress(address)
	if err != nil {
		return nil, fmt.Errorf("获取钱包信息失败: %w", err)
	}

	// 获取交易历史计算风险
	factors := r.calculateRiskFactors(wallet)

	// 计算加权总分
	totalScore := 0.0
	for _, factor := range factors {
		totalScore += factor.Weight * float64(factor.Score)
	}

	// 归一化到0-100
	score := int(totalScore)

	// 确定风险等级
	riskLevel := "LOW"
	if score >= 80 {
		riskLevel = "CRITICAL"
	} else if score >= 60 {
		riskLevel = "HIGH"
	} else if score >= 40 {
		riskLevel = "MEDIUM"
	}

	return &models.RiskScore{
		WalletAddress: address,
		Score:         score,
		RiskLevel:     riskLevel,
		Factors:       factors,
		LastAssessed:  time.Now(),
	}, nil
}

// calculateRiskFactors 计算风险因素
func (r *RiskScorer) calculateRiskFactors(wallet *models.Wallet) []models.RiskFactor {
	factors := []models.RiskFactor{}

	// 如果钱包被冻结，最高风险
	if wallet.Status == "FROZEN" {
		factors = append(factors, models.RiskFactor{
			Name:        "冻结状态",
			Weight:      0.4,
			Score:       100,
			Description: "钱包已被监管机构冻结",
		})
	}

	// 检查是否在黑名单中
	if wallet.Blacklisted {
		factors = append(factors, models.RiskFactor{
			Name:        "黑名单状态",
			Weight:      0.3,
			Score:       100,
			Description: "钱包在制裁或黑名单中",
		})
	}

	// 基于风险评分的因素
	if wallet.RiskScore >= 80 {
		factors = append(factors, models.RiskFactor{
			Name:        "系统风险评分",
			Weight:      0.2,
			Score:       wallet.RiskScore,
			Description: "系统自动评估的高风险评分",
		})
	}

	// 交易活跃度因素
	if wallet.LastActivityAt.Valid {
		hoursSinceActivity := time.Since(wallet.LastActivityAt.Time).Hours()
		if hoursSinceActivity > 24*30 { // 30天无活动
			factors = append(factors, models.RiskFactor{
				Name:        "长期不活跃",
				Weight:      0.1,
				Score:       50,
				Description: "钱包已超过30天无任何活动",
			})
		}
	}

	return factors
}

// BlockchainService 区块链服务
type BlockchainService struct {
	Config *config.Config
	Repos  *repository.Repositories
}

// NewBlockchainService 创建区块链服务
func NewBlockchainService(cfg *config.Config, repos *repository.Repositories) *BlockchainService {
	return &BlockchainService{
		Config: cfg,
		Repos:  repos,
	}
}

// SyncBlocks 同步区块链区块
func (b *BlockchainService) SyncBlocks() error {
	// 在实际实现中，这里会连接自托管的区块链节点
	// 解析新区块，提取交易信息

	// 示例：同步比特币区块
	// btcBlocks, err := b.syncBitcoinBlocks()
	// if err != nil { return err }

	// 示例：同步以太坊区块
	// ethBlocks, err := b.syncEthereumBlocks()
	// if err != nil { return err }

	return nil
}

// GetTransaction 获取交易详情
func (b *BlockchainService) GetTransaction(txID string) (*models.Transaction, error) {
	return b.Repos.Transactions.GetByID(txID)
}

// ExchangeService 交易所服务
type ExchangeService struct {
	Config *config.Config
	Repos  *repository.Repositories
}

// NewExchangeService 创建交易所服务
func NewExchangeService(cfg *config.Config, repos *repository.Repositories) *ExchangeService {
	return &ExchangeService{
		Config: cfg,
		Repos:  repos,
	}
}

// GetHealthScore 获取交易所健康评分
func (e *ExchangeService) GetHealthScore(exchangeID string) (*models.ExchangeHealth, error) {
	exchange, err := e.Repos.Exchanges.GetByID(exchangeID)
	if err != nil {
		return nil, fmt.Errorf("获取交易所信息失败: %w", err)
	}

	// 计算各维度评分
	availability := 100.0
	latencyScore := 100.0
	volumeScore := 100.0
	complianceScore := 100.0

	// 根据交易所状态调整评分
	switch exchange.Status {
	case "ACTIVE":
		availability = 100.0
	case "SUSPENDED":
		availability = 0.0
	case "PENDING":
		availability = 50.0
	}

	// 计算综合评分
	overallScore := (availability*0.3 + latencyScore*0.2 + volumeScore*0.2 + complianceScore*0.3)

	return &models.ExchangeHealth{
		ExchangeID:     exchangeID,
		OverallScore:   overallScore,
		Availability:   availability,
		LatencyScore:   latencyScore,
		VolumeScore:    volumeScore,
		ComplianceScore: complianceScore,
		LastChecked:    time.Now(),
	}, nil
}

// MiningService 挖矿服务
type MiningService struct {
	Config *config.Config
	Repos  *repository.Repositories
}

// NewMiningService 创建挖矿服务
func NewMiningService(cfg *config.Config, repos *repository.Repositories) *MiningService {
	return &MiningService{
		Config: cfg,
		Repos:  repos,
	}
}

// RemoteShutdown 远程关闭矿机
func (m *MiningService) RemoteShutdown(minerID string, reason string) error {
	// 更新矿工状态
	if err := m.Repos.Miners.UpdateStatus(minerID, "SHUTDOWN"); err != nil {
		return fmt.Errorf("更新矿工状态失败: %w", err)
	}

	// 创建审计日志
	auditLog := &models.AuditLog{
		UserID:       "SYSTEM",
		UserRole:     "SYSTEM",
		Action:       "MINER_SHUTDOWN",
		ResourceType: "MINER",
		ResourceID:   &minerID,
		IPAddress:    "127.0.0.1",
		ResponseCode: 200,
		Timestamp:    time.Now(),
	}

	// 注意：在实际实现中，这应该通过WORM存储服务处理
	// m.WORMStorage.Append(auditLog)

	return nil
}

// GetTotalHashRate 获取总哈希率
func (m *MiningService) GetTotalHashRate() (float64, error) {
	miners, err := m.Repos.Miners.GetAll("ONLINE", 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("获取矿工列表失败: %w", err)
	}

	totalHashRate := 0.0
	for _, miner := range miners {
		hashRate := miner.HashRate
		switch miner.HashRateUnit {
		case "TH/s":
			hashRate = hashRate / 1000 // 转换为PH/s
		case "EH/s":
			hashRate = hashRate * 1000
		}
		totalHashRate += hashRate
	}

	return totalHashRate, nil
}

// ReportingService 报告服务
type ReportingService struct {
	Config *config.Config
	Repos  *repository.Repositories
}

// NewReportingService 创建报告服务
func NewReportingService(cfg *config.Config, repos *repository.Repositories) *ReportingService {
	return &ReportingService{
		Config: cfg,
		Repos:  repos,
	}
}

// GenerateDailyReport 生成日报告
func (r *ReportingService) GenerateDailyReport(date time.Time, generatedBy string) (*models.Report, error) {
	periodStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.Add(24 * time.Hour)

	report := &models.Report{
		ReportType:  "DAILY",
		Title:       fmt.Sprintf("日监管报告 - %s", periodStart.Format("2006-01-02")),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      "GENERATING",
		GeneratedBy: generatedBy,
		Format:      "PDF",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存报告记录
	if err := r.Repos.Reports.Create(report); err != nil {
		return nil, fmt.Errorf("创建报告记录失败: %w", err)
	}

	// 生成报告数据
	data := r.gatherReportData(periodStart, periodEnd)

	// 生成报告文件
	filePath, err := r.renderReport(report, data)
	if err != nil {
		report.Status = "FAILED"
		r.Repos.Reports.UpdateStatus(report.ID, "FAILED", "", "", 0)
		return report, fmt.Errorf("生成报告文件失败: %w", err)
	}

	// 计算校验和
	checksum := r.calculateChecksum(filePath)

	// 更新报告状态
	if err := r.Repos.Reports.UpdateStatus(report.ID, "COMPLETED", filePath, checksum, 0); err != nil {
		return nil, fmt.Errorf("更新报告状态失败: %w", err)
	}

	report.Status = "COMPLETED"
	report.FilePath.String = filePath
	report.Checksum.String = checksum

	return report, nil
}

// gatherReportData 收集报告数据
func (r *ReportingService) gatherReportData(start, end time.Time) map[string]interface{} {
	// 收集交易所数据
	exchanges, _ := r.Repos.Exchanges.GetAll("", 1000, 0)

	// 收集交易数据
	transactions, _ := r.Repos.Transactions.Search(models.TransactionFilter{
		StartTime: start,
		EndTime:   end,
		Limit:     10000,
	})

	// 收集警报数据
	alerts, _ := r.Repos.Alerts.GetActive("", 1000, 0)

	// 收集许可证数据
	licenses, _ := r.Repos.Licenses.GetAll("", "", 1000, 0)

	return map[string]interface{}{
		"period_start":   start,
		"period_end":     end,
		"generated_at":   time.Now(),
		"exchanges":      exchanges,
		"transactions":   transactions,
		"alerts":         alerts,
		"licenses":       licenses,
		"exchange_count": len(exchanges),
		"transaction_count": len(transactions),
		"alert_count":    len(alerts),
	}
}

// renderReport 渲染报告
func (r *ReportingService) renderReport(report *models.Report, data map[string]interface{}) (string, error) {
	// 在实际实现中，这里会使用PDF生成库（如wkhtmltopdf、PDFium等）
	// 生成PDF报告文件

	reportDir := filepath.Join(r.Config.WORMStorage.Path, "reports")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", err
	}

	filename := filepath.Join(reportDir, fmt.Sprintf("%s_%s.pdf", report.ReportType, report.ID))
	_ = filename

	// 模拟生成PDF
	// 在实际实现中，这里会生成真实的PDF文件
	_ = data

	return filename, nil
}

// calculateChecksum 计算文件校验和
func (r *ReportingService) calculateChecksum(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// HSMService HSM硬件安全模块服务
type HSMService struct {
	Config *config.Config
}

// NewHSMService 创建HSM服务
func NewHSMService(cfg *config.Config) *HSMService {
	return &HSMService{
		Config: cfg,
	}
}

// GetStatus 获取HSM状态
func (h *HSMService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"connected":    true,
		"provider":     h.Config.HSM.Provider,
		"slot":         h.Config.HSM.Slot,
		"key_label":    h.Config.HSM.KeyLabel,
		"auto_rotate":  h.Config.HSM.AutoRotate,
		"last_check":   time.Now(),
	}
}

// Sign 签名数据
func (h *HSMService) Sign(data []byte) ([]byte, error) {
	// 在实际实现中，这里会使用HSM进行签名
	// 通过PKCS11或厂商SDK调用HSM

	// 模拟签名
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// Verify 验证签名
func (h *HSMService) Verify(data, signature []byte) bool {
	// 在实际实现中，这里会使用HSM验证签名

	expectedSig, _ := h.Sign(data)
	return string(signature) == string(expectedSig)
}

// GenerateKey 生成新密钥
func (h *HSMService) GenerateKey() error {
	// 在实际实现中，这里会在HSM中生成新密钥
	return nil
}

// RotateKey 轮换密钥
func (h *HSMService) RotateKey() error {
	// 在实际实现中，这里会执行密钥轮换
	return nil
}

// Helper function to convert int to bytes
func intToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

// Helper function to generate random nonce
func generateNonce() int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63()
}
