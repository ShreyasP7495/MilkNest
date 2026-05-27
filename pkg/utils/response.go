package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorBody is the standard error envelope defined in API Spec §15.
type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// OK writes a 200 with the payload as JSON.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// Created writes a 201.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

// NoContent writes a 204.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Fail writes the standard error envelope and aborts the request.
func Fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorBody{
		Error:   code,
		Message: message,
		Code:    status,
	})
}

// FailErr maps an error to the right HTTP status using the AppError taxonomy
// in pkg/utils/errors.go. Unknown errors map to 500.
func FailErr(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		Fail(c, appErr.Status, appErr.Code, appErr.Message)
		return
	}
	Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
}
