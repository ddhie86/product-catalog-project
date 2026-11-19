// Package domain (payment): kontrak integrasi payment gateway.
// Implementasi konkret (Xendit/Midtrans) berada di internal/platform/payment.
package domain

import "context"

type CreateInvoiceRequest struct {
	ExternalID  string // = invoice_number transaksi
	Amount      float64
	PayerEmail  string
	Description string
	SuccessURL  string
}

type Invoice struct {
	GatewayID  string
	PaymentURL string
}

// WebhookPayload adalah bentuk ternormalisasi dari callback gateway.
type WebhookPayload struct {
	ExternalID string `json:"external_id"` // invoice_number
	Status     string `json:"status"`      // PAID | EXPIRED | ...
}

type Gateway interface {
	CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error)
	// VerifyCallbackToken memvalidasi token header webhook.
	VerifyCallbackToken(token string) bool
}
