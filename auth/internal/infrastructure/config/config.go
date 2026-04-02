package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
}

type AppConfig struct {
	Port int
}

type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

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
		},
	}

	return configuration, nil
}

func readRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %q is required", key)
	}

	return value, nil
}

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
