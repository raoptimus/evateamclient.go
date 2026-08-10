/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRelationTypeCode_UnmarshalJSON covers both documented shapes of
// relation_type: a bare code string (plain field projection) and a nested
// object (fields: ["**"], per doc/dopolnitel_nye_optsii_api_zaprosov_doc-000695.pdf),
// plus JSON null.
func TestRelationTypeCode_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
		want RelationTypeCode
	}{
		{"string form", `"system.link"`, "system.link"},
		{"object form with code", `{"id":"CmfRelationType:1","code":"system.link"}`, "system.link"},
		{"object form without code falls back to id", `{"id":"CmfRelationType:1"}`, "CmfRelationType:1"},
		{"null", `null`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got RelationTypeCode
			require.NoError(t, json.Unmarshal([]byte(tt.json), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestTaskLink_RelationType_ObjectForm_UnmarshalsWithinTaskLink pins that a
// relation_type object embedded in a full TaskLink response (fields: ["**"])
// does not break parsing of the whole object — the original regression this
// finding guards against.
func TestTaskLink_RelationType_ObjectForm_UnmarshalsWithinTaskLink(t *testing.T) {
	raw := `{
		"id": "CmfRelationOption:123",
		"code": "RLO-001",
		"relation_type": {"id": "CmfRelationType:1", "code": "system.link"}
	}`

	var link TaskLink
	require.NoError(t, json.Unmarshal([]byte(raw), &link))
	assert.Equal(t, RelationTypeCode("system.link"), link.RelationType)
}
