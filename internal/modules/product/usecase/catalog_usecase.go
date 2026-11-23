// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
)

type catalogUsecase struct {
	repo domain.ProductRepository
}

func NewCatalogUsecase(repo domain.ProductRepository) domain.CatalogUsecase {
	return &catalogUsecase{repo: repo}
}

// FetchCatalog mengambil produk aktif dengan search, filter kategori, sorting & pagination.
func (u *catalogUsecase) FetchCatalog(ctx context.Context, f domain.CatalogFilter) ([]domain.Product, int, error) {
	f.Search = strings.TrimSpace(f.Search)
	f.Page, f.Limit = domain.NormalizePage(f.Page, f.Limit)
	return u.repo.FetchCatalog(ctx, f)
}

func (u *catalogUsecase) GetProductDetail(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *catalogUsecase) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return u.repo.ListCategories(ctx)
}
