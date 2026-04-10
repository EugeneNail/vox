package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App    AppConfig
	Images ImagesConfig
}

type AppConfig struct {
	Port int
}

type ImagesConfig struct {
	Directory string
}

func NewConfig() (*Config, error) {
	appPort, err := readIntEnv("APP_PORT")
	if err != nil {
		return nil, err
	}

	imagesDirectory, err := readRequiredEnv("IMAGES_DIR")
	if err != nil {
		return nil, err
	}

	configuration := &Config{
		App: AppConfig{
			Port: appPort,
		},
		Images: ImagesConfig{
			Directory: imagesDirectory,
		},
	}

	return configuration, nil
}

func (c ImagesConfig) CreateDirectory() error {
	if err := os.MkdirAll(c.Directory, 0o755); err != nil {
		return fmt.Errorf("create images directory %q: %w", c.Directory, err)
	}

	return nil
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
