package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raoptimus/evateamclient.go"
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
	}, wrapHandler(r.Task.TaskList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_get",
		Description: "Get a single task by code (e.g., 'PROJ-123') or ID",
	}, wrapHandler(r.Task.TaskGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_create",
		Description: "Create a new task",
	}, wrapHandler(r.Task.TaskCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_update",
		Description: "Update an existing task",
	}, wrapHandler(r.Task.TaskUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_delete",
		Description: "Delete a task",
	}, wrapHandler(r.Task.TaskDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_update_status",
		Description: "Update task status (OPEN, IN_PROGRESS, CLOSED)",
	}, wrapHandler(r.Task.TaskUpdateStatus))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_archive",
		Description: "Archive a task (soft delete)",
	}, wrapHandler(r.Task.TaskArchive))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_task_count",
		Description: "Count tasks matching filters",
	}, wrapHandler(r.Task.TaskCount))

	// Project tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_list",
		Description: "List projects",
	}, wrapHandler(r.Project.ProjectList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_get",
		Description: "Get a single project by code or ID",
	}, wrapHandler(r.Project.ProjectGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_create",
		Description: "Create a new project",
	}, wrapHandler(r.Project.ProjectCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_update",
		Description: "Update an existing project",
	}, wrapHandler(r.Project.ProjectUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_delete",
		Description: "Delete a project",
	}, wrapHandler(r.Project.ProjectDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_add_executor",
		Description: "Add an executor to a project",
	}, wrapHandler(r.Project.ProjectAddExecutor))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_remove_executor",
		Description: "Remove an executor from a project",
	}, wrapHandler(r.Project.ProjectRemoveExecutor))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_project_count",
		Description: "Count projects",
	}, wrapHandler(r.Project.ProjectCount))

	// List tools (sprints/releases)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_list",
		Description: "List all lists (sprints and releases) with optional filters",
	}, wrapHandler(r.List.ListList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_get",
		Description: "Get a single list by code (e.g., 'SPR-001543') or ID",
	}, wrapHandler(r.List.ListGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_create",
		Description: "Create a new list (sprint/release)",
	}, wrapHandler(r.List.ListCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_update",
		Description: "Update an existing list",
	}, wrapHandler(r.List.ListUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_close",
		Description: "Close a list (sprint/release)",
	}, wrapHandler(r.List.ListClose))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_delete",
		Description: "Delete a list",
	}, wrapHandler(r.List.ListDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_list_count",
		Description: "Count lists",
	}, wrapHandler(r.List.ListCount))

	// Sprint aliases
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_sprint_list",
		Description: "List sprints (alias for eva_list_list with type=sprint)",
	}, wrapHandler(r.List.SprintList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_sprint_get",
		Description: "Get a single sprint by code (e.g., 'SPR-001543')",
	}, wrapHandler(r.List.SprintGet))

	// Release aliases
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_release_list",
		Description: "List releases (alias for eva_list_list with type=release)",
	}, wrapHandler(r.List.ReleaseList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_release_get",
		Description: "Get a single release by code (e.g., 'REL-001641')",
	}, wrapHandler(r.List.ReleaseGet))

	// Document tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_list",
		Description: "List documents",
	}, wrapHandler(r.Document.DocumentList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_get",
		Description: "Get a single document by code or ID",
	}, wrapHandler(r.Document.DocumentGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_create",
		Description: "Create a new document",
	}, wrapHandler(r.Document.DocumentCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_update",
		Description: "Update an existing document",
	}, wrapHandler(r.Document.DocumentUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_delete",
		Description: "Delete a document",
	}, wrapHandler(r.Document.DocumentDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_document_count",
		Description: "Count documents",
	}, wrapHandler(r.Document.DocumentCount))

	// Person tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_person_list",
		Description: "List persons (users)",
	}, wrapHandler(r.Person.PersonList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_person_get",
		Description: "Get a single person by ID, login, or email",
	}, wrapHandler(r.Person.PersonGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_person_count",
		Description: "Count persons",
	}, wrapHandler(r.Person.PersonCount))

	// TimeLog tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_list",
		Description: "List time log entries",
	}, wrapHandler(r.TimeLog.TimeLogList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_get",
		Description: "Get a single time log entry by ID",
	}, wrapHandler(r.TimeLog.TimeLogGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_create",
		Description: "Create a new time log entry (time_spent in minutes)",
	}, wrapHandler(r.TimeLog.TimeLogCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_update",
		Description: "Update an existing time log entry",
	}, wrapHandler(r.TimeLog.TimeLogUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_delete",
		Description: "Delete a time log entry",
	}, wrapHandler(r.TimeLog.TimeLogDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_timelog_count",
		Description: "Count time log entries",
	}, wrapHandler(r.TimeLog.TimeLogCount))

	// Comment tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_list",
		Description: "List comments",
	}, wrapHandler(r.Comment.CommentList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_get",
		Description: "Get a single comment by ID",
	}, wrapHandler(r.Comment.CommentGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_create",
		Description: "Create a new comment on a task",
	}, wrapHandler(r.Comment.CommentCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_update",
		Description: "Update an existing comment",
	}, wrapHandler(r.Comment.CommentUpdate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_delete",
		Description: "Delete a comment",
	}, wrapHandler(r.Comment.CommentDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_comment_count",
		Description: "Count comments",
	}, wrapHandler(r.Comment.CommentCount))

	// Epic tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_epic_list",
		Description: "List epics",
	}, wrapHandler(r.Epic.EpicList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_epic_get",
		Description: "Get a single epic by code or ID",
	}, wrapHandler(r.Epic.EpicGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_epic_count",
		Description: "Count epics",
	}, wrapHandler(r.Epic.EpicCount))

	// TaskLink tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_list",
		Description: "List task links (relationships between tasks)",
	}, wrapHandler(r.TaskLink.TaskLinkList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_get",
		Description: "Get a single task link by ID",
	}, wrapHandler(r.TaskLink.TaskLinkGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_create",
		Description: "Create a new task link",
	}, wrapHandler(r.TaskLink.TaskLinkCreate))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_delete",
		Description: "Delete a task link",
	}, wrapHandler(r.TaskLink.TaskLinkDelete))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_tasklink_count",
		Description: "Count task links",
	}, wrapHandler(r.TaskLink.TaskLinkCount))

	// StatusHistory tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_statushistory_list",
		Description: "List status history entries",
	}, wrapHandler(r.StatusHistory.StatusHistoryList))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_statushistory_get",
		Description: "Get a single status history entry by ID",
	}, wrapHandler(r.StatusHistory.StatusHistoryGet))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_statushistory_count",
		Description: "Count status history entries",
	}, wrapHandler(r.StatusHistory.StatusHistoryCount))

	// Stats tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_stats_project",
		Description: "Get project statistics (total tasks, open tasks, active sprints, users)",
	}, wrapHandler(r.Stats.ProjectStats))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "eva_stats_sprint",
		Description: "Get sprint statistics (total tasks, tasks by status)",
	}, wrapHandler(r.Stats.SprintStats))
}
