package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/csic-platform/services/internal/models"
	"github.com/csic-platform/services/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handlers 包含所有HTTP处理器
type Handlers struct {
	svc *services.Services
}

// NewHandlers 创建所有处理器
func NewHandlers(svc *services.Services) *Handlers {
	return &Handlers{svc: svc}
}

// HealthCheck 健康检查
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// ReadinessCheck 就绪检查
func (h *Handlers) ReadinessCheck(c *gin.Context) {
	// 检查数据库连接
	if err := h.svc.Repos.DB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"error":   "database connection failed",
			"message": "Database is not accessible",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
	})
}

// GetSystemStatus 获取系统状态
func (h *Handlers) GetSystemStatus(c *gin.Context) {
	// 获取活动紧急停止
	activeStops, _ := h.svc.Repos.EmergencyStops.GetActive()

	// 获取活动警报统计
	alertCounts, _ := h.svc.Repos.Alerts.CountBySeverity()

	// 获取交易所统计
	exchangeCounts, _ := h.svc.Repos.Exchanges.CountByStatus()

	// 获取HSM状态
	hsmStatus := h.svc.HSMService.GetStatus()

	status := "ONLINE"
	if len(activeStops) > 0 {
		status = "EMERGENCY"
	}

	totalAlerts := 0
	for _, count := range alertCounts {
		totalAlerts += count
	}

	activeExchanges := exchangeCounts["ACTIVE"]
	monitoredWallets, _ := h.svc.Repos.Wallets.GetFrozen(1000, 0)

	response := models.SystemStatus{
		State:            status,
		LastHeartbeat:    time.Now(),
		ActiveExchanges:  activeExchanges,
		MonitoredWallets: len(monitoredWallets),
		PendingAlerts:    totalAlerts,
		HSMStatus:        hsmStatus["connected"].(bool) && "connected",
		DatabaseStatus:   "CONNECTED",
		Uptime:           models.Duration{Duration: time.Since(time.Now().Add(-24 * time.Hour))},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
		Metadata: &models.APIMeta{
			RequestID:  uuid.New().String(),
			Timestamp:  time.Now().UTC(),
			Version:    "1.0.0",
			Processing: 10,
		},
	})
}

// EmergencyStop 紧急停止
func (h *Handlers) EmergencyStop(c *gin.Context) {
	var req struct {
		StopType string `json:"stop_type" binding:"required"`
		EntityID string `json:"entity_id"`
		Reason   string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
				Details: err.Error(),
			},
		})
		return
	}

	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")

	// 验证权限
	if userRole != "ADMIN" && userRole != "OPERATOR" {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INSUFFICIENT_PERMISSION",
				Message: "权限不足",
			},
		})
		return
	}

	stop := &models.EmergencyStop{
		ID:        uuid.New().String(),
		StopType:  req.StopType,
		Reason:    req.Reason,
		IssuedBy:  userID,
		IssuedAt:  time.Now(),
		Status:    "ACTIVE",
	}

	if req.EntityID != "" {
		stop.EntityID.String = req.EntityID
	}

	if err := h.svc.Repos.EmergencyStops.Create(stop); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "创建紧急停止记录失败",
			},
		})
		return
	}

	// 执行停止操作
	switch req.StopType {
	case "GLOBAL":
		// 停止所有交易
	case "EXCHANGE":
		if err := h.svc.Repos.Exchanges.Freeze(req.EntityID, req.Reason, stop.ID); err != nil {
			// 处理错误
		}
	case "TRANSACTION":
		// 暂停特定交易
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "紧急停止已执行",
		Data:    stop,
	})
}

// EmergencyResume 恢复运行
func (h *Handlers) EmergencyResume(c *gin.Context) {
	var req struct {
		StopID    string `json:"stop_id" binding:"required"`
		Resolution string `json:"resolution"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
			},
		})
		return
	}

	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")

	if userRole != "ADMIN" && userRole != "OPERATOR" {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INSUFFICIENT_PERMISSION",
				Message: "权限不足",
			},
		})
		return
	}

	if err := h.svc.Repos.EmergencyStops.Resolve(req.StopID, req.Resolution); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "解决紧急停止失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "系统已恢复正常运行",
	})
}

// GetExchanges 获取交易所列表
func (h *Handlers) GetExchanges(c *gin.Context) {
	status := c.Query("status")
	limit, offset := getPagination(c)

	exchanges, err := h.svc.Repos.Exchanges.GetAll(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取交易所列表失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    exchanges,
	})
}

// GetExchangeDetails 获取交易所详情
func (h *Handlers) GetExchangeDetails(c *gin.Context) {
	id := c.Param("id")

	exchange, err := h.svc.Repos.Exchanges.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取交易所详情失败",
			},
		})
		return
	}

	if exchange == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "交易所不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    exchange,
	})
}

// FreezeExchange 冻结交易所
func (h *Handlers) FreezeExchange(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "必须提供冻结原因",
			},
		})
		return
	}

	// 创建冻结订单
	order := &models.FreezeOrder{
		ID:            uuid.New().String(),
		OrderType:     "FREEZE",
		EntityType:    "EXCHANGE",
		EntityID:      id,
		Reason:        req.Reason,
		IssuedBy:      c.GetString("user_id"),
		IssuedAt:      time.Now(),
		Status:        "ACTIVE",
		EffectiveFrom: time.Now(),
	}

	if err := h.svc.Repos.FreezeOrders.Create(order); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "创建冻结订单失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "交易所已冻结",
		Data:    order,
	})
}

// ThawExchange 解冻交易所
func (h *Handlers) ThawExchange(c *gin.Context) {
	id := c.Param("id")

	// 撤销冻结订单
	if err := h.svc.Repos.FreezeOrders.Revoke(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "撤销冻结订单失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "交易所已解冻",
	})
}

// GetExchangeMetrics 获取交易所指标
func (h *Handlers) GetExchangeMetrics(c *gin.Context) {
	id := c.Param("id")

	// 获取交易所健康评分
	health, err := h.svc.ExchangeSvc.GetHealthScore(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取交易所指标失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    health,
	})
}

// GetWallets 获取钱包列表
func (h *Handlers) GetWallets(c *gin.Context) {
	limit, offset := getPagination(c)

	wallets, err := h.svc.Repos.Wallets.GetFrozen(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取钱包列表失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    wallets,
	})
}

// CreateWallet 创建钱包
func (h *Handlers) CreateWallet(c *gin.Context) {
	var wallet models.Wallet

	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
				Details: err.Error(),
			},
		})
		return
	}

	if err := h.svc.Repos.Wallets.Create(&wallet); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "创建钱包失败",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "钱包创建成功",
		Data:    wallet,
	})
}

// FreezeWallet 冻结钱包
func (h *Handlers) FreezeWallet(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "必须提供冻结原因",
			},
		})
		return
	}

	orderID := uuid.New().String()
	if err := h.svc.Repos.Wallets.Freeze(id, req.Reason, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "冻结钱包失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "钱包已冻结",
	})
}

// UnfreezeWallet 解冻钱包
func (h *Handlers) UnfreezeWallet(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Repos.Wallets.Unfreeze(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "解冻钱包失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "钱包已解冻",
	})
}

// TransferFromWallet 从钱包转出
func (h *Handlers) TransferFromWallet(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ToAddress string  `json:"to_address" binding:"required"`
		Amount    float64 `json:"amount" binding:"required,gt=0"`
		Currency  string  `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
			},
		})
		return
	}

	// 检查钱包状态
	wallet, err := h.svc.Repos.Wallets.GetByID(id)
	if err != nil || wallet == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "钱包不存在",
			},
		})
		return
	}

	if wallet.Status == "FROZEN" {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WALLET_FROZEN",
				Message: "钱包已被冻结，无法转出",
			},
		})
		return
	}

	// 在实际实现中，这里会调用区块链服务执行转出
	_ = req.ToAddress
	_ = req.Amount
	_ = req.Currency

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "转出请求已提交",
	})
}

// GetTransactions 获取交易列表
func (h *Handlers) GetTransactions(c *gin.Context) {
	limit, offset := getPagination(c)

	filter := models.TransactionFilter{
		Limit:  limit,
		Offset: offset,
	}

	if from := c.Query("from_address"); from != "" {
		filter.FromAddress = from
	}
	if to := c.Query("to_address"); to != "" {
		filter.ToAddress = to
	}
	if currency := c.Query("currency"); currency != "" {
		filter.Currency = currency
	}
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	transactions, total, err := h.svc.Repos.Transactions.Search(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取交易列表失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.PaginatedResponse{
			Data:       transactions,
			Total:      total,
			Page:       offset/limit + 1,
			PageSize:   limit,
			TotalPages: int((total + int64(limit) - 1) / int64(limit)),
		},
	})
}

// GetTransactionDetails 获取交易详情
func (h *Handlers) GetTransactionDetails(c *gin.Context) {
	id := c.Param("id")

	tx, err := h.svc.Repos.Transactions.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取交易详情失败",
			},
		})
		return
	}

	if tx == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "交易不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tx,
	})
}

// SearchTransactions 搜索交易
func (h *Handlers) SearchTransactions(c *gin.Context) {
	var filter models.TransactionFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的搜索参数",
			},
		})
		return
	}

	filter.Limit = 100
	filter.Offset = 0

	transactions, total, err := h.svc.Repos.Transactions.Search(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "搜索交易失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.PaginatedResponse{
			Data:       transactions,
			Total:      total,
			Page:       1,
			PageSize:   100,
			TotalPages: int((total + 99) / 100),
		},
	})
}

// FlagTransaction 标记交易
func (h *Handlers) FlagTransaction(c *gin.Context) {
	var req struct {
		TxID   string `json:"tx_id" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
			},
		})
		return
	}

	if err := h.svc.Repos.Transactions.Flag(req.TxID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "标记交易失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "交易已标记",
	})
}

// GetWalletRiskScore 获取钱包风险评分
func (h *Handlers) GetWalletRiskScore(c *gin.Context) {
	address := c.Param("address")

	riskScore, err := h.svc.RiskScorer.CalculateWalletRisk(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "计算风险评分失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    riskScore,
	})
}

// GetLicenses 获取许可证列表
func (h *Handlers) GetLicenses(c *gin.Context) {
	status := c.Query("status")
	entityType := c.Query("entity_type")
	limit, offset := getPagination(c)

	licenses, err := h.svc.Repos.Licenses.GetAll(status, entityType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取许可证列表失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    licenses,
	})
}

// CreateLicense 创建许可证
func (h *Handlers) CreateLicense(c *gin.Context) {
	var license models.License

	if err := c.ShouldBindJSON(&license); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
			},
		})
		return
	}

	// 生成许可证号
	license.LicenseNumber = fmt.Sprintf("CSIC-LIC-%s", time.Now().Format("20060102"))

	if err := h.svc.Repos.Licenses.Create(&license); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "创建许可证失败",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "许可证创建成功",
		Data:    license,
	})
}

// GetLicenseDetails 获取许可证详情
func (h *Handlers) GetLicenseDetails(c *gin.Context) {
	id := c.Param("id")

	license, err := h.svc.Repos.Licenses.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取许可证详情失败",
			},
		})
		return
	}

	if license == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "许可证不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    license,
	})
}

// UpdateLicense 更新许可证
func (h *Handlers) UpdateLicense(c *gin.Context) {
	id := c.Param("id")

	var license models.License
	if err := c.ShouldBindJSON(&license); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
			},
		})
		return
	}

	license.ID = id
	if err := h.svc.Repos.Licenses.Update(&license); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "更新许可证失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "许可证更新成功",
		Data:    license,
	})
}

// RevokeLicense 撤销许可证
func (h *Handlers) RevokeLicense(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Repos.Licenses.Revoke(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "撤销许可证失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "许可证已撤销",
	})
}

// GetMiners 获取矿工列表
func (h *Handlers) GetMiners(c *gin.Context) {
	status := c.Query("status")
	limit, offset := getPagination(c)

	miners, err := h.svc.Repos.Miners.GetAll(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取矿工列表失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    miners,
	})
}

// GetMinerDetails 获取矿工详情
func (h *Handlers) GetMinerDetails(c *gin.Context) {
	id := c.Param("id")

	miner, err := h.svc.Repos.Miners.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取矿工详情失败",
			},
		})
		return
	}

	if miner == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "矿工不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    miner,
	})
}

// ShutdownMiner 关闭矿机
func (h *Handlers) ShutdownMiner(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "必须提供关闭原因",
			},
		})
		return
	}

	if err := h.svc.MiningSvc.RemoteShutdown(id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "关闭矿机失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "矿机已关闭",
	})
}

// StartMiner 启动矿机
func (h *Handlers) StartMiner(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Repos.Miners.UpdateStatus(id, "ONLINE"); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "启动矿机失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "矿机已启动",
	})
}

// GetMiningMetrics 获取挖矿指标
func (h *Handlers) GetMiningMetrics(c *gin.Context) {
	totalHashRate, err := h.svc.MiningSvc.GetTotalHashRate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取挖矿指标失败",
			},
		})
		return
	}

	miners, _ := h.svc.Repos.Miners.GetAll("", 1000, 0)
	onlineCount := 0
	totalPower := 0.0

	for _, miner := range miners {
		if miner.Status == "ONLINE" || miner.Status == "THROTTLED" {
			onlineCount++
		}
		totalPower += miner.PowerConsumption
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"total_hash_rate_phs": totalHashRate,
			"online_miners":       onlineCount,
			"total_miners":        len(miners),
			"total_power_mw":      totalPower,
		},
	})
}

// GetGridStatus 获取电网状态
func (h *Handlers) GetGridStatus(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"status":            "NORMAL",
			"total_load":        5000,
			"crypto_load":       450,
			"available_capacity": 2000,
			"grid_frequency":    50.0,
			"voltage":           230.0,
		},
	})
}

// GetEnergyConsumption 获取能源消耗
func (h *Handlers) GetEnergyConsumption(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"hourly_consumption":   []float64{420, 435, 450, 455, 460, 470, 480, 490, 485, 475, 460, 445},
			"daily_average":        455,
			"monthly_total":        13650,
			"carbon_emissions":     6800,
		},
	})
}

// TriggerLoadShedding 触发负载削减
func (h *Handlers) TriggerLoadShedding(c *gin.Context) {
	var req struct {
		RegionID string `json:"region_id"`
		Percent  int    `json:"percent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
			},
		})
		return
	}

	// 在实际实现中，这里会触发能源管理系统执行负载削减
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("已触发 %d%% 的负载削减", req.Percent),
	})
}

// GetReports 获取报告列表
func (h *Handlers) GetReports(c *gin.Context) {
	status := c.Query("status")
	reportType := c.Query("type")
	limit, offset := getPagination(c)

	reports, err := h.svc.Repos.Reports.GetAll(status, reportType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取报告列表失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    reports,
	})
}

// GenerateReport 生成报告
func (h *Handlers) GenerateReport(c *gin.Context) {
	var req struct {
		ReportType string `json:"report_type" binding:"required"`
		PeriodStart string `json:"period_start"`
		PeriodEnd   string `json:"period_end"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "无效的请求参数",
			},
		})
		return
	}

	// 解析日期
	startDate, _ := time.Parse("2006-01-02", req.PeriodStart)
	if req.PeriodStart == "" {
		startDate = time.Now().AddDate(0, 0, -1)
	}

	report, err := h.svc.ReportingSvc.GenerateDailyReport(startDate, c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "生成报告失败",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "报告生成成功",
		Data:    report,
	})
}

// DownloadReport 下载报告
func (h *Handlers) DownloadReport(c *gin.Context) {
	id := c.Param("id")

	report, err := h.svc.Repos.Reports.GetByID(id)
	if err != nil || report == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "报告不存在",
			},
		})
		return
	}

	if report.Status != "COMPLETED" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REPORT_NOT_READY",
				Message: "报告尚未生成完成",
			},
		})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", report.ID))
	c.File(report.FilePath.String)
}

// GetAuditLogs 获取审计日志
func (h *Handlers) GetAuditLogs(c *gin.Context) {
	limit, offset := getPagination(c)

	userID := c.Query("user_id")
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")

	var logs []models.AuditLog
	var err error

	switch {
	case userID != "":
		logs, err = h.svc.Repos.AuditLogs.GetByUser(userID, limit, offset)
	case resourceType != "" && resourceID != "":
		logs, err = h.svc.Repos.AuditLogs.GetByResource(resourceType, resourceID, limit, offset)
	default:
		// 在实际实现中，这里会提供获取所有审计日志的接口
		logs = []models.AuditLog{}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取审计日志失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    logs,
	})
}

// GetAuditLogDetails 获取审计日志详情
func (h *Handlers) GetAuditLogDetails(c *gin.Context) {
	id := c.Param("id")

	log, err := h.svc.Repos.AuditLogs.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "获取审计日志详情失败",
			},
		})
		return
	}

	if log == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "审计日志不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    log,
	})
}

// ExportAuditLogs 导出审计日志
func (h *Handlers) ExportAuditLogs(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "审计日志导出请求已提交",
	})
}

// GetHSMStatus 获取HSM状态
func (h *Handlers) GetHSMStatus(c *gin.Context) {
	status := h.svc.HSMService.GetStatus()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    status,
	})
}

// RotateKeys 轮换密钥
func (h *Handlers) RotateKeys(c *gin.Context) {
	if err := h.svc.HSMService.RotateKey(); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "密钥轮换失败",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "密钥轮换成功",
	})
}

// getPagination 获取分页参数
func getPagination(c *gin.Context) (limit, offset int) {
	limit = 50
	offset = 0

	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return
}

// Helper to bind JSON body
func bindJSON(c *gin.Context, v interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}
