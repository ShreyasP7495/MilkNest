package payments

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/logger"
	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Create handles POST /payments/create.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	if req.Method == "wallet" {
		// Wallet payments don't go via Razorpay; route to wallet handler instead.
		utils.FailErr(c, utils.NewBadRequest("use_wallet_endpoint",
			"to pay an order via wallet use POST /wallet/pay-order"))
		return
	}
	resp, err := h.svc.CreateForOrder(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.Created(c, resp)
}

// Get handles GET /payments/:id.
func (h *Handler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, p)
}

// Webhook handles POST /payments/webhook. It must read the raw body to verify
// the signature.
func (h *Handler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	signature := c.GetHeader("X-Razorpay-Signature")
	if err := h.svc.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		logger.L().Warn().Err(err).Msg("webhook handling failed")
		utils.FailErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
