package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/token"
)

type fakeCategoryUsecase struct{}

func (fakeCategoryUsecase) ListCategories(context.Context) ([]domain.Category, error) {
	return []domain.Category{}, nil
}
func (fakeCategoryUsecase) CreateCategory(context.Context, domain.CreateCategoryRequest) (*domain.Category, error) {
	return nil, nil
}
func (fakeCategoryUsecase) UpdateCategory(context.Context, uuid.UUID, domain.UpdateCategoryRequest) (*domain.Category, error) {
	return nil, nil
}
func (fakeCategoryUsecase) DeleteCategory(context.Context, uuid.UUID) error { return nil }

type activeAdminVerifier map[uuid.UUID]bool

func (v activeAdminVerifier) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	return v[id], nil
}

type allowAllPermissions struct{}

func (allowAllPermissions) HasPermission(context.Context, uuid.UUID, string) (bool, error) {
	return true, nil
}

func TestAdminCategoryRoutesRequireJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := token.NewManager("test-secret-with-more-than-16-characters", time.Hour)
	adminID := uuid.New()
	router := gin.New()
	admin := router.Group("/api/v1/admin", middleware.JWTAuth(manager, activeAdminVerifier{adminID: true}))
	NewCategoryHandler(fakeCategoryUsecase{}).RegisterAdminRoutes(admin, allowAllPermissions{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/categories", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	jwt, _, err := manager.Generate(adminID.String(), "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/categories", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}
