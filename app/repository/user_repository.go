package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/app/model"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
}

type userRepo struct{ pool *pgxpool.Pool }

func NewUserRepository(p *pgxpool.Pool) UserRepository { return &userRepo{p} }

func (r *userRepo) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx, `SELECT id, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE email = $1 AND is_active = TRUE`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

func (r *userRepo) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx, `SELECT id, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
