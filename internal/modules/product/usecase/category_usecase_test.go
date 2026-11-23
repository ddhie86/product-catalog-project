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

type fakeCategoryRepository struct {
	categories map[uuid.UUID]domain.Category
}

func newFakeCategoryRepository() *fakeCategoryRepository {
	return &fakeCategoryRepository{categories: make(map[uuid.UUID]domain.Category)}
}

func (r *fakeCategoryRepository) ListCategories(context.Context) ([]domain.Category, error) {
	categories := make([]domain.Category, 0, len(r.categories))
	for _, category := range r.categories {
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *fakeCategoryRepository) CreateCategory(_ context.Context, category *domain.Category) error {
	for _, existing := range r.categories {
		if existing.Name == category.Name {
			return apperror.Conflict("category name already exists")
		}
	}
	r.categories[category.ID] = *category
	return nil
}

func (r *fakeCategoryRepository) UpdateCategory(_ context.Context, category *domain.Category) error {
	if _, ok := r.categories[category.ID]; !ok {
		return apperror.NotFound("category not found")
	}
	for id, existing := range r.categories {
		if id != category.ID && existing.Name == category.Name {
			return apperror.Conflict("category name already exists")
		}
	}
	r.categories[category.ID] = *category
	return nil
}

func (r *fakeCategoryRepository) DeleteCategory(_ context.Context, id uuid.UUID) error {
	if _, ok := r.categories[id]; !ok {
		return apperror.NotFound("category not found")
	}
	delete(r.categories, id)
	return nil
}

func TestCreateCategory_TrimsName(t *testing.T) {
	uc := usecase.NewCategoryUsecase(newFakeCategoryRepository())
	category, err := uc.CreateCategory(context.Background(), domain.CreateCategoryRequest{Name: "  Minuman  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if category.Name != "Minuman" {
		t.Errorf("name = %q, want %q", category.Name, "Minuman")
	}
}

func TestCreateCategory_RejectsEmptyName(t *testing.T) {
	uc := usecase.NewCategoryUsecase(newFakeCategoryRepository())
	_, err := uc.CreateCategory(context.Background(), domain.CreateCategoryRequest{Name: "  "})
	if apperror.Code(err) != "BAD_REQUEST" {
		t.Fatalf("want BAD_REQUEST, got %v", err)
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	uc := usecase.NewCategoryUsecase(newFakeCategoryRepository())
	_, err := uc.UpdateCategory(context.Background(), uuid.New(), domain.UpdateCategoryRequest{Name: "Minuman"})
	if apperror.Code(err) != "NOT_FOUND" {
		t.Fatalf("want NOT_FOUND, got %v", err)
	}
}
