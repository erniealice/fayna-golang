package list

import (
	"context"
	"fmt"
	"log"

	jtr "github.com/erniealice/fayna-golang/domain/operation/job_template_relation"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	relationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes                   jtr.Routes
	ListJobTemplateRelations func(ctx context.Context, req *relationpb.ListJobTemplateRelationsRequest) (*relationpb.ListJobTemplateRelationsResponse, error)
	Labels                   jtr.Labels
	CommonLabels             pyeza.CommonLabels
	TableLabels              types.TableLabels
}

// PageData holds the data for the job template relation list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the job template relation list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_relation", "list") {
			return view.Forbidden("job_template_relation:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		var items []*relationpb.JobTemplateRelation
		if deps.ListJobTemplateRelations != nil {
			resp, err := deps.ListJobTemplateRelations(ctx, &relationpb.ListJobTemplateRelationsRequest{})
			if err != nil {
				log.Printf("Failed to list job template relations: %v", err)
				return view.Error(fmt.Errorf("failed to load template relations: %w", err))
			}
			items = resp.GetData()
		}

		l := deps.Labels
		columns := listColumns(l)
		rows := buildTableRows(items, status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "job-template-relations-table",
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
				Label:           l.Buttons.AddRelation,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job_template_relation", "create"),
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
				HeaderIcon:     "icon-git-branch",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-template-relation-list-content",
			Table:           tableConfig,
		}

		return view.OK("job-template-relation-list", pageData)
	})
}

func listColumns(l jtr.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "parent_template_id", Label: l.Columns.ParentTemplateID},
		{Key: "child_template_id", Label: l.Columns.ChildTemplateID},
		{Key: "relation_type", Label: l.Columns.RelationType},
		{Key: "sequence_order", Label: l.Columns.SequenceOrder},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*relationpb.JobTemplateRelation,
	status string,
	l jtr.Labels,
	routes jtr.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, rel := range items {
		activeStatus := "inactive"
		if rel.GetActive() {
			activeStatus = "active"
		}
		if activeStatus != status {
			continue
		}

		id := rel.GetId()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		statusVariant := "warning"
		if rel.GetActive() {
			statusVariant = "success"
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: rel.GetParentTemplateId()},
				{Type: "text", Value: rel.GetChildTemplateId()},
				{Type: "text", Value: rel.GetRelationType().String()},
				{Type: "number", Value: fmt.Sprintf("%d", rel.GetSequenceOrder())},
				{Type: "badge", Value: activeStatus, Variant: statusVariant},
			},
			DataAttrs: map[string]string{
				"parent_template_id": rel.GetParentTemplateId(),
				"child_template_id":  rel.GetChildTemplateId(),
				"status":             activeStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job_template_relation", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: id, Disabled: !perms.Can("job_template_relation", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func statusPageTitle(l jtr.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l jtr.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l jtr.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l jtr.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
