package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/utils"
)

// Handler exposes the Gin handlers for /auth/*.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// SendOTP handles POST /auth/send-otp.
func (h *Handler) SendOTP(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	resp, err := h.svc.SendOTP(c.Request.Context(), req.PhoneNumber)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, resp)
}

// VerifyOTP handles POST /auth/verify-otp.
func (h *Handler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	if err := utils.Validator().Struct(req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_request", err.Error()))
		return
	}
	resp, err := h.svc.VerifyOTP(c.Request.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, resp)
}

// Refresh handles POST /auth/refresh.
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailErr(c, utils.NewBadRequest("invalid_body", err.Error()))
		return
	}
	resp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.FailErr(c, err)
		return
	}
	utils.OK(c, resp)
}
