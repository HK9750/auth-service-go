package repository

import (
	"context"
	"database/sql"
	"errors"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type PermissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// CreatePermission inserts a new permission.
func (r *PermissionRepository) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	if perm == nil || perm.Name == "" {
		return ErrInvalidArgument
	}
	const query = `INSERT INTO permissions (name, description) VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, perm.Name, perm.Description).Scan(&perm.ID, &perm.CreatedAt)
}

// PermissionByName retrieves a permission by its name.
func (r *PermissionRepository) PermissionByName(ctx context.Context, name string) (domain.Permission, error) {
	if name == "" {
		return domain.Permission{}, ErrInvalidArgument
	}
	const query = `SELECT id, name, description, created_at FROM permissions WHERE name = $1`
	var perm domain.Permission
	err := r.db.QueryRowContext(ctx, query, name).Scan(&perm.ID, &perm.Name, &perm.Description, &perm.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Permission{}, ErrNotFound
	}
	return perm, err
}

// ListPermissions returns all permissions ordered by name.
func (r *PermissionRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	const query = `SELECT id, name, description, created_at FROM permissions ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
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

// DeletePermission removes a permission by ID.
func (r *PermissionRepository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidArgument
	}
	const query = `DELETE FROM permissions WHERE id = $1`
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
