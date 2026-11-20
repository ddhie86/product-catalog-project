package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type CategoryHandler struct {
	uc domain.CategoryUsecase
}

func NewCategoryHandler(uc domain.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

// RegisterAdminRoutes must be called with a JWT-protected route group.
func (h *CategoryHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	categories := rg.Group("/categories")
	categories.GET("", h.List)
	categories.POST("", h.Create)
	categories.PUT("/:id", h.Update)
	categories.DELETE("/:id", h.Delete)
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.uc.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, categories)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req domain.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	category, err := h.uc.CreateCategory(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid category id"))
		return
	}
	var req domain.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	category, err := h.uc.UpdateCategory(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid category id"))
		return
	}
	if err := h.uc.DeleteCategory(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
