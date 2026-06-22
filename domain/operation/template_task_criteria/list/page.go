package list

import (
	"context"
	"fmt"
	"log"

	ttc "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes                    ttc.Routes
	ListTemplateTaskCriterias func(ctx context.Context, req *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error)
	Labels                    ttc.Labels
	CommonLabels              pyeza.CommonLabels
	TableLabels               types.TableLabels
}

// PageData holds the data for the template task criteria list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the template task criteria list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "list") {
			return view.Forbidden("template_task_criteria:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListTemplateTaskCriterias(ctx, &ttcpb.ListTemplateTaskCriteriasRequest{})
		if err != nil {
			log.Printf("Failed to list template task criterias: %v", err)
			return view.Error(fmt.Errorf("failed to load template task criteria: %w", err))
		}

		l := deps.Labels
		columns := listColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "template-task-criteria-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "job_template_task_id",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   emptyTitle(l, status),
				Message: emptyMessage(l, status),
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddLink,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("template_task_criteria", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          statusPageTitle(l, status),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    statusPageTitle(l, status),
				HeaderSubtitle: statusPageCaption(l, status),
				HeaderIcon:     "icon-link",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "template-task-criteria-list-content",
			Table:           tableConfig,
		}

		return view.OK("template-task-criteria-list", pageData)
	})
}

func listColumns(l ttc.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "job_template_task_id", Label: l.Columns.JobTemplateTaskID},
		{Key: "outcome_criteria_id", Label: l.Columns.OutcomeCriteriaID},
		{Key: "sequence_order", Label: l.Columns.SequenceOrder},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*ttcpb.TemplateTaskCriteria,
	status string,
	l ttc.Labels,
	routes ttc.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, c := range items {
		activeStatus := "inactive"
		if c.GetActive() {
			activeStatus = "active"
		}
		if activeStatus != status {
			continue
		}

		id := c.GetId()
		taskID := c.GetJobTemplateTaskId()
		criteriaID := c.GetOutcomeCriteriaId()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		statusVariant := "warning"
		if c.GetActive() {
			statusVariant = "success"
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: taskID},
				{Type: "text", Value: criteriaID},
				{Type: "number", Value: fmt.Sprintf("%d", c.GetSequenceOrder())},
				{Type: "badge", Value: activeStatus, Variant: statusVariant},
			},
			DataAttrs: map[string]string{
				"job_template_task_id": taskID,
				"outcome_criteria_id":  criteriaID,
				"status":               activeStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("template_task_criteria", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: id, Disabled: !perms.Can("template_task_criteria", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func statusPageTitle(l ttc.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l ttc.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l ttc.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l ttc.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
