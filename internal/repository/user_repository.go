package repository

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	if email == "" || passwordHash == "" {
		return nil, errors.New("email and password_hash cannot be empty")
	}

	query := `INSERT INTO users (email, password_hash)
	          VALUES ($1, $2)
	          RETURNING id, email, password_hash, is_active, is_verified, created_at, updated_at`

	var user domain.User
	err := r.DB.QueryRowContext(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, email, password_hash, is_active, is_verified, created_at, updated_at
	          FROM users
	          WHERE id = $1`

	var user domain.User
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, is_active, is_verified, created_at, updated_at
	          FROM users
	          WHERE email = $1`

	var user domain.User
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	query := `UPDATE users
	          SET email=$1, password_hash=$2, is_active=$3, is_verified=$4
	          WHERE id=$5`
	res, err := r.DB.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.IsActive,
		user.IsVerified,
		user.ID,
	)
	if ra, err := res.RowsAffected(); err != nil && ra == 0 {
		return errors.New("no user updated")
	}
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id=$1`
	res, err := r.DB.ExecContext(ctx, query, id)
	if ra, err := res.RowsAffected(); err != nil && ra == 0 {
		return errors.New("no user deleted")
	}
	return err
}
