// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type RoleHandler struct {
	uc domain.RoleUsecase
}

func NewRoleHandler(uc domain.RoleUsecase) *RoleHandler {
	return &RoleHandler{uc: uc}
}

func (h *RoleHandler) RegisterAdminRoutes(rg *gin.RouterGroup, permissions middleware.PermissionChecker) {
	rg.GET("/permissions", h.CurrentPermissions)
	rg.GET("/roles", middleware.RequireAnyPermission(permissions, "roles:view", "users:assign_roles"), h.List)
	rg.POST("/roles", middleware.RequirePermission(permissions, "roles:create"), h.Create)
	rg.PUT("/roles/:id", middleware.RequirePermission(permissions, "roles:edit"), h.Update)
	rg.DELETE("/roles/:id", middleware.RequirePermission(permissions, "roles:delete"), h.Delete)
}

func (h *RoleHandler) CurrentPermissions(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString(middleware.CtxUserID))
	if err != nil {
		response.Error(c, apperror.Unauthorized("invalid admin identity"))
		return
	}
	permissions, err := h.uc.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, permissions)
}

func (h *RoleHandler) List(c *gin.Context) {
	result, err := h.uc.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req domain.SaveRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	role, err := h.uc.CreateRole(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid role id"))
		return
	}
	var req domain.SaveRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.BadRequest(err.Error()))
		return
	}
	role, err := h.uc.UpdateRole(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid role id"))
		return
	}
	if err := h.uc.DeleteRole(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
