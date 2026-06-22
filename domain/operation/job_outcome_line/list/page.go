package list

import (
	"context"
	"fmt"
	"log"

	job_outcome_line "github.com/erniealice/fayna-golang/domain/operation/job_outcome_line"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	joboutcomelinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes              job_outcome_line.Routes
	ListJobOutcomeLines func(ctx context.Context, req *joboutcomelinepb.ListJobOutcomeLinesRequest) (*joboutcomelinepb.ListJobOutcomeLinesResponse, error)
	Labels              job_outcome_line.Labels
	CommonLabels        pyeza.CommonLabels
	TableLabels         types.TableLabels
}

// PageData holds the data for the job outcome line list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the job outcome line list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_line", "list") {
			return view.Forbidden("job_outcome_line:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListJobOutcomeLines(ctx, &joboutcomelinepb.ListJobOutcomeLinesRequest{})
		if err != nil {
			log.Printf("Failed to list job outcome lines: %v", err)
			return view.Error(fmt.Errorf("failed to load outcome lines: %w", err))
		}

		l := deps.Labels
		columns := outcomeLineColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "job-outcome-lines-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "label",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   emptyTitle(l, status),
				Message: emptyMessage(l, status),
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddLine,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job_outcome_line", "create"),
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
				HeaderIcon:     "icon-list",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-outcome-line-list-content",
			Table:           tableConfig,
		}

		return view.OK("job-outcome-line-list", pageData)
	})
}

func outcomeLineColumns(l job_outcome_line.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "label", Label: l.Columns.Label},
		{Key: "reporting_role", Label: l.Columns.ReportingRole, WidthClass: "col-3xl"},
		{Key: "output_value", Label: l.Columns.OutputValue, WidthClass: "col-3xl"},
		{Key: "output_label", Label: l.Columns.OutputLabel, WidthClass: "col-3xl"},
		{Key: "active", Label: l.Columns.Active, WidthClass: "col-lg"},
	}
}

func buildTableRows(
	items []*joboutcomelinepb.JobOutcomeLine,
	status string,
	l job_outcome_line.Labels,
	routes job_outcome_line.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, item := range items {
		// Filter by active status
		itemStatus := "inactive"
		if item.GetActive() {
			itemStatus = "active"
		}
		if itemStatus != status {
			continue
		}

		id := item.GetId()
		label := item.GetLabel()
		reportingRole := reportingRoleString(item.GetReportingRole())
		outputValue := fmt.Sprintf("%.2f", item.GetOutputValue())
		outputLabel := item.GetOutputLabel()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		activeVariant := "default"
		if item.GetActive() {
			activeVariant = "success"
		}
		activeLabel := "Inactive"
		if item.GetActive() {
			activeLabel = "Active"
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: label},
				{Type: "badge", Value: reportingRole, Variant: reportingRoleVariant(item.GetReportingRole())},
				{Type: "text", Value: outputValue},
				{Type: "text", Value: outputLabel},
				{Type: "badge", Value: activeLabel, Variant: activeVariant},
			},
			DataAttrs: map[string]string{
				"label":          label,
				"reporting_role": reportingRole,
				"output_value":   outputValue,
				"output_label":   outputLabel,
				"active":         activeLabel,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job_outcome_line", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: label, Disabled: !perms.Can("job_outcome_line", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

// reportingRoleString converts ReportingRole enum to a display string.
func reportingRoleString(r enums.ReportingRole) string {
	switch r {
	case enums.ReportingRole_REPORTING_ROLE_PRIMARY:
		return "Primary"
	case enums.ReportingRole_REPORTING_ROLE_ALTERNATE:
		return "Alternate"
	case enums.ReportingRole_REPORTING_ROLE_TRANSCRIPT:
		return "Transcript"
	case enums.ReportingRole_REPORTING_ROLE_PERCENTILE:
		return "Percentile"
	default:
		return "Unspecified"
	}
}

func reportingRoleVariant(r enums.ReportingRole) string {
	switch r {
	case enums.ReportingRole_REPORTING_ROLE_PRIMARY:
		return "success"
	case enums.ReportingRole_REPORTING_ROLE_TRANSCRIPT:
		return "info"
	case enums.ReportingRole_REPORTING_ROLE_PERCENTILE:
		return "warning"
	default:
		return "default"
	}
}

func statusPageTitle(l job_outcome_line.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l job_outcome_line.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l job_outcome_line.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l job_outcome_line.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
