// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
// Command api: entry point & dependency injection wiring (Modular Monolith).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dedysetiawan/product-catalog-project/internal/config"
	activityhttp "github.com/dedysetiawan/product-catalog-project/internal/modules/activity/delivery/http"
	activitymiddleware "github.com/dedysetiawan/product-catalog-project/internal/modules/activity/middleware"
	activityrepo "github.com/dedysetiawan/product-catalog-project/internal/modules/activity/repository"
	"github.com/dedysetiawan/product-catalog-project/internal/platform/database"
	"github.com/dedysetiawan/product-catalog-project/internal/platform/mailer"
	platformpayment "github.com/dedysetiawan/product-catalog-project/internal/platform/payment"
	"github.com/dedysetiawan/product-catalog-project/pkg/logger"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/token"

	authhttp "github.com/dedysetiawan/product-catalog-project/internal/modules/auth/delivery/http"
	authrepo "github.com/dedysetiawan/product-catalog-project/internal/modules/auth/repository"
	authuc "github.com/dedysetiawan/product-catalog-project/internal/modules/auth/usecase"

	contacthttp "github.com/dedysetiawan/product-catalog-project/internal/modules/contact/delivery/http"
	contactrepo "github.com/dedysetiawan/product-catalog-project/internal/modules/contact/repository"
	contactuc "github.com/dedysetiawan/product-catalog-project/internal/modules/contact/usecase"

	notifdomain "github.com/dedysetiawan/product-catalog-project/internal/modules/notification/domain"
	notifsvc "github.com/dedysetiawan/product-catalog-project/internal/modules/notification/service"

	producthttp "github.com/dedysetiawan/product-catalog-project/internal/modules/product/delivery/http"
	productrepo "github.com/dedysetiawan/product-catalog-project/internal/modules/product/repository"
	productuc "github.com/dedysetiawan/product-catalog-project/internal/modules/product/usecase"

	trxhttp "github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/delivery/http"
	trxrepo "github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/repository"
	trxuc "github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.Init(cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- Infrastructure ----
	pool, err := database.NewPostgres(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	tokens := token.NewManager(cfg.JWTSecret, cfg.JWTTTL)
	gateway := platformpayment.NewXendit(cfg.XenditAPIKey, cfg.XenditCallbackToken)
	smtpSender := mailer.NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)

	// ---- Notification (async worker) ----
	emailChan := make(chan notifdomain.OrderEmailEvent, cfg.EmailQueueSize)
	worker := notifsvc.NewEmailWorker(emailChan, smtpSender, log)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); worker.Run() }()

	// ---- Auth module ----
	userRepo := authrepo.NewPostgresUserRepository(pool)
	roleRepo := authrepo.NewPostgresRoleRepository(pool)
	authHandler := authhttp.NewAuthHandler(authuc.NewAuthUsecase(userRepo, tokens))
	adminUserHandler := authhttp.NewAdminUserHandler(authuc.NewAdminUserUsecase(userRepo), roleRepo)
	roleHandler := authhttp.NewRoleHandler(authuc.NewRoleUsecase(roleRepo))

	// ---- Product module ----
	productRepo := productrepo.NewPostgresProductRepository(pool)
	productHandler := producthttp.NewProductHandler(productuc.NewProductUsecase(productRepo))
	catalogHandler := producthttp.NewCatalogHandler(productuc.NewCatalogUsecase(productRepo))
	categoryHandler := producthttp.NewCategoryHandler(productuc.NewCategoryUsecase(productRepo))
	productService := productuc.NewProductService(productRepo) // kontrak untuk modul lain

	// ---- Contact module ----
	contactRepo := contactrepo.NewPostgresContactRepository(pool)
	contactHandler := contacthttp.NewContactHandler(contactuc.NewContactUsecase(contactRepo, smtpSender, cfg.SupportEmail))
	activityRepo := activityrepo.NewPostgresActivityRepository(pool)
	activityHandler := activityhttp.NewActivityHandler(activityRepo)

	// ---- Transaction module (hanya tahu produk lewat ProductServiceContract) ----
	trxRepo := trxrepo.NewPostgresTransactionRepository(pool)
	reportHandler := trxhttp.NewTransactionHandler(trxuc.NewReportUsecase(trxRepo, productService, cfg.LowStockThreshold))
	checkoutHandler := trxhttp.NewCheckoutHandler(
		trxuc.NewCheckoutUsecase(trxRepo, productService, gateway, emailChan, cfg.PaymentRedirectURL),
		gateway,
	)

	// ---- HTTP ----
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.Recovery(), middleware.RequestLogger(), middleware.CORS(cfg.AllowedOrigins))

	r.GET("/healthz", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	catalogHandler.RegisterRoutes(v1)  // GET  /products, /products/:id
	checkoutHandler.RegisterRoutes(v1) // POST /checkout, /payments/webhook
	contactHandler.RegisterRoutes(v1)  // POST /contact-us

	admin := v1.Group("/admin")
	authHandler.RegisterRoutes(admin) // POST /admin/login
	protected := admin.Group("", middleware.JWTAuth(tokens, userRepo))
	protected.Use(activitymiddleware.Audit(activityRepo, roleRepo))
	adminUserHandler.RegisterAdminRoutes(protected, roleRepo)
	roleHandler.RegisterAdminRoutes(protected, roleRepo)
	categoryHandler.RegisterAdminRoutes(protected, roleRepo)
	productHandler.RegisterAdminRoutes(protected, roleRepo)
	reportHandler.RegisterAdminRoutes(protected, roleRepo)
	contactHandler.RegisterAdminRoutes(protected, roleRepo)
	activityHandler.RegisterAdminRoutes(protected, roleRepo)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server started", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	// HTTP sudah berhenti (tidak ada producer baru) -> tutup channel & tunggu worker menguras antrian.
	close(emailChan)
	wg.Wait()
	return nil
}
