// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/token"
)

type fakeProductImageStorage struct {
	result      domain.ProductImageUploadResult
	err         error
	content     []byte
	contentType string
	calls       int
}

func (s *fakeProductImageStorage) Upload(_ context.Context, content io.Reader, contentType, _ string) (domain.ProductImageUploadResult, error) {
	s.calls++
	s.contentType = contentType
	s.content, _ = io.ReadAll(content)
	return s.result, s.err
}

type imagePermissionChecker map[string]bool

func (p imagePermissionChecker) HasPermission(_ context.Context, _ uuid.UUID, permission string) (bool, error) {
	return p[permission], nil
}

func TestProductImageUploadReturnsServiceUnavailableWhenStorageIsNotConfigured(t *testing.T) {
	recorder, context := newUploadContext(t, "image.png", []byte("\x89PNG\r\n\x1a\n"))
	NewProductImageHandler(nil, 5<<20).Upload(context)
	if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), "image storage is not configured") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestProductImageUploadRejectsOversizedFile(t *testing.T) {
	storage := &fakeProductImageStorage{}
	recorder, context := newUploadContext(t, "image.png", []byte("\x89PNG\r\n\x1a\nx"))
	NewProductImageHandler(storage, 8).Upload(context)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
	if storage.calls != 0 {
		t.Fatal("oversized image reached storage")
	}
}

func TestProductImageUploadRejectsInvalidContentAndExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  []byte
	}{
		{name: "unsupported format", filename: "image.svg", content: []byte("<svg/>")},
		{name: "mismatched extension", filename: "image.jpg", content: []byte("\x89PNG\r\n\x1a\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &fakeProductImageStorage{}
			recorder, context := newUploadContext(t, tt.filename, tt.content)
			NewProductImageHandler(storage, 5<<20).Upload(context)
			if recorder.Code != http.StatusUnsupportedMediaType {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusUnsupportedMediaType, recorder.Body.String())
			}
			if storage.calls != 0 {
				t.Fatal("invalid image reached storage")
			}
		})
	}
}

func TestProductImageUploadSuccessUsesDetectedImageType(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\nimage-data")
	storage := &fakeProductImageStorage{
		result: domain.ProductImageUploadResult{
			ImageURL: "https://drive.google.com/uc?export=view&id=file-id",
			FileID:   "file-id",
			Name:     "safe-name.png",
		},
	}
	recorder, context := newUploadContext(t, "client-name.png", png)
	NewProductImageHandler(storage, 5<<20).Upload(context)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if storage.calls != 1 || storage.contentType != "image/png" || !bytes.Equal(storage.content, png) {
		t.Fatalf("storage upload = %#v", storage)
	}
	var response struct {
		Data domain.ProductImageUploadResult `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.ImageURL != "https://drive.google.com/uc?export=view&id=file-id" {
		t.Fatalf("response = %s", recorder.Body.String())
	}
}

func TestProductImageUploadSanitizesStorageErrors(t *testing.T) {
	storage := &fakeProductImageStorage{err: errors.New("private provider response")}
	recorder, context := newUploadContext(t, "image.png", []byte("\x89PNG\r\n\x1a\n"))
	NewProductImageHandler(storage, 5<<20).Upload(context)
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if strings.Contains(recorder.Body.String(), "private provider response") {
		t.Fatalf("provider error leaked: %s", recorder.Body.String())
	}
}

func TestProductImageRouteRequiresJWTAndEitherProductPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := token.NewManager("test-secret-with-more-than-16-characters", time.Hour)
	adminID := uuid.New()
	storage := &fakeProductImageStorage{
		result: domain.ProductImageUploadResult{
			ImageURL: "https://drive.google.com/uc?export=view&id=file-id",
			FileID:   "file-id",
			Name:     "safe-name.png",
		},
	}
	checker := imagePermissionChecker{"products:edit": true}
	router := gin.New()
	admin := router.Group("/api/v1/admin", middleware.JWTAuth(manager, activeAdminVerifier{adminID: true}))
	NewProductImageHandler(storage, 5<<20).RegisterAdminRoutes(admin, checker)

	unauthenticated := multipartUploadRequest(t, "image.png", []byte("\x89PNG\r\n\x1a\n"))
	unauthenticatedRecorder := httptest.NewRecorder()
	router.ServeHTTP(unauthenticatedRecorder, unauthenticated)
	if unauthenticatedRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", unauthenticatedRecorder.Code)
	}

	jwt, _, err := manager.Generate(adminID.String(), "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	checker["products:edit"] = false
	checker["products:create"] = false
	forbidden := multipartUploadRequest(t, "image.png", []byte("\x89PNG\r\n\x1a\n"))
	forbidden.Header.Set("Authorization", "Bearer "+jwt)
	forbiddenRecorder := httptest.NewRecorder()
	router.ServeHTTP(forbiddenRecorder, forbidden)
	if forbiddenRecorder.Code != http.StatusForbidden {
		t.Fatalf("insufficient permission status = %d", forbiddenRecorder.Code)
	}

	checker["products:edit"] = true
	authorized := multipartUploadRequest(t, "image.png", []byte("\x89PNG\r\n\x1a\n"))
	authorized.Header.Set("Authorization", "Bearer "+jwt)
	authorizedRecorder := httptest.NewRecorder()
	router.ServeHTTP(authorizedRecorder, authorized)
	if authorizedRecorder.Code != http.StatusCreated {
		t.Fatalf("products:edit should allow upload: %d %s", authorizedRecorder.Code, authorizedRecorder.Body.String())
	}
}

func newUploadContext(t *testing.T, filename string, content []byte) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = multipartUploadRequest(t, filename, content)
	return recorder, context
}

func multipartUploadRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/product-images", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
