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

// Create inserts a new user from registration data.
func (r *UserRepository) Create(ctx context.Context, userDto domain.RegisterDto) (domain.User, error) {
	if userDto.Email == "" || userDto.Password == "" {
		return domain.User{}, ErrInvalidArgument
	}

	const query = `INSERT INTO users (email, password_hash)
        VALUES ($1, $2)
        RETURNING id, email, password_hash, is_active, is_verified, created_at, updated_at`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, userDto.Email, userDto.Password).Scan(
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

// ByID retrieves a user by their UUID.
func (r *UserRepository) ByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const query = `SELECT id, email, password_hash, is_active, is_verified, created_at, updated_at
        FROM users WHERE id = $1`

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
	return user, err
}

// ByEmail retrieves a user by their email address.
func (r *UserRepository) ByEmail(ctx context.Context, email string) (domain.User, error) {
	if email == "" {
		return domain.User{}, ErrInvalidArgument
	}

	const query = `SELECT id, email, password_hash, is_active, is_verified, created_at, updated_at
        FROM users WHERE email = $1`

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
	return user, err
}

// Update modifies an existing user record.
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

// Delete removes a user by ID (cascades to sessions, user_roles, audit_logs via DB constraints).
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

// AssignRole grants a role to a user.
func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if userID == uuid.Nil || roleID == uuid.Nil {
		return ErrInvalidArgument
	}
	const query = `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

// RemoveRole revokes a role from a user.
func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if userID == uuid.Nil || roleID == uuid.Nil {
		return ErrInvalidArgument
	}
	const query = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	res, err := r.db.ExecContext(ctx, query, userID, roleID)
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

// RolesForUser returns all roles assigned to a user.
func (r *UserRepository) RolesForUser(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidArgument
	}
	const query = `
		SELECT r.id, r.name, r.description, r.created_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var rl domain.Role
		if err := rows.Scan(&rl.ID, &rl.Name, &rl.Description, &rl.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, rl)
	}
	return roles, rows.Err()
}

// HasPermission checks whether a user has a specific permission (directly or via roles).
func (r *UserRepository) HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	if userID == uuid.Nil || permissionName == "" {
		return false, ErrInvalidArgument
	}
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1 AND p.name = $2
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, permissionName).Scan(&exists)
	return exists, err
}

// PermissionsForUser returns all distinct permissions a user has through their roles.
func (r *UserRepository) PermissionsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidArgument
	}
	const query = `
		SELECT DISTINCT p.id, p.name, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.name
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}
