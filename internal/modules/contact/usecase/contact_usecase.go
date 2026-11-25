// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package usecase

import (
	"context"
	"fmt"
	"html"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/contact/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type contactUsecase struct {
	repo         domain.ContactRepository
	emailSender  domain.EmailSender
	supportEmail string
}

func NewContactUsecase(repo domain.ContactRepository, emailSender domain.EmailSender, supportEmail string) domain.ContactUsecase {
	return &contactUsecase{repo: repo, emailSender: emailSender, supportEmail: supportEmail}
}

func (u *contactUsecase) CreateMessage(ctx context.Context, req domain.CreateContactMessageRequest) error {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	messageText := strings.TrimSpace(req.Message)
	parsedEmail, err := mail.ParseAddress(email)
	if name == "" || utf8.RuneCountInString(name) > 255 {
		return apperror.BadRequest("name must be between 1 and 255 characters")
	}
	if err != nil || parsedEmail.Address != email || utf8.RuneCountInString(email) > 255 {
		return apperror.BadRequest("email must be valid")
	}
	if req.Type != domain.MessageQuestion && req.Type != domain.MessageComplaint {
		return apperror.BadRequest("type must be QUESTION or COMPLAINT")
	}
	if messageText == "" || utf8.RuneCountInString(messageText) > 5000 {
		return apperror.BadRequest("message must be between 1 and 5000 characters")
	}

	contactMessage := &domain.ContactMessage{
		ID:      uuid.New(),
		Name:    name,
		Email:   email,
		Type:    req.Type,
		Message: messageText,
	}
	if err := u.repo.Create(ctx, contactMessage); err != nil {
		return err
	}

	subject := fmt.Sprintf("[Contact Us] %s", contactMessage.Type)
	body := fmt.Sprintf(
		"<h2>Pesan Contact Us</h2><p><strong>Nama:</strong> %s</p><p><strong>Email:</strong> %s</p><p><strong>Tipe:</strong> %s</p><p><strong>Pesan:</strong></p><p>%s</p>",
		html.EscapeString(contactMessage.Name),
		html.EscapeString(contactMessage.Email),
		html.EscapeString(string(contactMessage.Type)),
		strings.ReplaceAll(html.EscapeString(contactMessage.Message), "\n", "<br>"),
	)
	if err := u.emailSender.Send(ctx, u.supportEmail, subject, body); err != nil {
		return fmt.Errorf("send contact message email: %w", err)
	}
	return nil
}

func (u *contactUsecase) ListMessages(ctx context.Context, page, limit int) ([]domain.ContactMessage, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return u.repo.List(ctx, page, limit)
}

func (u *contactUsecase) Reply(ctx context.Context, id, actorID uuid.UUID, rawReply string) (*domain.ContactMessage, error) {
	reply := strings.TrimSpace(rawReply)
	if reply == "" || utf8.RuneCountInString(reply) > 5000 {
		return nil, apperror.BadRequest("reply must be between 1 and 5000 characters")
	}
	message, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	body := fmt.Sprintf(
		"<p>%s</p><hr><p><strong>Pesan Anda:</strong></p><p>%s</p>",
		strings.ReplaceAll(html.EscapeString(reply), "\n", "<br>"),
		strings.ReplaceAll(html.EscapeString(message.Message), "\n", "<br>"),
	)
	if err := u.emailSender.Send(ctx, message.Email, "Re: Pesan Anda untuk Product Catalog", body); err != nil {
		return nil, fmt.Errorf("send contact reply email: %w", err)
	}
	message.ReplyMessage = reply
	message.RepliedBy = &actorID
	if err := u.repo.SaveReply(ctx, message); err != nil {
		return nil, err
	}
	return message, nil
}
