package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	Key       string `json:"key"`
	Menu      string `json:"menu"`
	MenuLabel string `json:"menu_label"`
	Action    string `json:"action"`
	Label     string `json:"label"`
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	IsSystem    bool      `json:"is_system"`
	IsDefault   bool      `json:"is_default"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoleListResponse struct {
	Roles       []Role       `json:"roles"`
	Permissions []Permission `json:"permissions"`
}

type SaveRoleRequest struct {
	Name        string   `json:"name" binding:"required,max=64"`
	Permissions []string `json:"permissions"`
}

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]Role, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
}

type RoleUsecase interface {
	ListRoles(ctx context.Context) (*RoleListResponse, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	CreateRole(ctx context.Context, req SaveRoleRequest) (*Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req SaveRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
}
