package products

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context())
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}

func (h *Handler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, p)
}
