/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import "github.com/gofrs/uuid"

type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	CallID  string      `json:"callid"`
	Args    interface{} `json:"args,omitempty"`
	Kwargs  interface{} `json:"kwargs,omitempty"`
}

var (
	AllBasicFields                  = []string{"*"}
	AllBasicAndRelationFields       = []string{"**"}
	AllBasicAndRelationAndM2MFields = []string{"***"}
)

func newCallID() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return id.String()
}
