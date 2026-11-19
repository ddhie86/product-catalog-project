package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	notification "github.com/dedysetiawan/product-catalog-project/internal/modules/notification/domain"
	payment "github.com/dedysetiawan/product-catalog-project/internal/modules/payment/domain"
	product "github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type checkoutUsecase struct {
	repo        domain.TransactionRepository
	products    product.ProductServiceContract // satu-satunya akses ke data produk
	gateway     payment.Gateway
	emailChan   chan<- notification.OrderEmailEvent
	redirectURL string
}

func NewCheckoutUsecase(
	repo domain.TransactionRepository,
	products product.ProductServiceContract,
	gateway payment.Gateway,
	emailChan chan<- notification.OrderEmailEvent,
	redirectURL string,
) domain.CheckoutUsecase {
	return &checkoutUsecase{repo: repo, products: products, gateway: gateway, emailChan: emailChan, redirectURL: redirectURL}
}

type reservedItem struct {
	productID uuid.UUID
	qty       int
}

// CreateCheckout:
//  1. ProductServiceContract.GetProductByID & DeductStock (reservasi stok)
//  2. Simpan transaksi berstatus PENDING
//  3. Buat invoice di payment gateway
//  4. Kirim event ke emailChan (non-blocking)
//
// Bila langkah 2/3 gagal, stok yang sudah direservasi dikembalikan.
func (u *checkoutUsecase) CreateCheckout(ctx context.Context, req domain.CheckoutRequest) (*domain.CheckoutResponse, error) {
	trxID := uuid.New()
	trx := &domain.Transaction{
		ID:              trxID,
		InvoiceNumber:   newInvoiceNumber(),
		CustomerName:    strings.TrimSpace(req.CustomerName),
		CustomerEmail:   strings.ToLower(strings.TrimSpace(req.CustomerEmail)),
		CustomerPhone:   req.CustomerPhone,
		ShippingAddress: req.ShippingAddress,
		Status:          domain.StatusPending,
	}

	var reserved []reservedItem
	rollback := func() {
		// context baru: rollback harus tetap jalan walau request sudah dibatalkan.
		rctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, r := range reserved {
			if err := u.products.RestoreStock(rctx, r.productID, r.qty); err != nil {
				slog.Error("restore stock failed", "product_id", r.productID, "qty", r.qty, "err", err)
			}
		}
	}

	for _, it := range req.Items {
		p, err := u.products.GetProductByID(ctx, it.ProductID)
		if err != nil {
			rollback()
			return nil, err
		}
		if err := u.products.DeductStock(ctx, it.ProductID, it.Quantity); err != nil {
			rollback()
			return nil, err
		}
		reserved = append(reserved, reservedItem{it.ProductID, it.Quantity})

		trx.TotalAmount += p.Price * float64(it.Quantity)
		trx.Items = append(trx.Items, domain.TransactionItem{
			ID:              uuid.New(),
			ProductID:       it.ProductID,
			Quantity:        it.Quantity,
			PriceAtPurchase: p.Price, // snapshot harga saat beli
		})
	}

	if err := u.repo.CreateWithItems(ctx, trx); err != nil {
		rollback()
		return nil, err
	}

	inv, err := u.gateway.CreateInvoice(ctx, payment.CreateInvoiceRequest{
		ExternalID:  trx.InvoiceNumber,
		Amount:      trx.TotalAmount,
		PayerEmail:  trx.CustomerEmail,
		Description: "Order " + trx.InvoiceNumber,
		SuccessURL:  u.redirectURL,
	})
	if err != nil {
		rollback()
		// Tandai CANCELLED agar tidak tertinggal sebagai PENDING tanpa payment URL.
		_, _ = u.repo.UpdateStatusIf(context.Background(), trx.ID, domain.StatusPending, domain.StatusCancelled)
		return nil, apperror.Internal(fmt.Errorf("create payment invoice: %w", err))
	}
	trx.PaymentURL = inv.PaymentURL
	if err := u.repo.UpdatePaymentURL(ctx, trx.ID, inv.PaymentURL); err != nil {
		return nil, err
	}

	u.emit(notification.OrderEmailEvent{
		Kind:          notification.EmailOrderConfirmation,
		InvoiceNumber: trx.InvoiceNumber,
		CustomerName:  trx.CustomerName,
		CustomerEmail: trx.CustomerEmail,
		TotalAmount:   trx.TotalAmount,
		PaymentURL:    trx.PaymentURL,
	})

	return &domain.CheckoutResponse{
		TransactionID: trx.ID,
		InvoiceNumber: trx.InvoiceNumber,
		TotalAmount:   trx.TotalAmount,
		Status:        trx.Status,
		PaymentURL:    trx.PaymentURL,
	}, nil
}

// HandleWebhook memperbarui status transaksi. Idempoten: callback ganda aman.
func (u *checkoutUsecase) HandleWebhook(ctx context.Context, p payment.WebhookPayload) error {
	trx, err := u.repo.FindByInvoice(ctx, p.ExternalID)
	if err != nil {
		return err
	}

	var target domain.Status
	switch strings.ToUpper(p.Status) {
	case "PAID", "SETTLED":
		target = domain.StatusPaid
	case "EXPIRED":
		target = domain.StatusExpired
	default:
		return nil // status lain diabaikan (ack agar gateway tidak retry)
	}

	changed, err := u.repo.UpdateStatusIf(ctx, trx.ID, domain.StatusPending, target)
	if err != nil {
		return err
	}
	if !changed {
		return nil // sudah final sebelumnya
	}

	switch target {
	case domain.StatusPaid:
		u.emit(notification.OrderEmailEvent{
			Kind:          notification.EmailPaymentReceipt,
			InvoiceNumber: trx.InvoiceNumber,
			CustomerName:  trx.CustomerName,
			CustomerEmail: trx.CustomerEmail,
			TotalAmount:   trx.TotalAmount,
		})
	case domain.StatusExpired:
		for _, it := range trx.Items {
			if err := u.products.RestoreStock(ctx, it.ProductID, it.Quantity); err != nil {
				slog.Error("restore stock on expiry failed", "invoice", trx.InvoiceNumber, "product_id", it.ProductID, "err", err)
			}
		}
	}
	return nil
}

// emit non-blocking: bila antrian penuh, event di-drop & di-log (webhook/checkout tidak boleh macet).
func (u *checkoutUsecase) emit(ev notification.OrderEmailEvent) {
	select {
	case u.emailChan <- ev:
	default:
		slog.Warn("email queue full, event dropped", "invoice", ev.InvoiceNumber, "kind", ev.Kind)
	}
}

func newInvoiceNumber() string {
	return fmt.Sprintf("INV-%s-%s", time.Now().UTC().Format("20060102"), strings.ToUpper(uuid.NewString()[:8]))
}
