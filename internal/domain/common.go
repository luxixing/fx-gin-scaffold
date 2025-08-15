package domain

import (
	"fmt"
	"net/http"
)

// Error represents a domain error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// Common error codes
const (
	// Validation errors
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeInvalid    = "INVALID_VALUE"

	// Authentication errors
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeInvalidToken    = "INVALID_TOKEN"
	ErrCodeInvalidPassword = "INVALID_PASSWORD"

	// Resource errors
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"

	// Internal errors
	ErrCodeInternal = "INTERNAL_ERROR"
	ErrCodeDatabase = "DATABASE_ERROR"
)

// Predefined errors
var (
	ErrUserNotFound    = &Error{Code: ErrCodeNotFound, Message: "User not found"}
	ErrUserExists      = &Error{Code: ErrCodeAlreadyExists, Message: "User already exists"}
	ErrInvalidPassword = &Error{Code: ErrCodeInvalidPassword, Message: "Invalid password"}
	ErrUnauthorized    = &Error{Code: ErrCodeUnauthorized, Message: "Unauthorized"}
	ErrForbidden       = &Error{Code: ErrCodeForbidden, Message: "Forbidden"}
	ErrInvalidToken    = &Error{Code: ErrCodeInvalidToken, Message: "Invalid token"}
	ErrValidation      = &Error{Code: ErrCodeValidation, Message: "Validation failed"}
	ErrInternalServer  = &Error{Code: ErrCodeInternal, Message: "Internal server error"}
)

// NewError creates a new domain error
func NewError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// NewErrorWithDetails creates a new domain error with details
func NewErrorWithDetails(code, message, details string) *Error {
	return &Error{Code: code, Message: message, Details: details}
}

// Response represents a standard API response
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Meta represents response metadata
type Meta struct {
	Total  int64 `json:"total,omitempty"`
	Offset int   `json:"offset,omitempty"`
	Limit  int   `json:"limit,omitempty"`
	Page   int   `json:"page,omitempty"`
	Pages  int   `json:"pages,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data any) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// NewSuccessResponseWithMeta creates a success response with metadata
func NewSuccessResponseWithMeta(data any, meta *Meta) *Response {
	return &Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err *Error) *Response {
	return &Response{
		Success: false,
		Error:   err,
	}
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page  int `form:"page,default=1" validate:"min=1"`
	Limit int `form:"limit,default=10" validate:"min=1,max=100"`
}

// GetOffset calculates the offset for pagination
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// GetMeta creates pagination metadata
func (p *PaginationRequest) GetMeta(total int64) *Meta {
	pages := int((total + int64(p.Limit) - 1) / int64(p.Limit))
	return &Meta{
		Total:  total,
		Offset: p.GetOffset(),
		Limit:  p.Limit,
		Page:   p.Page,
		Pages:  pages,
	}
}

// HTTPStatusFromError returns the appropriate HTTP status code for a domain error
func HTTPStatusFromError(err error) int {
	if domainErr, ok := err.(*Error); ok {
		switch domainErr.Code {
		case ErrCodeValidation, ErrCodeInvalid:
			return http.StatusBadRequest
		case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeInvalidPassword:
			return http.StatusUnauthorized
		case ErrCodeForbidden:
			return http.StatusForbidden
		case ErrCodeNotFound:
			return http.StatusNotFound
		case ErrCodeAlreadyExists:
			return http.StatusConflict
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// WrapError wraps a regular error as a domain error
func WrapError(err error, code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}

// ValidationError creates a validation error with field details
func ValidationError(field, message string) *Error {
	return &Error{
		Code:    ErrCodeValidation,
		Message: fmt.Sprintf("Validation failed for field '%s': %s", field, message),
		Details: fmt.Sprintf("field=%s", field),
	}
}
