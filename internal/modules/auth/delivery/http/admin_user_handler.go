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
	uc          domain.AdminUserUsecase
	permissions middleware.PermissionChecker
}

func NewAdminUserHandler(uc domain.AdminUserUsecase, permissions middleware.PermissionChecker) *AdminUserHandler {
	return &AdminUserHandler{uc: uc, permissions: permissions}
}

// RegisterAdminRoutes must be called with a JWT-protected route group.
func (h *AdminUserHandler) RegisterAdminRoutes(rg *gin.RouterGroup, permissions middleware.PermissionChecker) {
	rg.GET("/users", middleware.RequirePermission(permissions, "users:view"), h.List)
	rg.POST("/users", middleware.RequirePermission(permissions, "users:create"), h.Create)
	rg.PUT("/users/:id", middleware.RequirePermission(permissions, "users:edit"), h.Update)
	rg.DELETE("/users/:id", middleware.RequirePermission(permissions, "users:delete"), h.Delete)
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
	if req.RoleIDs != nil && !h.requireRoleAssignmentPermission(c) {
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
	if req.RoleIDs != nil && !h.requireRoleAssignmentPermission(c) {
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

func (h *AdminUserHandler) requireRoleAssignmentPermission(c *gin.Context) bool {
	userID, err := uuid.Parse(c.GetString(middleware.CtxUserID))
	if err != nil {
		response.Error(c, apperror.Unauthorized("invalid admin identity"))
		return false
	}
	allowed, err := h.permissions.HasPermission(c.Request.Context(), userID, "users:assign_roles")
	if err != nil {
		response.Error(c, apperror.Internal(err))
		return false
	}
	if !allowed {
		response.Error(c, apperror.Forbidden("assigning admin roles requires users:assign_roles"))
		return false
	}
	return true
}
