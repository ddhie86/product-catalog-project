// Package domain (transaction): entity, DTO, dan kontrak modul Transaction.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	payment "github.com/dedysetiawan/product-catalog-project/internal/modules/payment/domain"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPaid      Status = "PAID"
	StatusExpired   Status = "EXPIRED"
	StatusCancelled Status = "CANCELLED"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusPaid, StatusExpired, StatusCancelled:
		return true
	}
	return false
}

type Transaction struct {
	ID              uuid.UUID         `json:"id"`
	InvoiceNumber   string            `json:"invoice_number"`
	CustomerName    string            `json:"customer_name"`
	CustomerEmail   string            `json:"customer_email"`
	CustomerPhone   string            `json:"customer_phone"`
	ShippingAddress string            `json:"shipping_address"`
	TotalAmount     float64           `json:"total_amount"`
	Status          Status            `json:"status"`
	PaymentURL      string            `json:"payment_url"`
	Items           []TransactionItem `json:"items,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type TransactionItem struct {
	ID              uuid.UUID `json:"id"`
	TransactionID   uuid.UUID `json:"transaction_id"`
	ProductID       uuid.UUID `json:"product_id"`
	Quantity        int       `json:"quantity"`
	PriceAtPurchase float64   `json:"price_at_purchase"`
}

type TransactionSummary struct {
	TotalRevenue          float64 `json:"total_revenue"`
	TotalPaidTransactions int     `json:"total_paid_transactions"`
	PendingCount          int     `json:"pending_count"`
	LowStockAlerts        int     `json:"low_stock_alerts"`
}

type ListFilter struct {
	Status Status // kosong = semua
	Page   int
	Limit  int
}

// ---- Request DTO ----

type CheckoutItem struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
}

type CheckoutRequest struct {
	CustomerName    string         `json:"customer_name" binding:"required,max=255"`
	CustomerEmail   string         `json:"customer_email" binding:"required,email,max=255"`
	CustomerPhone   string         `json:"customer_phone" binding:"required,max=32"`
	ShippingAddress string         `json:"shipping_address" binding:"required"`
	Items           []CheckoutItem `json:"items" binding:"required,min=1,dive"`
}

type CheckoutResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	InvoiceNumber string    `json:"invoice_number"`
	TotalAmount   float64   `json:"total_amount"`
	Status        Status    `json:"status"`
	PaymentURL    string    `json:"payment_url"`
}

// ---- Contracts ----

type TransactionRepository interface {
	// CreateWithItems menyimpan transaksi + item secara atomik (1 DB transaction).
	CreateWithItems(ctx context.Context, t *Transaction) error
	UpdatePaymentURL(ctx context.Context, id uuid.UUID, url string) error
	FindByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	FindByInvoice(ctx context.Context, invoice string) (*Transaction, error)
	// UpdateStatusIf mengubah status hanya bila status saat ini == from (idempoten untuk webhook).
	// Mengembalikan true bila baris benar-benar berubah.
	UpdateStatusIf(ctx context.Context, id uuid.UUID, from, to Status) (bool, error)
	List(ctx context.Context, f ListFilter) ([]Transaction, int, error)
	Summary(ctx context.Context) (*TransactionSummary, error)
}

type CheckoutUsecase interface {
	CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
	HandleWebhook(ctx context.Context, payload payment.WebhookPayload) error
}

type ReportUsecase interface {
	GetTransactionList(ctx context.Context, f ListFilter) ([]Transaction, int, error)
	GetTransactionDetail(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetDashboardSummary(ctx context.Context) (*TransactionSummary, error)
}
