package handlers

import (
	"context"
	"time"

	pb "github.com/csic/oversight/api/proto/v1"
	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler handles gRPC requests for the oversight service
type GRPCHandler struct {
	pb.UnimplementedOversightServiceServer
	healthScorer  ports.HealthScorerService
	abuseDetector ports.OversightService
	exchangeRepo  ports.ExchangeRepository
	logger        *zap.Logger
}

// NewGRPCHandler creates a new GRPCHandler
func NewGRPCHandler(
	healthScorer ports.HealthScorerService,
	abuseDetector ports.OversightService,
	exchangeRepo ports.ExchangeRepository,
	logger *zap.Logger,
) *GRPCHandler {
	return &GRPCHandler{
		healthScorer:  healthScorer,
		abuseDetector: abuseDetector,
		exchangeRepo:  exchangeRepo,
		logger:        logger,
	}
}

// GetHealthStatus returns the health status of an exchange
func (h *GRPCHandler) GetHealthStatus(ctx context.Context, req *pb.GetHealthStatusRequest) (*pb.GetHealthStatusResponse, error) {
	health, err := h.healthScorer.GetExchangeHealth(ctx, req.ExchangeId)
	if err != nil {
		h.logger.Error("Failed to get health status",
			zap.String("exchange_id", req.ExchangeId),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Failed to retrieve health status")
	}

	return &pb.GetHealthStatusResponse{
		ExchangeId:     health.ExchangeID,
		HealthScore:    health.HealthScore,
		Status:         translateStatus(health.Status),
		LatencyMs:      health.LatencyMs,
		ErrorRate:      health.ErrorRate,
		UptimePercent:  health.UptimePercent,
		TradeVolume24H: health.TradeVolume24h,
		TradeCount24H:  health.TradeCount24h,
		LastUpdatedAt:  health.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetAllHealthStatus returns health status for all exchanges
func (h *GRPCHandler) GetAllHealthStatus(ctx context.Context, req *pb.GetAllHealthStatusRequest) (*pb.GetAllHealthStatusResponse, error) {
	healthRecords, err := h.healthScorer.GetAllExchangeHealth(ctx)
	if err != nil {
		h.logger.Error("Failed to get all health status", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to retrieve health status")
	}

	response := &pb.GetAllHealthStatusResponse{
		HealthRecords: make([]*pb.HealthStatus, 0, len(healthRecords)),
	}

	for _, health := range healthRecords {
		response.HealthRecords = append(response.HealthRecords, &pb.HealthStatus{
			ExchangeId:     health.ExchangeID,
			HealthScore:    health.HealthScore,
			Status:         translateStatus(health.Status),
			LatencyMs:      health.LatencyMs,
			ErrorRate:      health.ErrorRate,
			UptimePercent:  health.UptimePercent,
			TradeVolume24H: health.TradeVolume24h,
			TradeCount24H:  health.TradeCount24h,
			LastUpdatedAt:  health.UpdatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

// ForceThrottle initiates immediate throttling for an exchange
func (h *GRPCHandler) ForceThrottle(ctx context.Context, req *pb.ForceThrottleRequest) (*pb.ForceThrottleResponse, error) {
	cmd := domain.NewThrottleCommand(
		req.ExchangeId,
		domain.ThrottleActionLimit,
		req.Reason,
		req.TargetRatePercent,
		int(req.DurationSeconds),
	)

	// In a real implementation, this would call the throttle publisher
	h.logger.Info("Throttle command issued via gRPC",
		zap.String("exchange_id", req.ExchangeId),
		zap.Float64("target_rate_pct", req.TargetRatePercent),
		zap.Int("duration_secs", int(req.DurationSeconds)),
	)

	return &pb.ForceThrottleResponse{
		CommandId:   cmd.ID,
		Status:      "issued",
		IssuedAt:    cmd.IssuedAt.Format(time.RFC3339),
		ExpiresAt:   cmd.ExpiresAt.Format(time.RFC3339),
	}, nil
}

// SubmitTrade submits a trade event for processing
func (h *GRPCHandler) SubmitTrade(ctx context.Context, req *pb.SubmitTradeRequest) (*pb.SubmitTradeResponse, error) {
	trade := domain.TradeEvent{
		TradeID:       req.TradeId,
		ExchangeID:    req.ExchangeId,
		TradingPair:   req.TradingPair,
		Price:         req.Price,
		Volume:        req.Volume,
		QuoteVolume:   req.QuoteVolume,
		BuyerUserID:   req.BuyerUserId,
		SellerUserID:  req.SellerUserId,
		BuyerOrderID:  req.BuyerOrderId,
		SellerOrderID: req.SellerOrderId,
		Timestamp:     parseTimestamp(req.Timestamp),
		ReceivedAt:    time.Now().UTC(),
	}

	if err := h.abuseDetector.ProcessTradeStream(ctx, trade); err != nil {
		h.logger.Error("Failed to process trade via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to process trade")
	}

	return &pb.SubmitTradeResponse{
		Status:       "processed",
		ProcessedAt:  time.Now().UTC().Format(time.RFC3339),
		TradeId:      trade.TradeID,
	}, nil
}

// ListAlerts returns a list of alerts
func (h *GRPCHandler) ListAlerts(ctx context.Context, req *pb.ListAlertsRequest) (*pb.ListAlertsResponse, error) {
	filter := ports.AlertFilter{
		ExchangeID:   req.ExchangeId,
		Limit:        int(req.Limit),
		Offset:       int(req.Offset),
	}

	// This would use the alert repository
	h.logger.Debug("ListAlerts called via gRPC",
		zap.String("exchange_id", req.ExchangeId),
		zap.Int("limit", int(req.Limit)),
	)

	return &pb.ListAlertsResponse{
		Alerts:     []*pb.Alert{},
		TotalCount: 0,
	}, nil
}

// GetAlert returns a specific alert
func (h *GRPCHandler) GetAlert(ctx context.Context, req *pb.GetAlertRequest) (*pb.GetAlertResponse, error) {
	h.logger.Debug("GetAlert called via gRPC", zap.String("alert_id", req.AlertId))

	return &pb.GetAlertResponse{
		Alert: &pb.Alert{},
	}, nil
}

// UpdateAlertStatus updates an alert status
func (h *GRPCHandler) UpdateAlertStatus(ctx context.Context, req *pb.UpdateAlertStatusRequest) (*pb.UpdateAlertStatusResponse, error) {
	h.logger.Info("UpdateAlertStatus called via gRPC",
		zap.String("alert_id", req.AlertId),
		zap.String("new_status", req.Status.String()),
	)

	return &pb.UpdateAlertStatusResponse{
		Success: true,
	}, nil
}

// CreateRule creates a new detection rule
func (h *GRPCHandler) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleResponse, error) {
	rule := domain.DetectionRule{
		Name:        req.Name,
		Description: req.Description,
		AlertType:   domain.AlertType(req.AlertType),
		Severity:    domain.AlertSeverity(req.Severity),
		TimeWindow:  int(req.TimeWindowMs),
		Threshold:   req.Threshold,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	h.logger.Info("CreateRule called via gRPC", zap.String("rule_name", rule.Name))

	return &pb.CreateRuleResponse{
		RuleId:   rule.ID,
		Status:   "created",
		CreatedAt: rule.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ListRules returns all detection rules
func (h *GRPCHandler) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	h.logger.Debug("ListRules called via gRPC")

	return &pb.ListRulesResponse{
		Rules: []*pb.DetectionRule{},
		Count: 0,
	}, nil
}

// GetRule returns a specific rule
func (h *GRPCHandler) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleResponse, error) {
	h.logger.Debug("GetRule called via gRPC", zap.String("rule_id", req.RuleId))

	return &pb.GetRuleResponse{
		Rule: &pb.DetectionRule{},
	}, nil
}

// UpdateRule updates a detection rule
func (h *GRPCHandler) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleResponse, error) {
	h.logger.Info("UpdateRule called via gRPC", zap.String("rule_id", req.RuleId))

	return &pb.UpdateRuleResponse{
		Success: true,
	}, nil
}

// DeleteRule deletes a detection rule
func (h *GRPCHandler) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	h.logger.Info("DeleteRule called via gRPC", zap.String("rule_id", req.RuleId))

	return &pb.DeleteRuleResponse{
		Success: true,
	}, nil
}

// StreamTrades handles bidirectional streaming for trade submission
func (h *GRPCHandler) StreamTrades(stream pb.OversightService_StreamTradesServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			h.logger.Error("Error receiving trade stream", zap.Error(err))
			return err
		}

		trade := domain.TradeEvent{
			TradeID:      req.TradeId,
			ExchangeID:   req.ExchangeId,
			TradingPair:  req.TradingPair,
			Price:        req.Price,
			Volume:       req.Volume,
			Timestamp:    parseTimestamp(req.Timestamp),
			ReceivedAt:   time.Now().UTC(),
		}

		if err := h.abuseDetector.ProcessTradeStream(stream.Context(), trade); err != nil {
			h.logger.Error("Failed to process streamed trade", zap.Error(err))
		}

		if err := stream.Send(&pb.SubmitTradeResponse{
			Status:      "processed",
			ProcessedAt: time.Now().UTC().Format(time.RFC3339),
			TradeId:     trade.TradeID,
		}); err != nil {
			h.logger.Error("Error sending response", zap.Error(err))
			return err
		}
	}

	return nil
}

// StreamAlerts streams alerts to the client
func (h *GRPCHandler) StreamAlerts(req *pb.StreamAlertsRequest, stream pb.OversightService_StreamAlertsServer) error {
	h.logger.Info("StreamAlerts started", zap.String("exchange_id", req.ExchangeId))

	// This would implement real-time alert streaming
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			// Send periodic health updates
			health, err := h.healthScorer.GetExchangeHealth(stream.Context(), req.ExchangeId)
			if err != nil {
				continue
			}

			if err := stream.Send(&pb.StreamAlertsResponse{
				Alert: &pb.Alert{
					AlertId:   "health_update",
					AlertType: pb.AlertType_ALERT_TYPE_HEALTH,
					Severity:  translateSeverity(health.Status),
				},
				HealthUpdate: &pb.HealthStatus{
					ExchangeId:  health.ExchangeID,
					HealthScore: health.HealthScore,
					Status:      translateStatus(health.Status),
				},
			}); err != nil {
				return err
			}
		}
	}
}

// Helper function to translate domain status to proto status
func translateStatus(status domain.ExchangeStatus) pb.ExchangeStatus {
	switch status {
	case domain.ExchangeStatusActive:
		return pb.ExchangeStatus_EXCHANGE_STATUS_ACTIVE
	case domain.ExchangeStatusDegraded:
		return pb.ExchangeStatus_EXCHANGE_STATUS_DEGRADED
	case domain.ExchangeStatusThrottled:
		return pb.ExchangeStatus_EXCHANGE_STATUS_THROTTLED
	case domain.ExchangeStatusSuspended:
		return pb.ExchangeStatus_EXCHANGE_STATUS_SUSPENDED
	case domain.ExchangeStatusOffline:
		return pb.ExchangeStatus_EXCHANGE_STATUS_OFFLINE
	default:
		return pb.ExchangeStatus_EXCHANGE_STATUS_UNSPECIFIED
	}
}

// Helper function to translate status to severity
func translateSeverity(status domain.ExchangeStatus) pb.AlertSeverity {
	switch status {
	case domain.ExchangeStatusActive:
		return pb.AlertSeverity_ALERT_SEVERITY_LOW
	case domain.ExchangeStatusDegraded:
		return pb.AlertSeverity_ALERT_SEVERITY_MEDIUM
	case domain.ExchangeStatusThrottled:
		return pb.AlertSeverity_ALERT_SEVERITY_HIGH
	case domain.ExchangeStatusSuspended:
		return pb.AlertSeverity_ALERT_SEVERITY_CRITICAL
	default:
		return pb.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED
	}
}

// Helper function to parse timestamp
func parseTimestamp(ts string) time.Time {
	t, _ := time.Parse(time.RFC3339, ts)
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}
