// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
// Package payment: adapter payment gateway (Xendit).
package payment

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/payment/domain"
)

const xenditInvoiceURL = "https://api.xendit.co/v2/invoices"

type Xendit struct {
	apiKey        string
	callbackToken string
	client        *http.Client
}

func NewXendit(apiKey, callbackToken string) *Xendit {
	return &Xendit{
		apiKey:        apiKey,
		callbackToken: callbackToken,
		client:        &http.Client{Timeout: 15 * time.Second},
	}
}

var _ domain.Gateway = (*Xendit)(nil)

func (x *Xendit) CreateInvoice(ctx context.Context, req domain.CreateInvoiceRequest) (*domain.Invoice, error) {
	payload, _ := json.Marshal(map[string]any{
		"external_id":          req.ExternalID,
		"amount":               req.Amount,
		"payer_email":          req.PayerEmail,
		"description":          req.Description,
		"success_redirect_url": req.SuccessURL,
		"currency":             "IDR",
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, xenditInvoiceURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.SetBasicAuth(x.apiKey, "")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := x.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("xendit request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("xendit returned status %d", resp.StatusCode)
	}

	var out struct {
		ID         string `json:"id"`
		InvoiceURL string `json:"invoice_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode xendit response: %w", err)
	}
	return &domain.Invoice{GatewayID: out.ID, PaymentURL: out.InvoiceURL}, nil
}

func (x *Xendit) VerifyCallbackToken(token string) bool {
	if x.callbackToken == "" {
		return false // fail closed bila token belum dikonfigurasi
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(x.callbackToken)) == 1
}
