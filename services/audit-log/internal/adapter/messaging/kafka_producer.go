package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"audit-log/internal/core/domain"
	"audit-log/internal/core/ports"
)

// KafkaAuditProducer implements ports.KafkaProducer for Kafka
type KafkaAuditProducer struct {
	writers     map[string]*kafka.Writer
	brokers     []string
	logger      ports.Logger
	topicPrefix string
}

// NewKafkaAuditProducer creates a new KafkaAuditProducer
func NewKafkaAuditProducer(brokers []string, topicPrefix string, logger ports.Logger) *KafkaAuditProducer {
	return &KafkaAuditProducer{
		brokers:     brokers,
		topicPrefix: topicPrefix,
		logger:      logger,
		writers:     make(map[string]*kafka.Writer),
	}
}

// getTopicName returns the full topic name with prefix
func (p *KafkaAuditProducer) getTopicName(topic string) string {
	if p.topicPrefix != "" {
		return fmt.Sprintf("%s_%s", p.topicPrefix, topic)
	}
	return topic
}

// getWriter gets or creates a Kafka writer for a topic
func (p *KafkaAuditProducer) getWriter(topic string) *kafka.Writer {
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

// PublishEntry publishes an audit entry to Kafka
func (p *KafkaAuditProducer) PublishEntry(ctx context.Context, entry *domain.AuditEntry) error {
	topic := "audit.entries"

	event := map[string]interface{}{
		"event_type":   "AUDIT_ENTRY_CREATED",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"entry_id":     entry.ID,
		"trace_id":     entry.TraceID,
		"actor_id":     entry.ActorID,
		"actor_type":   entry.ActorType,
		"action":       entry.Action,
		"resource":     entry.Resource,
		"resource_id":  entry.ResourceID,
		"operation":    entry.Operation,
		"outcome":      entry.Outcome,
		"severity":     entry.Severity,
		"source_ip":    entry.SourceIP,
		"user_agent":   entry.UserAgent,
		"entry_time":   entry.Timestamp.Format(time.RFC3339),
		"hash":         entry.CurrentHash,
	}

	return p.publishMessage(ctx, topic, entry.ID, event)
}

// PublishVerificationEvent publishes a verification event to Kafka
func (p *KafkaAuditProducer) PublishVerificationEvent(ctx context.Context, entryID string, result *domain.VerificationResult) error {
	topic := "audit.verifications"

	event := map[string]interface{}{
		"event_type":    "AUDIT_VERIFICATION",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"entry_id":      entryID,
		"valid":         result.Valid,
		"block_number":  result.BlockNumber,
		"verified_at":   result.Timestamp.Format(time.RFC3339),
		"message":       result.Message,
	}

	return p.publishMessage(ctx, topic, entryID, event)
}

// publishMessage publishes a message to Kafka
func (p *KafkaAuditProducer) publishMessage(ctx context.Context, topic, key string, value map[string]interface{}) error {
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
func (p *KafkaAuditProducer) Close() error {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Failed to close writer", "error", err, "topic", topic)
		}
	}
	p.writers = make(map[string]*kafka.Writer)
	return nil
}

// KafkaAuditConsumer implements ports.KafkaConsumer for Kafka
type KafkaAuditConsumer struct {
	reader  *kafka.Reader
	handler func(ctx context.Context, event domain.AuditEvent) error
	logger  ports.Logger
}

// NewKafkaAuditConsumer creates a new KafkaAuditConsumer
func NewKafkaAuditConsumer(brokers []string, topic string, consumerGroup string, logger ports.Logger) *KafkaAuditConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        consumerGroup,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	return &KafkaAuditConsumer{
		reader: reader,
		logger: logger,
	}
}

// SetHandler sets the event handler for processing messages
func (c *KafkaAuditConsumer) SetHandler(handler func(ctx context.Context, event domain.AuditEvent) error) {
	c.handler = handler
}

// Consume starts consuming messages from Kafka
func (c *KafkaAuditConsumer) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Kafka consumer context cancelled")
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c.logger.Error("Failed to fetch message", "error", err)
				continue
			}

			// Parse the message
			var event domain.AuditEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.logger.Error("Failed to unmarshal message", "error", err)
				if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
					c.logger.Error("Failed to commit message", "error", commitErr)
				}
				continue
			}

			// Process the message
			if c.handler != nil {
				if err := c.handler(ctx, event); err != nil {
					c.logger.Error("Failed to process message", "error", err, "key", string(msg.Key))
					// Don't commit on error - message will be reprocessed
					continue
				}
			}

			// Commit the message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("Failed to commit message", "error", err)
			}
		}
	}
}

// Close closes the Kafka consumer
func (c *KafkaAuditConsumer) Close() error {
	return c.reader.Close()
}

// EnsureTopics creates the required Kafka topics
func EnsureTopics(ctx context.Context, brokers []string, topics []string, topicPrefix string) error {
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get Kafka controller: %w", err)
	}

	controllerConn, err := kafka.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafka.TopicConfig, len(topics))
	for i, topic := range topics {
		fullTopic := topic
		if topicPrefix != "" {
			fullTopic = fmt.Sprintf("%s_%s", topicPrefix, topic)
		}
		topicConfigs[i] = kafka.TopicConfig{
			Topic:             fullTopic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		}
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		// Topics may already exist
		return nil
	}

	return nil
}
