/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raoptimus/evateamclient.go"
)

func boolPtr(b bool) *bool { return &b }

var (
	readOnlyAnnotations = &mcp.ToolAnnotations{
		ReadOnlyHint: true,
	}
	writeAnnotations = &mcp.ToolAnnotations{
		DestructiveHint: boolPtr(false),
		OpenWorldHint:   boolPtr(true),
	}
	idempotentWriteAnnotations = &mcp.ToolAnnotations{
		DestructiveHint: boolPtr(false),
		IdempotentHint:  true,
		OpenWorldHint:   boolPtr(true),
	}
	destructiveAnnotations = &mcp.ToolAnnotations{
		DestructiveHint: boolPtr(true),
		OpenWorldHint:   boolPtr(true),
	}
)

// Registry holds all tool handlers.
type Registry struct {
	Task          *TaskTools
	Project       *ProjectTools
	List          *ListTools
	Document      *DocumentTools
	Person        *PersonTools
	TimeLog       *TimeLogTools
	Comment       *CommentTools
	Epic          *EpicTools
	TaskLink      *TaskLinkTools
	StatusHistory *StatusHistoryTools
	Stats         *StatsTools
	LogicType     *LogicTypeTools
	Tag           *TagTools
}

// NewRegistry creates a new Registry with all tools initialized.
func NewRegistry(client *evateamclient.Client) *Registry {
	return &Registry{
		Task:          NewTaskTools(client),
		Project:       NewProjectTools(client),
		List:          NewListTools(client),
		Document:      NewDocumentTools(client),
		Person:        NewPersonTools(client),
		TimeLog:       NewTimeLogTools(client),
		Comment:       NewCommentTools(client),
		Epic:          NewEpicTools(client),
		TaskLink:      NewTaskLinkTools(client),
		StatusHistory: NewStatusHistoryTools(client),
		Stats:         NewStatsTools(client),
		LogicType:     NewLogicTypeTools(client),
		Tag:           NewTagTools(client),
	}
}

// wrapHandler wraps a typed handler function to work with MCP's generic interface.
func wrapHandler[In, Out any](handler func(context.Context, In) (Out, error)) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args In) (*mcp.CallToolResult, Out, error) {
		result, err := handler(ctx, args)
		if err != nil {
			var zero Out
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: FormatToolError(err)},
				},
				IsError: true,
			}, zero, nil
		}

		// Serialize result to JSON for text content
		jsonBytes, jsonErr := json.MarshalIndent(result, "", "  ")
		if jsonErr != nil {
			var zero Out
			return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "Failed to serialize result: " + jsonErr.Error()},
					},
					IsError: true,
				},
				zero,
				jsonErr
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			},
			result,
			nil
	}
}

// addTool registers a typed handler, defaulting its input schema to one that
// tolerates JSON-encoded strings in place of arrays (see relaxArraySchema).
// Many MCP clients stringify array arguments; without the relaxed schema the
// SDK rejects them before the handler runs. A tool that sets InputSchema
// explicitly (e.g. eva_task_create_tree) keeps its own schema.
func addTool[In, Out any](
	server *mcp.Server,
	tool *mcp.Tool,
	handler func(context.Context, In) (Out, error),
) {
	if tool.InputSchema == nil {
		tool.InputSchema = relaxedInputSchema[In]()
	}
	mcp.AddTool(server, tool, wrapHandler(handler))
}

// relaxedInputSchema builds the JSON schema for the input type In and relaxes
// every array node to also accept a JSON-encoded string. Pointer types are
// dereferenced. Panics on failure since the type is fixed and reflection is
// deterministic.
func relaxedInputSchema[In any]() *jsonschema.Schema {
	t := reflect.TypeFor[In]()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	s, err := jsonschema.ForType(t, &jsonschema.ForOptions{})
	if err != nil {
		panic("build input schema for " + t.String() + ": " + err.Error())
	}
	relaxArraySchema(s)
	return s
}

// RegisterAll registers all tools with the MCP server.
func (r *Registry) RegisterAll(server *mcp.Server) {
	// Task tools
	addTool(server, &mcp.Tool{
		Name:        "eva_task_list",
		Description: "List tasks with optional filters (project, status, sprint, responsible)",
		Annotations: readOnlyAnnotations,
	}, r.Task.TaskList)

	addTool(server, &mcp.Tool{
		Name: "eva_task_get",
		Description: "Get a single task by code (e.g., 'PROJ-123') or ID. " +
			"Returns a not-found error for a non-existent code/id. " +
			"To find which Story/task a task hangs under, use parent_task_id (stable) — " +
			"epic_id is server-denormalized and may read back as the ROOT epic, not the immediate parent.",
		Annotations: readOnlyAnnotations,
	}, r.Task.TaskGet)

	addTool(server, &mcp.Tool{
		Name: "eva_task_create",
		Description: "Create a new task. " +
			"project_id accepts a project ID or code (e.g. 'epud'). " +
			"epic and parent_task accept the parent's task ID. " +
			"logic_type_id is required to set the task type — resolve a code via eva_logic_type_get. " +
			"tags accepts tag codes (e.g. 'TAG-000004') — use eva_tag_list to find available tags.",
		Annotations: writeAnnotations,
	}, r.Task.TaskCreate)

	addTool(server, &mcp.Tool{
		Name: "eva_task_create_with_subtasks",
		Description: "Create a parent task and multiple child tasks in one call. " +
			"Children are created in parallel (workers, default 3, max 10). " +
			"Each child inherits project_id and gets parent_task set to the new parent. " +
			"Returns parent and children as full task objects; failed children include an error field.",
		Annotations: writeAnnotations,
	}, r.Task.TaskCreateWithSubtasks)

	addTool(server, &mcp.Tool{
		Name: "eva_task_create_tree",
		Description: "Create a whole epic→story→task hierarchy in one call to save round-trips. " +
			"Provide one or more epics in epics[], each with its stories, and the tasks nested under each story. " +
			"epic_id chains up the hierarchy: a story's epic_id is its epic, a task's epic_id is its story; every task also gets parent_task set to its story. " +
			"Optional status (per node, or a tree-wide default) is applied right after creation via a workflow transition that preserves epic_id (e.g. 'Backlog', 'OPEN', 'CLOSED'). " +
			"Logic types are resolved automatically (epic/userstory/agile-task defaults); override per level with " +
			"epic_logic_type/story_logic_type/task_logic_type (accepts a code or a CmfLogicType: ID). " +
			"Stories and tasks are created in parallel (workers, default 3, max 10). " +
			"Returns a compact tree of created ids/codes; failed nodes include an error field.",
		Annotations: writeAnnotations,
		InputSchema: taskCreateTreeInputSchema(),
	}, r.Task.TaskCreateTree)

	addTool(server, &mcp.Tool{
		Name: "eva_task_update",
		Description: "Update an existing task. " +
			"Pass fields to change in updates (e.g. name, priority, deadline). " +
			"To set tags pass an array of tag codes in updates: {\"tags\": [\"TAG-000004\"]}. " +
			"Use eva_tag_list to find available tags. " +
			"epic_id is preserved automatically (the server may otherwise reset it on a partial " +
			"update); to move a task, pass epic explicitly. KNOWN SERVER ISSUE: a partial update " +
			"may still occasionally reset responsible_id or status — verify the result if those matter.",
		Annotations: idempotentWriteAnnotations,
	}, r.Task.TaskUpdate)

	addTool(server, &mcp.Tool{
		Name:        "eva_task_delete",
		Description: "Delete a task",
		Annotations: destructiveAnnotations,
	}, r.Task.TaskDelete)

	addTool(server, &mcp.Tool{
		Name: "eva_task_update_status",
		Description: "Update task status. Accepts OPEN, IN_PROGRESS, CLOSED (and 'Backlog'); " +
			"moves to the first substatus of that type, not a specific substatus. " +
			"Preserves epic_id across the transition. A concrete substatus via " +
			"eva_task_update updates.status_id is not supported (status_id is readonly).",
		Annotations: idempotentWriteAnnotations,
	}, r.Task.TaskUpdateStatus)

	addTool(server, &mcp.Tool{
		Name:        "eva_task_archive",
		Description: "Archive a task (soft delete)",
		Annotations: destructiveAnnotations,
	}, r.Task.TaskArchive)

	addTool(server, &mcp.Tool{
		Name:        "eva_task_count",
		Description: "Count tasks matching filters",
		Annotations: readOnlyAnnotations,
	}, r.Task.TaskCount)

	// Project tools
	addTool(server, &mcp.Tool{
		Name:        "eva_project_list",
		Description: "List projects",
		Annotations: readOnlyAnnotations,
	}, r.Project.ProjectList)

	addTool(server, &mcp.Tool{
		Name:        "eva_project_get",
		Description: "Get a single project by code or ID",
		Annotations: readOnlyAnnotations,
	}, r.Project.ProjectGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_project_create",
		Description: "Create a new project",
		Annotations: writeAnnotations,
	}, r.Project.ProjectCreate)

	addTool(server, &mcp.Tool{
		Name:        "eva_project_update",
		Description: "Update an existing project",
		Annotations: idempotentWriteAnnotations,
	}, r.Project.ProjectUpdate)

	addTool(server, &mcp.Tool{
		Name:        "eva_project_delete",
		Description: "Delete a project",
		Annotations: destructiveAnnotations,
	}, r.Project.ProjectDelete)

	addTool(server, &mcp.Tool{
		Name:        "eva_project_add_executor",
		Description: "Add an executor to a project",
		Annotations: writeAnnotations,
	}, r.Project.ProjectAddExecutor)

	addTool(server, &mcp.Tool{
		Name:        "eva_project_remove_executor",
		Description: "Remove an executor from a project",
		Annotations: writeAnnotations,
	}, r.Project.ProjectRemoveExecutor)

	addTool(server, &mcp.Tool{
		Name:        "eva_project_count",
		Description: "Count projects",
		Annotations: readOnlyAnnotations,
	}, r.Project.ProjectCount)

	// List tools (sprints/releases)
	addTool(server, &mcp.Tool{
		Name:        "eva_list_list",
		Description: "List all lists (sprints and releases) with optional filters",
		Annotations: readOnlyAnnotations,
	}, r.List.ListList)

	addTool(server, &mcp.Tool{
		Name:        "eva_list_get",
		Description: "Get a single list by code (e.g., 'SPR-001543') or ID",
		Annotations: readOnlyAnnotations,
	}, r.List.ListGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_list_create",
		Description: "Create a new list (sprint/release)",
		Annotations: writeAnnotations,
	}, r.List.ListCreate)

	addTool(server, &mcp.Tool{
		Name:        "eva_list_update",
		Description: "Update an existing list",
		Annotations: idempotentWriteAnnotations,
	}, r.List.ListUpdate)

	addTool(server, &mcp.Tool{
		Name:        "eva_list_close",
		Description: "Close a list (sprint/release)",
		Annotations: destructiveAnnotations,
	}, r.List.ListClose)

	addTool(server, &mcp.Tool{
		Name:        "eva_list_delete",
		Description: "Delete a list",
		Annotations: destructiveAnnotations,
	}, r.List.ListDelete)

	addTool(server, &mcp.Tool{
		Name:        "eva_list_count",
		Description: "Count lists",
		Annotations: readOnlyAnnotations,
	}, r.List.ListCount)

	// Sprint aliases
	addTool(server, &mcp.Tool{
		Name:        "eva_sprint_list",
		Description: "List sprints (alias for eva_list_list with type=sprint)",
		Annotations: readOnlyAnnotations,
	}, r.List.SprintList)

	addTool(server, &mcp.Tool{
		Name:        "eva_sprint_get",
		Description: "Get a single sprint by code (e.g., 'SPR-001543')",
		Annotations: readOnlyAnnotations,
	}, r.List.SprintGet)

	// Release aliases
	addTool(server, &mcp.Tool{
		Name:        "eva_release_list",
		Description: "List releases (alias for eva_list_list with type=release)",
		Annotations: readOnlyAnnotations,
	}, r.List.ReleaseList)

	addTool(server, &mcp.Tool{
		Name:        "eva_release_get",
		Description: "Get a single release by code (e.g., 'REL-001641')",
		Annotations: readOnlyAnnotations,
	}, r.List.ReleaseGet)

	// Document tools
	addTool(server, &mcp.Tool{
		Name:        "eva_document_list",
		Description: "List documents",
		Annotations: readOnlyAnnotations,
	}, r.Document.DocumentList)

	addTool(server, &mcp.Tool{
		Name:        "eva_document_get",
		Description: "Get a single document by code or ID",
		Annotations: readOnlyAnnotations,
	}, r.Document.DocumentGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_document_create",
		Description: "Create a new document",
		Annotations: writeAnnotations,
	}, r.Document.DocumentCreate)

	addTool(server, &mcp.Tool{
		Name:        "eva_document_update",
		Description: "Update an existing document",
		Annotations: idempotentWriteAnnotations,
	}, r.Document.DocumentUpdate)

	addTool(server, &mcp.Tool{
		Name:        "eva_document_delete",
		Description: "Delete a document",
		Annotations: destructiveAnnotations,
	}, r.Document.DocumentDelete)

	addTool(server, &mcp.Tool{
		Name:        "eva_document_count",
		Description: "Count documents",
		Annotations: readOnlyAnnotations,
	}, r.Document.DocumentCount)

	addTool(server, &mcp.Tool{
		Name:        "eva_document_page_tree",
		Description: "Get document page tree hierarchy by root node ID. Returns flat list with parent_id and tree_node_is_branch for building tree structure",
		Annotations: readOnlyAnnotations,
	}, r.Document.DocumentPageTree)

	// Person tools
	addTool(server, &mcp.Tool{
		Name:        "eva_person_list",
		Description: "List persons (users)",
		Annotations: readOnlyAnnotations,
	}, r.Person.PersonList)

	addTool(server, &mcp.Tool{
		Name:        "eva_person_get",
		Description: "Get a single person by ID, login, or email",
		Annotations: readOnlyAnnotations,
	}, r.Person.PersonGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_person_count",
		Description: "Count persons",
		Annotations: readOnlyAnnotations,
	}, r.Person.PersonCount)

	// TimeLog tools
	addTool(server, &mcp.Tool{
		Name:        "eva_timelog_list",
		Description: "List time log entries",
		Annotations: readOnlyAnnotations,
	}, r.TimeLog.TimeLogList)

	addTool(server, &mcp.Tool{
		Name:        "eva_timelog_get",
		Description: "Get a single time log entry by ID",
		Annotations: readOnlyAnnotations,
	}, r.TimeLog.TimeLogGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_timelog_create",
		Description: "Create a new time log entry (time_spent in minutes)",
		Annotations: writeAnnotations,
	}, r.TimeLog.TimeLogCreate)

	addTool(server, &mcp.Tool{
		Name:        "eva_timelog_update",
		Description: "Update an existing time log entry",
		Annotations: idempotentWriteAnnotations,
	}, r.TimeLog.TimeLogUpdate)

	addTool(server, &mcp.Tool{
		Name:        "eva_timelog_delete",
		Description: "Delete a time log entry",
		Annotations: destructiveAnnotations,
	}, r.TimeLog.TimeLogDelete)

	addTool(server, &mcp.Tool{
		Name:        "eva_timelog_count",
		Description: "Count time log entries",
		Annotations: readOnlyAnnotations,
	}, r.TimeLog.TimeLogCount)

	// Comment tools
	addTool(server, &mcp.Tool{
		Name:        "eva_comment_list",
		Description: "List comments",
		Annotations: readOnlyAnnotations,
	}, r.Comment.CommentList)

	addTool(server, &mcp.Tool{
		Name:        "eva_comment_get",
		Description: "Get a single comment by ID",
		Annotations: readOnlyAnnotations,
	}, r.Comment.CommentGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_comment_create",
		Description: "Create a new comment on a task",
		Annotations: writeAnnotations,
	}, r.Comment.CommentCreate)

	addTool(server, &mcp.Tool{
		Name:        "eva_comment_update",
		Description: "Update an existing comment",
		Annotations: idempotentWriteAnnotations,
	}, r.Comment.CommentUpdate)

	addTool(server, &mcp.Tool{
		Name:        "eva_comment_delete",
		Description: "Delete a comment",
		Annotations: destructiveAnnotations,
	}, r.Comment.CommentDelete)

	addTool(server, &mcp.Tool{
		Name:        "eva_comment_count",
		Description: "Count comments",
		Annotations: readOnlyAnnotations,
	}, r.Comment.CommentCount)

	// Epic tools
	addTool(server, &mcp.Tool{
		Name:        "eva_epic_list",
		Description: "List epics",
		Annotations: readOnlyAnnotations,
	}, r.Epic.EpicList)

	addTool(server, &mcp.Tool{
		Name:        "eva_epic_get",
		Description: "Get a single epic by code or ID",
		Annotations: readOnlyAnnotations,
	}, r.Epic.EpicGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_epic_count",
		Description: "Count epics",
		Annotations: readOnlyAnnotations,
	}, r.Epic.EpicCount)

	// TaskLink tools
	addTool(server, &mcp.Tool{
		Name: "eva_tasklink_list",
		Description: "List task links (relationships between tasks). " +
			"Items include in_link/out_link (each a task ID, or an object with id/code/name " +
			"when fields includes '**') and relation_type — the link type code.",
		Annotations: readOnlyAnnotations,
	}, r.TaskLink.TaskLinkList)

	addTool(server, &mcp.Tool{
		Name:        "eva_tasklink_get",
		Description: "Get a single task link by ID",
		Annotations: readOnlyAnnotations,
	}, r.TaskLink.TaskLinkGet)

	addTool(server, &mcp.Tool{
		Name: "eva_tasklink_create",
		Description: "Create a link between two tasks. " +
			"source_task_id/target_task_id accept a task code (e.g. 'TSK-000001') or ID. " +
			"relation_type is the code of the link type, defaulting to 'system.link' ('Относится к').",
		Annotations: writeAnnotations,
	}, r.TaskLink.TaskLinkCreate)

	addTool(server, &mcp.Tool{
		Name:        "eva_tasklink_delete",
		Description: "Delete a task link",
		Annotations: destructiveAnnotations,
	}, r.TaskLink.TaskLinkDelete)

	addTool(server, &mcp.Tool{
		Name:        "eva_tasklink_count",
		Description: "Count task links",
		Annotations: readOnlyAnnotations,
	}, r.TaskLink.TaskLinkCount)

	// StatusHistory tools
	addTool(server, &mcp.Tool{
		Name:        "eva_statushistory_list",
		Description: "List status history entries",
		Annotations: readOnlyAnnotations,
	}, r.StatusHistory.StatusHistoryList)

	addTool(server, &mcp.Tool{
		Name:        "eva_statushistory_get",
		Description: "Get a single status history entry by ID",
		Annotations: readOnlyAnnotations,
	}, r.StatusHistory.StatusHistoryGet)

	addTool(server, &mcp.Tool{
		Name:        "eva_statushistory_count",
		Description: "Count status history entries",
		Annotations: readOnlyAnnotations,
	}, r.StatusHistory.StatusHistoryCount)

	// Stats tools
	addTool(server, &mcp.Tool{
		Name:        "eva_stats_project",
		Description: "Get project statistics (total tasks, open tasks, active sprints, users)",
		Annotations: readOnlyAnnotations,
	}, r.Stats.ProjectStats)

	addTool(server, &mcp.Tool{
		Name:        "eva_stats_sprint",
		Description: "Get sprint statistics (total tasks, tasks by status)",
		Annotations: readOnlyAnnotations,
	}, r.Stats.SprintStats)

	addTool(server, &mcp.Tool{
		Name:        "eva_stats_timespent",
		Description: "Get time spent report grouped by person and task",
		Annotations: readOnlyAnnotations,
	}, r.Stats.TimeSpentStats)

	addTool(server, &mcp.Tool{
		Name:        "eva_stats_sprint_executors_kpi",
		Description: "Get KPI of closed sprint tasks by executor (requires project_code; if sprint_code is empty, aggregates across all project sprints; excludes tasks added during sprint)",
		Annotations: readOnlyAnnotations,
	}, r.Stats.SprintExecutorsKPI)

	// LogicType tools
	addTool(server, &mcp.Tool{
		Name:        "eva_logic_type_list",
		Description: "List logic types (task subtypes like epic/story/task/bug). Filter by cmf_model_name (e.g. 'CmfTask') or code (e.g. 'task.epic:default')",
		Annotations: readOnlyAnnotations,
	}, r.LogicType.LogicTypeList)

	addTool(server, &mcp.Tool{
		Name:        "eva_logic_type_get",
		Description: "Get a single logic type by code (e.g. 'task.epic:default', 'task.userstory:story', 'task.agile:task', 'task.bug:default'). Returns LogicType with ID for use as logic_type_id when creating tasks",
		Annotations: readOnlyAnnotations,
	}, r.LogicType.LogicTypeGet)

	// Tag tools
	addTool(server, &mcp.Tool{
		Name:        "eva_tag_list",
		Description: "List tags available for tasks. Returns tag code (e.g. 'TAG-000004') and name/aliases. Use tag code in the tags field of eva_task_create. Filter by project_id or name.",
		Annotations: readOnlyAnnotations,
	}, r.Tag.TagList)
}
