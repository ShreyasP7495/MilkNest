package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/internal/addresses"
	"github.com/milknest/backend/internal/admin"
	"github.com/milknest/backend/internal/auth"
	"github.com/milknest/backend/internal/complaints"
	"github.com/milknest/backend/internal/delivery"
	"github.com/milknest/backend/internal/inventory"
	"github.com/milknest/backend/internal/notifications"
	"github.com/milknest/backend/internal/offers"
	"github.com/milknest/backend/internal/orders"
	"github.com/milknest/backend/internal/payments"
	"github.com/milknest/backend/internal/products"
	"github.com/milknest/backend/internal/subscriptions"
	"github.com/milknest/backend/internal/users"
	"github.com/milknest/backend/internal/wallet"

	"github.com/milknest/backend/pkg/config"
	"github.com/milknest/backend/pkg/db"
	"github.com/milknest/backend/pkg/logger"
	rediscli "github.com/milknest/backend/pkg/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger.Init(cfg.App.Env)
	log := logger.L()
	log.Info().Str("env", cfg.App.Env).Str("port", cfg.App.Port).Msg("starting milknest")

	pg, err := db.New(cfg.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres init failed")
	}
	defer pg.Close()

	rdb, err := rediscli.New(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("redis init failed")
	}
	defer rdb.Close()

	// --- Repositories -------------------------------------------------------
	authRepo := auth.NewRepository(pg)
	addrRepo := addresses.NewRepository(pg)
	productRepo := products.NewRepository(pg)
	invRepo := inventory.NewRepository(pg)
	subsRepo := subscriptions.NewRepository(pg)
	ordersRepo := orders.NewRepository(pg)
	payRepo := payments.NewRepository(pg)
	walletRepo := wallet.NewRepository(pg)
	deliveryRepo := delivery.NewRepository(pg)
	complaintsRepo := complaints.NewRepository(pg)
	offersRepo := offers.NewRepository(pg)
	notifRepo := notifications.NewRepository(pg)
	adminRepo := admin.NewRepository(pg)

	// --- Services -----------------------------------------------------------
	notifSvc := notifications.NewService(cfg, notifRepo)
	tokenSvc := auth.NewTokenService(cfg.JWT)
	authSvc := auth.NewService(authRepo, tokenSvc, rdb, cfg, notifSvc)
	usersSvc := users.NewService(authRepo)
	addrSvc := addresses.NewService(addrRepo)
	productSvc := products.NewService(productRepo)
	invSvc := inventory.NewService(invRepo)
	offersSvc := offers.NewService(offersRepo)
	subsSvc := subscriptions.NewService(subsRepo, invSvc, offersSvc, cfg)
	ordersSvc := orders.NewService(ordersRepo)
	gw := payments.NewRazorpay(cfg.Razorpay)
	paySvc := payments.NewService(payRepo, ordersRepo, gw, cfg)
	walletSvc := wallet.NewService(walletRepo, paySvc)
	paySvc.SetWallet(walletSvc) // resolves the cycle at runtime
	deliverySvc := delivery.NewService(deliveryRepo, ordersRepo, rdb)
	complaintsSvc := complaints.NewService(complaintsRepo)
	adminSvc := admin.NewService(adminRepo)

	// --- Handlers -----------------------------------------------------------
	h := &handlers{
		auth:          auth.NewHandler(authSvc),
		users:         users.NewHandler(usersSvc),
		addresses:     addresses.NewHandler(addrSvc),
		products:      products.NewHandler(productSvc),
		subscriptions: subscriptions.NewHandler(subsSvc),
		orders:        orders.NewHandler(ordersSvc),
		payments:      payments.NewHandler(paySvc),
		wallet:        wallet.NewHandler(walletSvc),
		delivery:      delivery.NewHandler(deliverySvc),
		complaints:    complaints.NewHandler(complaintsSvc),
		notifications: notifications.NewHandler(notifSvc),
		offers:        offers.NewHandler(offersSvc),
		inventory:     inventory.NewHandler(invSvc),
		admin:         admin.NewHandler(adminSvc),
	}

	// --- Cron ---------------------------------------------------------------
	generator := orders.NewOrderGenerator(pg, subsRepo, invRepo, ordersRepo)
	cronRunner, err := generator.Start(context.Background(), cfg.Cron)
	if err != nil {
		log.Fatal().Err(err).Msg("cron start failed")
	}
	defer cronRunner.Stop()

	// --- HTTP server --------------------------------------------------------
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := buildRouter(cfg, rdb, h)
	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http listen failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}
