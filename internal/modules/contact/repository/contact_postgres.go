// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/contact/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

const contactMessageColumns = `id, name, email, type, message, COALESCE(reply_message, ''), replied_at, replied_by, created_at`

type PostgresContactRepository struct {
	db *pgxpool.Pool
}

func NewPostgresContactRepository(db *pgxpool.Pool) *PostgresContactRepository {
	return &PostgresContactRepository{db: db}
}

var _ domain.ContactRepository = (*PostgresContactRepository)(nil)

func (r *PostgresContactRepository) Create(ctx context.Context, message *domain.ContactMessage) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO contact_message (id, name, email, type, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`,
		message.ID, message.Name, message.Email, message.Type, message.Message,
	).Scan(&message.CreatedAt)
}

func (r *PostgresContactRepository) List(ctx context.Context, page, limit int) ([]domain.ContactMessage, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contact_message`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT `+contactMessageColumns+`
		FROM contact_message
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	messages := make([]domain.ContactMessage, 0, limit)
	for rows.Next() {
		var message domain.ContactMessage
		if err := rows.Scan(
			&message.ID, &message.Name, &message.Email, &message.Type, &message.Message,
			&message.ReplyMessage, &message.RepliedAt, &message.RepliedBy, &message.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		messages = append(messages, message)
	}
	return messages, total, rows.Err()
}

func (r *PostgresContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ContactMessage, error) {
	var message domain.ContactMessage
	err := r.db.QueryRow(ctx, `SELECT `+contactMessageColumns+` FROM contact_message WHERE id=$1`, id).Scan(
		&message.ID, &message.Name, &message.Email, &message.Type, &message.Message,
		&message.ReplyMessage, &message.RepliedAt, &message.RepliedBy, &message.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("contact message not found")
	}
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *PostgresContactRepository) SaveReply(ctx context.Context, message *domain.ContactMessage) error {
	err := r.db.QueryRow(ctx, `
		UPDATE contact_message
		SET reply_message=$2, replied_at=now(), replied_by=$3
		WHERE id=$1
		RETURNING replied_at`, message.ID, message.ReplyMessage, message.RepliedBy).Scan(&message.RepliedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("contact message not found")
	}
	return err
}
