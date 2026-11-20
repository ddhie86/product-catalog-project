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

type PostgresUserRepository struct{ db *pgxpool.Pool }

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

var _ domain.UserRepository = (*PostgresUserRepository)(nil)

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

func (r *PostgresUserRepository) ListAdmins(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, email, created_at, updated_at FROM users ORDER BY created_at, email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *PostgresUserRepository) UpdateAdmin(ctx context.Context, u *domain.User) (*domain.User, error) {
	const q = `UPDATE users
		SET email=$2,
			password_hash=CASE WHEN $3 = '' THEN password_hash ELSE $3 END,
			updated_at=now()
		WHERE id=$1
		RETURNING id, email, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, u.ID, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("admin not found")
	}
	if err != nil {
		return nil, uniqueEmailError(err)
	}
	return u, nil
}

func (r *PostgresUserRepository) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `LOCK TABLE users IN SHARE ROW EXCLUSIVE MODE`); err != nil {
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

	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("admin not found")
	}
	return tx.Commit(ctx)
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
