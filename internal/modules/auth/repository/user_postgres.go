package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type PostgresUserRepository struct{ db *pgxpool.Pool }

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

var _ domain.UserRepository = (*PostgresUserRepository)(nil)
var _ domain.AdminUserRepository = (*PostgresUserRepository)(nil)

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("user not found")
	}
	return &u, err
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1,$2,$3) RETURNING created_at, updated_at`,
		u.ID, u.Email, u.PasswordHash).Scan(&u.CreatedAt, &u.UpdatedAt)
	return uniqueEmailError(err)
}

func (r *PostgresUserRepository) CreateAdmin(ctx context.Context, u *domain.User, roleIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1,$2,$3) RETURNING created_at, updated_at`,
		u.ID, u.Email, u.PasswordHash).Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return uniqueEmailError(err)
	}
	if len(roleIDs) == 0 {
		tag, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE is_default=TRUE`, u.ID)
		if err != nil {
			return roleAssignmentError(err)
		}
		if tag.RowsAffected() == 0 {
			return apperror.Internal(errors.New("default role is missing"))
		}
	} else if err := replaceUserRoles(ctx, tx, u.ID, roleIDs); err != nil {
		return roleAssignmentError(err)
	}
	u.Roles, err = loadUserRoles(ctx, tx, u.ID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresUserRepository) ListAdmins(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.email, u.created_at, u.updated_at,
			COALESCE(jsonb_agg(jsonb_build_object('id', r.id, 'name', r.name) ORDER BY r.name)
				FILTER (WHERE r.id IS NOT NULL), '[]'::jsonb)::text
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id=u.id
		LEFT JOIN roles r ON r.id=ur.role_id
		GROUP BY u.id
		ORDER BY u.created_at, u.email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		var rolesJSON string
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt, &rolesJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(rolesJSON), &u.Roles); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *PostgresUserRepository) UpdateAdmin(ctx context.Context, u *domain.User, roleIDs []uuid.UUID) (*domain.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `LOCK TABLE users, user_roles IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return nil, err
	}
	const q = `UPDATE users
		SET email=$2,
			password_hash=CASE WHEN $3 = '' THEN password_hash ELSE $3 END,
			updated_at=now()
		WHERE id=$1
		RETURNING id, email, created_at, updated_at`
	err = tx.QueryRow(ctx, q, u.ID, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("admin not found")
	}
	if err != nil {
		return nil, uniqueEmailError(err)
	}
	if roleIDs != nil {
		if err := replaceUserRoles(ctx, tx, u.ID, roleIDs); err != nil {
			return nil, roleAssignmentError(err)
		}
	}
	var superAdminCount int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_roles ur JOIN roles r ON r.id=ur.role_id
		WHERE r.is_system=TRUE`).Scan(&superAdminCount); err != nil {
		return nil, err
	}
	if superAdminCount == 0 {
		return nil, apperror.Conflict("at least one super admin must remain")
	}
	u.Roles, err = loadUserRoles(ctx, tx, u.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *PostgresUserRepository) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `LOCK TABLE users, user_roles IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return err
	}
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return apperror.NotFound("admin not found")
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return apperror.Conflict("cannot delete the last admin")
	}

	var hasSuperAdminRole bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id
			WHERE ur.user_id=$1 AND r.is_system=TRUE
		)`, id).Scan(&hasSuperAdminRole); err != nil {
		return err
	}
	if hasSuperAdminRole {
		var superAdminCount int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM user_roles ur JOIN roles r ON r.id=ur.role_id
			WHERE r.is_system=TRUE`).Scan(&superAdminCount); err != nil {
			return err
		}
		if superAdminCount <= 1 {
			return apperror.Conflict("cannot delete the last super admin")
		}
	}

	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("admin not found")
	}
	return tx.Commit(ctx)
}

func replaceUserRoles(ctx context.Context, tx pgx.Tx, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if len(roleIDs) == 0 {
		return apperror.BadRequest("at least one role must be assigned")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1`, userID); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1,$2)`, userID, roleID); err != nil {
			return err
		}
	}
	return nil
}

func loadUserRoles(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]domain.Role, error) {
	rows, err := tx.Query(ctx, `SELECT r.id, r.name FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1 ORDER BY r.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func roleAssignmentError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return apperror.BadRequest("one or more roles do not exist")
	}
	return err
}

func (r *PostgresUserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}

func uniqueEmailError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.Conflict("email already registered")
	}
	return err
}
