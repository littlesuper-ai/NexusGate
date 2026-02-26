package config

import (
	"fmt"
	"os"
)

type Config struct {
	ListenAddr string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	MQTTBroker string
	JWTSecret  string
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr: getEnv("LISTEN_ADDR", ":8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "nexusgate"),
		DBPassword: getEnv("DB_PASSWORD", "nexusgate"),
		DBName:     getEnv("DB_NAME", "nexusgate"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		MQTTBroker: getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
