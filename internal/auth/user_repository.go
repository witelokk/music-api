package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, name, email string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, name, email string) (*User, error) {
	var (
		id        = uuid.NewString()
		createdAt time.Time
	)

	err := r.pool.
		QueryRow(
			ctx,
			`INSERT INTO users (id, name, email, created_at) 
			 VALUES ($1, $2, $3, now()) 
			 RETURNING created_at`,
			id, name, email,
		).
		Scan(&createdAt)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: createdAt,
	}, nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var (
		id        string
		name      string
		createdAt time.Time
	)

	err := r.pool.
		QueryRow(
			ctx,
			`SELECT id, name, email, created_at 
			 FROM users 
			 WHERE email = $1`,
			email,
		).
		Scan(&id, &name, &email, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: createdAt,
	}, nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var (
		name      string
		email     string
		createdAt time.Time
	)

	err := r.pool.
		QueryRow(
			ctx,
			`SELECT id, name, email, created_at 
			 FROM users 
			 WHERE id = $1`,
			id,
		).
		Scan(&id, &name, &email, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: createdAt,
	}, nil
}
