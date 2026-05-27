package notifications

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Send handles POST /notifications/send. It is admin-only - registered behind
// the admin role guard in the router.
func (h *Handler) Send(c *gin.Context) {
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	if _, err := h.svc.SendInternal(c.Request.Context(), req); err != nil {
		utils.FailErr(c, utils.NewInternal(err))
		return
	}
	utils.OK(c, gin.H{"message": "queued"})
}

// List handles GET /notifications for the authenticated user.
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		utils.FailErr(c, utils.NewInternal(err))
		return
	}
	utils.OK(c, items)
}
