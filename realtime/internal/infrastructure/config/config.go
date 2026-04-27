package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App AppConfig
}

// AppConfig contains application runtime settings.
type AppConfig struct {
	Port int
}

// NewConfig reads application configuration from environment variables.
func NewConfig() (*Config, error) {
	appPort, err := readIntEnv("APP_PORT")
	if err != nil {
		return nil, err
	}

	return &Config{
		App: AppConfig{
			Port: appPort,
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
