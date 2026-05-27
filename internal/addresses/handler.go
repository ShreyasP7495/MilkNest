package addresses

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Create(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	a, err := h.svc.Create(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.Created(c, a)
}

func (h *Handler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}

func (h *Handler) Update(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	a, err := h.svc.Update(c.Request.Context(), middleware.UserID(c), c.Param("id"), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, a)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.NoContent(c)
}
