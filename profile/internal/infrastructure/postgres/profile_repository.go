package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/profile/internal/domain"
	"github.com/google/uuid"
)

type ProfileRepository struct {
	database *sql.DB
}

// NewProfileRepository constructs a PostgreSQL-backed profile repository.
func NewProfileRepository(database *sql.DB) *ProfileRepository {
	return &ProfileRepository{
		database: database,
	}
}

// FindByUserUuid returns a profile by auth user UUID or nil when the profile does not exist.
func (repository *ProfileRepository) FindByUserUuid(ctx context.Context, userUuid uuid.UUID) (*domain.Profile, error) {
	const query = `
		SELECT user_uuid, avatar, name, nickname, created_at, updated_at
		FROM profiles
		WHERE user_uuid = $1
	`

	var profile domain.Profile

	if err := repository.database.QueryRowContext(ctx, query, userUuid).Scan(
		&profile.UserUuid,
		&profile.Avatar,
		&profile.Name,
		&profile.Nickname,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding profile by user uuid %q: %w", userUuid, err)
	}

	return &profile, nil
}

// Search returns profiles matched by name or nickname.
func (repository *ProfileRepository) Search(ctx context.Context, query string, limit int) ([]domain.Profile, error) {
	const sqlQuery = `
		SELECT user_uuid, avatar, name, nickname, created_at, updated_at
		FROM profiles
		WHERE name ILIKE '%' || $1 || '%'
		   OR nickname ILIKE '%' || $1 || '%'
		ORDER BY
			CASE
				WHEN nickname ILIKE $1 || '%' THEN 0
				WHEN name ILIKE $1 || '%' THEN 1
				ELSE 2
			END,
			name ASC,
			nickname ASC
		LIMIT $2
	`

	rows, err := repository.database.QueryContext(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("searching profiles by query %q: %w", query, err)
	}
	defer rows.Close()

	profiles := make([]domain.Profile, 0)
	for rows.Next() {
		var profile domain.Profile
		if err := rows.Scan(
			&profile.UserUuid,
			&profile.Avatar,
			&profile.Name,
			&profile.Nickname,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning searched profile row for query %q: %w", query, err)
		}

		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating searched profiles for query %q: %w", query, err)
	}

	return profiles, nil
}

// Create persists a new public user profile in PostgreSQL.
func (repository *ProfileRepository) Create(ctx context.Context, profile domain.Profile) error {
	_, err := repository.database.ExecContext(ctx,
		`INSERT INTO profiles (user_uuid, avatar, name, nickname, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		profile.UserUuid, profile.Avatar, profile.Name, profile.Nickname, profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating profile for user %q: %w", profile.UserUuid, err)
	}

	return nil
}

// Update replaces public user profile data in PostgreSQL.
func (repository *ProfileRepository) Update(ctx context.Context, profile domain.Profile) error {
	_, err := repository.database.ExecContext(ctx,
		`UPDATE profiles SET avatar = $2, name = $3, nickname = $4, updated_at = $5 WHERE user_uuid = $1`,
		profile.UserUuid, profile.Avatar, profile.Name, profile.Nickname, profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating profile for user %q: %w", profile.UserUuid, err)
	}

	return nil
}
