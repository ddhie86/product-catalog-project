package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	notification "github.com/dedysetiawan/product-catalog-project/internal/modules/notification/domain"
	payment "github.com/dedysetiawan/product-catalog-project/internal/modules/payment/domain"
	product "github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/usecase"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

// ---- fakes ----

type fakeProducts struct{ items map[uuid.UUID]*product.Product }

func (f *fakeProducts) GetProductByID(_ context.Context, id uuid.UUID) (*product.Product, error) {
	if p, ok := f.items[id]; ok {
		return p, nil
	}
	return nil, apperror.NotFound("product not found")
}
func (f *fakeProducts) DeductStock(_ context.Context, id uuid.UUID, qty int) error {
	p := f.items[id]
	if p.Stock < qty {
		return apperror.InsufficientStock("insufficient stock")
	}
	p.Stock -= qty
	return nil
}
func (f *fakeProducts) RestoreStock(_ context.Context, id uuid.UUID, qty int) error {
	f.items[id].Stock += qty
	return nil
}
func (f *fakeProducts) CountLowStock(context.Context, int) (int, error) { return 0, nil }

type fakeTrxRepo struct{ byInvoice map[string]*domain.Transaction }

func (f *fakeTrxRepo) CreateWithItems(_ context.Context, t *domain.Transaction) error {
	f.byInvoice[t.InvoiceNumber] = t
	return nil
}
func (f *fakeTrxRepo) UpdatePaymentURL(_ context.Context, id uuid.UUID, url string) error {
	for _, t := range f.byInvoice {
		if t.ID == id {
			t.PaymentURL = url
		}
	}
	return nil
}
func (f *fakeTrxRepo) FindByID(context.Context, uuid.UUID) (*domain.Transaction, error) {
	return nil, apperror.NotFound("transaction not found")
}
func (f *fakeTrxRepo) FindByInvoice(_ context.Context, inv string) (*domain.Transaction, error) {
	if t, ok := f.byInvoice[inv]; ok {
		return t, nil
	}
	return nil, apperror.NotFound("transaction not found")
}
func (f *fakeTrxRepo) UpdateStatusIf(_ context.Context, id uuid.UUID, from, to domain.Status) (bool, error) {
	for _, t := range f.byInvoice {
		if t.ID == id && t.Status == from {
			t.Status = to
			return true, nil
		}
	}
	return false, nil
}
func (f *fakeTrxRepo) List(context.Context, domain.ListFilter) ([]domain.Transaction, int, error) {
	return nil, 0, nil
}
func (f *fakeTrxRepo) Summary(context.Context) (*domain.TransactionSummary, error) {
	return &domain.TransactionSummary{}, nil
}

type fakeGateway struct{}

func (fakeGateway) CreateInvoice(_ context.Context, r payment.CreateInvoiceRequest) (*payment.Invoice, error) {
	return &payment.Invoice{GatewayID: "gw-1", PaymentURL: "https://pay.example/" + r.ExternalID}, nil
}
func (fakeGateway) VerifyCallbackToken(string) bool { return true }

func setup() (domain.CheckoutUsecase, *fakeProducts, *fakeTrxRepo, chan notification.OrderEmailEvent, uuid.UUID, uuid.UUID) {
	idA, idB := uuid.New(), uuid.New()
	products := &fakeProducts{items: map[uuid.UUID]*product.Product{
		idA: {ID: idA, Name: "A", Price: 10000, Stock: 5},
		idB: {ID: idB, Name: "B", Price: 20000, Stock: 1},
	}}
	repo := &fakeTrxRepo{byInvoice: map[string]*domain.Transaction{}}
	emails := make(chan notification.OrderEmailEvent, 10)
	uc := usecase.NewCheckoutUsecase(repo, products, fakeGateway{}, emails, "https://shop.example/success")
	return uc, products, repo, emails, idA, idB
}

func validReq(items ...domain.CheckoutItem) domain.CheckoutRequest {
	return domain.CheckoutRequest{
		CustomerName: "Budi", CustomerEmail: "budi@example.com",
		CustomerPhone: "0812", ShippingAddress: "Jl. Mawar 1", Items: items,
	}
}

func TestCreateCheckout_Success(t *testing.T) {
	uc, products, _, emails, idA, _ := setup()

	res, err := uc.CreateCheckout(context.Background(), validReq(domain.CheckoutItem{ProductID: idA, Quantity: 2}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalAmount != 20000 || res.Status != domain.StatusPending || res.PaymentURL == "" {
		t.Errorf("unexpected response: %+v", res)
	}
	if products.items[idA].Stock != 3 {
		t.Errorf("stock should be reserved, got %d", products.items[idA].Stock)
	}
	if ev := <-emails; ev.Kind != notification.EmailOrderConfirmation {
		t.Errorf("want order confirmation event, got %s", ev.Kind)
	}
}

func TestCreateCheckout_InsufficientStock_RestoresReservedItems(t *testing.T) {
	uc, products, repo, _, idA, idB := setup()

	_, err := uc.CreateCheckout(context.Background(), validReq(
		domain.CheckoutItem{ProductID: idA, Quantity: 2}, // berhasil direservasi
		domain.CheckoutItem{ProductID: idB, Quantity: 5}, // stok hanya 1 -> gagal
	))
	if apperror.Code(err) != "INSUFFICIENT_STOCK" {
		t.Fatalf("want INSUFFICIENT_STOCK, got %v", err)
	}
	if products.items[idA].Stock != 5 {
		t.Errorf("stock A must be restored to 5, got %d", products.items[idA].Stock)
	}
	if len(repo.byInvoice) != 0 {
		t.Error("no transaction should be persisted")
	}
}

func TestHandleWebhook_Paid_IsIdempotent(t *testing.T) {
	uc, _, repo, emails, idA, _ := setup()
	res, err := uc.CreateCheckout(context.Background(), validReq(domain.CheckoutItem{ProductID: idA, Quantity: 1}))
	if err != nil {
		t.Fatal(err)
	}
	<-emails // buang email konfirmasi order

	payload := payment.WebhookPayload{ExternalID: res.InvoiceNumber, Status: "PAID"}
	for i := 0; i < 2; i++ { // callback ganda
		if err := uc.HandleWebhook(context.Background(), payload); err != nil {
			t.Fatal(err)
		}
	}
	if repo.byInvoice[res.InvoiceNumber].Status != domain.StatusPaid {
		t.Error("status should be PAID")
	}
	if len(emails) != 1 {
		t.Errorf("receipt must be sent exactly once, got %d", len(emails))
	}
}

func TestHandleWebhook_Expired_RestoresStock(t *testing.T) {
	uc, products, _, _, idA, _ := setup()
	res, _ := uc.CreateCheckout(context.Background(), validReq(domain.CheckoutItem{ProductID: idA, Quantity: 2}))

	if err := uc.HandleWebhook(context.Background(), payment.WebhookPayload{ExternalID: res.InvoiceNumber, Status: "EXPIRED"}); err != nil {
		t.Fatal(err)
	}
	if products.items[idA].Stock != 5 {
		t.Errorf("stock should be restored, got %d", products.items[idA].Stock)
	}
}
