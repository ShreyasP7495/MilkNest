package utils

import (
	"fmt"
	"net/http"
)

// AppError is the canonical typed error used across services. Handlers map it
// to the JSON envelope from API Spec §15.
type AppError struct {
	Status  int    // HTTP status
	Code    string // machine code, e.g. "max_packets_exceeded"
	Message string // human-readable message
	Err     error  // wrapped underlying error (optional)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// Constructors

func NewBadRequest(code, msg string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: msg}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: msg}
}

func NewForbidden(msg string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: "forbidden", Message: msg}
}

func NewNotFound(resource string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Code:    "not_found",
		Message: resource + " not found",
	}
}

func NewConflict(code, msg string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: code, Message: msg}
}

func NewUnprocessable(code, msg string) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Code: code, Message: msg}
}

func NewInternal(err error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "internal server error",
		Err:     err,
	}
}
