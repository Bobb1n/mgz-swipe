package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	KafkaBrokers string

	GeoRadiusKm float64
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	redisDB := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			redisDB = n
		}
	}

	radiusKm := 50.0
	if v := os.Getenv("GEO_RADIUS_KM"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			radiusKm = f
		}
	}

	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8084"),
		DatabaseURL:   dbURL,
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "kafka:9092"),
		GeoRadiusKm:   radiusKm,
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
