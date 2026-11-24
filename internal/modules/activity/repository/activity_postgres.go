// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/activity/domain"
)

type PostgresActivityRepository struct {
	db *pgxpool.Pool
}

func NewPostgresActivityRepository(db *pgxpool.Pool) *PostgresActivityRepository {
	return &PostgresActivityRepository{db: db}
}

var _ domain.Repository = (*PostgresActivityRepository)(nil)

func (r *PostgresActivityRepository) Record(ctx context.Context, userID uuid.UUID, action, entity string, before, after json.RawMessage) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO activity_log (id, user_id, action, entity, data_before_updated, data_after_updated)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, action, entity, nullableJSON(before), nullableJSON(after))
	return err
}

func (r *PostgresActivityRepository) List(ctx context.Context, filter domain.ListFilter) ([]domain.Activity, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM activity_log`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.datetime, COALESCE(u.email, 'Admin dihapus'), a.action, a.entity,
			a.data_before_updated, a.data_after_updated
		FROM activity_log a
		LEFT JOIN users u ON u.id = a.user_id
		ORDER BY a.datetime DESC, a.id DESC
		LIMIT $1 OFFSET $2`, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Activity, 0, filter.Limit)
	for rows.Next() {
		var item domain.Activity
		var before, after []byte
		if err := rows.Scan(&item.ID, &item.DateTime, &item.User, &item.Action, &item.Entity, &before, &after); err != nil {
			return nil, 0, err
		}
		item.DataBefore = before
		item.DataAfter = after
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *PostgresActivityRepository) Snapshot(ctx context.Context, entity string, id uuid.UUID) (json.RawMessage, error) {
	var query string
	switch entity {
	case "products":
		query = `SELECT jsonb_build_object('id',id,'name',name,'sku',sku,'description',description,'price',price,'stock',stock,'category_id',category_id,'image_url',image_url,'created_at',created_at,'updated_at',updated_at) FROM products WHERE id=$1 AND deleted_at IS NULL`
	case "categories":
		query = `SELECT jsonb_build_object('id',id,'name',name,'created_at',created_at) FROM categories WHERE id=$1`
	case "users":
		query = `SELECT jsonb_build_object('id',u.id,'email',u.email,'roles',COALESCE((SELECT jsonb_agg(jsonb_build_object('id',r.id,'name',r.name) ORDER BY r.name) FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=u.id),'[]'::jsonb),'created_at',u.created_at,'updated_at',u.updated_at) FROM users u WHERE u.id=$1`
	case "roles":
		query = `SELECT jsonb_build_object('id',r.id,'name',r.name,'is_system',r.is_system,'is_default',r.is_default,'permissions',COALESCE((SELECT jsonb_agg(rp.permission_key ORDER BY rp.permission_key) FROM role_permissions rp WHERE rp.role_id=r.id),'[]'::jsonb),'created_at',r.created_at,'updated_at',r.updated_at) FROM roles r WHERE r.id=$1`
	case "contact_messages":
		query = `SELECT jsonb_build_object('id',id,'reply_message',reply_message,'replied_at',replied_at,'replied_by',replied_by) FROM contact_message WHERE id=$1`
	default:
		return nil, nil
	}

	var data []byte
	err := r.db.QueryRow(ctx, query, id).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func nullableJSON(data json.RawMessage) any {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	return json.RawMessage(data)
}
