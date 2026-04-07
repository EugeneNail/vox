package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
)

func main() {
	configuration, err := newMigratorConfig()
	if err != nil {
		log.Fatal(err)
	}

	if configuration.Command == "up" {
		if err := ensureDatabaseExists(configuration); err != nil {
			log.Fatal(err)
		}
	}

	migrationRunner, err := migrate.New("file:///migrations", configuration.databaseURL())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		sourceErr, databaseErr := migrationRunner.Close()
		if sourceErr != nil {
			log.Printf("closing migration source: %v", sourceErr)
		}
		if databaseErr != nil {
			log.Printf("closing migration database: %v", databaseErr)
		}
	}()

	switch configuration.Command {
	case "up":
		if err := migrationRunner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	case "down":
		if err := migrationRunner.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unsupported migrator command %q", configuration.Command)
	}
}

type migratorConfig struct {
	Command  string
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

func newMigratorConfig() (*migratorConfig, error) {
	command := readEnvOrDefault("MIGRATOR_COMMAND", "up")

	port, err := readIntEnv("POSTGRES_PORT")
	if err != nil {
		return nil, err
	}

	host, err := readRequiredEnv("POSTGRES_HOST")
	if err != nil {
		return nil, err
	}

	database, err := readRequiredEnv("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	user, err := readRequiredEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}

	password, err := readRequiredEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	return &migratorConfig{
		Command:  command,
		Host:     host,
		Port:     port,
		Database: database,
		User:     user,
		Password: password,
		SSLMode:  readEnvOrDefault("POSTGRES_SSLMODE", "disable"),
	}, nil
}

func (configuration *migratorConfig) databaseURL() string {
	return configuration.databaseURLForDatabase(configuration.Database)
}

func (configuration *migratorConfig) maintenanceDatabaseURL() string {
	return configuration.databaseURLForDatabase("postgres")
}

func (configuration *migratorConfig) databaseURLForDatabase(database string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(configuration.User),
		url.QueryEscape(configuration.Password),
		configuration.Host,
		configuration.Port,
		database,
		url.QueryEscape(configuration.SSLMode),
	)
}

func ensureDatabaseExists(configuration *migratorConfig) error {
	connection, err := sql.Open("postgres", configuration.maintenanceDatabaseURL())
	if err != nil {
		return fmt.Errorf("opening postgres maintenance database: %w", err)
	}
	defer func() {
		if closeErr := connection.Close(); closeErr != nil {
			log.Printf("closing postgres maintenance database connection: %v", closeErr)
		}
	}()

	if err := connection.Ping(); err != nil {
		return fmt.Errorf("pinging postgres maintenance database: %w", err)
	}

	var exists bool
	if err := connection.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		configuration.Database,
	).Scan(&exists); err != nil {
		return fmt.Errorf("checking whether database %q exists: %w", configuration.Database, err)
	}

	if exists {
		return nil
	}

	if _, err := connection.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(configuration.Database))); err != nil {
		return fmt.Errorf("creating database %q: %w", configuration.Database, err)
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

func readEnvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
