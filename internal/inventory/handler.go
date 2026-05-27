package inventory

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Update handles POST /admin/inventory/update.
func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	inv, err := h.svc.Update(c.Request.Context(), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, inv)
}

// List handles GET /admin/inventory (admin).
func (h *Handler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context())
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}
