// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
// Package usecase (product): business logic CMS produk (admin).
package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type productUsecase struct {
	repo domain.ProductRepository
}

func NewProductUsecase(repo domain.ProductRepository) domain.ProductUsecase {
	return &productUsecase{repo: repo}
}

func (u *productUsecase) CreateProduct(ctx context.Context, req domain.CreateProductRequest) (*domain.Product, error) {
	sku := strings.ToUpper(strings.TrimSpace(req.SKU))

	// Validasi keunikan SKU.
	if _, err := u.repo.FindBySKU(ctx, sku); err == nil {
		return nil, apperror.Conflict("sku already exists")
	} else if apperror.Code(err) != "NOT_FOUND" {
		return nil, err
	}
	if err := u.ensureCategory(ctx, req.CategoryID); err != nil {
		return nil, err
	}

	p := &domain.Product{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(req.Name),
		SKU:         sku,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
	}
	if err := u.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (u *productUsecase) UpdateProduct(ctx context.Context, id uuid.UUID, req domain.UpdateProductRequest) (*domain.Product, error) {
	p, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.ensureCategory(ctx, req.CategoryID); err != nil {
		return nil, err
	}
	p.Name = strings.TrimSpace(req.Name)
	p.Description = req.Description
	p.Price = req.Price
	p.Stock = req.Stock
	p.CategoryID = req.CategoryID
	p.ImageURL = req.ImageURL

	if err := u.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeleteProduct melakukan soft delete agar data historis transaksi tetap utuh.
func (u *productUsecase) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *productUsecase) GetAdminProducts(ctx context.Context, f domain.AdminListFilter) ([]domain.Product, int, error) {
	f.Page, f.Limit = domain.NormalizePage(f.Page, f.Limit)
	return u.repo.FetchAdminList(ctx, f)
}

func (u *productUsecase) ensureCategory(ctx context.Context, id *uuid.UUID) error {
	if id == nil {
		return nil
	}
	ok, err := u.repo.CategoryExists(ctx, *id)
	if err != nil {
		return err
	}
	if !ok {
		return apperror.BadRequest("category not found")
	}
	return nil
}
