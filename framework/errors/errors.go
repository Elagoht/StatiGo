// Package errors provides custom error types for the Statigo framework.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents different categories of errors.
type ErrorType string

const (
	ErrorTypeNetwork        ErrorType = "NETWORK"
	ErrorTypeTimeout        ErrorType = "TIMEOUT"
	ErrorTypeHTTP           ErrorType = "HTTP"
	ErrorTypeParse          ErrorType = "PARSE"
	ErrorTypeRetryExhausted ErrorType = "RETRY_EXHAUSTED"
)

// AppError represents a custom application error with context.
type AppError struct {
	Type       ErrorType
	Message    string
	StatusCode int   // For HTTP errors
	Err        error // Wrapped error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewNetworkError creates a network-related error.
func NewNetworkError(err error, message string) *AppError {
	return &AppError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Err:     err,
	}
}

// NewTimeoutError creates a timeout-related error.
func NewTimeoutError(err error, message string) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Err:     err,
	}
}

// NewHTTPError creates an HTTP-related error.
func NewHTTPError(statusCode int, message string) *AppError {
	return &AppError{
		Type:       ErrorTypeHTTP,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewParseError creates a parsing-related error.
func NewParseError(err error, message string) *AppError {
	return &AppError{
		Type:    ErrorTypeParse,
		Message: message,
		Err:     err,
	}
}

// NewRetryExhaustedError creates a retry exhausted error.
func NewRetryExhaustedError(err error, attempts int) *AppError {
	return &AppError{
		Type:    ErrorTypeRetryExhausted,
		Message: fmt.Sprintf("all %d retry attempts failed", attempts),
		Err:     err,
	}
}

// IsNetworkError checks if an error is a network error.
func IsNetworkError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeNetwork
	}
	return false
}

// IsTimeoutError checks if an error is a timeout error.
func IsTimeoutError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeTimeout
	}
	return false
}

// IsHTTPError checks if an error is an HTTP error.
func IsHTTPError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeHTTP
	}
	return false
}

// GetHTTPStatusCode extracts the HTTP status code from an error.
func GetHTTPStatusCode(err error) (int, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Type == ErrorTypeHTTP {
		return appErr.StatusCode, true
	}
	return 0, false
}
