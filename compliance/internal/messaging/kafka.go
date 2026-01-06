package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/platform/compliance/internal/domain"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers       []string
	ConsumerGroup string
	Topics        KafkaTopicsConfig
}

// KafkaTopicsConfig holds Kafka topic names
type KafkaTopicsConfig struct {
	Transactions string
	Violations   string
	Audit        string
}

// KafkaConsumer handles Kafka message consumption
type KafkaConsumer struct {
	reader       *kafka.Reader
	logger       *zap.Logger
	topics       KafkaTopicsConfig
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(cfg KafkaConfig, logger *zap.Logger) (*KafkaConsumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topics.Transactions,
		GroupID:        cfg.ConsumerGroup,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 1 * time.Second,
	})

	logger.Info("Kafka consumer initialized",
		zap.String("brokers", fmt.Sprintf("%v", cfg.Brokers)),
		zap.String("topic", cfg.Topics.Transactions),
		zap.String("consumer_group", cfg.ConsumerGroup))

	return &KafkaConsumer{
		reader: reader,
		logger: logger,
		topics: cfg.Topics,
	}, nil
}

// Consume starts consuming messages
func (c *KafkaConsumer) Consume(ctx context.Context, handler func(context.Context, *domain.Transaction) error) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Kafka consumer stopping due to context cancellation")
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c.logger.Error("Error fetching message", zap.Error(err))
				continue
			}

			// Parse transaction from message
			var tx domain.Transaction
			if err := json.Unmarshal(msg.Value, &tx); err != nil {
				c.logger.Warn("Invalid message format",
					zap.Error(err),
					zap.ByteString("value", msg.Value))
				// Commit message to avoid reprocessing invalid messages
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					c.logger.Error("Error committing message", zap.Error(err))
				}
				continue
			}

			// Process message
			if err := handler(ctx, &tx); err != nil {
				c.logger.Error("Error processing message",
					zap.Error(err),
					zap.String("transaction_id", tx.ID))
				// Don't commit on error - message will be reprocessed
				continue
			}

			// Commit message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("Error committing message", zap.Error(err))
			}

			c.logger.Debug("Message processed",
				zap.String("transaction_id", tx.ID),
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset))
		}
	}
}

// Close closes the consumer
func (c *KafkaConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// KafkaProducerConfig holds Kafka producer configuration
type KafkaProducerConfig struct {
	Brokers      []string
	RequiredAcks string
	RetryMax     int
}

// KafkaProducer handles Kafka message production
type KafkaProducer struct {
	writers      map[string]*kafka.Writer
	logger       *zap.Logger
	requiredAcks kafka.RequiredAcks
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(cfg KafkaProducerConfig, logger *zap.Logger) (*KafkaProducer, error) {
	// Parse required acks
	var requiredAcks kafka.RequiredAcks
	switch cfg.RequiredAcks {
	case "all", "-1":
		requiredAcks = kafka.RequireAll
	case "1":
		requiredAcks = kafka.RequireOne
	case "0":
		requiredAcks = kafka.RequireNone
	default:
		requiredAcks = kafka.RequireAll
	}

	producer := &KafkaProducer{
		writers:      make(map[string]*kafka.Writer),
		logger:       logger,
		requiredAcks: requiredAcks,
	}

	logger.Info("Kafka producer initialized",
		zap.String("brokers", fmt.Sprintf("%v", cfg.Brokers)),
		zap.String("required_acks", cfg.RequiredAcks))

	return producer, nil
}

// Send sends a message to a Kafka topic
func (p *KafkaProducer) Send(ctx context.Context, topic, message string) error {
	// Get or create writer for topic
	writer, ok := p.writers[topic]
	if !ok {
		// Get brokers from first writer or use default
		brokers := []string{"localhost:9092"}
		if len(p.writers) > 0 {
			for _, w := range p.writers {
				brokers = w.Addr.TCP
				break
			}
		}

		writer = &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    1,
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: p.requiredAcks,
			Async:        false,
		}
		p.writers[topic] = writer
	}

	// Create message
	msg := kafka.Message{
		Key:   []byte(time.Now().Format(time.RFC3339)),
		Value: []byte(message),
		Time:  time.Now(),
	}

	// Send message
	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.logger.Debug("Message sent to Kafka",
		zap.String("topic", topic),
		zap.String("message_preview", truncate(message, 100)))

	return nil
}

// SendToTopic sends a message to a specific topic with a key
func (p *KafkaProducer) SendToTopic(ctx context.Context, topic, key, message string) error {
	writer, ok := p.writers[topic]
	if !ok {
		brokers := []string{"localhost:9092"}
		if len(p.writers) > 0 {
			for _, w := range p.writers {
				brokers = w.Addr.TCP
				break
			}
		}

		writer = &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    1,
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: p.requiredAcks,
			Async:        false,
		}
		p.writers[topic] = writer
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(message),
		Time:  time.Now(),
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// Close closes all writers
func (p *KafkaProducer) Close() error {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Error closing writer",
				zap.Error(err),
				zap.String("topic", topic))
		}
	}
	return nil
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
