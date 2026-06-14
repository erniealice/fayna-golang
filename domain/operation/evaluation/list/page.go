package list

import (
	"context"
	"fmt"
	"log"

	evaluation "github.com/erniealice/fayna-golang/domain/operation/evaluation"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
)

// evaluationSortableSQLCols is the sort allowlist (SEC-3: NO raw column from
// the query string). Only these columns may drive ORDER BY (§C.1).
var evaluationSortableSQLCols = map[string]string{
	"submitted_at":  "submitted_at",
	"period_start":  "period_start",
	"overall_score": "overall_score",
	"client":        "client_id",
}

// ListViewDeps holds view dependencies for the staff evaluation (Reviews) list.
// ListEvaluations is servicing-gated for staff (CR-5) inside the closure.
type ListViewDeps struct {
	Routes          evaluation.Routes
	Labels          evaluation.Labels
	CommonLabels    pyeza.CommonLabels
	TableLabels     types.TableLabels
	ListEvaluations func(ctx context.Context, req *evalpb.ListEvaluationsRequest) (*evalpb.ListEvaluationsResponse, error)
	GetListPageData func(ctx context.Context, req *evalpb.GetEvaluationListPageDataRequest) (*evalpb.GetEvaluationListPageDataResponse, error)
}

// PageData holds the data for the evaluation list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	Labels          evaluation.Labels
	StatusTabs      []StatusTab
	ActiveStatus    string
}

// StatusTab represents a status filter tab.
type StatusTab struct {
	Key    string
	Label  string
	Href   string
	HxGet  string
	Active bool
	TestID string
}

// NewView creates the staff evaluation (Reviews) list view (Surface 1).
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "list") {
			return view.Forbidden("evaluation:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "submitted"
		}

		resp, err := deps.ListEvaluations(ctx, &evalpb.ListEvaluationsRequest{})
		if err != nil {
			log.Printf("Failed to list evaluations: %v", err)
			return view.Error(fmt.Errorf("failed to load reviews: %w", err))
		}

		l := deps.Labels
		columns := evaluationColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := buildTableConfig(columns, rows, status, l, deps, perms, true)
		statusTabs := buildStatusTabs(l, status, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    l.Page.Heading,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-clipboard-check",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "evaluation-list-content",
			Table:           tableConfig,
			Labels:          l,
			StatusTabs:      statusTabs,
			ActiveStatus:    status,
		}

		return view.OK("evaluation-list", pageData)
	})
}

// NewTableView creates the table-only partial view for HTMX table swaps.
func NewTableView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "list") {
			return view.Forbidden("evaluation:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "submitted"
		}

		resp, err := deps.ListEvaluations(ctx, &evalpb.ListEvaluationsRequest{})
		if err != nil {
			log.Printf("Failed to list evaluations: %v", err)
			return view.Error(fmt.Errorf("failed to load reviews: %w", err))
		}

		l := deps.Labels
		columns := evaluationColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := buildTableConfig(columns, rows, status, l, deps, perms, false)
		return view.OK("table-card", tableConfig)
	})
}

func buildTableConfig(
	columns []types.TableColumn,
	rows []types.TableRow,
	status string,
	l evaluation.Labels,
	deps *ListViewDeps,
	perms *types.UserPermissions,
	full bool,
) *types.TableConfig {
	tableConfig := &types.TableConfig{
		ID:                   "evaluation-list-table",
		Columns:              columns,
		Rows:                 rows,
		ShowSearch:           true,
		ShowActions:          true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "submitted_at",
		DefaultSortDirection: "desc",
		RefreshURL:           route.ResolveURL(deps.Routes.TableURL, "status", status),
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
		// §C.1: NO primary add — staff do not create from this list
		// (creation is the portal "Rate My Team" / perf-panel "Start review").
	}
	if full {
		tableConfig.BulkActions = &types.BulkActionsConfig{
			Enabled:        true,
			SelectAllLabel: "Select all",
			SelectedLabel:  "selected",
			CancelLabel:    "Cancel",
			Actions: []types.BulkAction{
				{
					Key:              "archive",
					Label:            l.Actions.Bulk,
					Icon:             "icon-archive",
					Variant:          "secondary",
					Endpoint:         deps.Routes.BulkArchiveURL,
					RequiresDataAttr: "bulk-archivable",
				},
			},
		}
	}
	types.ApplyTableSettings(tableConfig)
	return tableConfig
}

func evaluationColumns(l evaluation.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "associate", Label: l.Columns.Associate},
		{Key: "client", Label: l.Columns.Client},
		{Key: "period", Label: l.Columns.Period, WidthClass: "col-3xl"},
		{Key: "type", Label: l.Columns.Type, WidthClass: "col-3xl"},
		{Key: "overall", Label: l.Columns.Overall, WidthClass: "col-lg"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
		{Key: "submitted", Label: l.Columns.Submitted, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*evalpb.Evaluation,
	status string,
	l evaluation.Labels,
	routes evaluation.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, e := range items {
		evStatus := statusString(e.GetStatus())
		if !matchesStatusTab(evStatus, status, e.GetActive()) {
			continue
		}

		id := e.GetId()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)
		period := formatPeriod(e.GetPeriodStart(), e.GetPeriodEnd())
		typeLabel := typeString(e.GetEvaluationType(), l)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: stringOrDash(e.GetSubjectStaffId())},
				{Type: "text", Value: stringOrDash(e.GetClientId())},
				{Type: "text", Value: period},
				{Type: "badge", Value: typeLabel, Variant: typeVariant(e.GetEvaluationType())},
				{Type: "text", Value: overallScore(e)},
				{Type: "badge", Value: statusLabel(e.GetStatus(), l), Variant: statusVariant(evStatus)},
				types.DateTimeCell(submittedString(e), types.DateReadable),
			},
			DataAttrs: map[string]string{
				"status":           evStatus,
				"subject-staff-id": e.GetSubjectStaffId(),
				"client-id":        e.GetClientId(),
				"evaluation-type":  typeLabel,
				// bulk-archivable gates the Bulk Archive action to SUBMITTED rows
				// (RequiresDataAttr must be "true" on ALL selected rows).
				"bulk-archivable": boolAttr(evStatus == "submitted"),
			},
			Actions: rowActions(e, evStatus, id, l, routes, perms),
		})
	}
	return rows
}

func rowActions(
	e *evalpb.Evaluation,
	evStatus string,
	id string,
	l evaluation.Labels,
	routes evaluation.Routes,
	perms *types.UserPermissions,
) []types.TableAction {
	detailURL := route.ResolveURL(routes.DetailURL, "id", id)
	actions := []types.TableAction{
		{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
	}
	// Sign off — SUBMITTED only, evaluation:sign_off + is_owner (CR-5, server-side)
	actions = append(actions, types.TableAction{
		Type:            "edit",
		Label:           l.Actions.SignOff,
		Action:          "sign-off",
		URL:             route.ResolveURL(routes.SignOffURL, "id", id),
		Disabled:        !perms.Can("evaluation", "sign_off") || evStatus != "submitted",
		DisabledTooltip: l.Errors.PermissionDenied,
	})
	// Archive — SUBMITTED→ARCHIVED, evaluation:update
	actions = append(actions, types.TableAction{
		Type:            "edit",
		Label:           l.Actions.Archive,
		Action:          "archive",
		URL:             route.ResolveURL(routes.ArchiveURL, "id", id),
		Disabled:        !perms.Can("evaluation", "update") || evStatus != "submitted",
		DisabledTooltip: l.Errors.PermissionDenied,
	})
	// Delete — staff {1,2}, evaluation:delete
	actions = append(actions, types.TableAction{
		Type:            "delete",
		Label:           l.Actions.Delete,
		Action:          "delete",
		URL:             route.ResolveURL(routes.DeleteURL, "id", id),
		Disabled:        !perms.Can("evaluation", "delete"),
		DisabledTooltip: l.Errors.PermissionDenied,
	})
	return actions
}

// matchesStatusTab maps the snapshot status to a tab. "all" matches everything;
// "active" omits archived; an explicit slug matches exactly.
func matchesStatusTab(evStatus, tabStatus string, active bool) bool {
	switch tabStatus {
	case "all":
		return true
	case "active":
		return active
	default:
		return evStatus == tabStatus
	}
}

func buildStatusTabs(l evaluation.Labels, active string, routes evaluation.Routes) []StatusTab {
	tabs := []struct {
		key   string
		label string
	}{
		{"draft", l.Status.Draft},
		{"submitted", l.Status.Submitted},
		{"signed_off", l.Status.SignedOff},
		{"archived", l.Status.Archived},
		{"all", "All"},
	}
	result := make([]StatusTab, 0, len(tabs))
	for _, t := range tabs {
		result = append(result, StatusTab{
			Key:    t.key,
			Label:  t.label,
			Href:   route.ResolveURL(routes.ListURL, "status", t.key),
			HxGet:  route.ResolveURL(routes.TableURL, "status", t.key),
			Active: t.key == active,
			TestID: "eval-filter-status-" + t.key,
		})
	}
	return result
}

func statusString(s evalpb.EvaluationStatus) string {
	switch s {
	case evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT:
		return "draft"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED:
		return "submitted"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF:
		return "signed_off"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_ARCHIVED:
		return "archived"
	default:
		return "draft"
	}
}

func statusLabel(s evalpb.EvaluationStatus, l evaluation.Labels) string {
	switch s {
	case evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT:
		return l.Status.Draft
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED:
		return l.Status.Submitted
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF:
		return l.Status.SignedOff
	case evalpb.EvaluationStatus_EVALUATION_STATUS_ARCHIVED:
		return l.Status.Archived
	default:
		return l.Status.Draft
	}
}

func statusVariant(status string) string {
	switch status {
	case "draft":
		return "default"
	case "submitted":
		return "info"
	case "signed_off":
		return "success"
	case "archived":
		return "secondary"
	default:
		return "default"
	}
}

func typeString(t evalpb.EvaluationType, l evaluation.Labels) string {
	switch t {
	case evalpb.EvaluationType_EVALUATION_TYPE_PERFORMANCE_REVIEW:
		return l.Type.PerformanceReview
	case evalpb.EvaluationType_EVALUATION_TYPE_CSAT:
		return l.Type.CSAT
	case evalpb.EvaluationType_EVALUATION_TYPE_COURSE_EVAL:
		return l.Type.CourseEval
	case evalpb.EvaluationType_EVALUATION_TYPE_VENDOR_SCORECARD:
		return l.Type.VendorScorecard
	default:
		return ""
	}
}

func typeVariant(t evalpb.EvaluationType) string {
	switch t {
	case evalpb.EvaluationType_EVALUATION_TYPE_PERFORMANCE_REVIEW:
		return "info"
	case evalpb.EvaluationType_EVALUATION_TYPE_CSAT:
		return "success"
	case evalpb.EvaluationType_EVALUATION_TYPE_VENDOR_SCORECARD:
		return "warning"
	default:
		return "default"
	}
}

// overallScore renders the computed score (blank if nil — §C.1 "blank if nil").
func overallScore(e *evalpb.Evaluation) string {
	if e.GetStatus() == evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT {
		return "—"
	}
	if e.GetOverallScore() == 0 {
		return "—"
	}
	return fmt.Sprintf("%.2f", e.GetOverallScore())
}

func submittedString(e *evalpb.Evaluation) string {
	if e.GetSubmittedAt() == 0 {
		return ""
	}
	return e.GetDateModifiedString()
}

func formatPeriod(start, end string) string {
	if start == "" && end == "" {
		return "—"
	}
	if end == "" {
		return start
	}
	return start + " – " + end
}

func stringOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func boolAttr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
