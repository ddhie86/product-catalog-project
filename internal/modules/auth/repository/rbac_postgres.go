package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type PostgresRoleRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRoleRepository(db *pgxpool.Pool) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

var _ domain.RoleRepository = (*PostgresRoleRepository)(nil)

func (r *PostgresRoleRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.name, r.is_system, r.is_default, r.created_at, r.updated_at,
			COALESCE(array_agg(rp.permission_key ORDER BY rp.permission_key)
				FILTER (WHERE rp.permission_key IS NOT NULL), '{}')
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id=r.id
		GROUP BY r.id
		ORDER BY r.is_system DESC, r.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.IsSystem, &role.IsDefault, &role.CreatedAt, &role.UpdatedAt, &role.Permissions); err != nil {
			return nil, err
		}
		if role.Permissions == nil {
			role.Permissions = []string{}
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *PostgresRoleRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT key, menu, menu_label, action, label
		FROM permissions
		ORDER BY menu, CASE action
			WHEN 'view' THEN 0
			WHEN 'create' THEN 1
			WHEN 'edit' THEN 2
			WHEN 'delete' THEN 3
			ELSE 4
		END`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]domain.Permission, 0)
	for rows.Next() {
		var permission domain.Permission
		if err := rows.Scan(&permission.Key, &permission.Menu, &permission.MenuLabel, &permission.Action, &permission.Label); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *PostgresRoleRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT rp.permission_key
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id=ur.role_id
		WHERE ur.user_id=$1
		ORDER BY rp.permission_key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *PostgresRoleRepository) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	var allowed bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN role_permissions rp ON rp.role_id=ur.role_id
			WHERE ur.user_id=$1 AND rp.permission_key=$2
		)`, userID, permission).Scan(&allowed)
	return allowed, err
}

func (r *PostgresRoleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO roles (id, name) VALUES ($1, $2)
		RETURNING created_at, updated_at`, role.ID, role.Name).Scan(&role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return roleWriteError(err)
	}
	if err := replaceRolePermissions(ctx, tx, role.ID, role.Permissions); err != nil {
		return roleWriteError(err)
	}
	return tx.Commit(ctx)
}

func (r *PostgresRoleRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE roles SET name=$2, updated_at=now()
		WHERE id=$1 AND is_system=FALSE
		RETURNING is_default, created_at, updated_at`, role.ID, role.Name).
		Scan(&role.IsDefault, &role.CreatedAt, &role.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r.roleUnavailable(ctx, tx, role.ID)
	}
	if err != nil {
		return roleWriteError(err)
	}
	if err := replaceRolePermissions(ctx, tx, role.ID, role.Permissions); err != nil {
		return roleWriteError(err)
	}
	return tx.Commit(ctx)
}

func (r *PostgresRoleRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `LOCK TABLE users, user_roles IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return err
	}
	var systemRole, defaultRole bool
	if err := tx.QueryRow(ctx, `SELECT is_system, is_default FROM roles WHERE id=$1 FOR UPDATE`, id).Scan(&systemRole, &defaultRole); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("role not found")
		}
		return err
	}
	if defaultRole {
		return apperror.Forbidden("default role cannot be deleted")
	}
	if systemRole {
		return apperror.Forbidden("system role cannot be deleted")
	}
	var assigned bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles WHERE role_id=$1)`, id).Scan(&assigned); err != nil {
		return err
	}
	if assigned {
		return apperror.Conflict("role is assigned to one or more admin users")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM roles WHERE id=$1`, id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperror.Conflict("role is assigned to one or more admin users")
		}
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRoleRepository) roleUnavailable(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1)`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return apperror.NotFound("role not found")
	}
	return apperror.Forbidden("system role cannot be modified")
}

func replaceRolePermissions(ctx context.Context, tx pgx.Tx, roleID uuid.UUID, permissions []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id=$1`, roleID); err != nil {
		return err
	}
	for _, permission := range permissions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_permissions (role_id, permission_key) VALUES ($1, $2)`,
			roleID, permission); err != nil {
			return err
		}
	}
	return nil
}

func roleWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return apperror.Conflict("role name already exists")
		case "23503":
			return apperror.BadRequest("role contains an unknown permission")
		}
	}
	return err
}
