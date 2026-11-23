// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type AuthHandler struct{ uc domain.AuthUsecase }

func NewAuthHandler(uc domain.AuthUsecase) *AuthHandler { return &AuthHandler{uc: uc} }

// RegisterRoutes: rg = /api/v1/admin (endpoint login TIDAK diproteksi JWT).
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	res, err := h.uc.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}
