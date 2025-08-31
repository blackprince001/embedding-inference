// internal/domain/errors/errors.go
package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different types of errors that can occur
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeTokenizer  ErrorType = "tokenizer"
	ErrorTypeBackend    ErrorType = "backend"
	ErrorTypeOverloaded ErrorType = "overloaded"
	ErrorTypeUnhealthy  ErrorType = "unhealthy"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeUnknown    ErrorType = "unknown"
)

// TEIError represents an error from the Text Embeddings Inference service
type TEIError struct {
	Message   string    `json:"message"`
	Type      ErrorType `json:"error_type"`
	Code      int       `json:"code,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e *TEIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("TEI error [%s] (request: %s): %s", e.Type, e.RequestID, e.Message)
	}
	return fmt.Sprintf("TEI error [%s]: %s", e.Type, e.Message)
}

// IsRetryable returns true if the error indicates a retryable condition
func (e *TEIError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeOverloaded, ErrorTypeNetwork, ErrorTypeTimeout:
		return true
	case ErrorTypeBackend:
		// Some backend errors might be retryable (5xx status codes)
		return e.Code >= 500
	default:
		return false
	}
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", v.Field, v.Message)
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (m *MultiValidationError) Error() string {
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors", len(m.Errors))
}

// Add appends a validation error
func (m *MultiValidationError) Add(field, message string, value any) {
	m.Errors = append(m.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (m *MultiValidationError) HasErrors() bool {
	return len(m.Errors) > 0
}

// NewTEIError creates a new TEI error
func NewTEIError(message string, errorType ErrorType) *TEIError {
	return &TEIError{
		Message: message,
		Type:    errorType,
	}
}

// NewTEIErrorFromHTTP creates a TEI error from HTTP response
func NewTEIErrorFromHTTP(statusCode int, message string) *TEIError {
	var errorType ErrorType

	switch {
	case statusCode == http.StatusRequestEntityTooLarge:
		errorType = ErrorTypeValidation
	case statusCode == http.StatusUnprocessableEntity:
		errorType = ErrorTypeTokenizer
	case statusCode == http.StatusFailedDependency:
		errorType = ErrorTypeBackend
	case statusCode == http.StatusTooManyRequests:
		errorType = ErrorTypeOverloaded
	case statusCode == http.StatusServiceUnavailable:
		errorType = ErrorTypeUnhealthy
	case statusCode >= 500:
		errorType = ErrorTypeBackend
	case statusCode >= 400:
		errorType = ErrorTypeValidation
	default:
		errorType = ErrorTypeUnknown
	}

	return &TEIError{
		Message: message,
		Type:    errorType,
		Code:    statusCode,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value any) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}
