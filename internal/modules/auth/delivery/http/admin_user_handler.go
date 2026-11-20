package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type AdminUserHandler struct {
	uc domain.AdminUserUsecase
}

func NewAdminUserHandler(uc domain.AdminUserUsecase) *AdminUserHandler {
	return &AdminUserHandler{uc: uc}
}

// RegisterAdminRoutes must be called with a JWT-protected route group.
func (h *AdminUserHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/users", h.List)
	rg.POST("/users", h.Create)
	rg.PUT("/users/:id", h.Update)
	rg.DELETE("/users/:id", h.Delete)
}

func (h *AdminUserHandler) List(c *gin.Context) {
	users, err := h.uc.ListAdmins(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, users)
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	var req domain.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	user, err := h.uc.CreateAdmin(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, user)
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid admin id"))
		return
	}
	var req domain.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	user, err := h.uc.UpdateAdmin(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid admin id"))
		return
	}
	actorID, err := uuid.Parse(c.GetString(middleware.CtxUserID))
	if err != nil {
		response.Error(c, apperror.Unauthorized("invalid admin identity"))
		return
	}
	if err := h.uc.DeleteAdmin(c.Request.Context(), id, actorID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
