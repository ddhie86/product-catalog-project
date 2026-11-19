// Package domain (notification): event & kontrak pengiriman email.
package domain

import "context"

type EmailKind string

const (
	EmailOrderConfirmation EmailKind = "ORDER_CONFIRMATION"
	EmailPaymentReceipt    EmailKind = "PAYMENT_RECEIPT"
)

// OrderEmailEvent dikirim ke channel & diproses worker secara asynchronous.
type OrderEmailEvent struct {
	Kind          EmailKind
	InvoiceNumber string
	CustomerName  string
	CustomerEmail string
	TotalAmount   float64
	PaymentURL    string
}

// EmailSender diimplementasi oleh adapter (SMTP) di internal/platform/mailer.
type EmailSender interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}
