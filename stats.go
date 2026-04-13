/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go/models"
)

// TasksCount returns total tasks matching filters.
func (c *Client) TasksCount(ctx context.Context, kwargs map[string]any) (int64, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.count",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CountResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return 0, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// ProjectTasksCount returns total tasks in project.
func (c *Client) ProjectTasksCount(ctx context.Context, projectID string) (int64, *models.Meta, error) {
	kwargs := map[string]any{
		"filter": []any{TaskFieldProjectID, "==", projectID},
	}
	return c.TasksCount(ctx, kwargs)
}

// SprintTasksCount returns total tasks in sprint by list code.
func (c *Client) SprintTasksCount(ctx context.Context, sprintCode string) (int64, *models.Meta, error) {
	kwargs := map[string]any{
		"filter": []any{TaskFieldLists, "IN", []string{sprintCode}},
	}
	return c.TasksCount(ctx, kwargs)
}

// ListTasksCount returns total tasks in list (sprint/release) by list code.
func (c *Client) ListTasksCount(ctx context.Context, listCode string) (int64, *models.Meta, error) {
	kwargs := map[string]any{
		"filter": []any{TaskFieldLists, "contains", listCode},
	}
	return c.TasksCount(ctx, kwargs)
}

// SprintStats retrieves sprint statistics.
func (c *Client) SprintStats(ctx context.Context, sprintCode string) (*models.SprintStats, error) {
	tasks, _, err := c.SprintTasks(ctx, sprintCode, []string{TaskFieldCacheStatusType})
	if err != nil {
		return nil, err
	}

	stats := &models.SprintStats{
		SprintID:   sprintCode,
		TotalTasks: len(tasks),
	}

	statusCount := make(map[string]int)
	for i := range tasks {
		statusCount[tasks[i].CacheStatusType]++
	}
	stats.TasksByStatus = statusCount

	return stats, nil
}

// ProjectStats retrieves project statistics.
func (c *Client) ProjectStats(ctx context.Context, projectID string) (*models.ProjectStats, *models.Meta, error) {
	stats := &models.ProjectStats{ProjectID: projectID}

	// Total tasks
	count, _, err := c.ProjectTasksCount(ctx, projectID)
	if err == nil {
		stats.TotalTasks = int(count)
	}

	// Open tasks
	qb := NewQueryBuilder().
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Where(sq.Eq{TaskFieldCacheStatusType: models.StatusTypeOpen})
	openCount, err := c.TaskCount(ctx, qb)
	if err == nil {
		stats.OpenTasks = openCount
	}

	// Active sprints
	sprints, _, err := c.OpenProjectSprints(ctx, projectID, []string{ListFieldID})
	if err == nil {
		stats.ActiveSprints = len(sprints)
	}

	// Total users (from project executors)
	qb = NewQueryBuilder().
		Select("executors").
		From(EntityProject).
		Where(sq.Eq{ProjectFieldID: projectID}).
		Limit(1)
	project, _, err := c.ProjectQuery(ctx, qb)
	if err == nil && project != nil {
		stats.TotalUsers = len(project.Executors)
	}

	return stats, nil, nil
}

// TimeSpentStatsParams contains parameters for time spent stats aggregation.
type TimeSpentStatsParams struct {
	ProjectID string // required: "CmfProject:uuid"
	DateFrom  string // optional: "2025-01-01"
	DateTo    string // optional: "2025-12-31"
}

// SprintExecutorsKPIParams contains KPI report settings for sprint executors.
type SprintExecutorsKPIParams struct {
	// SprintCode is optional (example: "SPR-001543").
	SprintCode string
	// ProjectCode is required; if set, must match project.
	ProjectCode string
	// SprintStartDate optionally.
	SprintStartDate time.Time
	// SprintEndDate optionally.
	SprintEndDate time.Time
}

type sprintKPIExecutorAgg struct {
	personID      string
	baselineTasks int
	closedTasks   int
	taskCodes     []string
	sprintNames   map[string]string
}

// TimeSpentStats retrieves aggregated time spent report grouped by person and task.
func (c *Client) TimeSpentStats(ctx context.Context, params TimeSpentStatsParams) (*models.TimeSpentStats, error) {
	logs, err := c.fetchAllProjectTimeLogs(ctx, params)
	if err != nil {
		return nil, err
	}

	personIDs := collectUniquePersonIDs(logs)

	personNames := c.fetchPersonNames(ctx, personIDs)

	return aggregateTimeSpent(logs, personNames, params), nil
}

// SprintExecutorsKPI builds executor KPI for closed tasks in a sprint.
//
// Scope rules:
// - Task belongs to sprint list (`task.lists` contains sprint code or list id)
// - Task closed in sprint date range (`status_closed_at` in [start_date, end_date])
// - Tasks appeared during sprint are excluded:
//   - if BaselineTaskIDs provided: only those IDs are counted
//   - otherwise: task.cmf_created_at must be <= sprint.start_date
func (c *Client) SprintExecutorsKPI(ctx context.Context, params *SprintExecutorsKPIParams) (*models.SprintExecutorsKPI, error) {
	if params.ProjectCode == "" {
		return nil, errors.New("project_code is required")
	}

	project, _, err := c.Project(ctx, params.ProjectCode, []string{ProjectFieldID})
	if err != nil {
		return nil, err
	}

	// get sprints
	qbLists := NewQueryBuilder().
		Select(
			ListFieldID,
		).
		Where(sq.Eq{ListFieldProjectID: project.ID}).
		OrderBy(ListFieldID)

	if params.SprintCode != "" {
		qbLists.Where(sq.Eq{ListFieldCode: params.SprintCode}).Limit(1)
	} else {
		qbLists.Where(sq.Like{ListFieldCode: models.ListSprintPrefix + "%"})
	}

	if !params.SprintStartDate.IsZero() {
		qbLists.Where(sq.GtOrEq{ListFieldPlanStartDate: params.SprintStartDate})
	}

	if !params.SprintEndDate.IsZero() {
		qbLists.Where(sq.LtOrEq{ListFieldPlanEndDate: params.SprintEndDate})
	}

	sprints, _, err := c.ListsList(ctx, qbLists)
	if err != nil {
		return nil, err
	}

	executorAgg := make(map[string]*sprintKPIExecutorAgg)
	baselineTasks := 0
	closedTasks := 0
	unassignedTasks := 0

	for i := range sprints {
		qbSprint := NewQueryBuilder().
			Select(
				ListFieldID,
				ListFieldCode,
				ListFieldName,
				ListFieldProjectID,
				ListFieldPlanStartDate,
				ListFieldPlanEndDate,
			).
			Where(sq.Eq{ListFieldID: sprints[i].ID})
		sprint, _, err := c.ListQuery(ctx, qbSprint)
		if err != nil {
			return nil, err
		}

		qbTasks := NewQueryBuilder().
			Select(
				TaskFieldID,
				TaskFieldCode,
				TaskFieldName,
				TaskFieldCmfCreatedAt,
				TaskFieldStatusClosedAt,
				TaskFieldCacheStatusType,
				TaskFieldLists,
			).
			Where(sq.Eq{TaskFieldProjectID: project.ID}).
			Where(sq.Eq{TaskFieldLists: []string{sprint.ID}}).
			OrderBy(TaskFieldID)
		allSprintTasks, _, err := c.TasksList(ctx, qbTasks)
		if err != nil {
			return nil, err
		}

		for k := range allSprintTasks {
			task := allSprintTasks[k]

			// find assignee for task
			assigneeID, err := c.resolveTaskFirstInProgressOwner(ctx, task.ID)
			if err != nil {
				return nil, err
			}
			if assigneeID == "" {
				topLoggerID, err := c.resolveTaskTopLoggerDuringSprint(ctx, task.ID, sprint.PlanStartDate, sprint.PlanEndDate)
				if err != nil {
					return nil, err
				}
				assigneeID = topLoggerID
			}

			if assigneeID == "" {
				unassignedTasks++
				continue
			}

			agg, ok := executorAgg[assigneeID]
			if !ok {
				agg = &sprintKPIExecutorAgg{
					personID:    assigneeID,
					sprintNames: make(map[string]string),
					taskCodes:   make([]string, 0),
				}
				executorAgg[assigneeID] = agg
			}

			agg.taskCodes = append(agg.taskCodes, task.Code)
			agg.sprintNames[sprint.ID] = sprint.Name
			agg.baselineTasks++
			baselineTasks++
			if task.IsClosedBetween(sprint.PlanStartDate, sprint.PlanEndDate) {
				agg.closedTasks++
				closedTasks++
			}
		}
	}

	personIDs := make([]string, 0, len(executorAgg))
	for personID := range executorAgg {
		personIDs = append(personIDs, personID)
	}
	sort.Strings(personIDs)

	personNames := c.fetchPersonNames(ctx, personIDs)

	report := &models.SprintExecutorsKPI{
		ProjectCode:      params.ProjectCode,
		SprintCode:       params.SprintCode,
		SprintStartDate:  params.SprintStartDate,
		SprintEndDate:    params.SprintEndDate,
		BaselineTasks:    baselineTasks,
		ExcludedNewTasks: 0,
		TotalClosedTasks: closedTasks,
		UnassignedClosed: unassignedTasks,
	}

	report.Executors = make([]models.SprintExecutorKPIEntry, 0, len(executorAgg))
	for _, personID := range personIDs {
		agg := executorAgg[personID]
		sort.Strings(agg.taskCodes)

		personName := personNames[personID]
		if personName == "" {
			personName = personID
		}

		report.Executors = append(report.Executors, models.SprintExecutorKPIEntry{
			PersonID:      personID,
			PersonName:    personName,
			BaselineTasks: agg.baselineTasks,
			ClosedTasks:   agg.closedTasks,
			TaskCodes:     agg.taskCodes,
		})
	}

	return report, nil
}

func (c *Client) resolveTaskFirstInProgressOwner(ctx context.Context, taskID string) (string, error) {
	comment, _, err := c.CommentQuery(ctx,
		NewQueryBuilder().
			Select(CommentFieldAuthorID).
			From(EntityComment).
			Where(sq.Eq{CommentFieldText: "Работа начата"}).
			Where(sq.Eq{CommentFieldParentID: taskID}).
			OrderBy(CommentFieldCmfCreatedAt+" DESC").
			Limit(1),
	)
	if err != nil {
		return "", err
	}

	return comment.AuthorID, nil
}

func (c *Client) resolveTaskTopLoggerDuringSprint(ctx context.Context, taskID string, dateFrom, dateTo time.Time) (string, error) {
	timeByPerson := make(map[string]int)

	for offset := 0; ; offset += timeLogPageSize {
		filters := [][]any{
			{TimeLogFieldParentID, "==", taskID},
		}
		if !dateFrom.IsZero() {
			filters = append(filters, []any{TimeLogFieldCmfCreatedAt, ">=", dateFrom})
		}
		if !dateTo.IsZero() {
			filters = append(filters, []any{TimeLogFieldCmfCreatedAt, "<=", dateTo})
		}

		kwargs := map[string]any{
			"filter": filters,
			"fields": []string{
				TimeLogFieldID,
				TimeLogFieldTimeSpent,
				TimeLogFieldCmfOwnerID,
			},
			"order_by": []string{"-" + TimeLogFieldCmfCreatedAt},
			"slice":    []int{offset, offset + timeLogPageSize},
		}

		page, _, err := c.TimeLogs(ctx, kwargs)
		if err != nil {
			return "", err
		}

		for i := range page {
			personID := strings.TrimSpace(page[i].CmfOwnerID)
			if personID == "" {
				continue
			}
			timeByPerson[personID] += page[i].TimeSpent
		}

		if len(page) < timeLogPageSize {
			break
		}
	}

	if len(timeByPerson) == 0 {
		return "", nil
	}

	topPersonID := ""
	topMinutes := -1
	for personID, minutes := range timeByPerson {
		if minutes > topMinutes {
			topPersonID = personID
			topMinutes = minutes
			continue
		}
		if minutes == topMinutes && personID < topPersonID {
			topPersonID = personID
		}
	}

	return topPersonID, nil
}

const timeLogPageSize = 200

func (c *Client) fetchAllProjectTimeLogs(ctx context.Context, params TimeSpentStatsParams) ([]models.TimeLog, error) {
	var allLogs []models.TimeLog

	for offset := 0; ; offset += timeLogPageSize {
		filters := [][]any{
			{"parent.project_id", "==", params.ProjectID},
		}
		if params.DateFrom != "" {
			filters = append(filters, []any{"cmf_created_at", ">=", params.DateFrom})
		}
		if params.DateTo != "" {
			filters = append(filters, []any{"cmf_created_at", "<=", params.DateTo})
		}

		kwargs := map[string]any{
			"filter":   filters,
			"fields":   []string{"id", "time_spent", "cmf_owner_id", "parent", "parent_id", "cmf_created_at"},
			"order_by": []string{"-cmf_created_at"},
			"slice":    []int{offset, offset + timeLogPageSize},
		}

		page, _, err := c.TimeLogs(ctx, kwargs)
		if err != nil {
			return nil, err
		}

		allLogs = append(allLogs, page...)

		if len(page) < timeLogPageSize {
			break
		}
	}

	return allLogs, nil
}

func collectUniquePersonIDs(logs []models.TimeLog) []string {
	seen := make(map[string]struct{})
	var ids []string

	for i := range logs {
		pid := logs[i].CmfOwnerID
		if pid == "" {
			continue
		}
		if _, ok := seen[pid]; !ok {
			seen[pid] = struct{}{}
			ids = append(ids, pid)
		}
	}

	return ids
}

func (c *Client) fetchPersonNames(ctx context.Context, ids []string) map[string]string {
	names := make(map[string]string, len(ids))

	for _, id := range ids {
		person, _, err := c.Person(ctx, id, []string{"id", "name"})
		if err != nil {
			names[id] = id
			continue
		}
		names[id] = person.Name
	}

	return names
}

func aggregateTimeSpent(logs []models.TimeLog, personNames map[string]string, params TimeSpentStatsParams) *models.TimeSpentStats {
	personTasks := make(map[string]map[string]*models.TimeSpentTaskEntry)

	for i := range logs {
		log := &logs[i]
		pid := log.CmfOwnerID
		if pid == "" {
			continue
		}

		taskID := log.ParentID
		var taskCode, taskName string
		if log.Parent != nil {
			taskCode = log.Parent.Code
			taskName = log.Parent.Name
		}

		tasks, ok := personTasks[pid]
		if !ok {
			tasks = make(map[string]*models.TimeSpentTaskEntry)
			personTasks[pid] = tasks
		}

		entry, ok := tasks[taskID]
		if !ok {
			entry = &models.TimeSpentTaskEntry{
				TaskID:   taskID,
				TaskCode: taskCode,
				TaskName: taskName,
			}
			tasks[taskID] = entry
		}
		entry.TimeSpent += log.TimeSpent
	}

	persons := make([]models.TimeSpentPersonEntry, 0, len(personTasks))
	grandTotal := 0

	for pid, tasks := range personTasks {
		var taskEntries []models.TimeSpentTaskEntry
		personTotal := 0

		for _, entry := range tasks {
			taskEntries = append(taskEntries, *entry)
			personTotal += entry.TimeSpent
		}

		sort.Slice(taskEntries, func(i, j int) bool {
			return taskEntries[i].TimeSpent > taskEntries[j].TimeSpent
		})

		name := personNames[pid]
		if name == "" {
			name = pid
		}

		persons = append(persons, models.TimeSpentPersonEntry{
			PersonID:   pid,
			PersonName: name,
			Tasks:      taskEntries,
			TotalTime:  personTotal,
		})

		grandTotal += personTotal
	}

	sort.Slice(persons, func(i, j int) bool {
		return persons[i].TotalTime > persons[j].TotalTime
	})

	return &models.TimeSpentStats{
		ProjectID:      params.ProjectID,
		DateFrom:       params.DateFrom,
		DateTo:         params.DateTo,
		Persons:        persons,
		GrandTotalTime: grandTotal,
	}
}
