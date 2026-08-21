/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/raoptimus/evateamclient.go"
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

	var apiErr *evateamclient.APIError
	if errors.As(err, &apiErr) {
		return fmt.Errorf("%s: %w: %s", operation, classifyStatus(apiErr), apiErr.Body)
	}

	errMsg := err.Error()

	// Check for common EVA RPC error patterns
	switch {
	case strings.Contains(errMsg, "not found"):
		return fmt.Errorf("%s: %w: %s", operation, ErrNotFound, errMsg)
	case strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized"):
		return fmt.Errorf("%s: %w: %s", operation, ErrUnauthorized, errMsg)
	case strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Forbidden"):
		return fmt.Errorf("%s: %w: %s", operation, ErrForbidden, errMsg)
	case strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid"):
		return fmt.Errorf("%s: %w: %s", operation, ErrInvalidInput, errMsg)
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

// classifyStatus maps an API response to a sentinel error.
// EVA answers an unusable token with 403 and the body "Invalid API token", so a
// 403 is only a permission problem when the body does not blame the token.
func classifyStatus(apiErr *evateamclient.APIError) error {
	switch apiErr.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		if strings.Contains(strings.ToLower(apiErr.Body), "token") {
			return ErrUnauthorized
		}

		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return ErrInvalidInput
	default:
		return ErrInternalServer
	}
}

// FormatToolError formats error for MCP tool response.
func FormatToolError(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return fmt.Sprintf("Resource not found. Please check the ID or code and try again: %v", err)
	case errors.Is(err, ErrUnauthorized):
		return fmt.Sprintf("Authentication failed. Please check EVA_API_TOKEN: %v", err)
	case errors.Is(err, ErrForbidden):
		return fmt.Sprintf("Access denied. You don't have permission for this operation: %v", err)
	case errors.Is(err, ErrInvalidInput):
		return fmt.Sprintf("Invalid input: %v", err)
	default:
		return fmt.Sprintf("Operation failed: %v", err)
	}
}
