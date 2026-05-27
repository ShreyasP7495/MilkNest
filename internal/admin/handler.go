package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Dashboard(c *gin.Context) {
	stats, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, stats)
}
