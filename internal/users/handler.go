package users

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Me handles GET /users/me.
func (h *Handler) Me(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, u)
}

// UpdateMe handles PUT /users/me.
func (h *Handler) UpdateMe(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	u, err := h.svc.Update(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, u)
}
