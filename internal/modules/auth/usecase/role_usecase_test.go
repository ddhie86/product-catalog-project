package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/usecase"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type fakeRoleRepository struct {
	roles       map[uuid.UUID]domain.Role
	permissions []domain.Permission
}

func newFakeRoleRepository() *fakeRoleRepository {
	return &fakeRoleRepository{
		roles: make(map[uuid.UUID]domain.Role),
		permissions: []domain.Permission{
			{Key: "products:view", Menu: "products", Action: "view"},
			{Key: "products:edit", Menu: "products", Action: "edit"},
		},
	}
}

func (r *fakeRoleRepository) ListRoles(context.Context) ([]domain.Role, error) {
	roles := make([]domain.Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}
	return roles, nil
}
func (r *fakeRoleRepository) ListPermissions(context.Context) ([]domain.Permission, error) {
	return r.permissions, nil
}
func (r *fakeRoleRepository) GetUserPermissions(context.Context, uuid.UUID) ([]string, error) {
	return []string{}, nil
}
func (r *fakeRoleRepository) HasPermission(context.Context, uuid.UUID, string) (bool, error) {
	return false, nil
}
func (r *fakeRoleRepository) CreateRole(_ context.Context, role *domain.Role) error {
	r.roles[role.ID] = *role
	return nil
}
func (r *fakeRoleRepository) UpdateRole(_ context.Context, role *domain.Role) error {
	if _, ok := r.roles[role.ID]; !ok {
		return apperror.NotFound("role not found")
	}
	r.roles[role.ID] = *role
	return nil
}
func (r *fakeRoleRepository) DeleteRole(_ context.Context, id uuid.UUID) error {
	delete(r.roles, id)
	return nil
}
func TestCreateRole_TrimsNameAndDeduplicatesPermissions(t *testing.T) {
	repo := newFakeRoleRepository()
	uc := usecase.NewRoleUsecase(repo)

	role, err := uc.CreateRole(context.Background(), domain.SaveRoleRequest{
		Name:        "  Catalog Editor  ",
		Permissions: []string{"products:view", "products:edit", "products:view"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "Catalog Editor" {
		t.Errorf("name = %q, want trimmed name", role.Name)
	}
	if len(role.Permissions) != 2 {
		t.Fatalf("got %d permissions, want 2", len(role.Permissions))
	}
}

func TestCreateRole_RejectsUnknownPermission(t *testing.T) {
	uc := usecase.NewRoleUsecase(newFakeRoleRepository())
	_, err := uc.CreateRole(context.Background(), domain.SaveRoleRequest{
		Name: "Invalid", Permissions: []string{"users:root"},
	})
	if apperror.Code(err) != "BAD_REQUEST" {
		t.Fatalf("want BAD_REQUEST, got %v", err)
	}
}

func TestCreateRole_AddsViewPermissionForManagementActions(t *testing.T) {
	uc := usecase.NewRoleUsecase(newFakeRoleRepository())
	role, err := uc.CreateRole(context.Background(), domain.SaveRoleRequest{
		Name: "Product Editor", Permissions: []string{"products:edit"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(role.Permissions) != 2 || role.Permissions[1] != "products:view" {
		t.Fatalf("permissions = %v, want edit and view", role.Permissions)
	}
}
