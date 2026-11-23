// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

const trxCols = `id, invoice_number, customer_name, customer_email, customer_phone, shipping_address,
	total_amount, status, payment_url, created_at, updated_at`

type PostgresTransactionRepository struct{ db *pgxpool.Pool }

func NewPostgresTransactionRepository(db *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

var _ domain.TransactionRepository = (*PostgresTransactionRepository)(nil)

func scanTrx(row pgx.Row) (*domain.Transaction, error) {
	var t domain.Transaction
	err := row.Scan(&t.ID, &t.InvoiceNumber, &t.CustomerName, &t.CustomerEmail, &t.CustomerPhone,
		&t.ShippingAddress, &t.TotalAmount, &t.Status, &t.PaymentURL, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("transaction not found")
	}
	return &t, err
}

func (r *PostgresTransactionRepository) CreateWithItems(ctx context.Context, t *domain.Transaction) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	err = tx.QueryRow(ctx, `INSERT INTO transactions
		(id, invoice_number, customer_name, customer_email, customer_phone, shipping_address, total_amount, status, payment_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING created_at, updated_at`,
		t.ID, t.InvoiceNumber, t.CustomerName, t.CustomerEmail, t.CustomerPhone, t.ShippingAddress,
		t.TotalAmount, t.Status, t.PaymentURL).Scan(&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return err
	}
	for _, it := range t.Items {
		if _, err := tx.Exec(ctx, `INSERT INTO transaction_items
			(id, transaction_id, product_id, quantity, price_at_purchase) VALUES ($1,$2,$3,$4,$5)`,
			it.ID, t.ID, it.ProductID, it.Quantity, it.PriceAtPurchase); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresTransactionRepository) UpdatePaymentURL(ctx context.Context, id uuid.UUID, url string) error {
	_, err := r.db.Exec(ctx, `UPDATE transactions SET payment_url=$2, updated_at=now() WHERE id=$1`, id, url)
	return err
}

func (r *PostgresTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	t, err := scanTrx(r.db.QueryRow(ctx, `SELECT `+trxCols+` FROM transactions WHERE id=$1`, id))
	if err != nil {
		return nil, err
	}
	return t, r.loadItems(ctx, t)
}

func (r *PostgresTransactionRepository) FindByInvoice(ctx context.Context, invoice string) (*domain.Transaction, error) {
	t, err := scanTrx(r.db.QueryRow(ctx, `SELECT `+trxCols+` FROM transactions WHERE invoice_number=$1`, invoice))
	if err != nil {
		return nil, err
	}
	return t, r.loadItems(ctx, t)
}

func (r *PostgresTransactionRepository) loadItems(ctx context.Context, t *domain.Transaction) error {
	rows, err := r.db.Query(ctx,
		`SELECT id, transaction_id, product_id, quantity, price_at_purchase FROM transaction_items WHERE transaction_id=$1`, t.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var it domain.TransactionItem
		if err := rows.Scan(&it.ID, &it.TransactionID, &it.ProductID, &it.Quantity, &it.PriceAtPurchase); err != nil {
			return err
		}
		t.Items = append(t.Items, it)
	}
	return rows.Err()
}

func (r *PostgresTransactionRepository) UpdateStatusIf(ctx context.Context, id uuid.UUID, from, to domain.Status) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE transactions SET status=$3, updated_at=now() WHERE id=$1 AND status=$2`, id, from, to)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *PostgresTransactionRepository) List(ctx context.Context, f domain.ListFilter) ([]domain.Transaction, int, error) {
	where, args := "TRUE", []any{}
	if f.Status != "" {
		args = append(args, f.Status)
		where = fmt.Sprintf("status = $%d", len(args))
	}
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, (f.Page-1)*f.Limit)
	rows, err := r.db.Query(ctx, fmt.Sprintf(
		`SELECT %s FROM transactions WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		trxCols, where, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Transaction, 0, f.Limit)
	for rows.Next() {
		t, err := scanTrx(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *t)
	}
	return items, total, rows.Err()
}

func (r *PostgresTransactionRepository) Summary(ctx context.Context) (*domain.TransactionSummary, error) {
	var s domain.TransactionSummary
	err := r.db.QueryRow(ctx, `SELECT
		COALESCE(SUM(total_amount) FILTER (WHERE status='PAID'), 0),
		COUNT(*) FILTER (WHERE status='PAID'),
		COUNT(*) FILTER (WHERE status='PENDING')
		FROM transactions`).Scan(&s.TotalRevenue, &s.TotalPaidTransactions, &s.PendingCount)
	return &s, err
}
