// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/usecase"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

// fakeRepo: in-memory ProductRepository untuk unit test.
type fakeRepo struct {
	bySKU      map[string]*domain.Product
	categories map[uuid.UUID]bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{bySKU: map[string]*domain.Product{}, categories: map[uuid.UUID]bool{}}
}

func (f *fakeRepo) Create(_ context.Context, p *domain.Product) error { f.bySKU[p.SKU] = p; return nil }
func (f *fakeRepo) Update(context.Context, *domain.Product) error      { return nil }
func (f *fakeRepo) Delete(context.Context, uuid.UUID) error            { return nil }
func (f *fakeRepo) FindByID(context.Context, uuid.UUID) (*domain.Product, error) {
	return nil, apperror.NotFound("product not found")
}
func (f *fakeRepo) FindBySKU(_ context.Context, sku string) (*domain.Product, error) {
	if p, ok := f.bySKU[sku]; ok {
		return p, nil
	}
	return nil, apperror.NotFound("product not found")
}
func (f *fakeRepo) FetchAdminList(context.Context, domain.AdminListFilter) ([]domain.Product, int, error) {
	return nil, 0, nil
}
func (f *fakeRepo) FetchCatalog(context.Context, domain.CatalogFilter) ([]domain.Product, int, error) {
	return nil, 0, nil
}
func (f *fakeRepo) CategoryExists(_ context.Context, id uuid.UUID) (bool, error) {
	return f.categories[id], nil
}
func (f *fakeRepo) ListCategories(context.Context) ([]domain.Category, error) { return nil, nil }
func (f *fakeRepo) DeductStock(context.Context, uuid.UUID, int) error  { return nil }
func (f *fakeRepo) RestoreStock(context.Context, uuid.UUID, int) error { return nil }
func (f *fakeRepo) CountLowStock(context.Context, int) (int, error)    { return 0, nil }

func TestCreateProduct_Success(t *testing.T) {
	uc := usecase.NewProductUsecase(newFakeRepo())
	p, err := uc.CreateProduct(context.Background(), domain.CreateProductRequest{
		Name: "Kopi", SKU: " kopi-001 ", Price: 25000, Stock: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.SKU != "KOPI-001" {
		t.Errorf("sku should be normalized, got %q", p.SKU)
	}
}

func TestCreateProduct_DuplicateSKU(t *testing.T) {
	uc := usecase.NewProductUsecase(newFakeRepo())
	req := domain.CreateProductRequest{Name: "Kopi", SKU: "KOPI-001", Price: 25000}
	if _, err := uc.CreateProduct(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	_, err := uc.CreateProduct(context.Background(), req)
	if apperror.Code(err) != "CONFLICT" {
		t.Fatalf("want CONFLICT, got %v", err)
	}
}

func TestCreateProduct_UnknownCategory(t *testing.T) {
	uc := usecase.NewProductUsecase(newFakeRepo())
	cat := uuid.New()
	_, err := uc.CreateProduct(context.Background(), domain.CreateProductRequest{
		Name: "Kopi", SKU: "K-1", Price: 1000, CategoryID: &cat,
	})
	if apperror.Code(err) != "BAD_REQUEST" {
		t.Fatalf("want BAD_REQUEST, got %v", err)
	}
}
