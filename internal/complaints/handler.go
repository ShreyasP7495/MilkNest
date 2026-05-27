package complaints

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.Created(c, out)
}

func (h *Handler) List(c *gin.Context) {
	role := middleware.Role(c)
	var (
		out []Complaint
		err error
	)
	if role == middleware.RoleAdmin {
		out, err = h.svc.ListAll(c.Request.Context())
	} else {
		out, err = h.svc.ListForUser(c.Request.Context(), middleware.UserID(c))
	}
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}

func (h *Handler) Resolve(c *gin.Context) {
	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	out, err := h.svc.Resolve(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, out)
}
