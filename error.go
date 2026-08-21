/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrOptionIsRequired    = errors.New("option is required")
	ErrBodyIsRequired      = errors.New("body is required")
	ErrRPCMethodIsRequired = errors.New("RPCRequest.Method is required")
)

// APIError represents a non-2xx HTTP response from the EVA API.
// StatusCode and Body are kept separately so callers can classify the failure
// without parsing the error message.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Body)
}

// RPCError represents JSON-RPC error response
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return e.Message
}

// rpcErrorResponse is used to check for RPC errors in 200 OK responses
type rpcErrorResponse struct {
	Error *RPCError `json:"error,omitempty"`
}
