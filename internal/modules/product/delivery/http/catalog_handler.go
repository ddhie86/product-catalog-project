// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type CatalogHandler struct {
	uc domain.CatalogUsecase
}

func NewCatalogHandler(uc domain.CatalogUsecase) *CatalogHandler { return &CatalogHandler{uc: uc} }

func (h *CatalogHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/products")
	g.GET("", h.List)
	g.GET("/:id", h.Detail)
	rg.GET("/categories", h.Categories)
}

// GET /api/v1/categories — dipakai filter katalog & form produk admin.
func (h *CatalogHandler) Categories(c *gin.Context) {
	items, err := h.uc.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

// GET /api/v1/products?q=&category=&sort=&page=&limit=
func (h *CatalogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))

	f := domain.CatalogFilter{
		Search: c.Query("q"),
		Sort:   c.Query("sort"),
		Page:   page,
		Limit:  limit,
	}
	if raw := c.Query("category"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(c, apperror.BadRequest("invalid category id"))
			return
		}
		f.CategoryID = &id
	}

	items, total, err := h.uc.FetchCatalog(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	page, limit = domain.NormalizePage(page, limit)
	response.Paginated(c, items, response.NewMeta(page, limit, total))
}

func (h *CatalogHandler) Detail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid product id"))
		return
	}
	p, err := h.uc.GetProductDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, p)
}
