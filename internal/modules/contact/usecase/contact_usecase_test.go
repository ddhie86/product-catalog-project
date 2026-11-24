// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/contact/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/contact/usecase"
	"github.com/google/uuid"
)

type fakeContactRepository struct {
	message *domain.ContactMessage
}

func (r *fakeContactRepository) Create(_ context.Context, message *domain.ContactMessage) error {
	copy := *message
	r.message = &copy
	return nil
}

func (r *fakeContactRepository) List(context.Context, int, int) ([]domain.ContactMessage, int, error) {
	return nil, 0, nil
}

func (r *fakeContactRepository) FindByID(_ context.Context, id uuid.UUID) (*domain.ContactMessage, error) {
	if r.message == nil || r.message.ID != id {
		return nil, errors.New("not found")
	}
	copy := *r.message
	return &copy, nil
}

func (r *fakeContactRepository) SaveReply(_ context.Context, message *domain.ContactMessage) error {
	copy := *message
	r.message = &copy
	return nil
}

type fakeEmailSender struct {
	to, subject, body string
	err               error
}

func (s *fakeEmailSender) Send(_ context.Context, to, subject, body string) error {
	s.to, s.subject, s.body = to, subject, body
	return s.err
}

func TestCreateMessagePersistsAndForwardsEscapedMessage(t *testing.T) {
	repo := &fakeContactRepository{}
	sender := &fakeEmailSender{}
	uc := usecase.NewContactUsecase(repo, sender, "support@example.com")

	err := uc.CreateMessage(context.Background(), domain.CreateContactMessageRequest{
		Name:    "  Ada Lovelace  ",
		Email:   " ada@example.com ",
		Type:    domain.MessageQuestion,
		Message: "<script>alert('hi')</script>\nNeed help",
	})
	if err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}
	if repo.message == nil || repo.message.Name != "Ada Lovelace" || repo.message.Email != "ada@example.com" {
		t.Fatalf("saved message = %#v", repo.message)
	}
	if sender.to != "support@example.com" || sender.subject != "[Contact Us] QUESTION" {
		t.Fatalf("email destination/subject = %q / %q", sender.to, sender.subject)
	}
	if !strings.Contains(sender.body, "&lt;script&gt;") || !strings.Contains(sender.body, "Need help") {
		t.Fatalf("email body was not escaped/formatted: %s", sender.body)
	}
}

func TestCreateMessageRejectsInvalidTypeBeforeSaving(t *testing.T) {
	repo := &fakeContactRepository{}
	sender := &fakeEmailSender{}
	uc := usecase.NewContactUsecase(repo, sender, "support@example.com")

	err := uc.CreateMessage(context.Background(), domain.CreateContactMessageRequest{
		Name: "Ada", Email: "ada@example.com", Type: "OTHER", Message: "Help",
	})
	if err == nil {
		t.Fatal("CreateMessage() expected validation error")
	}
	if repo.message != nil || sender.to != "" {
		t.Fatal("invalid contact message was persisted or emailed")
	}
}

func TestCreateMessageReturnsEmailFailureAfterSaving(t *testing.T) {
	repo := &fakeContactRepository{}
	sender := &fakeEmailSender{err: errors.New("smtp unavailable")}
	uc := usecase.NewContactUsecase(repo, sender, "support@example.com")

	err := uc.CreateMessage(context.Background(), domain.CreateContactMessageRequest{
		Name: "Ada", Email: "ada@example.com", Type: domain.MessageComplaint, Message: "Help",
	})
	if err == nil || !strings.Contains(err.Error(), "smtp unavailable") {
		t.Fatalf("CreateMessage() error = %v, want SMTP failure", err)
	}
	if repo.message == nil {
		t.Fatal("message must remain persisted when email forwarding fails")
	}
}

func TestReplyEmailsSenderAndStoresAdminReply(t *testing.T) {
	repo := &fakeContactRepository{}
	sender := &fakeEmailSender{}
	uc := usecase.NewContactUsecase(repo, sender, "support@example.com")
	if err := uc.CreateMessage(context.Background(), domain.CreateContactMessageRequest{
		Name: "Ada", Email: "ada@example.com", Type: domain.MessageQuestion, Message: "Help",
	}); err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}

	actorID := uuid.New()
	updated, err := uc.Reply(context.Background(), repo.message.ID, actorID, "  <strong>Thanks</strong>  ")
	if err != nil {
		t.Fatalf("Reply() error = %v", err)
	}
	if sender.to != "ada@example.com" || !strings.Contains(sender.body, "&lt;strong&gt;Thanks&lt;/strong&gt;") {
		t.Fatalf("reply email destination/body = %q / %q", sender.to, sender.body)
	}
	if updated.ReplyMessage != "<strong>Thanks</strong>" || updated.RepliedBy == nil || *updated.RepliedBy != actorID {
		t.Fatalf("saved reply = %#v", updated)
	}
}
