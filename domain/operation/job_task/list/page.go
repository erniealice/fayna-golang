// Package list provides the job_task list page view.
// This is a power-user/debugging surface — there is no sidebar entry for it.
// Tasks are normally accessed via the JobPhase detail Tasks tab deep links.
package list

import (
	"context"
	"fmt"
	"log"

	job_task "github.com/erniealice/fayna-golang/domain/operation/job_task"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// ListViewDeps holds view dependencies for the job_task list page.
type ListViewDeps struct {
	Routes       job_task.Routes
	ListJobTasks func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	// GetInUseIDs blocks deletion of tasks that are referenced by job_activity rows.
	GetInUseIDs  func(ctx context.Context, ids []string) (map[string]bool, error)
	Labels       job_task.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels
}

// PageData holds the data for the job_task list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the job_task list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		// 2026-05-14 permission-gates P2a: job_task catalog row added in Phase 1b.
		if !perms.Can("job_task", "list") {
			return view.Forbidden("job_task:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "pending"
		}

		if deps.ListJobTasks == nil {
			return view.Error(fmt.Errorf("list tasks not available"))
		}
		resp, err := deps.ListJobTasks(ctx, &jobtaskpb.ListJobTasksRequest{})
		if err != nil {
			log.Printf("Failed to list job tasks: %v", err)
			return view.Error(fmt.Errorf("failed to load tasks: %w", err))
		}

		var inUseIDs map[string]bool
		if deps.GetInUseIDs != nil {
			var ids []string
			for _, t := range resp.GetData() {
				ids = append(ids, t.GetId())
			}
			inUseIDs, _ = deps.GetInUseIDs(ctx, ids)
		}

		l := deps.Labels
		columns := taskColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, inUseIDs, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "job-tasks-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "name",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddTask,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job_task", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
			BulkActions: &types.BulkActionsConfig{
				Enabled:        true,
				SelectAllLabel: "Select all",
				SelectedLabel:  "{count} selected",
				CancelLabel:    "Cancel",
				Actions: []types.BulkAction{
					{
						Key:              "delete",
						Label:            "Delete",
						Icon:             "icon-trash-2",
						Variant:          "danger",
						Endpoint:         deps.Routes.BulkDeleteURL,
						ConfirmTitle:     "Delete tasks?",
						ConfirmMessage:   "Delete {count} task(s)?",
						RequiresDataAttr: "deletable",
					},
					{
						Key:      "set_status_pending",
						Label:    "Set Pending",
						Icon:     "icon-clock",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=TASK_STATUS_PENDING",
					},
					{
						Key:      "set_status_in_progress",
						Label:    "Set In Progress",
						Icon:     "icon-play",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=TASK_STATUS_IN_PROGRESS",
					},
					{
						Key:      "set_status_completed",
						Label:    "Set Completed",
						Icon:     "icon-check-circle",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=TASK_STATUS_COMPLETED",
					},
					{
						Key:      "set_status_hold",
						Label:    "Set Hold",
						Icon:     "icon-pause",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=TASK_STATUS_HOLD",
					},
					{
						Key:      "set_status_rework",
						Label:    "Set Rework",
						Icon:     "icon-refresh-cw",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=TASK_STATUS_REWORK",
					},
				},
			},
		}
		types.ApplyTableSettings(tableConfig)

		heading := l.Page.Heading
		switch status {
		case "pending":
			heading = l.Page.HeadingPending
		case "in_progress":
			heading = l.Page.HeadingInProgress
		case "completed":
			heading = l.Page.HeadingCompleted
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    heading,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-check-square",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-task-list-content",
			Table:           tableConfig,
		}

		return view.OK("job-task-list", pageData)
	})
}

func taskColumns(l job_task.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "phase", Label: l.Columns.Phase},
		{Key: "step_order", Label: l.Columns.StepOrder, WidthClass: "col-sm"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
		{Key: "assigned_to", Label: l.Columns.AssignedTo, WidthClass: "col-4xl"},
		{Key: "percent_complete", Label: l.Columns.PercentComplete, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	tasks []*jobtaskpb.JobTask,
	status string,
	l job_task.Labels,
	routes job_task.Routes,
	inUseIDs map[string]bool,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, t := range tasks {
		taskStatus := taskStatusString(t.GetStatus())
		if taskStatus != status {
			continue
		}

		id := t.GetId()
		name := t.GetName()
		phaseName := ""
		if p := t.GetJobPhase(); p != nil {
			phaseName = p.GetName()
		}
		assignedTo := ""
		if t.AssignedTo != nil {
			assignedTo = *t.AssignedTo
		}
		percentComplete := float64(0)
		if t.PercentComplete != nil {
			percentComplete = *t.PercentComplete
		}

		detailURL := route.ResolveURL(routes.DetailURL, "id", id)
		inUse := inUseIDs[id]
		deleteDisabled := inUse || !perms.Can("job_task", "delete")
		deleteTooltip := l.Errors.PermissionDenied
		if inUse {
			deleteTooltip = "Cannot delete: referenced by activities"
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: phaseName},
				{Type: "text", Value: fmt.Sprintf("%d", t.GetStepOrder())},
				{Type: "badge", Value: taskStatus, Variant: taskStatusVariant(taskStatus)},
				{Type: "text", Value: assignedTo},
				{Type: "text", Value: fmt.Sprintf("%.0f%%", percentComplete)},
			},
			DataAttrs: map[string]string{
				"name":      name,
				"phase":     phaseName,
				"status":    taskStatus,
				"deletable": boolAttr(!inUse),
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job_task", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: deleteDisabled, DisabledTooltip: deleteTooltip},
			},
		})
	}
	return rows
}

func taskStatusString(s jobtaskpb.TaskStatus) string {
	switch s {
	case jobtaskpb.TaskStatus_TASK_STATUS_PENDING:
		return "pending"
	case jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return "in_progress"
	case jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED:
		return "completed"
	case jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED:
		return "skipped"
	case jobtaskpb.TaskStatus_TASK_STATUS_HOLD:
		return "hold"
	case jobtaskpb.TaskStatus_TASK_STATUS_REWORK:
		return "rework"
	default:
		return "pending"
	}
}

func taskStatusVariant(status string) string {
	switch status {
	case "pending":
		return "warning"
	case "in_progress":
		return "info"
	case "completed":
		return "success"
	case "skipped":
		return "default"
	case "hold":
		return "warning"
	case "rework":
		return "danger"
	default:
		return "default"
	}
}

func boolAttr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
