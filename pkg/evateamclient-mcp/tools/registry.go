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

// RegisterAll registers all tools with the MCP server.
func (r *Registry) RegisterAll(server *mcp.Server) {
	// Task tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_list",
		Description: "List tasks with optional filters (project, status, sprint, responsible)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Task.TaskList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_get",
		Description: "Get a single task by code (e.g., 'PROJ-123') or ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Task.TaskGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_create",
		Description: "Create a new task",
		Annotations: writeAnnotations,
	}, wrapHandler(r.Task.TaskCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_update",
		Description: "Update an existing task",
		Annotations: idempotentWriteAnnotations,
	}, wrapHandler(r.Task.TaskUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_delete",
		Description: "Delete a task",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.Task.TaskDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_update_status",
		Description: "Update task status (OPEN, IN_PROGRESS, CLOSED)",
		Annotations: idempotentWriteAnnotations,
	}, wrapHandler(r.Task.TaskUpdateStatus))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_archive",
		Description: "Archive a task (soft delete)",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.Task.TaskArchive))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_count",
		Description: "Count tasks matching filters",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Task.TaskCount))

	// Project tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_list",
		Description: "List projects",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Project.ProjectList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_get",
		Description: "Get a single project by code or ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Project.ProjectGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_create",
		Description: "Create a new project",
		Annotations: writeAnnotations,
	}, wrapHandler(r.Project.ProjectCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_update",
		Description: "Update an existing project",
		Annotations: idempotentWriteAnnotations,
	}, wrapHandler(r.Project.ProjectUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_delete",
		Description: "Delete a project",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.Project.ProjectDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_add_executor",
		Description: "Add an executor to a project",
		Annotations: writeAnnotations,
	}, wrapHandler(r.Project.ProjectAddExecutor))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_remove_executor",
		Description: "Remove an executor from a project",
		Annotations: writeAnnotations,
	}, wrapHandler(r.Project.ProjectRemoveExecutor))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_count",
		Description: "Count projects",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Project.ProjectCount))

	// List tools (sprints/releases)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_list",
		Description: "List all lists (sprints and releases) with optional filters",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.List.ListList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_get",
		Description: "Get a single list by code (e.g., 'SPR-001543') or ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.List.ListGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_create",
		Description: "Create a new list (sprint/release)",
		Annotations: writeAnnotations,
	}, wrapHandler(r.List.ListCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_update",
		Description: "Update an existing list",
		Annotations: idempotentWriteAnnotations,
	}, wrapHandler(r.List.ListUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_close",
		Description: "Close a list (sprint/release)",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.List.ListClose))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_delete",
		Description: "Delete a list",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.List.ListDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_count",
		Description: "Count lists",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.List.ListCount))

	// Sprint aliases
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_sprint_list",
		Description: "List sprints (alias for eva_list_list with type=sprint)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.List.SprintList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_sprint_get",
		Description: "Get a single sprint by code (e.g., 'SPR-001543')",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.List.SprintGet))

	// Release aliases
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_release_list",
		Description: "List releases (alias for eva_list_list with type=release)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.List.ReleaseList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_release_get",
		Description: "Get a single release by code (e.g., 'REL-001641')",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.List.ReleaseGet))

	// Document tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_list",
		Description: "List documents",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Document.DocumentList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_get",
		Description: "Get a single document by code or ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Document.DocumentGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_create",
		Description: "Create a new document",
		Annotations: writeAnnotations,
	}, wrapHandler(r.Document.DocumentCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_update",
		Description: "Update an existing document",
		Annotations: idempotentWriteAnnotations,
	}, wrapHandler(r.Document.DocumentUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_delete",
		Description: "Delete a document",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.Document.DocumentDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_count",
		Description: "Count documents",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Document.DocumentCount))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_page_tree",
		Description: "Get document page tree hierarchy by root node ID. Returns flat list with parent_id and tree_node_is_branch for building tree structure",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Document.DocumentPageTree))

	// Person tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_person_list",
		Description: "List persons (users)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Person.PersonList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_person_get",
		Description: "Get a single person by ID, login, or email",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Person.PersonGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_person_count",
		Description: "Count persons",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Person.PersonCount))

	// TimeLog tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_list",
		Description: "List time log entries",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.TimeLog.TimeLogList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_get",
		Description: "Get a single time log entry by ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.TimeLog.TimeLogGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_create",
		Description: "Create a new time log entry (time_spent in minutes)",
		Annotations: writeAnnotations,
	}, wrapHandler(r.TimeLog.TimeLogCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_update",
		Description: "Update an existing time log entry",
		Annotations: idempotentWriteAnnotations,
	}, wrapHandler(r.TimeLog.TimeLogUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_delete",
		Description: "Delete a time log entry",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.TimeLog.TimeLogDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_count",
		Description: "Count time log entries",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.TimeLog.TimeLogCount))

	// Comment tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_list",
		Description: "List comments",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Comment.CommentList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_get",
		Description: "Get a single comment by ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Comment.CommentGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_create",
		Description: "Create a new comment on a task",
		Annotations: writeAnnotations,
	}, wrapHandler(r.Comment.CommentCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_update",
		Description: "Update an existing comment",
		Annotations: idempotentWriteAnnotations,
	}, wrapHandler(r.Comment.CommentUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_delete",
		Description: "Delete a comment",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.Comment.CommentDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_count",
		Description: "Count comments",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Comment.CommentCount))

	// Epic tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_epic_list",
		Description: "List epics",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Epic.EpicList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_epic_get",
		Description: "Get a single epic by code or ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Epic.EpicGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_epic_count",
		Description: "Count epics",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Epic.EpicCount))

	// TaskLink tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_list",
		Description: "List task links (relationships between tasks)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.TaskLink.TaskLinkList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_get",
		Description: "Get a single task link by ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.TaskLink.TaskLinkGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_create",
		Description: "Create a new task link",
		Annotations: writeAnnotations,
	}, wrapHandler(r.TaskLink.TaskLinkCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_delete",
		Description: "Delete a task link",
		Annotations: destructiveAnnotations,
	}, wrapHandler(r.TaskLink.TaskLinkDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_count",
		Description: "Count task links",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.TaskLink.TaskLinkCount))

	// StatusHistory tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_statushistory_list",
		Description: "List status history entries",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.StatusHistory.StatusHistoryList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_statushistory_get",
		Description: "Get a single status history entry by ID",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.StatusHistory.StatusHistoryGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_statushistory_count",
		Description: "Count status history entries",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.StatusHistory.StatusHistoryCount))

	// Stats tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_stats_project",
		Description: "Get project statistics (total tasks, open tasks, active sprints, users)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Stats.ProjectStats))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_stats_sprint",
		Description: "Get sprint statistics (total tasks, tasks by status)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Stats.SprintStats))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_stats_timespent",
		Description: "Get time spent report grouped by person and task",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Stats.TimeSpentStats))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_stats_sprint_executors_kpi",
		Description: "Get KPI of closed sprint tasks by executor (requires project_code; if sprint_code is empty, aggregates across all project sprints; excludes tasks added during sprint)",
		Annotations: readOnlyAnnotations,
	}, wrapHandler(r.Stats.SprintExecutorsKPI))
}
