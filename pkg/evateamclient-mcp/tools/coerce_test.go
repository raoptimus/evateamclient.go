/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoerceTags_StringSlice(t *testing.T) {
	input := []string{"TAG-001", "TAG-002"}
	assert.Equal(t, []string{"TAG-001", "TAG-002"}, coerceTags(input))
}

func TestCoerceTags_AnySlice(t *testing.T) {
	input := []any{"TAG-001", "TAG-002"}
	assert.Equal(t, []string{"TAG-001", "TAG-002"}, coerceTags(input))
}

func TestCoerceTags_JSONEncodedString(t *testing.T) {
	input := `["TAG-004379","TAG-004400","TAG-004402"]`
	assert.Equal(t, []string{"TAG-004379", "TAG-004400", "TAG-004402"}, coerceTags(input))
}

func TestCoerceTags_Nil(t *testing.T) {
	assert.Nil(t, coerceTags(nil))
}

func TestCoerceTags_InvalidString(t *testing.T) {
	assert.Nil(t, coerceTags("not-json"))
}
