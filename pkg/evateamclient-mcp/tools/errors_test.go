/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools_test

import (
	"errors"
	"net/http"
	"testing"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
)

func TestWrapError_NilError(t *testing.T) {
	err := tools.WrapError("test", nil)

	assert.Nil(t, err)
}

func TestWrapError_NotFoundError(t *testing.T) {
	originalErr := errors.New("resource not found")

	err := tools.WrapError("test_operation", originalErr)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, tools.ErrNotFound))
}

func TestWrapError_UnauthorizedError(t *testing.T) {
	originalErr := errors.New("401 Unauthorized")

	err := tools.WrapError("test_operation", originalErr)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, tools.ErrUnauthorized))
}

func TestWrapError_ForbiddenError(t *testing.T) {
	originalErr := errors.New("403 Forbidden")

	err := tools.WrapError("test_operation", originalErr)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, tools.ErrForbidden))
}

func TestWrapError_ValidationError(t *testing.T) {
	originalErr := errors.New("validation failed: invalid input")

	err := tools.WrapError("test_operation", originalErr)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, tools.ErrInvalidInput))
}

func TestWrapError_GenericError(t *testing.T) {
	originalErr := errors.New("some unknown error")

	err := tools.WrapError("test_operation", originalErr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_operation")
	assert.Contains(t, err.Error(), "some unknown error")
}

func TestFormatToolError_NilError(t *testing.T) {
	result := tools.FormatToolError(nil)

	assert.Empty(t, result)
}

func TestFormatToolError_NotFoundError(t *testing.T) {
	err := tools.WrapError("test", errors.New("not found"))

	result := tools.FormatToolError(err)

	assert.Contains(t, result, "not found")
}

func TestFormatToolError_UnauthorizedError(t *testing.T) {
	err := tools.WrapError("test", errors.New("401 Unauthorized"))

	result := tools.FormatToolError(err)

	assert.Contains(t, result, "Authentication failed")
}

func TestFormatToolError_ForbiddenError(t *testing.T) {
	err := tools.WrapError("test", errors.New("403 Forbidden"))

	result := tools.FormatToolError(err)

	assert.Contains(t, result, "Access denied")
}

func TestFormatToolError_InvalidInputError(t *testing.T) {
	err := tools.WrapError("test", errors.New("validation error"))

	result := tools.FormatToolError(err)

	assert.Contains(t, result, "Invalid input")
}

func TestFormatToolError_GenericError(t *testing.T) {
	err := errors.New("some error")

	result := tools.FormatToolError(err)

	assert.Contains(t, result, "Operation failed")
}

func TestWrapError_APIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{
			name:       "invalid token is reported as auth failure, not as missing permission",
			statusCode: http.StatusForbidden,
			body:       "Invalid API token",
			wantErr:    tools.ErrUnauthorized,
		},
		{
			name:       "forbidden without token mention stays a permission problem",
			statusCode: http.StatusForbidden,
			body:       "You have no access to this project",
			wantErr:    tools.ErrForbidden,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       "Unauthorized",
			wantErr:    tools.ErrUnauthorized,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       "Unknown method",
			wantErr:    tools.ErrNotFound,
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			body:       "Malformed kwargs",
			wantErr:    tools.ErrInvalidInput,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			body:       "Internal error",
			wantErr:    tools.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &evateamclient.APIError{StatusCode: tt.statusCode, Body: tt.body}

			err := tools.WrapError("test_operation", apiErr)

			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.Contains(t, err.Error(), tt.body)
		})
	}
}

func TestFormatToolError_InvalidTokenKeepsResponseBody(t *testing.T) {
	apiErr := &evateamclient.APIError{StatusCode: http.StatusForbidden, Body: "Invalid API token"}

	result := tools.FormatToolError(tools.WrapError("person_list", apiErr))

	assert.Contains(t, result, "EVA_API_TOKEN")
	assert.Contains(t, result, "Invalid API token")
}
