// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/token"
)

type fakeAdminUsecase struct {
	createCalls int
}

func (u *fakeAdminUsecase) ListAdmins(context.Context) ([]domain.User, error) {
	return []domain.User{}, nil
}
func (u *fakeAdminUsecase) CreateAdmin(_ context.Context, req domain.CreateAdminRequest) (*domain.User, error) {
	u.createCalls++
	return &domain.User{ID: uuid.New(), Email: req.Email}, nil
}
func (u *fakeAdminUsecase) UpdateAdmin(context.Context, uuid.UUID, domain.UpdateAdminRequest) (*domain.User, error) {
	return nil, nil
}
func (u *fakeAdminUsecase) DeleteAdmin(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type activeAdmin map[uuid.UUID]bool

func (a activeAdmin) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	return a[id], nil
}

type testPermissions map[string]bool

func (p testPermissions) HasPermission(_ context.Context, _ uuid.UUID, key string) (bool, error) {
	return p[key], nil
}

func TestCreateAdminCannotAssignRolesWithoutPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminID := uuid.New()
	tokens := token.NewManager("test-secret-with-more-than-16-characters", time.Hour)
	jwt, _, err := tokens.Generate(adminID.String(), "operator@example.com")
	if err != nil {
		t.Fatal(err)
	}
	permissions := testPermissions{
		"users:create":       true,
		"users:assign_roles": false,
	}
	uc := &fakeAdminUsecase{}
	router := gin.New()
	protected := router.Group("/api/v1/admin", middleware.JWTAuth(tokens, activeAdmin{adminID: true}))
	NewAdminUserHandler(uc, permissions).RegisterAdminRoutes(protected, permissions)

	body := `{"email":"new@example.com","password":"correct-horse-battery","role_ids":["00000000-0000-4000-8000-000000000001"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if uc.createCalls != 0 {
		t.Fatal("admin user was created despite missing role-assignment permission")
	}
}
