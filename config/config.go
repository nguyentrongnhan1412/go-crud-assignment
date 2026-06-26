package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
}

func Load() *Config {
	databaseURL := firstEnv("DATABASE_URL", "POSTGRES_URL")

	serverPort := getEnv("PORT", "")
	if serverPort == "" {
		serverPort = getEnv("SERVER_PORT", "8080")
	}

	return &Config{
		ServerPort:  serverPort,
		DatabaseURL: databaseURL,
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "product_management"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
	}
}

func (c *Config) DatabaseDSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) DatabaseTarget() string {
	if c.DatabaseURL != "" {
		return "DATABASE_URL (or POSTGRES_URL)"
	}
	return fmt.Sprintf("%s:%s/%s", c.DBHost, c.DBPort, c.DBName)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
