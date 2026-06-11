// Package list provides the job_phase list page view.
// This is a power-user/debugging surface — there is no sidebar entry for it.
// Phases are normally accessed via the Job detail Phases tab deep links.
package list

import (
	"context"
	"fmt"
	"log"

	job_phase "github.com/erniealice/fayna-golang/domain/operation/job_phase"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
)

// ListViewDeps holds view dependencies for the job_phase list page.
type ListViewDeps struct {
	Routes        job_phase.Routes
	ListJobPhases func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)
	// GetInUseIDs blocks deletion of phases that are referenced by job_task rows.
	GetInUseIDs  func(ctx context.Context, ids []string) (map[string]bool, error)
	Labels       job_phase.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels
}

// PageData holds the data for the job_phase list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the job_phase list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		// 2026-05-14 permission-gates P2a.
		if !perms.Can("job_phase", "list") {
			return view.Forbidden("job_phase:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "pending"
		}

		if deps.ListJobPhases == nil {
			return view.Error(fmt.Errorf("list phases not available"))
		}
		resp, err := deps.ListJobPhases(ctx, &jobphasepb.ListJobPhasesRequest{})
		if err != nil {
			log.Printf("Failed to list job phases: %v", err)
			return view.Error(fmt.Errorf("failed to load phases: %w", err))
		}

		var inUseIDs map[string]bool
		if deps.GetInUseIDs != nil {
			var ids []string
			for _, p := range resp.GetData() {
				ids = append(ids, p.GetId())
			}
			inUseIDs, _ = deps.GetInUseIDs(ctx, ids)
		}

		l := deps.Labels
		columns := phaseColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, inUseIDs, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "job-phases-table",
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
				Label:           l.Buttons.AddPhase,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job_phase", "create"),
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
						ConfirmTitle:     "Delete phases?",
						ConfirmMessage:   "Delete {count} phase(s)?",
						RequiresDataAttr: "deletable",
					},
					{
						Key:      "set_status_pending",
						Label:    "Set Pending",
						Icon:     "icon-clock",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=PHASE_STATUS_PENDING",
					},
					{
						Key:      "set_status_active",
						Label:    "Set Active",
						Icon:     "icon-play",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=PHASE_STATUS_ACTIVE",
					},
					{
						Key:      "set_status_completed",
						Label:    "Set Completed",
						Icon:     "icon-check-circle",
						Endpoint: deps.Routes.BulkSetStatusURL + "?target_status=PHASE_STATUS_COMPLETED",
					},
				},
			},
		}
		types.ApplyTableSettings(tableConfig)

		heading := l.Page.Heading
		switch status {
		case "pending":
			heading = l.Page.HeadingPending
		case "active":
			heading = l.Page.HeadingActive
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
				HeaderIcon:     "icon-layers",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-phase-list-content",
			Table:           tableConfig,
		}

		return view.OK("job-phase-list", pageData)
	})
}

func phaseColumns(l job_phase.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "job", Label: l.Columns.Job},
		{Key: "phase_order", Label: l.Columns.PhaseOrder, WidthClass: "col-sm"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
		{Key: "planned_start", Label: l.Columns.PlannedStart, WidthClass: "col-4xl"},
	}
}

func buildTableRows(
	phases []*jobphasepb.JobPhase,
	status string,
	l job_phase.Labels,
	routes job_phase.Routes,
	inUseIDs map[string]bool,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, p := range phases {
		phaseStatus := phaseStatusString(p.GetStatus())
		if phaseStatus != status {
			continue
		}

		id := p.GetId()
		name := p.GetName()
		jobName := ""
		if j := p.GetJob(); j != nil {
			jobName = j.GetName()
		}

		detailURL := route.ResolveURL(routes.DetailURL, "id", id)
		inUse := inUseIDs[id]
		deleteDisabled := inUse || !perms.Can("job_phase", "delete")
		deleteTooltip := l.Errors.PermissionDenied
		if inUse {
			deleteTooltip = "Cannot delete: referenced by tasks"
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: jobName},
				{Type: "text", Value: fmt.Sprintf("%d", p.GetPhaseOrder())},
				{Type: "badge", Value: phaseStatus, Variant: phaseStatusVariant(phaseStatus)},
				types.DateTimeCell(p.GetPlannedStartString(), types.DateReadable),
			},
			DataAttrs: map[string]string{
				"name":      name,
				"job":       jobName,
				"status":    phaseStatus,
				"deletable": boolAttr(!inUse),
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job_phase", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: deleteDisabled, DisabledTooltip: deleteTooltip},
			},
		})
	}
	return rows
}

func phaseStatusString(s jobphasepb.PhaseStatus) string {
	switch s {
	case jobphasepb.PhaseStatus_PHASE_STATUS_PENDING:
		return "pending"
	case jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE:
		return "active"
	case jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED:
		return "completed"
	default:
		return "pending"
	}
}

func phaseStatusVariant(status string) string {
	switch status {
	case "pending":
		return "warning"
	case "active":
		return "success"
	case "completed":
		return "info"
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
