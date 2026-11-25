// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/contact/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type ContactHandler struct {
	uc domain.ContactUsecase
}

func NewContactHandler(uc domain.ContactUsecase) *ContactHandler {
	return &ContactHandler{uc: uc}
}

func (h *ContactHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/contact-us", h.Create)
}

func (h *ContactHandler) RegisterAdminRoutes(rg *gin.RouterGroup, permissions middleware.PermissionChecker) {
	messages := rg.Group("/contact-messages")
	messages.GET("", middleware.RequirePermission(permissions, "contact_messages:view"), h.List)
	messages.POST("/:id/reply", middleware.RequirePermission(permissions, "contact_messages:reply"), h.Reply)
}

func (h *ContactHandler) Create(c *gin.Context) {
	var req domain.CreateContactMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	if err := h.uc.CreateMessage(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, gin.H{"message": "Pesan berhasil dikirim."})
}

func (h *ContactHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	items, total, err := h.uc.ListMessages(c.Request.Context(), page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	response.Paginated(c, items, response.NewMeta(page, limit, total))
}

func (h *ContactHandler) Reply(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid contact message id"))
		return
	}
	actorID, err := uuid.Parse(c.GetString(middleware.CtxUserID))
	if err != nil {
		response.Error(c, apperror.Unauthorized("invalid admin identity"))
		return
	}
	var req domain.ReplyContactMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	message, err := h.uc.Reply(c.Request.Context(), id, actorID, req.Message)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, message)
}
