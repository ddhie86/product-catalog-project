// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageQuestion  MessageType = "QUESTION"
	MessageComplaint MessageType = "COMPLAINT"
)

type ContactMessage struct {
	ID           uuid.UUID   `json:"id"`
	Name         string      `json:"name"`
	Email        string      `json:"email"`
	Type         MessageType `json:"type"`
	Message      string      `json:"message"`
	ReplyMessage string      `json:"reply_message,omitempty"`
	RepliedAt    *time.Time  `json:"replied_at,omitempty"`
	RepliedBy    *uuid.UUID  `json:"replied_by,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

type CreateContactMessageRequest struct {
	Name    string      `json:"name" binding:"required,max=255"`
	Email   string      `json:"email" binding:"required,email,max=255"`
	Type    MessageType `json:"type" binding:"required,oneof=QUESTION COMPLAINT"`
	Message string      `json:"message" binding:"required,max=5000"`
}

type ReplyContactMessageRequest struct {
	Message string `json:"message" binding:"required,max=5000"`
}

type ContactRepository interface {
	Create(ctx context.Context, message *ContactMessage) error
	List(ctx context.Context, page, limit int) ([]ContactMessage, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*ContactMessage, error)
	SaveReply(ctx context.Context, message *ContactMessage) error
}

type EmailSender interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

type ContactUsecase interface {
	CreateMessage(ctx context.Context, req CreateContactMessageRequest) error
	ListMessages(ctx context.Context, page, limit int) ([]ContactMessage, int, error)
	Reply(ctx context.Context, id, actorID uuid.UUID, reply string) (*ContactMessage, error)
}
