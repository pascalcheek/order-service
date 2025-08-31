package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"order-service/internal/kafka"
)

func main() {
	var n int
	flag.IntVar(&n, "n", 3, "number of messages to send")
	flag.Parse()

	broker := "kafka:9092"
	if os.Getenv("KAFKA_BROKER") != "" {
		broker = os.Getenv("KAFKA_BROKER")
	} else {
		if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
			broker = "localhost:9093"
		}
	}

	fmt.Printf("Connecting to Kafka broker: %s\n", broker)
	kafka.InitProducer(broker)
	defer kafka.Close()

	fmt.Printf("Sending %d messages...\n", n)
	err := kafka.SendTestData(context.Background(), n)
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println("Done! Check http://localhost:8080")
}
