package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"forensic-tools/internal/core/domain"
	"forensic-tools/internal/core/ports"
)

// KafkaProducer implements ports.MessagingClient using Kafka
type KafkaProducer struct {
	writers      map[string]*kafka.Writer
	brokers      []string
	logger       ports.Logger
	topicPrefix  string
}

// NewKafkaProducer creates a new KafkaProducer
func NewKafkaProducer(brokers []string, topicPrefix string, logger ports.Logger) *KafkaProducer {
	return &KafkaProducer{
		brokers:     brokers,
		topicPrefix: topicPrefix,
		logger:      logger,
		writers:     make(map[string]*kafka.Writer),
	}
}

// getTopicName returns the full topic name with prefix
func (p *KafkaProducer) getTopicName(topic string) string {
	if p.topicPrefix != "" {
		return fmt.Sprintf("%s_%s", p.topicPrefix, topic)
	}
	return topic
}

// getWriter gets or creates a kafka writer for a topic
func (p *KafkaProducer) getWriter(topic string) *kafka.Writer {
	fullTopic := p.getTopicName(topic)

	if writer, exists := p.writers[fullTopic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        fullTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	p.writers[fullTopic] = writer
	return writer
}

// PublishEvidenceCollected publishes evidence collected event
func (p *KafkaProducer) PublishEvidenceCollected(ctx context.Context, evidence *domain.Evidence) error {
	topic := "evidence.collected"

	event := map[string]interface{}{
		"event_type":   "EVIDENCE_COLLECTED",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"evidence_id":  evidence.ID,
		"evidence_name": evidence.Name,
		"evidence_type": evidence.Type,
		"source":       evidence.Source,
		"collected_by": evidence.CollectedBy,
		"collected_at": evidence.CollectedAt.Format(time.RFC3339),
		"hash":         evidence.Hash,
		"size":         evidence.Size,
		"tags":         evidence.Tags,
	}

	return p.publishMessage(ctx, topic, evidence.ID, event)
}

// PublishAnalysisStarted publishes analysis started event
func (p *KafkaProducer) PublishAnalysisStarted(ctx context.Context, analysis *domain.Analysis) error {
	topic := "analysis.started"

	event := map[string]interface{}{
		"event_type":   "ANALYSIS_STARTED",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"analysis_id":  analysis.ID,
		"evidence_id":  analysis.EvidenceID,
		"evidence_name": analysis.EvidenceName,
		"analysis_type": analysis.AnalysisType,
		"processed_by": analysis.ProcessedBy,
		"started_at":   analysis.StartedAt.Format(time.RFC3339),
	}

	return p.publishMessage(ctx, topic, analysis.ID, event)
}

// PublishAnalysisCompleted publishes analysis completed event
func (p *KafkaProducer) PublishAnalysisCompleted(ctx context.Context, analysis *domain.Analysis) error {
	topic := "analysis.completed"

	event := map[string]interface{}{
		"event_type":   "ANALYSIS_COMPLETED",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"analysis_id":  analysis.ID,
		"evidence_id":  analysis.EvidenceID,
		"analysis_type": analysis.AnalysisType,
		"status":       analysis.Status,
		"findings":     analysis.Findings,
		"completed_at": analysis.CompletedAt.Format(time.RFC3339),
		"processed_by": analysis.ProcessedBy,
	}

	return p.publishMessage(ctx, topic, analysis.ID, event)
}

// PublishCustodyTransfer publishes custody transfer event
func (p *KafkaProducer) PublishCustodyTransfer(ctx context.Context, evidenceID string, record *domain.CustodyRecord) error {
	topic := "custody.transfer"

	event := map[string]interface{}{
		"event_type":    "CUSTODY_TRANSFER",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"evidence_id":   evidenceID,
		"handler":       record.Handler,
		"action":        record.Action,
		"location":      record.Location,
		"notes":         record.Notes,
		"digital_sig":   record.DigitalSig,
		"record_hash":   record.RecordHash,
		"transfer_time": record.Timestamp.Format(time.RFC3339),
	}

	return p.publishMessage(ctx, topic, evidenceID, event)
}

// PublishSecurityEvent publishes a security event
func (p *KafkaProducer) PublishSecurityEvent(ctx context.Context, event *ports.SecurityEvent) error {
	topic := "security.events"

	eventData := map[string]interface{}{
		"event_type":   event.EventType,
		"severity":     event.Severity,
		"timestamp":    event.Timestamp.Format(time.RFC3339),
		"source":       event.Source,
		"description":  event.Description,
		"evidence_id":  event.EvidenceID,
		"analysis_id":  event.AnalysisID,
		"details":      event.Details,
	}

	return p.publishMessage(ctx, topic, event.EventType, eventData)
}

// publishMessage publishes a message to Kafka
func (p *KafkaProducer) publishMessage(ctx context.Context, topic, key string, value map[string]interface{}) error {
	writer := p.getWriter(topic)

	data, err := json.Marshal(value)
	if err != nil {
		p.logger.Error("Failed to marshal event", "error", err, "topic", topic)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Time:  time.Now().UTC(),
	}

	if err := writer.WriteMessages(ctx, message); err != nil {
		p.logger.Error("Failed to publish message", "error", err, "topic", topic, "key", key)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Debug("Message published", "topic", topic, "key", key)
	return nil
}

// Close closes all Kafka writers
func (p *KafkaProducer) Close() error {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Failed to close writer", "error", err, "topic", topic)
		}
	}
	p.writers = make(map[string]*kafka.Writer)
	return nil
}

// EnsureTopics creates the required Kafka topics
func (p *KafkaProducer) EnsureTopics(ctx context.Context, topics []string) error {
	conn, err := kafka.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		p.logger.Error("Failed to connect to Kafka", "error", err)
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		p.logger.Error("Failed to get Kafka controller", "error", err)
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		p.logger.Error("Failed to connect to Kafka controller", "error", err)
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafka.TopicConfig, len(topics))
	for i, topic := range topics {
		topicConfigs[i] = kafka.TopicConfig{
			Topic:             p.getTopicName(topic),
			NumPartitions:     3,
			ReplicationFactor: 1,
		}
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		p.logger.Warn("Failed to create topics (may already exist)", "error", err)
		// Don't return error - topics may already exist
	}

	return nil
}
