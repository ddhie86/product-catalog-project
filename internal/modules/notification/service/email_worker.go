// Package service (notification): worker email asynchronous (goroutine + channel).
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/notification/domain"
)

type EmailWorker struct {
	events  <-chan domain.OrderEmailEvent
	sender  domain.EmailSender
	log     *slog.Logger
	retries int
}

func NewEmailWorker(events <-chan domain.OrderEmailEvent, sender domain.EmailSender, log *slog.Logger) *EmailWorker {
	return &EmailWorker{events: events, sender: sender, log: log, retries: 3}
}

// Run memproses event sampai channel ditutup (graceful shutdown: tutup channel
// setelah HTTP server berhenti, sehingga sisa antrian tetap terkirim).
func (w *EmailWorker) Run() {
	for ev := range w.events {
		w.handle(ev)
	}
	w.log.Info("email worker stopped")
}

func (w *EmailWorker) handle(ev domain.OrderEmailEvent) {
	subject, body := render(ev)
	var err error
	for attempt := 1; attempt <= w.retries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err = w.sender.Send(ctx, ev.CustomerEmail, subject, body)
		cancel()
		if err == nil {
			w.log.Info("email sent", "kind", ev.Kind, "invoice", ev.InvoiceNumber)
			return
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	w.log.Error("email failed", "kind", ev.Kind, "invoice", ev.InvoiceNumber, "err", err)
}

// render: template minimal. TODO: pindahkan ke html/template + embed.FS.
func render(ev domain.OrderEmailEvent) (subject, body string) {
	switch ev.Kind {
	case domain.EmailPaymentReceipt:
		subject = fmt.Sprintf("Tanda terima pembayaran %s", ev.InvoiceNumber)
		body = fmt.Sprintf("<p>Halo %s,</p><p>Pembayaran untuk pesanan <b>%s</b> sebesar Rp%.0f telah kami terima. Terima kasih!</p>",
			ev.CustomerName, ev.InvoiceNumber, ev.TotalAmount)
	default:
		subject = fmt.Sprintf("Konfirmasi pesanan %s", ev.InvoiceNumber)
		body = fmt.Sprintf("<p>Halo %s,</p><p>Pesanan <b>%s</b> sebesar Rp%.0f berhasil dibuat. Selesaikan pembayaran di: <a href=\"%s\">%s</a></p>",
			ev.CustomerName, ev.InvoiceNumber, ev.TotalAmount, ev.PaymentURL, ev.PaymentURL)
	}
	return
}
