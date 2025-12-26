package evateamclient

import "github.com/gofrs/uuid"

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	CallID  string      `json:"callid"`
	Args    interface{} `json:"args,omitempty"`
	Kwargs  interface{} `json:"kwargs,omitempty"`
}

func newCallID() string {
	id, _ := uuid.NewV4()
	return id.String()
}
