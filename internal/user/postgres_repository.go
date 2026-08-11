package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

// contrsuctor method
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, email string, passwordHash string) (User, error) {

	var user User
	var pgErr *pgconn.PgError

	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email,password_hash) 
		VALUES ($1,$2)
		RETURNING id, email, password_hash, created_at
	`, email, passwordHash).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrEmailAlreadyExists
	}

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, `

	SELECT id, email, password_hash, created_at 
	FROM users 
	WHERE email = $1
	
	`,
		email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}

	if err != nil {
		return User{}, err
	}
	return user, nil
}
