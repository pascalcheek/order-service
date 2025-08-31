package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"order-service/internal/model"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(host, port, user, password, dbname, sslmode string) (*Postgres, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to PostgreSQL database")
	return &Postgres{db: db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) SaveOrder(order *model.Order) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to delete existing items: %v", err)
	}

	_, err = tx.Exec("DELETE FROM payments WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to delete existing payments: %v", err)
	}

	_, err = tx.Exec("DELETE FROM deliveries WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to delete existing deliveries: %v", err)
	}

	_, err = tx.Exec("DELETE FROM orders WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to delete existing order: %v", err)
	}

	orderQuery := `INSERT INTO orders (
		order_uid, track_number, entry, locale, internal_signature, 
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(orderQuery,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("failed to insert order: %v", err)
	}

	deliveryQuery := `INSERT INTO deliveries (
		order_uid, name, phone, zip, city, address, region, email
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(deliveryQuery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %v", err)
	}

	paymentQuery := `INSERT INTO payments (
		order_uid, transaction, request_id, currency, provider, amount,
		payment_dt, bank, delivery_cost, goods_total, custom_fee
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(paymentQuery,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %v", err)
	}

	itemQuery := `INSERT INTO items (
		order_uid, chrt_id, track_number, price, rid, name, sale, size,
		total_price, nm_id, brand, status
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	for _, item := range order.Items {
		_, err = tx.Exec(itemQuery,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("failed to insert item: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (p *Postgres) GetOrderByUID(orderUID string) (*model.Order, error) {
	var order model.Order

	orderQuery := `SELECT order_uid, track_number, entry, locale, internal_signature, 
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard 
		FROM orders WHERE order_uid = $1`

	err := p.db.QueryRow(orderQuery, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %v", err)
	}

	var delivery model.Delivery
	deliveryQuery := `SELECT name, phone, zip, city, address, region, email 
		FROM deliveries WHERE order_uid = $1`

	err = p.db.QueryRow(deliveryQuery, orderUID).Scan(
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
		&delivery.Address, &delivery.Region, &delivery.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get delivery: %v", err)
	}
	order.Delivery = delivery

	var payment model.Payment
	paymentQuery := `SELECT transaction, request_id, currency, provider, amount,
		payment_dt, bank, delivery_cost, goods_total, custom_fee 
		FROM payments WHERE order_uid = $1`

	err = p.db.QueryRow(paymentQuery, orderUID).Scan(
		&payment.Transaction, &payment.RequestID, &payment.Currency,
		&payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank,
		&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get payment: %v", err)
	}
	order.Payment = payment

	itemsQuery := `SELECT chrt_id, track_number, price, rid, name, sale, size,
		total_price, nm_id, brand, status FROM items WHERE order_uid = $1`

	rows, err := p.db.Query(itemsQuery, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %v", err)
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var item model.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %v", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %v", err)
	}

	order.Items = items

	return &order, nil
}

func (p *Postgres) GetAllOrders() ([]model.Order, error) {
	rows, err := p.db.Query("SELECT order_uid FROM orders")
	if err != nil {
		return nil, fmt.Errorf("failed to get order UIDs: %v", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("failed to scan order UID: %v", err)
		}

		order, err := p.GetOrderByUID(orderUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get order %s: %v", orderUID, err)
		}
		if order != nil {
			orders = append(orders, *order)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %v", err)
	}

	return orders, nil
}

func (p *Postgres) Exec(query string, args ...interface{}) error {
	_, err := p.db.Exec(query, args...)
	return err
}
