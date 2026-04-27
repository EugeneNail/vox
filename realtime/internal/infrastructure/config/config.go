package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App   AppConfig
	Redis RedisConfig
}

// AppConfig contains application runtime settings.
type AppConfig struct {
	Port int
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	Host string
	Port int
}

// NewConfig reads application configuration from environment variables.
func NewConfig() (*Config, error) {
	appPort, err := readIntEnv("APP_PORT")
	if err != nil {
		return nil, err
	}

	redisHost, err := readRequiredEnv("REDIS_HOST")
	if err != nil {
		return nil, err
	}

	redisPort, err := readIntEnv("REDIS_PORT")
	if err != nil {
		return nil, err
	}

	return &Config{
		App: AppConfig{
			Port: appPort,
		},
		Redis: RedisConfig{
			Host: redisHost,
			Port: redisPort,
		},
	}, nil
}

// readRequiredEnv reads a required string environment variable.
func readRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %q is required", key)
	}

	return value, nil
}

// readIntEnv reads and parses a required integer environment variable.
func readIntEnv(key string) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("environment variable %q is required", key)
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parsing environment variable %q as int: %w", key, err)
	}

	return parsedValue, nil
}
