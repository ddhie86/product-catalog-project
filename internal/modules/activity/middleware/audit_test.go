// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	activitydomain "github.com/dedysetiawan/product-catalog-project/internal/modules/activity/domain"
	authmiddleware "github.com/dedysetiawan/product-catalog-project/pkg/middleware"
)

type fakeActivityRepository struct {
	snapshots int
	recorded  bool
	action    string
	entity    string
	before    json.RawMessage
	after     json.RawMessage
}

func (r *fakeActivityRepository) Record(_ context.Context, _ uuid.UUID, action, entity string, before, after json.RawMessage) error {
	r.recorded, r.action, r.entity, r.before, r.after = true, action, entity, before, after
	return nil
}

func (*fakeActivityRepository) List(context.Context, activitydomain.ListFilter) ([]activitydomain.Activity, int, error) {
	return nil, 0, nil
}

func (r *fakeActivityRepository) Snapshot(_ context.Context, _ string, id uuid.UUID) (json.RawMessage, error) {
	r.snapshots++
	return json.RawMessage(`{"id":"` + id.String() + `"}`), nil
}

type fakePermissionChecker bool

func (p fakePermissionChecker) HasPermission(context.Context, uuid.UUID, string) (bool, error) {
	return bool(p), nil
}

func TestAuditCapturesCreatedEntityAndAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeActivityRepository{}
	userID := uuid.New()
	productID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(authmiddleware.CtxUserID, userID.String())
		c.Next()
	}, Audit(repo, fakePermissionChecker(true)))
	router.POST("/api/v1/admin/products", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": productID}})
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/products", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if !repo.recorded || repo.action != "CREATE" || repo.entity != "products" {
		t.Fatalf("recorded activity = %#v", repo)
	}
	if repo.snapshots != 1 || repo.before != nil || string(repo.after) != `{"id":"`+productID.String()+`"}` {
		t.Fatalf("snapshot state: snapshots=%d before=%s after=%s", repo.snapshots, repo.before, repo.after)
	}
}

func TestAuditDoesNotRunBeforePermissionIsGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeActivityRepository{}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(authmiddleware.CtxUserID, uuid.NewString())
		c.Next()
	}, Audit(repo, fakePermissionChecker(false)))
	router.PUT("/api/v1/admin/products/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPut, "/api/v1/admin/products/"+uuid.NewString(), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if repo.snapshots != 0 || repo.recorded {
		t.Fatalf("unauthorized operation reached audit/data layer: %#v", repo)
	}
}
