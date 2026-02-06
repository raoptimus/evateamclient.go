package tools

import (
	"errors"
	"fmt"
	"strings"
)

// Common error types for MCP tools.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInternalServer = errors.New("internal server error")
)

// WrapError wraps an error with context for MCP response.
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for common EVA RPC error patterns
	switch {
	case strings.Contains(errMsg, "not found"):
		return fmt.Errorf("%s: %w", operation, ErrNotFound)
	case strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized"):
		return fmt.Errorf("%s: %w", operation, ErrUnauthorized)
	case strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Forbidden"):
		return fmt.Errorf("%s: %w", operation, ErrForbidden)
	case strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid"):
		return fmt.Errorf("%s: %w: %s", operation, ErrInvalidInput, errMsg)
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

// FormatToolError formats error for MCP tool response.
func FormatToolError(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return "Resource not found. Please check the ID or code and try again."
	case errors.Is(err, ErrUnauthorized):
		return "Authentication failed. Please check EVA_API_TOKEN."
	case errors.Is(err, ErrForbidden):
		return "Access denied. You don't have permission for this operation."
	case errors.Is(err, ErrInvalidInput):
		return fmt.Sprintf("Invalid input: %v", err)
	default:
		return fmt.Sprintf("Operation failed: %v", err)
	}
}
