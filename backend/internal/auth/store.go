package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"quotetrack/backend/internal/models"
)

// Store is a pgx-backed implementation of the auth Store interface.
type userStore struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *userStore {
	return &userStore{pool: pool}
}

func (s *userStore) CreateUser(ctx context.Context, email, passwordHash, businessName string) (*models.User, error) {
	u := models.User{
		Email:        email,
		BusinessName: businessName,
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, business_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, business_name, created_at, updated_at`,
		email, passwordHash, businessName,
	).Scan(&u.ID, &u.Email, &u.BusinessName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (s *userStore) UserByEmail(ctx context.Context, email string) (*models.User, string, error) {
	var u models.User
	var hash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, business_name, password_hash, created_at, updated_at
		FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.BusinessName, &hash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("find user: %w", err)
	}
	return &u, hash, nil
}

func (s *userStore) UserByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, business_name, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.BusinessName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &u, nil
}

func (s *userStore) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	err := s.pool.QueryRow(ctx, `
		UPDATE users SET email = $2, business_name = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, email, business_name, created_at, updated_at`,
		user.ID, user.Email, user.BusinessName,
	).Scan(&user.ID, &user.Email, &user.BusinessName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}