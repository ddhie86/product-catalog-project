// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
// Package domain (product) berisi entity, DTO, dan kontrak modul Product.
package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type ProductImageUploadResult struct {
	ImageURL string `json:"image_url"`
	FileID   string `json:"file_id"`
	Name     string `json:"name"`
}

type ProductImageStorage interface {
	Upload(ctx context.Context, content io.Reader, contentType, extension string) (ProductImageUploadResult, error)
}

type Product struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	SKU         string     `json:"sku"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	ImageURL    string     `json:"image_url"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Category struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// Sort options untuk katalog publik.
const (
	SortNewest    = "newest"
	SortPriceAsc  = "price_asc"
	SortPriceDesc = "price_desc"
	SortPopular   = "popular"
)

type CatalogFilter struct {
	Search     string
	CategoryID *uuid.UUID
	Sort       string
	Page       int
	Limit      int
}

type AdminListFilter struct {
	Search string
	Page   int
	Limit  int
}

// ---- Request DTO ----

type CreateProductRequest struct {
	Name        string     `json:"name" binding:"required,max=255"`
	SKU         string     `json:"sku" binding:"required,max=64"`
	Description string     `json:"description"`
	Price       float64    `json:"price" binding:"required,gt=0"`
	Stock       int        `json:"stock" binding:"gte=0"`
	CategoryID  *uuid.UUID `json:"category_id"`
	ImageURL    string     `json:"image_url" binding:"omitempty,max=512"`
}

type UpdateProductRequest struct {
	Name        string     `json:"name" binding:"required,max=255"`
	Description string     `json:"description"`
	Price       float64    `json:"price" binding:"required,gt=0"`
	Stock       int        `json:"stock" binding:"gte=0"`
	CategoryID  *uuid.UUID `json:"category_id"`
	ImageURL    string     `json:"image_url" binding:"omitempty,max=512"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,max=255"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required,max=255"`
}

// ---- Contracts ----

// ProductRepository: akses data produk (diimplementasi di layer repository).
type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	// Delete melakukan soft delete agar histori transaksi tetap utuh.
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Product, error)
	FindBySKU(ctx context.Context, sku string) (*Product, error)
	FetchAdminList(ctx context.Context, f AdminListFilter) ([]Product, int, error)
	FetchCatalog(ctx context.Context, f CatalogFilter) ([]Product, int, error)
	CategoryExists(ctx context.Context, id uuid.UUID) (bool, error)
	ListCategories(ctx context.Context) ([]Category, error)

	// DeductStock atomik; mengembalikan apperror INSUFFICIENT_STOCK bila stok kurang.
	DeductStock(ctx context.Context, id uuid.UUID, qty int) error
	RestoreStock(ctx context.Context, id uuid.UUID, qty int) error
	CountLowStock(ctx context.Context, threshold int) (int, error)
}

// ProductUsecase: dipakai oleh handler admin.
type ProductUsecase interface {
	CreateProduct(ctx context.Context, req CreateProductRequest) (*Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	GetAdminProducts(ctx context.Context, f AdminListFilter) ([]Product, int, error)
}

// CatalogUsecase: dipakai oleh handler katalog publik.
type CatalogUsecase interface {
	FetchCatalog(ctx context.Context, f CatalogFilter) ([]Product, int, error)
	GetProductDetail(ctx context.Context, id uuid.UUID) (*Product, error)
	ListCategories(ctx context.Context) ([]Category, error)
}

type CategoryRepository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, category *Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

type CategoryUsecase interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, req CreateCategoryRequest) (*Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, req UpdateCategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// ProductServiceContract adalah satu-satunya pintu modul lain (mis. transaction)
// untuk mengakses data produk. Modul lain TIDAK boleh menyentuh tabel/repository produk.
type ProductServiceContract interface {
	GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error)
	DeductStock(ctx context.Context, id uuid.UUID, qty int) error
	RestoreStock(ctx context.Context, id uuid.UUID, qty int) error
	CountLowStock(ctx context.Context, threshold int) (int, error)
}

// NormalizePage menormalkan page (>=1) dan limit (1..100, default 10).
func NormalizePage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
