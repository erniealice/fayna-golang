package list

import (
	"context"
	"fmt"
	"log"

	evaluation_template "github.com/erniealice/fayna-golang/domain/operation/evaluation_template"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
	templatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template"
)

// ListViewDeps holds view dependencies for the evaluation template list.
type ListViewDeps struct {
	Routes                      evaluation_template.Routes
	ListEvaluationTemplates     func(ctx context.Context, req *templatepb.ListEvaluationTemplatesRequest) (*templatepb.ListEvaluationTemplatesResponse, error)
	ListEvaluationTemplateItems func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)
	Labels                      evaluation_template.Labels
	CommonLabels                pyeza.CommonLabels
	TableLabels                 types.TableLabels
}

// PageData holds the data for the evaluation template list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the evaluation template list view (staff-only; L3 gate).
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "list") {
			return view.Forbidden("evaluation_template:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListEvaluationTemplates(ctx, &templatepb.ListEvaluationTemplatesRequest{})
		if err != nil {
			log.Printf("Failed to list evaluation templates: %v", err)
			return view.Error(fmt.Errorf("failed to load templates: %w", err))
		}

		l := deps.Labels
		columns := templateColumns(l)
		rows := buildTableRows(ctx, deps, resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := buildTableConfig(deps, l, columns, rows, status, perms, true)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          statusPageTitle(l, status),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    statusPageTitle(l, status),
				HeaderSubtitle: statusPageCaption(l, status),
				HeaderIcon:     "icon-clipboard",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "evaluation_template-list-content",
			Table:           tableConfig,
		}

		return view.OK("evaluation_template-list", pageData)
	})
}

// NewTableView creates the table-only partial view for HTMX table swaps
// (GET /action/evaluation_template/table/{status}).
func NewTableView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "list") {
			return view.Forbidden("evaluation_template:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListEvaluationTemplates(ctx, &templatepb.ListEvaluationTemplatesRequest{})
		if err != nil {
			log.Printf("Failed to list evaluation templates: %v", err)
			return view.Error(fmt.Errorf("failed to load templates: %w", err))
		}

		l := deps.Labels
		columns := templateColumns(l)
		rows := buildTableRows(ctx, deps, resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := buildTableConfig(deps, l, columns, rows, status, perms, false)

		return view.OK("table-card", tableConfig)
	})
}

func buildTableConfig(
	deps *ListViewDeps,
	l evaluation_template.Labels,
	columns []types.TableColumn,
	rows []types.TableRow,
	status string,
	perms *types.UserPermissions,
	withPrimary bool,
) *types.TableConfig {
	cfg := &types.TableConfig{
		ID:                   "evaluation_template-list-table",
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
		RefreshURL:           route.ResolveURL(deps.Routes.TableURL, "status", status),
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   emptyTitle(l, status),
			Message: emptyMessage(l, status),
		},
		BulkActions: &types.BulkActionsConfig{
			Enabled:        true,
			SelectAllLabel: "Select all",
			SelectedLabel:  "selected",
			CancelLabel:    "Cancel",
			Actions: []types.BulkAction{
				{
					Key:              "bulk-deprecate",
					Label:            l.Actions.Deprecate,
					Icon:             "icon-archive",
					Variant:          "warning",
					Endpoint:         deps.Routes.BulkDeprecateURL,
					RequiresDataAttr: "deprecatable",
					Disabled:         !perms.Can("evaluation_template", "update"),
					DisabledTooltip:  l.Errors.PermissionDenied,
				},
			},
		},
	}
	if withPrimary {
		cfg.PrimaryAction = &types.PrimaryAction{
			Label:           l.Buttons.AddTemplate,
			ActionURL:       deps.Routes.AddURL,
			Icon:            "icon-plus",
			Disabled:        !perms.Can("evaluation_template", "create"),
			DisabledTooltip: l.Errors.PermissionDenied,
		}
	}
	types.ApplyTableSettings(cfg)
	return cfg
}

func templateColumns(l evaluation_template.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "evaluation_type", Label: l.Columns.EvaluationType, WidthClass: "col-3xl"},
		{Key: "relationship_type", Label: l.Columns.RelationshipType, WidthClass: "col-4xl"},
		{Key: "version", Label: l.Columns.Version, WidthClass: "col-lg"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
		{Key: "visibility", Label: l.Columns.Visibility, WidthClass: "col-3xl"},
		{Key: "items", Label: l.Columns.Items, WidthClass: "col-lg", SortKind: "number"},
		{Key: "created", Label: l.Columns.Created, WidthClass: "col-3xl", SortKind: "date"},
	}
}

func buildTableRows(
	ctx context.Context,
	deps *ListViewDeps,
	items []*templatepb.EvaluationTemplate,
	status string,
	l evaluation_template.Labels,
	routes evaluation_template.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	// One items query for the whole page; bucket counts by template id so the
	// Items column does not issue N round-trips.
	itemCounts := loadItemCounts(ctx, deps)

	rows := []types.TableRow{}
	for _, t := range items {
		st := statusString(t.GetStatus())
		if status != "all" && st != status {
			continue
		}

		id := t.GetId()
		name := t.GetName()
		evalType := evaluationTypeString(t.GetEvaluationType())
		relType := relationshipTypeString(t.GetRelationshipType())
		visibility := visibilityString(t.GetVisibilityType())
		version := fmt.Sprintf("v%d", t.GetVersion())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		actions := []types.TableAction{
			{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
		}
		// Activate (DRAFT only — BLOCKER-2 weighted-non-numeric guard is enforced
		// server-side in the activate action).
		if st == "draft" {
			actions = append(actions, types.TableAction{
				Type: "check", Label: l.Actions.Activate, Action: "activate",
				URL: route.ResolveURL(routes.ActivateURL, "id", id), ItemName: name,
				Disabled: !perms.Can("evaluation_template", "update"), DisabledTooltip: l.Errors.PermissionDenied,
			})
		}
		// Deprecate (ACTIVE only).
		if st == "active" {
			actions = append(actions, types.TableAction{
				Type: "archive", Label: l.Actions.Deprecate, Action: "deprecate",
				URL: route.ResolveURL(routes.DeprecateURL, "id", id), ItemName: name,
				Disabled: !perms.Can("evaluation_template", "update"), DisabledTooltip: l.Errors.PermissionDenied,
			})
		}
		// Clone (always — clone-flips-to-create per permission-reflection).
		actions = append(actions, types.TableAction{
			Type: "clone", Label: l.Actions.Clone, Action: "clone",
			URL: route.ResolveURL(routes.CloneURL, "id", id), ItemName: name,
			Disabled: !perms.Can("evaluation_template", "create"), DisabledTooltip: l.Errors.PermissionDenied,
		})

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "badge", Value: evalType, Variant: "info"},
				{Type: "text", Value: relType},
				{Type: "text", Value: version},
				{Type: "badge", Value: st, Variant: statusVariant(t.GetStatus())},
				{Type: "text", Value: visibility},
				{Type: "number", Value: fmt.Sprintf("%d", itemCounts[id])},
				{Type: "text", Value: t.GetDateCreatedString()},
			},
			DataAttrs: map[string]string{
				"name":          name,
				"status":        st,
				"version":       version,
				"deprecatable":  boolAttr(st == "active"),
			},
			Actions: actions,
		})
	}
	return rows
}

// loadItemCounts returns active-item counts keyed by template id (single query).
func loadItemCounts(ctx context.Context, deps *ListViewDeps) map[string]int {
	counts := map[string]int{}
	if deps.ListEvaluationTemplateItems == nil {
		return counts
	}
	resp, err := deps.ListEvaluationTemplateItems(ctx, &itempb.ListEvaluationTemplateItemsRequest{})
	if err != nil {
		log.Printf("Failed to load template item counts: %v", err)
		return counts
	}
	for _, it := range resp.GetData() {
		if it.GetActive() {
			counts[it.GetEvaluationTemplateId()]++
		}
	}
	return counts
}

func boolAttr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func statusString(s templatepb.EvaluationTemplateStatus) string {
	switch s {
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DRAFT:
		return "draft"
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_ACTIVE:
		return "active"
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DEPRECATED:
		return "deprecated"
	default:
		return "draft"
	}
}

func statusVariant(s templatepb.EvaluationTemplateStatus) string {
	switch s {
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_ACTIVE:
		return "success"
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DEPRECATED:
		return "warning"
	default:
		return "default"
	}
}

func evaluationTypeString(t evalpb.EvaluationType) string {
	switch t {
	case evalpb.EvaluationType_EVALUATION_TYPE_PERFORMANCE_REVIEW:
		return "Performance Review"
	case evalpb.EvaluationType_EVALUATION_TYPE_CSAT:
		return "CSAT"
	case evalpb.EvaluationType_EVALUATION_TYPE_COURSE_EVAL:
		return "Course Eval"
	case evalpb.EvaluationType_EVALUATION_TYPE_VENDOR_SCORECARD:
		return "Vendor Scorecard"
	default:
		return "Unspecified"
	}
}

func relationshipTypeString(t evalpb.RelationshipType) string {
	switch t {
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_CLIENT_TO_ASSOCIATE:
		return "Client -> Associate"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_STAFF_TO_CLIENT:
		return "Staff -> Client"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_SELF:
		return "Self"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_PEER:
		return "Peer"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_MANAGER:
		return "Manager"
	default:
		return "Unspecified"
	}
}

func visibilityString(v evalpb.VisibilityType) string {
	switch v {
	case evalpb.VisibilityType_VISIBILITY_TYPE_INTERNAL_ONLY:
		return "Internal Only"
	case evalpb.VisibilityType_VISIBILITY_TYPE_VISIBLE_TO_SUBJECT:
		return "Visible to Subject"
	case evalpb.VisibilityType_VISIBILITY_TYPE_VISIBLE_TO_SUBJECT_ANON:
		return "Visible (Anon)"
	default:
		return "Unspecified"
	}
}

func statusPageTitle(l evaluation_template.Labels, status string) string {
	switch status {
	case "draft":
		return l.Page.HeadingDraft
	case "active":
		return l.Page.HeadingActive
	case "deprecated":
		return l.Page.HeadingDeprecated
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l evaluation_template.Labels, status string) string {
	switch status {
	case "draft":
		return l.Page.CaptionDraft
	case "active":
		return l.Page.CaptionActive
	case "deprecated":
		return l.Page.CaptionDeprecated
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l evaluation_template.Labels, status string) string {
	switch status {
	case "draft":
		return l.Empty.DraftTitle
	case "active":
		return l.Empty.ActiveTitle
	case "deprecated":
		return l.Empty.DeprecatedTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l evaluation_template.Labels, status string) string {
	switch status {
	case "draft":
		return l.Empty.DraftMessage
	case "active":
		return l.Empty.ActiveMessage
	case "deprecated":
		return l.Empty.DeprecatedMessage
	default:
		return l.Empty.Message
	}
}
