package dto

import (
	"time"

	"coin-radar/internal/domain/signal"
)

// APIError represents the detail of an error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Fields holds per-field validation messages, keyed by JSON field name.
	// Empty for non-validation errors.
	Fields map[string]string `json:"fields,omitempty"`
}

// Response represents a standardized API response structure
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
}

// Success creates a new success Response with the provided data
func Success(data interface{}) Response {
	return Response{
		Success:   true,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Error creates a new error Response with the provided code and message
func Error(code string, message string) Response {
	return Response{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
	}
}

// ValidationError creates an error Response carrying per-field messages.
func ValidationError(code, message string, fields map[string]string) Response {
	return Response{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
		Timestamp: time.Now(),
	}
}

// SymbolsResponse represents the DTO for GET /symbols
type SymbolsResponse struct {
	Symbols []string `json:"symbols"`
}

// SignalsResponse represents the DTO for GET /signals
type SignalsResponse struct {
	Signals []signal.Signal `json:"signals"`
}
