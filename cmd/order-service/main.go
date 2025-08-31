package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/internal/config"
	"order-service/internal/database"
	"order-service/internal/handler"
	"order-service/internal/kafka"
	"order-service/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	consumer := kafka.NewConsumer(cfg)
	defer consumer.Close()

	orderService := service.NewOrderService(db, consumer)

	apiHandler := handler.NewAPIHandler(orderService)
	webHandler, err := handler.NewWebHandler("./templates")
	if err != nil {
		log.Fatalf("Failed to initialize web handler: %v", err)
	}

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/order/{order_uid}", apiHandler.GetOrder).Methods("GET")
	api.HandleFunc("/health", apiHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/", webHandler.ServeIndex).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orderService.Start(ctx)

	go func() {
		log.Printf("Server starting on http://%s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	go func() {
		time.Sleep(15 * time.Second)
		fmt.Println("Sending initial test data...")
		kafka.InitProducer("kafka:9092")
		defer kafka.Close()

		if err := kafka.SendTestData(context.Background(), 2); err != nil {
			fmt.Printf("Failed to send test data: %v\n", err)
		} else {
			fmt.Println("Test data sent successfully!")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server exited properly")
}
