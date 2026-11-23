// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"github.com/gin-gonic/gin"

	payment "github.com/dedysetiawan/product-catalog-project/internal/modules/payment/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type CheckoutHandler struct {
	uc      domain.CheckoutUsecase
	gateway payment.Gateway
}

func NewCheckoutHandler(uc domain.CheckoutUsecase, gateway payment.Gateway) *CheckoutHandler {
	return &CheckoutHandler{uc: uc, gateway: gateway}
}

// RegisterRoutes: rg = /api/v1 (public).
func (h *CheckoutHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/checkout", h.Checkout)
	rg.POST("/payments/webhook", h.Webhook)
}

func (h *CheckoutHandler) Checkout(c *gin.Context) {
	var req domain.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	res, err := h.uc.CreateCheckout(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, res)
}

// Webhook: diverifikasi lewat header X-Callback-Token (Xendit).
// Proses berat (email) dilakukan async oleh worker, jadi response cepat.
func (h *CheckoutHandler) Webhook(c *gin.Context) {
	if !h.gateway.VerifyCallbackToken(c.GetHeader("X-Callback-Token")) {
		response.Error(c, apperror.Unauthorized("invalid callback token"))
		return
	}
	var payload payment.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	if err := h.uc.HandleWebhook(c.Request.Context(), payload); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"received": true})
}
