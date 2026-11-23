// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
// Package repository (product): implementasi PostgreSQL untuk ProductRepository.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

const productCols = `id, name, sku, description, price, stock, category_id, image_url, created_at, updated_at`

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProductRepository(db *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

var _ domain.ProductRepository = (*PostgresProductRepository)(nil)
var _ domain.CategoryRepository = (*PostgresProductRepository)(nil)

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var p domain.Product
	err := row.Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.Price, &p.Stock,
		&p.CategoryID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("product not found")
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *domain.Product) error {
	const q = `INSERT INTO products (id, name, sku, description, price, stock, category_id, image_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING created_at, updated_at`
	err := r.db.QueryRow(ctx, q, p.ID, p.Name, p.SKU, p.Description, p.Price, p.Stock, p.CategoryID, p.ImageURL).
		Scan(&p.CreatedAt, &p.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.Conflict("sku already exists")
	}
	return err
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *domain.Product) error {
	const q = `UPDATE products
		SET name=$2, description=$3, price=$4, stock=$5, category_id=$6, image_url=$7, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL RETURNING updated_at`
	err := r.db.QueryRow(ctx, q, p.ID, p.Name, p.Description, p.Price, p.Stock, p.CategoryID, p.ImageURL).
		Scan(&p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("product not found")
	}
	return err
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE products SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("product not found")
	}
	return nil
}

func (r *PostgresProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return scanProduct(r.db.QueryRow(ctx,
		`SELECT `+productCols+` FROM products WHERE id=$1 AND deleted_at IS NULL`, id))
}

func (r *PostgresProductRepository) FindBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return scanProduct(r.db.QueryRow(ctx,
		`SELECT `+productCols+` FROM products WHERE sku=$1 AND deleted_at IS NULL`, sku))
}

func (r *PostgresProductRepository) FetchAdminList(ctx context.Context, f domain.AdminListFilter) ([]domain.Product, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR sku ILIKE $%d)", len(args), len(args)))
	}
	return r.list(ctx, strings.Join(where, " AND "), "created_at DESC", args, f.Page, f.Limit)
}

func (r *PostgresProductRepository) FetchCatalog(ctx context.Context, f domain.CatalogFilter) ([]domain.Product, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", len(args), len(args)))
	}
	if f.CategoryID != nil {
		args = append(args, *f.CategoryID)
		where = append(where, fmt.Sprintf("category_id = $%d", len(args)))
	}
	// Whitelist: nilai sort TIDAK PERNAH diinterpolasi dari input mentah.
	order := "created_at DESC"
	switch f.Sort {
	case domain.SortPriceAsc:
		order = "price ASC"
	case domain.SortPriceDesc:
		order = "price DESC"
	case domain.SortPopular:
		// TODO: urutkan berdasarkan jumlah terjual (agregasi transaction_items)
		// setelah kontrak lintas-modul untuk statistik penjualan ditetapkan.
		order = "created_at DESC"
	}
	return r.list(ctx, strings.Join(where, " AND "), order, args, f.Page, f.Limit)
}

// list: where & order dibangun dari konstanta internal, nilai user selalu lewat args.
func (r *PostgresProductRepository) list(ctx context.Context, where, order string, args []any, page, limit int) ([]domain.Product, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, (page-1)*limit)
	q := fmt.Sprintf(`SELECT %s FROM products WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		productCols, where, order, len(args)-1, len(args))
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Product, 0, limit)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *p)
	}
	return items, total, rows.Err()
}

func (r *PostgresProductRepository) CategoryExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id=$1)`, id).Scan(&ok)
	return ok, err
}

func (r *PostgresProductRepository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Category{}
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *PostgresProductRepository) CreateCategory(ctx context.Context, category *domain.Category) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO categories (id, name) VALUES ($1, $2) RETURNING id, name`,
		category.ID, category.Name,
	).Scan(&category.ID, &category.Name)
	return categoryConflict(err)
}

func (r *PostgresProductRepository) UpdateCategory(ctx context.Context, category *domain.Category) error {
	err := r.db.QueryRow(ctx,
		`UPDATE categories SET name=$2 WHERE id=$1 RETURNING id, name`,
		category.ID, category.Name,
	).Scan(&category.ID, &category.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("category not found")
	}
	return categoryConflict(err)
}

func (r *PostgresProductRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperror.Conflict("category is assigned to one or more products")
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("category not found")
	}
	return nil
}

func categoryConflict(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.Conflict("category name already exists")
	}
	return err
}

func (r *PostgresProductRepository) DeductStock(ctx context.Context, id uuid.UUID, qty int) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE products SET stock = stock - $2, updated_at = now()
		 WHERE id=$1 AND stock >= $2 AND deleted_at IS NULL`, id, qty)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.InsufficientStock("insufficient stock or product unavailable")
	}
	return nil
}

func (r *PostgresProductRepository) RestoreStock(ctx context.Context, id uuid.UUID, qty int) error {
	_, err := r.db.Exec(ctx, `UPDATE products SET stock = stock + $2, updated_at = now() WHERE id=$1`, id, qty)
	return err
}

func (r *PostgresProductRepository) CountLowStock(ctx context.Context, threshold int) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM products WHERE stock < $1 AND deleted_at IS NULL`, threshold).Scan(&n)
	return n, err
}
