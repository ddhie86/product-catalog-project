// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

const multipartOverheadBytes int64 = 64 << 10

type ProductImageHandler struct {
	storage        domain.ProductImageStorage
	maxUploadBytes int64
}

func NewProductImageHandler(storage domain.ProductImageStorage, maxUploadBytes int64) *ProductImageHandler {
	return &ProductImageHandler{storage: storage, maxUploadBytes: maxUploadBytes}
}

func (h *ProductImageHandler) RegisterAdminRoutes(rg *gin.RouterGroup, permissions middleware.PermissionChecker) {
	rg.POST("/product-images",
		middleware.RequireAnyPermission(permissions, "products:create", "products:edit"),
		h.Upload,
	)
}

func (h *ProductImageHandler) Upload(c *gin.Context) {
	if h.storage == nil {
		response.Error(c, apperror.New(
			"IMAGE_STORAGE_NOT_CONFIGURED",
			"image storage is not configured",
			http.StatusServiceUnavailable,
		))
		return
	}
	if h.maxUploadBytes <= 0 {
		response.Error(c, apperror.Internal(errors.New("invalid image upload size configuration")))
		return
	}
	contentType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || contentType != "multipart/form-data" {
		response.Error(c, apperror.BadRequest("request must use multipart/form-data"))
		return
	}

	bodyLimit := h.maxUploadBytes + multipartOverheadBytes
	if c.Request.ContentLength > bodyLimit {
		response.Error(c, apperror.New("PAYLOAD_TOO_LARGE", "image exceeds the maximum upload size", http.StatusRequestEntityTooLarge))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, bodyLimit)
	if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			response.Error(c, apperror.New("PAYLOAD_TOO_LARGE", "image exceeds the maximum upload size", http.StatusRequestEntityTooLarge))
			return
		}
		response.Error(c, apperror.BadRequest("invalid multipart upload"))
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}

	form := c.Request.MultipartForm
	if form == nil || len(form.File) != 1 || len(form.File["file"]) != 1 || len(form.Value) != 0 {
		response.Error(c, apperror.BadRequest("provide exactly one file in the 'file' field"))
		return
	}
	fileHeader := form.File["file"][0]
	if fileHeader.Size > h.maxUploadBytes {
		response.Error(c, apperror.New("PAYLOAD_TOO_LARGE", "image exceeds the maximum upload size", http.StatusRequestEntityTooLarge))
		return
	}

	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !extensionMatchesImageType(extension) {
		response.Error(c, unsupportedImageError())
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, apperror.BadRequest("unable to read uploaded image"))
		return
	}
	defer file.Close()

	sample := make([]byte, 512)
	n, err := io.ReadFull(file, sample)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		response.Error(c, apperror.BadRequest("unable to read uploaded image"))
		return
	}
	detectedType, detectedExtension, ok := detectImageType(sample[:n])
	if !ok || !extensionsMatch(extension, detectedExtension) {
		response.Error(c, unsupportedImageError())
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.Error(c, apperror.BadRequest("unable to read uploaded image"))
		return
	}

	result, err := h.storage.Upload(c.Request.Context(), io.LimitReader(file, h.maxUploadBytes+1), detectedType, detectedExtension)
	if err != nil || result.ImageURL == "" || result.FileID == "" || result.Name == "" {
		slog.Error("product image upload failed", "outcome", "failure")
		response.Error(c, apperror.New("IMAGE_UPLOAD_FAILED", "image upload failed", http.StatusBadGateway))
		return
	}
	slog.Info("product image upload succeeded", "outcome", "success", "file_id", result.FileID)
	response.Created(c, result)
}

func unsupportedImageError() *apperror.Error {
	return apperror.New("UNSUPPORTED_IMAGE", "only JPEG, PNG, and WebP images are supported", http.StatusUnsupportedMediaType)
}

func extensionMatchesImageType(extension string) bool {
	return extension == ".jpg" || extension == ".jpeg" || extension == ".png" || extension == ".webp"
}

func extensionsMatch(submitted, detected string) bool {
	if detected == ".jpg" {
		return submitted == ".jpg" || submitted == ".jpeg"
	}
	return submitted == detected
}

func detectImageType(content []byte) (string, string, bool) {
	switch {
	case len(content) >= 3 && content[0] == 0xff && content[1] == 0xd8 && content[2] == 0xff:
		return "image/jpeg", ".jpg", true
	case len(content) >= 8 && string(content[:8]) == "\x89PNG\r\n\x1a\n":
		return "image/png", ".png", true
	case len(content) >= 12 && string(content[:4]) == "RIFF" && string(content[8:12]) == "WEBP":
		return "image/webp", ".webp", true
	default:
		return "", "", false
	}
}
