// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
// Package usecase (transaction): checkout, webhook, dan reporting.
package usecase

import (
	"context"

	"github.com/google/uuid"

	product "github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type reportUsecase struct {
	repo              domain.TransactionRepository
	products          product.ProductServiceContract
	lowStockThreshold int
}

func NewReportUsecase(repo domain.TransactionRepository, products product.ProductServiceContract, lowStockThreshold int) domain.ReportUsecase {
	return &reportUsecase{repo: repo, products: products, lowStockThreshold: lowStockThreshold}
}

func (u *reportUsecase) GetTransactionList(ctx context.Context, f domain.ListFilter) ([]domain.Transaction, int, error) {
	if f.Status != "" && !f.Status.Valid() {
		return nil, 0, apperror.BadRequest("invalid status filter")
	}
	f.Page, f.Limit = product.NormalizePage(f.Page, f.Limit)
	return u.repo.List(ctx, f)
}

func (u *reportUsecase) GetTransactionDetail(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	return u.repo.FindByID(ctx, id)
}

// GetDashboardSummary: agregasi pendapatan + jumlah produk stok menipis (< threshold).
func (u *reportUsecase) GetDashboardSummary(ctx context.Context) (*domain.TransactionSummary, error) {
	s, err := u.repo.Summary(ctx)
	if err != nil {
		return nil, err
	}
	low, err := u.products.CountLowStock(ctx, u.lowStockThreshold)
	if err != nil {
		return nil, err
	}
	s.LowStockAlerts = low
	return s, nil
}
