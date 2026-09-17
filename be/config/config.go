package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	AppPort  string
	LogLevel string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSchema   string
	DBSSLMode  string

	RedisHost     string
	RedisPassword string
}

func LoadConfig() (*Config, error) {
	// Try loading .env file (ignore error if not present in production/environment)
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AppPort:       getEnv("APP_PORT", "8080"),
		LogLevel:      getEnv("LOG_LEVEL", "debug"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "postgres"),
		DBName:        getEnv("DB_NAME", "rebutin_db"),
		DBSchema:      getEnv("DB_SCHEMA", "public"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		RedisHost:     getEnv("REDIS_HOST", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}

	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode, c.DBSchema,
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strVal := getEnv(key, "")
	if val, err := strconv.Atoi(strVal); err == nil {
		return val
	}
	return fallback
}
