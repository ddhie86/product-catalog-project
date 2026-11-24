// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	ID         uuid.UUID       `json:"id"`
	DateTime   time.Time       `json:"datetime"`
	User       string          `json:"user"`
	Action     string          `json:"action"`
	Entity     string          `json:"entity"`
	DataBefore json.RawMessage `json:"data_before_updated"`
	DataAfter  json.RawMessage `json:"data_after_updated"`
}

type ListFilter struct {
	Page  int
	Limit int
}

type Repository interface {
	Record(ctx context.Context, userID uuid.UUID, action, entity string, before, after json.RawMessage) error
	List(ctx context.Context, filter ListFilter) ([]Activity, int, error)
	Snapshot(ctx context.Context, entity string, id uuid.UUID) (json.RawMessage, error)
}
