package repository

import (
	"context"
	"database/sql"
	"errors"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (domain.User, error) {
	if email == "" || passwordHash == "" {
		return domain.User{}, ErrInvalidArgument
	}

	const query = `INSERT INTO users (email, password_hash)
        VALUES ($1, $2)
        RETURNING id, email, password_hash, is_active, is_verified, created_at, updated_at`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) ByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const query = `SELECT id, email, password_hash, is_active, is_verified, created_at, updated_at
        FROM users
        WHERE id = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (domain.User, error) {
	if email == "" {
		return domain.User{}, ErrInvalidArgument
	}

	const query = `SELECT id, email, password_hash, is_active, is_verified, created_at, updated_at
        FROM users
        WHERE email = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	if user.ID == uuid.Nil {
		return ErrInvalidArgument
	}

	const query = `UPDATE users
        SET email=$1, password_hash=$2, is_active=$3, is_verified=$4, updated_at=NOW()
        WHERE id=$5`

	res, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.IsActive,
		user.IsVerified,
		user.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidArgument
	}

	const query = `DELETE FROM users WHERE id=$1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
