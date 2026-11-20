package usecase

import (
	"context"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type adminUserUsecase struct {
	repo domain.AdminUserRepository
}

func NewAdminUserUsecase(repo domain.AdminUserRepository) domain.AdminUserUsecase {
	return &adminUserUsecase{repo: repo}
}

func (u *adminUserUsecase) ListAdmins(ctx context.Context) ([]domain.User, error) {
	return u.repo.ListAdmins(ctx)
}

func (u *adminUserUsecase) CreateAdmin(ctx context.Context, req domain.CreateAdminRequest) (*domain.User, error) {
	email, err := normalizeAdminEmail(req.Email)
	if err != nil {
		return nil, err
	}
	hash, err := hashAdminPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{ID: uuid.New(), Email: email, PasswordHash: hash}
	if err := u.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *adminUserUsecase) UpdateAdmin(ctx context.Context, id uuid.UUID, req domain.UpdateAdminRequest) (*domain.User, error) {
	email, err := normalizeAdminEmail(req.Email)
	if err != nil {
		return nil, err
	}
	var passwordHash string
	if req.Password != "" {
		passwordHash, err = hashAdminPassword(req.Password)
		if err != nil {
			return nil, err
		}
	}
	return u.repo.UpdateAdmin(ctx, &domain.User{ID: id, Email: email, PasswordHash: passwordHash})
}

func (u *adminUserUsecase) DeleteAdmin(ctx context.Context, id, actorID uuid.UUID) error {
	if id == actorID {
		return apperror.Forbidden("cannot delete your own admin account")
	}
	return u.repo.DeleteAdmin(ctx, id)
}

func normalizeAdminEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 255 {
		return "", apperror.BadRequest("a valid email address is required")
	}
	return email, nil
}

func hashAdminPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" || len(password) < 12 || len(password) > 72 {
		return "", apperror.BadRequest("password must be between 12 and 72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.Internal(err)
	}
	return string(hash), nil
}
