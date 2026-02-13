package evateamclient

import (
	"context"
	"errors"
	"fmt"
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
		"filter": []any{TaskFieldLists, "contains", sprintCode},
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
	// SprintStartDate optionally overrides sprint.start_date from EVA list.
	SprintStartDate time.Time
	// SprintEndDate optionally overrides sprint.end_date from EVA list.
	SprintEndDate time.Time
}

type sprintKPIExecutorAgg struct {
	personID      string
	baselineTasks int
	closedTasks   int
	taskCodes     []string
	sprints       map[string]models.List
}

// TimeSpentStats retrieves aggregated time spent report grouped by person and task.
func (c *Client) TimeSpentStats(ctx context.Context, params TimeSpentStatsParams) (*models.TimeSpentStats, error) {
	logs, err := c.fetchAllProjectTimeLogs(ctx, params)
	if err != nil {
		return nil, err
	}

	personIDs := collectUniquePersonIDs(logs)

	personNames, err := c.fetchPersonNames(ctx, personIDs)
	if err != nil {
		return nil, err
	}

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
func (c *Client) SprintExecutorsKPI(ctx context.Context, params SprintExecutorsKPIParams) (*models.SprintExecutorsKPI, error) {
	if params.ProjectCode == "" {
		return nil, errors.New("project_id is required")
	}

	project, _, err := c.Project(ctx, params.ProjectCode, []string{TaskFieldProjectID})
	if err != nil {
		return nil, err
	}

	// get sprints
	qb := NewQueryBuilder().
		Select(
			ListFieldID,
			ListFieldCode,
			ListFieldName,
			ListFieldProjectID,
			ListFieldStartDate,
			ListFieldEndDate,
		).
		Where(sq.Eq{ListFieldProjectID: project.ID}).
		OrderBy(ListFieldID)

	if params.SprintCode != "" {
		qb.Where(sq.Eq{ListFieldCode: params.SprintCode}).Limit(1)
	} else {
		qb.Where(sq.Like{ListFieldCode: models.ListSprintPrefix})
	}

	if !params.SprintStartDate.IsZero() {
		qb.Where(sq.GtOrEq{ListFieldStartDate: params.SprintStartDate})
	}

	if !params.SprintEndDate.IsZero() {
		qb.Where(sq.LtOrEq{ListFieldEndDate: params.SprintStartDate})
	}

	sprints, _, err := c.ListsList(ctx, qb)
	if err != nil {
		return nil, err
	}

	executorAgg := make(map[string]*sprintKPIExecutorAgg)
	baselineTasks := 0
	closedTasks := 0
	unassignedTasks := 0

	for i := range sprints {
		sprint := sprints[i]

		allSprintTasks, _, err := c.TasksList(
			ctx,
			NewQueryBuilder().
				Select(
					TaskFieldID,
					TaskFieldCode,
					TaskFieldName,
					TaskFieldCmfCreatedAt,
					TaskFieldStatusClosedAt,
					TaskFieldCacheStatusType,
					TaskFieldLists,
				).
				Where(sq.Eq{TaskFieldProjectID: sprint.ProjectID}).
				Where(sq.Eq{TaskFieldLists: []string{sprint.ID}}).
				OrderBy(TaskFieldID),
		)
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
				topLoggerID, err := c.resolveTaskTopLoggerDuringSprint(ctx, task.ID, sprint.StartDate, sprint.EndDate)
				if err != nil {
					return nil, err
				}
				assigneeID = topLoggerID
			}

			if assigneeID != "" {
				unassignedTasks++
				continue
			}

			agg, ok := executorAgg[assigneeID]
			if !ok {
				agg = &sprintKPIExecutorAgg{personID: assigneeID}
				executorAgg[assigneeID] = agg
			}

			agg.taskCodes = append(agg.taskCodes, task.Code)
			agg.sprints[sprint.ID] = sprint
			agg.baselineTasks++
			baselineTasks++
			if task.IsClosedBetween(sprint.StartDate, sprint.EndDate) {
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

	personNames, err := c.fetchPersonNames(ctx, personIDs)
	if err != nil {
		return nil, err
	}

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
			PersonID:    personID,
			PersonName:  personName,
			ClosedTasks: agg.closedTasks,
			TaskCodes:   agg.taskCodes,
		})
	}

	return report, nil
}

func parseEVATime(raw string) (time.Time, bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false, fmt.Errorf("empty time value")
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		ts, err := time.Parse(layout, value)
		if err == nil {
			return ts, layout == "2006-01-02", nil
		}
	}

	return time.Time{}, false, fmt.Errorf("unsupported time format: %q", value)
}

func (c *Client) resolveTaskFirstInProgressOwner(ctx context.Context, taskID string) (string, error) {
	for offset := 0; ; offset += statusHistoryPageSize {
		kwargs := map[string]any{
			"filter": [][]any{
				{StatusHistoryFieldParentID, "==", taskID},
				{StatusHistoryFieldNewStatus, "==", models.StatusTypeInProgress},
			},
			"fields": []string{
				StatusHistoryFieldID,
				StatusHistoryFieldParentID,
				StatusHistoryFieldNewStatus,
				StatusHistoryFieldCmfOwnerID,
				StatusHistoryFieldCmfCreatedAt,
			},
			"order_by": []string{StatusHistoryFieldCmfCreatedAt},
			"slice":    []int{offset, offset + statusHistoryPageSize},
		}

		page, _, err := c.StatusHistories(ctx, kwargs)
		if err != nil {
			return "", err
		}

		for i := range page {
			if strings.TrimSpace(page[i].CmfOwnerID) != "" {
				return page[i].CmfOwnerID, nil
			}
		}

		if len(page) < statusHistoryPageSize {
			break
		}
	}

	return "", nil
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
const statusHistoryPageSize = 200

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

func (c *Client) fetchPersonNames(ctx context.Context, ids []string) (map[string]string, error) {
	names := make(map[string]string, len(ids))

	for _, id := range ids {
		person, _, err := c.Person(ctx, id, []string{"id", "name"})
		if err != nil {
			names[id] = id
			continue
		}
		names[id] = person.Name
	}

	return names, nil
}

func aggregateTimeSpent(logs []models.TimeLog, personNames map[string]string, params TimeSpentStatsParams) *models.TimeSpentStats {
	type taskKey struct {
		personID string
		taskID   string
	}

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

	var persons []models.TimeSpentPersonEntry
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
