package repository

import (
	"context"
	"database/sql"
	"errors"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// CreateRole inserts a new role.
func (r *RoleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	if role == nil || role.Name == "" {
		return ErrInvalidArgument
	}
	const query = `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, role.Name, role.Description).Scan(&role.ID, &role.CreatedAt)
}

// RoleByName retrieves a role by its name.
func (r *RoleRepository) RoleByName(ctx context.Context, name string) (domain.Role, error) {
	if name == "" {
		return domain.Role{}, ErrInvalidArgument
	}
	const query = `SELECT id, name, description, created_at FROM roles WHERE name = $1`
	var role domain.Role
	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Role{}, ErrNotFound
	}
	return role, err
}

// ListRoles returns all roles ordered by name.
func (r *RoleRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	const query = `SELECT id, name, description, created_at FROM roles ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
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

// DeleteRole removes a role by ID.
func (r *RoleRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidArgument
	}
	const query = `DELETE FROM roles WHERE id = $1`
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

// AssignPermissionToRole grants a permission to a role.
func (r *RoleRepository) AssignPermissionToRole(ctx context.Context, roleID, permID uuid.UUID) error {
	if roleID == uuid.Nil || permID == uuid.Nil {
		return ErrInvalidArgument
	}
	const query = `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, roleID, permID)
	return err
}

// RemovePermissionFromRole revokes a permission from a role.
func (r *RoleRepository) RemovePermissionFromRole(ctx context.Context, roleID, permID uuid.UUID) error {
	if roleID == uuid.Nil || permID == uuid.Nil {
		return ErrInvalidArgument
	}
	const query = `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	res, err := r.db.ExecContext(ctx, query, roleID, permID)
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

// PermissionsForRole returns all permissions assigned to a role.
func (r *RoleRepository) PermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	if roleID == uuid.Nil {
		return nil, ErrInvalidArgument
	}
	const query = `
		SELECT p.id, p.name, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name
	`
	rows, err := r.db.QueryContext(ctx, query, roleID)
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
