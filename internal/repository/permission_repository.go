package repository

import (
	"context"
	"database/sql"
	"errors"

	"auth-service/internal/domain"

	"github.com/google/uuid"
)

type permissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) *permissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	query := `
        INSERT INTO permissions (name, description)
        VALUES ($1, $2)
        RETURNING id, created_at
    `
	row := r.db.QueryRowContext(ctx, query, perm.Name, perm.Description)
	err := row.Scan(&perm.ID, &perm.CreatedAt)
	return err
}

func (r *permissionRepository) GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error) {
	var perm domain.Permission
	query := `SELECT id, name, description, created_at FROM permissions WHERE name = $1`
	row := r.db.QueryRowContext(ctx, query, name)
	err := row.Scan(&perm.ID, &perm.Name, &perm.Description, &perm.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	query := `SELECT id, name, description, created_at FROM permissions ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *permissionRepository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// --- Roles ---
func (r *permissionRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	query := `
        INSERT INTO roles (name, description)
        VALUES ($1, $2)
        RETURNING id, created_at
    `
	row := r.db.QueryRowContext(ctx, query, role.Name, role.Description)
	err := row.Scan(&role.ID, &role.CreatedAt)
	return err
}

func (r *permissionRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	query := `SELECT id, name, description, created_at FROM roles WHERE name = $1`
	row := r.db.QueryRowContext(ctx, query, name)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *permissionRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var rl domain.Role
		err := rows.Scan(&rl.ID, &rl.Name, &rl.Description, &rl.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, rl)
	}
	return roles, rows.Err()
}

func (r *permissionRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// --- Role-Permission assignments ---
func (r *permissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permID uuid.UUID) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, roleID, permID)
	return err
}

func (r *permissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permID uuid.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.ExecContext(ctx, query, roleID, permID)
	return err
}

func (r *permissionRepository) GetPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	query := `
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
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

// --- User-Role assignments ---
func (r *permissionRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *permissionRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *permissionRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	query := `
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
		err := rows.Scan(&rl.ID, &rl.Name, &rl.Description, &rl.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, rl)
	}
	return roles, rows.Err()
}

// --- Permission checking ---
func (r *permissionRepository) UserHasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM user_roles ur
            JOIN role_permissions rp ON ur.role_id = rp.role_id
            JOIN permissions p ON rp.permission_id = p.id
            WHERE ur.user_id = $1 AND p.name = $2
        )
    `
	var exists bool
	row := r.db.QueryRowContext(ctx, query, userID, permissionName)
	err := row.Scan(&exists)
	return exists, err
}

func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	query := `
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
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}
