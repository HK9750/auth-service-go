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

func (r *PermissionRepository) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	if perm == nil || perm.Name == "" {
		return ErrInvalidArgument
	}

	const query = `INSERT INTO permissions (name, description)
        VALUES ($1, $2)
        RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query, perm.Name, perm.Description)
	return row.Scan(&perm.ID, &perm.CreatedAt)
}

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
	if err != nil {
		return domain.Permission{}, err
	}
	return perm, nil
}

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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}

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

func (r *PermissionRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	if role == nil || role.Name == "" {
		return ErrInvalidArgument
	}

	const query = `INSERT INTO roles (name, description)
        VALUES ($1, $2)
        RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query, role.Name, role.Description)
	return row.Scan(&role.ID, &role.CreatedAt)
}

func (r *PermissionRepository) RoleByName(ctx context.Context, name string) (domain.Role, error) {
	if name == "" {
		return domain.Role{}, ErrInvalidArgument
	}

	const query = `SELECT id, name, description, created_at FROM roles WHERE name = $1`
	var role domain.Role
	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Role{}, ErrNotFound
	}
	if err != nil {
		return domain.Role{}, err
	}
	return role, nil
}

func (r *PermissionRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *PermissionRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
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

func (r *PermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permID uuid.UUID) error {
	if roleID == uuid.Nil || permID == uuid.Nil {
		return ErrInvalidArgument
	}

	const query = `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, roleID, permID)
	return err
}

func (r *PermissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permID uuid.UUID) error {
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

func (r *PermissionRepository) PermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	if roleID == uuid.Nil {
		return nil, ErrInvalidArgument
	}

	const query = `SELECT p.id, p.name, p.description, p.created_at
        FROM permissions p
        JOIN role_permissions rp ON p.id = rp.permission_id
        WHERE rp.role_id = $1
        ORDER BY p.name`

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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *PermissionRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if userID == uuid.Nil || roleID == uuid.Nil {
		return ErrInvalidArgument
	}

	const query = `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *PermissionRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
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

func (r *PermissionRepository) RolesForUser(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidArgument
	}

	const query = `SELECT r.id, r.name, r.description, r.created_at
        FROM roles r
        JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1
        ORDER BY r.name`

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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *PermissionRepository) UserHasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	if userID == uuid.Nil || permissionName == "" {
		return false, ErrInvalidArgument
	}

	const query = `SELECT EXISTS (
        SELECT 1
        FROM user_roles ur
        JOIN role_permissions rp ON ur.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE ur.user_id = $1 AND p.name = $2
    )`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, permissionName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PermissionRepository) PermissionsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidArgument
	}

	const query = `SELECT DISTINCT p.id, p.name, p.description, p.created_at
        FROM permissions p
        JOIN role_permissions rp ON p.id = rp.permission_id
        JOIN user_roles ur ON rp.role_id = ur.role_id
        WHERE ur.user_id = $1
        ORDER BY p.name`

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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}
