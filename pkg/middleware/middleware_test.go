package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/pkg/token"
)

type fakeAdminUserVerifier map[uuid.UUID]bool

func (f fakeAdminUserVerifier) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	return f[id], nil
}

type fakePermissionChecker map[string]bool

func (f fakePermissionChecker) HasPermission(_ context.Context, _ uuid.UUID, permission string) (bool, error) {
	return f[permission], nil
}

func TestJWTAuthRequiresActiveAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := token.NewManager("test-secret-with-more-than-16-characters", time.Hour)
	activeID, deletedID := uuid.New(), uuid.New()
	activeToken, _, err := manager.Generate(activeID.String(), "active@example.com")
	if err != nil {
		t.Fatal(err)
	}
	deletedToken, _, err := manager.Generate(deletedID.String(), "deleted@example.com")
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/admin", JWTAuth(manager, fakeAdminUserVerifier{activeID: true}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "missing token", wantStatus: http.StatusUnauthorized},
		{name: "deleted admin token", token: deletedToken, wantStatus: http.StatusUnauthorized},
		{name: "active admin token", token: activeToken, wantStatus: http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequirePermissionBlocksUnauthorizedAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := token.NewManager("test-secret-with-more-than-16-characters", time.Hour)
	adminID := uuid.New()
	jwt, _, err := manager.Generate(adminID.String(), "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.POST("/admin/products", JWTAuth(manager, fakeAdminUserVerifier{adminID: true}),
		RequirePermission(fakePermissionChecker{"products:view": true}, "products:create"),
		func(c *gin.Context) { c.Status(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodPost, "/admin/products", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}
