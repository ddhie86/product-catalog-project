// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type categoryUsecase struct {
	repo domain.CategoryRepository
}

func NewCategoryUsecase(repo domain.CategoryRepository) domain.CategoryUsecase {
	return &categoryUsecase{repo: repo}
}

func (u *categoryUsecase) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return u.repo.ListCategories(ctx)
}

func (u *categoryUsecase) CreateCategory(ctx context.Context, req domain.CreateCategoryRequest) (*domain.Category, error) {
	name, err := normalizeCategoryName(req.Name)
	if err != nil {
		return nil, err
	}
	category := &domain.Category{ID: uuid.New(), Name: name}
	if err := u.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (u *categoryUsecase) UpdateCategory(ctx context.Context, id uuid.UUID, req domain.UpdateCategoryRequest) (*domain.Category, error) {
	name, err := normalizeCategoryName(req.Name)
	if err != nil {
		return nil, err
	}
	category := &domain.Category{ID: id, Name: name}
	if err := u.repo.UpdateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (u *categoryUsecase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteCategory(ctx, id)
}

func normalizeCategoryName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || utf8.RuneCountInString(name) > 255 {
		return "", apperror.BadRequest("category name must be between 1 and 255 characters")
	}
	return name, nil
}
