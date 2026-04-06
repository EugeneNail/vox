package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/EugeneNail/vox/message/internal/infrastructure/config"
	_ "github.com/lib/pq"
)

// NewDatabase opens and verifies a PostgreSQL connection.
func NewDatabase(configuration config.PostgresConfig) (*sql.DB, error) {
	database, err := sql.Open("postgres", buildConnectionString(configuration))
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}

	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("pinging postgres connection: %w", err)
	}

	return database, nil
}

// buildConnectionString builds a PostgreSQL DSN from application config.
func buildConnectionString(configuration config.PostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(configuration.User),
		url.QueryEscape(configuration.Password),
		configuration.Host,
		configuration.Port,
		configuration.Database,
		url.QueryEscape(configuration.SSLMode),
	)
}
