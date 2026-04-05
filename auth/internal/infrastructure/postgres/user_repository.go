package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/auth/internal/domain"
)

type UserRepository struct {
	database *sql.DB
}

// NewUserRepository constructs a PostgreSQL-backed user repository.
func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{
		database: database,
	}
}

// FindByEmail returns a user by email or nil when the user does not exist.
func (repository *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT uuid, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User

	if err := repository.database.QueryRowContext(ctx, query, email).Scan(
		&user.Uuid,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding user by email %q: %w", email, err)
	}

	return &user, nil
}

// Create persists a new user in PostgreSQL.
func (repository *UserRepository) Create(ctx context.Context, user domain.User) error {
	_, err := repository.database.ExecContext(ctx,
		`INSERT INTO users (uuid, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		user.Uuid, user.Email, user.Password, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating user with email %q: %w", user.Email, err)
	}

	return nil
}
