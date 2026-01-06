package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/IBM/sarama"
)

// KafkaProducer interface for publishing messages
type KafkaProducer interface {
	Publish(ctx context.Context, topic string, message interface{}) error
	Close() error
}

// kafkaProducer implements KafkaProducer using Sarama
type kafkaProducer struct {
	producer sarama.SyncProducer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(cfg config.KafkaConfig) (KafkaProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = cfg.Producer.Retries

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &kafkaProducer{producer: producer}, nil
}

// Publish publishes a message to a Kafka topic
func (p *kafkaProducer) Publish(ctx context.Context, topic string, message interface{}) error {
	var data []byte
	var err error

	switch m := message.(type) {
	case string:
		data = []byte(m)
	case []byte:
		data = m
	default:
		data, err = json.Marshal(m)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka producer
func (p *kafkaProducer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
