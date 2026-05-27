package delivery

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Assign handles POST /delivery/assign (admin).
func (h *Handler) Assign(c *gin.Context) {
	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	o, err := h.svc.Assign(c.Request.Context(), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, o)
}

// UpdateStatus handles POST /delivery/status (delivery partner only).
func (h *Handler) UpdateStatus(c *gin.Context) {
	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	t, err := h.svc.UpdateStatus(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, t)
}

// Track handles GET /delivery/track/:order_id (any authenticated user, but
// customers can only see their own orders - enforced in service).
func (h *Handler) Track(c *gin.Context) {
	resp, err := h.svc.Track(c.Request.Context(), middleware.UserID(c), middleware.Role(c), c.Param("order_id"))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, resp)
}
