package list

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	checkpointpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/reporting_checkpoint"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes                   reporting_checkpoint.Routes
	ListReportingCheckpoints func(ctx context.Context, req *checkpointpb.ListReportingCheckpointsRequest) (*checkpointpb.ListReportingCheckpointsResponse, error)
	Labels                   reporting_checkpoint.Labels
	CommonLabels             pyeza.CommonLabels
	TableLabels              types.TableLabels
}

// PageData holds the data for the reporting checkpoint list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the reporting checkpoint list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("reporting_checkpoint", "list") {
			return view.Forbidden("reporting_checkpoint:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListReportingCheckpoints(ctx, &checkpointpb.ListReportingCheckpointsRequest{})
		if err != nil {
			log.Printf("Failed to list reporting checkpoints: %v", err)
			return view.Error(fmt.Errorf("failed to load checkpoints: %w", err))
		}

		l := deps.Labels
		columns := checkpointColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "reporting-checkpoint-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "sequence_order",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   emptyTitle(l, status),
				Message: emptyMessage(l, status),
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddCheckpoint,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("reporting_checkpoint", "create"),
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
				HeaderIcon:     "icon-flag",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "reporting-checkpoint-list-content",
			Table:           tableConfig,
		}

		return view.OK("reporting-checkpoint-list", pageData)
	})
}

func checkpointColumns(l reporting_checkpoint.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "label", Label: l.Columns.Label},
		{Key: "role_code", Label: l.Columns.RoleCode, WidthClass: "col-3xl"},
		{Key: "sequence_order", Label: l.Columns.SequenceOrder, WidthClass: "col-lg"},
		{Key: "version", Label: l.Columns.Version, WidthClass: "col-lg"},
		{Key: "version_status", Label: l.Columns.VersionStatus, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*checkpointpb.ReportingCheckpoint,
	status string,
	l reporting_checkpoint.Labels,
	routes reporting_checkpoint.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, c := range items {
		vsStr := versionStatusString(c.GetVersionStatus())
		// filter by status tab
		if !matchesStatus(vsStr, status) {
			continue
		}

		id := c.GetId()
		label := c.GetLabel()
		roleCode := c.GetRoleCode()
		seqOrder := fmt.Sprintf("%d", c.GetSequenceOrder())
		version := fmt.Sprintf("v%d", c.GetVersion())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: label},
				{Type: "text", Value: roleCode},
				{Type: "text", Value: seqOrder},
				{Type: "text", Value: version},
				{Type: "badge", Value: vsStr, Variant: versionStatusVariant(c.GetVersionStatus())},
			},
			DataAttrs: map[string]string{
				"label":          label,
				"role_code":      roleCode,
				"sequence_order": seqOrder,
				"version":        version,
				"version_status": vsStr,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("reporting_checkpoint", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: label, Disabled: !perms.Can("reporting_checkpoint", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func versionStatusString(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "draft"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "active"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "inactive"
	default:
		return "draft"
	}
}

func versionStatusVariant(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "default"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "warning"
	default:
		return "default"
	}
}

// matchesStatus returns true when the checkpoint's derived status matches the tab.
func matchesStatus(vsStr, status string) bool {
	switch status {
	case "active":
		return vsStr == "active"
	case "inactive":
		return vsStr == "inactive" || vsStr == "draft"
	default:
		return true
	}
}

func statusPageTitle(l reporting_checkpoint.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l reporting_checkpoint.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l reporting_checkpoint.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l reporting_checkpoint.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
