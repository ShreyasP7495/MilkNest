package orders

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) ListMine(c *gin.Context) {
	out, err := h.svc.ListForUser(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}

func (h *Handler) Get(c *gin.Context) {
	o, err := h.svc.Get(c.Request.Context(), middleware.UserID(c), middleware.Role(c), c.Param("id"))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, o)
}

func (h *Handler) ListAll(c *gin.Context) {
	out, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}
