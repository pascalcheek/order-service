package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"order-service/internal/model"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitProducer(broker string) {
	writer = &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
}

func SendTestData(ctx context.Context, n int) error {
	if writer == nil {
		return fmt.Errorf("kafka writer not initialized")
	}

	for i := 0; i < n; i++ {
		order := generateTestOrder(i)
		if err := sendMessage(ctx, order); err != nil {
			return err
		}
	}
	return nil
}

func generateTestOrder(index int) model.Order {
	now := time.Now()
	orderUID := fmt.Sprintf("test_order_%d_%d", now.Unix(), index)

	return model.Order{
		OrderUID:    orderUID,
		TrackNumber: fmt.Sprintf("WBILTESTTRACK%d", index),
		Entry:       "WBIL",
		Delivery: model.Delivery{
			Name:    fmt.Sprintf("Test User %d", index),
			Phone:   "+1234567890",
			Zip:     fmt.Sprintf("1000%d", index),
			City:    fmt.Sprintf("City %d", index),
			Address: fmt.Sprintf("Test Address %d", index),
			Region:  fmt.Sprintf("Region %d", index),
			Email:   fmt.Sprintf("test%d@example.com", index),
		},
		Payment: model.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "testpay",
			Amount:       rand.Intn(10000) + 1000,
			PaymentDt:    now.Unix(),
			Bank:         "testbank",
			DeliveryCost: 500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      rand.Intn(10000000),
				TrackNumber: fmt.Sprintf("WBILTESTTRACK%d", index),
				Price:       rand.Intn(1000) + 100,
				Rid:         fmt.Sprintf("test_rid_%d", index),
				Name:        fmt.Sprintf("Test Product %d", index),
				Sale:        rand.Intn(50),
				Size:        "M",
				TotalPrice:  rand.Intn(900) + 100,
				NmID:        rand.Intn(1000000),
				Brand:       fmt.Sprintf("Test Brand %d", index),
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("test_customer_%d", index),
		DeliveryService:   "test_delivery",
		Shardkey:          fmt.Sprintf("%d", index%10),
		SmID:              rand.Intn(100),
		DateCreated:       now,
		OofShard:          "1",
	}
}

func sendMessage(ctx context.Context, order model.Order) error {
	jsonData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("error marshaling order: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(order.OrderUID),
		Value: jsonData,
	}

	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	log.Printf("Sent order: %s", order.OrderUID)
	return nil
}

func Close() error {
	if writer != nil {
		return writer.Close()
	}
	return nil
}
