// Package domain (auth): entity & kontrak autentikasi admin.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=12,max=72"`
}

type UpdateAdminRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"omitempty,min=12,max=72"`
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User) error
}

type AdminUserRepository interface {
	Create(ctx context.Context, u *User) error
	ListAdmins(ctx context.Context) ([]User, error)
	UpdateAdmin(ctx context.Context, u *User) (*User, error)
	DeleteAdmin(ctx context.Context, id uuid.UUID) error
}

type AuthUsecase interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

type AdminUserUsecase interface {
	ListAdmins(ctx context.Context) ([]User, error)
	CreateAdmin(ctx context.Context, req CreateAdminRequest) (*User, error)
	UpdateAdmin(ctx context.Context, id uuid.UUID, req UpdateAdminRequest) (*User, error)
	DeleteAdmin(ctx context.Context, id, actorID uuid.UUID) error
}
