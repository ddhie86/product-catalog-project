// Package mailer: adapter pengiriman email via SMTP.
package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/notification/domain"
)

type SMTP struct {
	host, user, pass, from string
	port                   int
}

func NewSMTP(host string, port int, user, pass, from string) *SMTP {
	return &SMTP{host: host, port: port, user: user, pass: pass, from: from}
}

var _ domain.EmailSender = (*SMTP)(nil)

func (s *SMTP) Send(ctx context.Context, to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}
	// Cegah header injection.
	clean := func(v string) string { return strings.NewReplacer("\r", "", "\n", "").Replace(v) }

	msg := "From: " + clean(s.from) + "\r\n" +
		"To: " + clean(to) + "\r\n" +
		"Subject: " + clean(subject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" + htmlBody

	done := make(chan error, 1)
	go func() { done <- smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
