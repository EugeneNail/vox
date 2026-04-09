package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Redis    RedisConfig
}

// AppConfig contains application runtime settings.
type AppConfig struct {
	Port int
}

// PostgresConfig contains PostgreSQL connection settings.
type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
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

	postgresPort, err := readIntEnv("POSTGRES_PORT")
	if err != nil {
		return nil, err
	}

	postgresHost, err := readRequiredEnv("POSTGRES_HOST")
	if err != nil {
		return nil, err
	}

	postgresDatabase, err := readRequiredEnv("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	postgresUser, err := readRequiredEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}

	postgresPassword, err := readRequiredEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	postgresSSLMode, err := readRequiredEnv("POSTGRES_SSLMODE")
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

	configuration := &Config{
		App: AppConfig{
			Port: appPort,
		},
		Postgres: PostgresConfig{
			Host:     postgresHost,
			Port:     postgresPort,
			Database: postgresDatabase,
			User:     postgresUser,
			Password: postgresPassword,
			SSLMode:  postgresSSLMode,
		},
		Redis: RedisConfig{
			Host: redisHost,
			Port: redisPort,
		},
	}

	return configuration, nil
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
