package config

import "os"

type Config struct {
	Server struct {
		Host string
		Port string
	}
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Kafka struct {
		Brokers []string
		Topic   string
		GroupID string
	}
}

func Load() *Config {
	var cfg Config

	cfg.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Database.Host = getEnv("DB_HOST", "postgres")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "password")
	cfg.Database.Name = getEnv("DB_NAME", "order_service")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Kafka.Brokers = []string{getEnv("KAFKA_BROKER", "kafka:9092")}
	cfg.Kafka.Topic = getEnv("KAFKA_TOPIC", "orders")
	cfg.Kafka.GroupID = getEnv("KAFKA_GROUP_ID", "order-service-group")

	return &cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
