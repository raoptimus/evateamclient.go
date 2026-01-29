package tools_test

import (
	"errors"
	"testing"

	"github.com/raoptimus/evateamclient/pkg/evateamclient-mcp/tools"
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
