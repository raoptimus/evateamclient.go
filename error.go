package evateamclient

import "github.com/pkg/errors"

var (
	ErrOptionIsRequired    = errors.New("option is required")
	ErrBodyIsRequired      = errors.New("body is required")
	ErrRPCMethodIsRequired = errors.New("RPCRequest.Method is required")
)

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
