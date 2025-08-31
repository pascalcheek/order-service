package service

import (
	"context"
	"log"

	"order-service/internal/cache"
	"order-service/internal/database"
	"order-service/internal/kafka"
	"order-service/internal/model"
)

type OrderService struct {
	db       *database.Postgres
	cache    *cache.Cache
	consumer *kafka.Consumer
}

func NewOrderService(db *database.Postgres, consumer *kafka.Consumer) *OrderService {
	service := &OrderService{
		db:       db,
		cache:    cache.New(1000), // Кэш на 1000 записей
		consumer: consumer,
	}

	if err := service.loadCacheFromDB(); err != nil {
		log.Printf("Warning: failed to load cache from DB: %v", err)
	}

	return service
}

func (s *OrderService) loadCacheFromDB() error {
	orders, err := s.db.GetAllOrders()
	if err != nil {
		return err
	}

	s.cache.LoadFromOrders(orders)
	log.Printf("Loaded %d orders into cache", len(orders))
	return nil
}

func (s *OrderService) Start(ctx context.Context) {
	s.consumer.Start(ctx)
	go s.processMessages(ctx)
}

func (s *OrderService) processMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-s.consumer.Messages():
			if !ok {
				return
			}
			s.processOrder(order)
		}
	}
}

func (s *OrderService) processOrder(order model.Order) {
	if err := s.db.SaveOrder(&order); err != nil {
		log.Printf("Error saving order to database: %v", err)
		return
	}

	s.cache.Set(order)
	log.Printf("Processed and cached order: %s", order.OrderUID)
}

func (s *OrderService) GetOrder(orderUID string) (*model.Order, error) {
	if cachedOrder, exists := s.cache.Get(orderUID); exists {
		return &cachedOrder, nil
	}

	order, err := s.db.GetOrderByUID(orderUID)
	if err != nil {
		return nil, err
	}

	if order != nil {
		s.cache.Set(*order)
	}

	return order, nil
}

func (s *OrderService) GetCacheStats() (int, int) {
	return s.cache.Size(), 0
}
