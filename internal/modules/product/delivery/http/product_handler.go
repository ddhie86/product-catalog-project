// Package http (product): HTTP handlers & routes modul Product.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type ProductHandler struct {
	uc domain.ProductUsecase
}

func NewProductHandler(uc domain.ProductUsecase) *ProductHandler { return &ProductHandler{uc: uc} }

// RegisterAdminRoutes: rg harus sudah dipasangi middleware JWT.
func (h *ProductHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/products")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	items, total, err := h.uc.GetAdminProducts(c.Request.Context(), domain.AdminListFilter{
		Search: c.Query("q"), Page: page, Limit: limit,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	page, limit = domain.NormalizePage(page, limit)
	response.Paginated(c, items, response.NewMeta(page, limit, total))
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req domain.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	p, err := h.uc.CreateProduct(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, p)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid product id"))
		return
	}
	var req domain.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	p, err := h.uc.UpdateProduct(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid product id"))
		return
	}
	if err := h.uc.DeleteProduct(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
