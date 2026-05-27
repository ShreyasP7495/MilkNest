package offers

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List handles GET /offers.
func (h *Handler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context())
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}
