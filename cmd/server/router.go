package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

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
	"github.com/milknest/backend/pkg/middleware"
)

// handlers bundles all module HTTP handlers for the router wiring.
type handlers struct {
	auth          *auth.Handler
	users         *users.Handler
	addresses     *addresses.Handler
	products      *products.Handler
	subscriptions *subscriptions.Handler
	orders        *orders.Handler
	payments      *payments.Handler
	wallet        *wallet.Handler
	delivery      *delivery.Handler
	complaints    *complaints.Handler
	notifications *notifications.Handler
	offers        *offers.Handler
	inventory     *inventory.Handler
	admin         *admin.Handler
}

// buildRouter mounts every endpoint described in the API Spec under /api/v1.
func buildRouter(cfg *config.Config, rdb *redis.Client, h *handlers) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recover(), gin.Logger())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	v1 := r.Group("/api/v1")

	// --- Auth (public, OTP rate-limited by IP) -----------------------------
	otpLimit := middleware.RedisRateLimit(rdb, "otp", 5, time.Minute)
	v1.POST("/auth/send-otp", otpLimit, h.auth.SendOTP)
	v1.POST("/auth/verify-otp", h.auth.VerifyOTP)
	v1.POST("/auth/refresh", h.auth.Refresh)

	// --- Webhook (public, signature-verified) -----------------------------
	v1.POST("/payments/webhook", h.payments.Webhook)

	// --- Authenticated --------------------------------------------------------
	authed := v1.Group("")
	authed.Use(middleware.JWTAuth(cfg.JWT.AccessSecret))

	// Users
	authed.GET("/users/me", h.users.Me)
	authed.PUT("/users/me", h.users.UpdateMe)

	// Addresses
	authed.POST("/addresses", h.addresses.Create)
	authed.GET("/addresses", h.addresses.List)
	authed.PUT("/addresses/:id", h.addresses.Update)
	authed.DELETE("/addresses/:id", h.addresses.Delete)

	// Products
	authed.GET("/products", h.products.List)
	authed.GET("/products/:id", h.products.Get)

	// Subscriptions
	authed.POST("/subscriptions", h.subscriptions.Create)
	authed.GET("/subscriptions", h.subscriptions.List)
	authed.GET("/subscriptions/:id", h.subscriptions.Get)
	authed.PUT("/subscriptions/:id", h.subscriptions.Update)
	authed.POST("/subscriptions/:id/pause", h.subscriptions.Pause)
	authed.POST("/subscriptions/:id/cancel", h.subscriptions.Cancel)

	// Orders
	authed.GET("/orders", h.orders.ListMine)
	authed.GET("/orders/:id", h.orders.Get)

	// Payments
	authed.POST("/payments/create", h.payments.Create)
	authed.GET("/payments/:id", h.payments.Get)

	// Wallet
	authed.GET("/wallet", h.wallet.Balance)
	authed.POST("/wallet/topup", h.wallet.Topup)
	authed.GET("/wallet/transactions", h.wallet.Transactions)

	// Delivery
	authed.GET("/delivery/track/:order_id", h.delivery.Track)

	// Delivery partner only
	deliveryRole := authed.Group("")
	deliveryRole.Use(middleware.Roles(middleware.RoleDelivery))
	deliveryRole.POST("/delivery/status", h.delivery.UpdateStatus)

	// Complaints
	authed.POST("/complaints", h.complaints.Create)
	authed.GET("/complaints", h.complaints.List)

	// Notifications
	authed.GET("/notifications", h.notifications.List)

	// Offers
	authed.GET("/offers", h.offers.List)

	// --- Admin --------------------------------------------------------------
	adminGrp := authed.Group("/admin")
	adminGrp.Use(middleware.Roles(middleware.RoleAdmin))
	adminGrp.GET("/dashboard", h.admin.Dashboard)
	adminGrp.GET("/orders", h.orders.ListAll)
	adminGrp.GET("/inventory", h.inventory.List)
	adminGrp.POST("/inventory/update", h.inventory.Update)

	// Delivery assign + complaint resolve are admin-only too.
	authed.POST("/delivery/assign", middleware.Roles(middleware.RoleAdmin), h.delivery.Assign)
	authed.POST("/complaints/:id/resolve", middleware.Roles(middleware.RoleAdmin), h.complaints.Resolve)
	authed.POST("/notifications/send", middleware.Roles(middleware.RoleAdmin), h.notifications.Send)

	return r
}
