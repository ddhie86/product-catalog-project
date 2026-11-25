// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	activitydomain "github.com/dedysetiawan/product-catalog-project/internal/modules/activity/domain"
	productdomain "github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type ActivityHandler struct {
	repo activitydomain.Repository
}

func NewActivityHandler(repo activitydomain.Repository) *ActivityHandler {
	return &ActivityHandler{repo: repo}
}

func (h *ActivityHandler) RegisterAdminRoutes(rg *gin.RouterGroup, permissions middleware.PermissionChecker) {
	rg.GET("/activity-logs", middleware.RequirePermission(permissions, "activity_logs:view"), h.List)
}

func (h *ActivityHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, limit = productdomain.NormalizePage(page, limit)

	items, total, err := h.repo.List(c.Request.Context(), activitydomain.ListFilter{Page: page, Limit: limit})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, items, response.NewMeta(page, limit, total))
}
