package wallet

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Balance handles GET /wallet.
func (h *Handler) Balance(c *gin.Context) {
	w, err := h.svc.Balance(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, w)
}

// Topup handles POST /wallet/topup. Returns the Razorpay order id - the
// actual credit happens via the webhook.
func (h *Handler) Topup(c *gin.Context) {
	var req TopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	resp, err := h.svc.Topup(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.Created(c, resp)
}

// Transactions handles GET /wallet/transactions.
func (h *Handler) Transactions(c *gin.Context) {
	out, err := h.svc.ListTxns(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}
