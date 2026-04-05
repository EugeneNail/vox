package domain

import "context"

// UserRepository describes user persistence required by the domain and application.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user User) error
}
