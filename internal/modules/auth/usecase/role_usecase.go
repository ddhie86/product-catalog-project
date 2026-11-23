// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type roleUsecase struct {
	repo domain.RoleRepository
}

func NewRoleUsecase(repo domain.RoleRepository) domain.RoleUsecase {
	return &roleUsecase{repo: repo}
}

func (u *roleUsecase) ListRoles(ctx context.Context) (*domain.RoleListResponse, error) {
	roles, err := u.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	permissions, err := u.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.RoleListResponse{Roles: roles, Permissions: permissions}, nil
}

func (u *roleUsecase) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return u.repo.GetUserPermissions(ctx, userID)
}

func (u *roleUsecase) CreateRole(ctx context.Context, req domain.SaveRoleRequest) (*domain.Role, error) {
	role, err := u.newRole(ctx, req)
	if err != nil {
		return nil, err
	}
	role.ID = uuid.New()
	if err := u.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (u *roleUsecase) UpdateRole(ctx context.Context, id uuid.UUID, req domain.SaveRoleRequest) (*domain.Role, error) {
	role, err := u.newRole(ctx, req)
	if err != nil {
		return nil, err
	}
	role.ID = id
	if err := u.repo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (u *roleUsecase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteRole(ctx, id)
}

func (u *roleUsecase) newRole(ctx context.Context, req domain.SaveRoleRequest) (*domain.Role, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || utf8.RuneCountInString(name) > 64 {
		return nil, apperror.BadRequest("role name must be between 1 and 64 characters")
	}
	permissions, err := u.normalizePermissions(ctx, req.Permissions)
	if err != nil {
		return nil, err
	}
	return &domain.Role{Name: name, Permissions: permissions}, nil
}

func (u *roleUsecase) normalizePermissions(ctx context.Context, keys []string) ([]string, error) {
	definitions, err := u.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	valid := make(map[string]domain.Permission, len(definitions))
	for _, permission := range definitions {
		valid[permission.Key] = permission
	}

	unique := make(map[string]struct{}, len(keys))
	permissions := make([]string, 0, len(keys))
	for _, key := range keys {
		permission, ok := valid[key]
		if !ok {
			return nil, apperror.BadRequest("unknown permission: " + key)
		}
		if _, ok := unique[key]; !ok {
			unique[key] = struct{}{}
			permissions = append(permissions, key)
		}
		if permission.Action != "view" {
			viewKey := permission.Menu + ":view"
			if _, hasView := valid[viewKey]; hasView {
				if _, alreadyAdded := unique[viewKey]; !alreadyAdded {
					unique[viewKey] = struct{}{}
					permissions = append(permissions, viewKey)
				}
			}
		}
	}
	return permissions, nil
}
