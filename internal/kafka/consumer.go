package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"order-service/internal/config"
	"order-service/internal/model"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader      *kafka.Reader
	messageChan chan model.Order
}

func NewConsumer(cfg *config.Config) *Consumer {
	config := kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		GroupID:        cfg.Kafka.GroupID,
		Topic:          cfg.Kafka.Topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	}

	log.Printf("Creating Kafka consumer for topic: %s", cfg.Kafka.Topic)
	return &Consumer{
		reader:      kafka.NewReader(config),
		messageChan: make(chan model.Order, 100),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")
	go c.consumeMessages(ctx)
}

func (c *Consumer) consumeMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			c.reader.Close()
			close(c.messageChan)
			return
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			var order model.Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if order.OrderUID == "" {
				log.Printf("Invalid order: missing order_uid")
				continue
			}

			log.Printf("Received order: %s", order.OrderUID)
			c.messageChan <- order
		}
	}
}

func (c *Consumer) Messages() <-chan model.Order {
	return c.messageChan
}

func (c *Consumer) Close() error {
	log.Println("Closing Kafka consumer...")
	return c.reader.Close()
}
