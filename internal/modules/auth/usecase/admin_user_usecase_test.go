package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/usecase"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type fakeAdminUserRepository struct {
	users map[uuid.UUID]*domain.User
}

func newFakeAdminUserRepository(users ...*domain.User) *fakeAdminUserRepository {
	repo := &fakeAdminUserRepository{users: make(map[uuid.UUID]*domain.User)}
	for _, user := range users {
		repo.users[user.ID] = user
	}
	return repo
}

func (r *fakeAdminUserRepository) Create(_ context.Context, user *domain.User) error {
	for _, existing := range r.users {
		if existing.Email == user.Email {
			return apperror.Conflict("email already registered")
		}
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	r.users[user.ID] = user
	return nil
}

func (r *fakeAdminUserRepository) CreateAdmin(ctx context.Context, user *domain.User, roleIDs []uuid.UUID) error {
	if err := r.Create(ctx, user); err != nil {
		return err
	}
	user.Roles = make([]domain.Role, len(roleIDs))
	for i, id := range roleIDs {
		user.Roles[i].ID = id
	}
	return nil
}

func (r *fakeAdminUserRepository) ListAdmins(context.Context) ([]domain.User, error) {
	users := make([]domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}
	return users, nil
}

func (r *fakeAdminUserRepository) UpdateAdmin(_ context.Context, user *domain.User, roleIDs []uuid.UUID) (*domain.User, error) {
	existing, ok := r.users[user.ID]
	if !ok {
		return nil, apperror.NotFound("admin not found")
	}
	for id, other := range r.users {
		if id != user.ID && other.Email == user.Email {
			return nil, apperror.Conflict("email already registered")
		}
	}
	existing.Email = user.Email
	if user.PasswordHash != "" {
		existing.PasswordHash = user.PasswordHash
	}
	if roleIDs != nil {
		existing.Roles = make([]domain.Role, len(roleIDs))
		for i, id := range roleIDs {
			existing.Roles[i].ID = id
		}
	}
	existing.UpdatedAt = time.Now()
	return existing, nil
}

func (r *fakeAdminUserRepository) DeleteAdmin(_ context.Context, id uuid.UUID) error {
	if _, ok := r.users[id]; !ok {
		return apperror.NotFound("admin not found")
	}
	if len(r.users) <= 1 {
		return apperror.Conflict("cannot delete the last admin")
	}
	delete(r.users, id)
	return nil
}

func TestCreateAdmin_NormalizesEmailAndHashesPassword(t *testing.T) {
	repo := newFakeAdminUserRepository()
	uc := usecase.NewAdminUserUsecase(repo)

	user, err := uc.CreateAdmin(context.Background(), domain.CreateAdminRequest{
		Email: "  ADMIN@example.com ", Password: "strong-admin-password", RoleIDs: []uuid.UUID{uuid.New()},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "admin@example.com" {
		t.Errorf("email = %q, want normalized email", user.Email)
	}
	if user.PasswordHash == "strong-admin-password" {
		t.Fatal("password was stored without hashing")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("strong-admin-password")); err != nil {
		t.Fatalf("stored hash does not match password: %v", err)
	}
}

func TestUpdateAdmin_LeavesPasswordUnchangedWhenOmitted(t *testing.T) {
	id := uuid.New()
	repo := newFakeAdminUserRepository(&domain.User{
		ID: id, Email: "admin@example.com", PasswordHash: "existing-hash",
	})
	uc := usecase.NewAdminUserUsecase(repo)

	user, err := uc.UpdateAdmin(context.Background(), id, domain.UpdateAdminRequest{
		Email: "updated@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "updated@example.com" || user.PasswordHash != "existing-hash" {
		t.Fatalf("unexpected updated admin: %+v", user)
	}
}

func TestDeleteAdmin_RejectsSelfDeletion(t *testing.T) {
	id := uuid.New()
	repo := newFakeAdminUserRepository(&domain.User{ID: id, Email: "admin@example.com"})
	uc := usecase.NewAdminUserUsecase(repo)

	err := uc.DeleteAdmin(context.Background(), id, id)
	if apperror.Code(err) != "FORBIDDEN" {
		t.Fatalf("want FORBIDDEN, got %v", err)
	}
	if len(repo.users) != 1 {
		t.Fatal("admin account should remain")
	}
}

func TestCreateAdmin_RejectsShortPassword(t *testing.T) {
	uc := usecase.NewAdminUserUsecase(newFakeAdminUserRepository())
	_, err := uc.CreateAdmin(context.Background(), domain.CreateAdminRequest{
		Email: "admin@example.com", Password: "short",
	})
	if apperror.Code(err) != "BAD_REQUEST" {
		t.Fatalf("want BAD_REQUEST, got %v", err)
	}
}
