package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

// productService mengimplementasi ProductServiceContract untuk modul lain.
type productService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) domain.ProductServiceContract {
	return &productService{repo: repo}
}

func (s *productService) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *productService) DeductStock(ctx context.Context, id uuid.UUID, qty int) error {
	if qty <= 0 {
		return apperror.BadRequest("quantity must be greater than zero")
	}
	return s.repo.DeductStock(ctx, id, qty)
}

func (s *productService) RestoreStock(ctx context.Context, id uuid.UUID, qty int) error {
	return s.repo.RestoreStock(ctx, id, qty)
}

func (s *productService) CountLowStock(ctx context.Context, threshold int) (int, error) {
	return s.repo.CountLowStock(ctx, threshold)
}
