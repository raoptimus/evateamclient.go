/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"bytes"
	"context"
	encjson "encoding/json"

	"github.com/pkg/errors"
)

// parseWriteResult parses the `result` of a create/update RPC call.
//
// The EVA server's response shape for create/update isn't fixed across
// resources: some methods (e.g. CmfTask.create) return a bare ID string that
// needs a follow-up `.get`, others return the full object inline. A silent
// failure can also surface as an empty result (`null`/`""`/`false`/missing)
// instead of an error — that must not be mistaken for a zero-value success.
//
// fetch resolves a bare ID/code string to the full object (the follow-up
// `.get`). emptyID reports whether a resolved object's ID is unset; it lets
// callers detect the "silent empty object" failure without this function
// depending on reflection to reach an arbitrary T's ID field.
func parseWriteResult[T any](
	ctx context.Context,
	raw encjson.RawMessage,
	method string,
	fetch func(ctx context.Context, id string) (*T, error),
	emptyID func(*T) bool,
) (*T, error) {
	errEmptyResult := errors.Errorf("%s returned empty result", method)

	trimmed := bytes.TrimSpace(raw)
	switch {
	case len(trimmed) == 0, string(trimmed) == "null", string(trimmed) == "false", string(trimmed) == `""`:
		return nil, errEmptyResult
	case trimmed[0] == '"':
		var id string
		if err := encjson.Unmarshal(trimmed, &id); err != nil {
			return nil, errors.WithMessagef(err, "parse %s result", method)
		}

		result, err := fetch(ctx, id)
		if err != nil {
			return nil, errors.WithMessagef(err, "fetch %s result %s", method, id)
		}
		if result == nil || emptyID(result) {
			return nil, errEmptyResult
		}
		return result, nil
	case trimmed[0] == '{':
		var result T
		if err := encjson.Unmarshal(trimmed, &result); err != nil {
			return nil, errors.WithMessagef(err, "parse %s result", method)
		}
		if emptyID(&result) {
			return nil, errEmptyResult
		}
		return &result, nil
	default:
		return nil, errors.Errorf("%s returned unexpected result shape", method)
	}
}
