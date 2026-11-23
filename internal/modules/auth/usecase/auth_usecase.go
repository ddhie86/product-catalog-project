// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package usecase

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/token"
)

type authUsecase struct {
	users  domain.UserRepository
	tokens *token.Manager
}

func NewAuthUsecase(users domain.UserRepository, tokens *token.Manager) domain.AuthUsecase {
	return &authUsecase{users: users, tokens: tokens}
}

func (u *authUsecase) Login(ctx context.Context, req domain.LoginRequest) (*domain.LoginResponse, error) {
	// Pesan error sengaja generik agar tidak membocorkan keberadaan email.
	invalid := apperror.Unauthorized("invalid email or password")

	user, err := u.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if apperror.Code(err) == "NOT_FOUND" {
			return nil, invalid
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, invalid
	}
	tok, exp, err := u.tokens.Generate(user.ID.String(), user.Email)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return &domain.LoginResponse{AccessToken: tok, ExpiresAt: exp}, nil
}
