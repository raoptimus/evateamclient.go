/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/models"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const integrationProjectCode = "epud"

// TestIntegration_TaskCreate_TagsViaMCP verifies end-to-end that a task created
// through the MCP eva_task_create handler keeps its tags, even when the MCP
// client serialises the tags array as a JSON-encoded string — the exact shape
// that used to silently drop tags before the TagList coercion fix.
func TestIntegration_TaskCreate_TagsViaMCP(t *testing.T) {
	ctx := context.Background()

	client, err := evateamclient.NewClient(&evateamclient.Config{
		BaseURL:  os.Getenv("EVA_API_URL"),
		APIToken: os.Getenv("EVA_API_TOKEN"),
		Debug:    true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })

	// Resolve the integration project ID.
	project, _, err := client.Project(ctx, integrationProjectCode, evateamclient.DefaultProjectFields)
	require.NoError(t, err)
	require.NotNil(t, project)
	require.NotEmpty(t, project.ID)

	// Resolve a real tag code to attach.
	tagsList, _, err := client.TagList(ctx, evateamclient.NewQueryBuilder().
		From(evateamclient.EntityTag).Limit(1))
	require.NoError(t, err)
	require.NotEmpty(t, tagsList, "project must have at least one tag for this test")
	tagCode := tagsList[0].Code
	require.NotEmpty(t, tagCode)

	// Resolve the task logic type.
	logicType, err := client.LogicTypeByCode(ctx, evateamclient.LogicTypeCodeTask)
	require.NoError(t, err)
	require.NotEmpty(t, logicType.ID)

	tt := tools.NewTaskTools(client)
	name := fmt.Sprintf("[TEST] tags via MCP %d", time.Now().UnixNano())

	// Reproduce the MCP boundary: tags arrive as a JSON-encoded string inside
	// the tool arguments, decoded into TaskCreateInput via json.Unmarshal.
	payload := fmt.Sprintf(
		`{"name":%q,"project_id":%q,"logic_type_id":%q,"tags":%q}`,
		name, project.ID, logicType.ID, `["`+tagCode+`"]`,
	)
	var input tools.TaskCreateInput
	require.NoError(t, json.Unmarshal([]byte(payload), &input))

	raw, err := tt.TaskCreate(ctx, &input)
	require.NoError(t, err)

	created, ok := raw.(*models.Task)
	require.True(t, ok, "TaskCreate must return *models.Task")
	require.NotEmpty(t, created.ID)
	require.NotEmpty(t, created.Code)
	t.Cleanup(func() { _ = client.TaskDelete(context.Background(), created.ID) })

	// Re-fetch the task and assert the tag was actually persisted.
	fetched, _, err := client.Task(ctx, created.Code, evateamclient.DefaultTaskFields)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.NotEmpty(t, fetched.Tags, "created task must have the tag persisted")

	codes := make([]string, 0, len(fetched.Tags))
	for _, tg := range fetched.Tags {
		codes = append(codes, tg.Code)
	}
	assert.Contains(t, codes, tagCode, "task %s should carry tag %s", fetched.Code, tagCode)
}
