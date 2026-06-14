package list

import (
	"context"
	"fmt"
	"log"

	evaluation_cycle "github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	cyclepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle"
)

// evaluationCycleSortableSQLCols is the sort allowlist (SEC-3: NO raw column
// from query string). Every sortable column is enumerated.
var evaluationCycleSortableSQLCols = map[string]string{
	"name":              "name",
	"period_start":      "period_start",
	"sign_off_due_date": "sign_off_due_date",
	"close_date":        "close_date",
}

// ListViewDeps holds view dependencies for the evaluation cycle list.
type ListViewDeps struct {
	Routes               evaluation_cycle.Routes
	Labels               evaluation_cycle.Labels
	CommonLabels         pyeza.CommonLabels
	TableLabels          types.TableLabels
	ListEvaluationCycles func(ctx context.Context, req *cyclepb.ListEvaluationCyclesRequest) (*cyclepb.ListEvaluationCyclesResponse, error)
	// GetCycleProgress supplies the per-row "X of Y" cell. Optional; when nil the
	// Progress column renders "—" (the detail/banner compute it precisely).
	GetCycleProgress func(ctx context.Context, cycleID string) (*evaluation_cycle.CycleProgress, error)
}

// PageData holds the data for the evaluation cycle list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	Labels          evaluation_cycle.Labels
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

// NewView creates the evaluation cycle list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_cycle", "list") {
			return view.Forbidden("evaluation_cycle:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "open"
		}

		tableConfig, err := buildTable(ctx, deps, status, perms, true)
		if err != nil {
			return view.Error(err)
		}

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    l.Page.Heading,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-refresh-cw",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "evaluation-cycle-list-content",
			Table:           tableConfig,
			Labels:          l,
			StatusTabs:      buildStatusTabs(l, status, deps.Routes),
			ActiveStatus:    status,
		}

		return view.OK("evaluation-cycle-list", pageData)
	})
}

// NewTableView creates the table-only partial view for HTMX table swaps.
func NewTableView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_cycle", "list") {
			return view.Forbidden("evaluation_cycle:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "open"
		}

		tableConfig, err := buildTable(ctx, deps, status, perms, false)
		if err != nil {
			return view.Error(err)
		}

		return view.OK("table-card", tableConfig)
	})
}

func buildTable(ctx context.Context, deps *ListViewDeps, status string, perms *types.UserPermissions, withPrimary bool) (*types.TableConfig, error) {
	resp, err := deps.ListEvaluationCycles(ctx, &cyclepb.ListEvaluationCyclesRequest{})
	if err != nil {
		log.Printf("Failed to list evaluation cycles: %v", err)
		return nil, fmt.Errorf("failed to load evaluation cycles: %w", err)
	}

	l := deps.Labels
	columns := cycleColumns(l)
	rows := buildTableRows(ctx, deps, resp.GetData(), status, l, deps.Routes, perms)
	types.ApplyColumnStyles(columns, rows)

	tableConfig := &types.TableConfig{
		ID:                   "evaluation_cycle-list-table",
		Columns:              columns,
		Rows:                 rows,
		ShowSearch:           true,
		ShowActions:          true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "period_start",
		DefaultSortDirection: "desc",
		RefreshURL:           route.ResolveURL(deps.Routes.TableURL, "status", status),
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}
	if withPrimary {
		tableConfig.PrimaryAction = &types.PrimaryAction{
			Label:           l.Buttons.NewCycle,
			ActionURL:       deps.Routes.AddURL,
			Icon:            "icon-plus",
			Disabled:        !perms.Can("evaluation_cycle", "create"),
			DisabledTooltip: l.Errors.PermissionDenied,
		}
	}
	types.ApplyTableSettings(tableConfig)
	return tableConfig, nil
}

func cycleColumns(l evaluation_cycle.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "engagement", Label: l.Columns.Engagement, WidthClass: "col-3xl"},
		{Key: "period", Label: l.Columns.Period, WidthClass: "col-3xl"},
		{Key: "progress", Label: l.Columns.Progress, WidthClass: "col-3xl"},
		{Key: "sign_off_due", Label: l.Columns.SignOffDue, WidthClass: "col-3xl"},
		{Key: "closes", Label: l.Columns.Closes, WidthClass: "col-3xl"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	ctx context.Context,
	deps *ListViewDeps,
	items []*cyclepb.EvaluationCycle,
	status string,
	l evaluation_cycle.Labels,
	routes evaluation_cycle.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, c := range items {
		cycleStatus := cycleStatusString(c.GetStatus())
		if !matchesStatusTab(cycleStatus, status) {
			continue
		}

		id := c.GetId()
		name := c.GetName()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)
		period := periodRange(c.GetPeriodStart(), c.GetPeriodEnd())
		progress := progressCell(ctx, deps, id, l)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: dashIfEmpty(c.GetSubscriptionId())},
				{Type: "text", Value: period},
				{Type: "text", Value: progress},
				{Type: "text", Value: dashIfEmpty(c.GetSignOffDueDate())},
				{Type: "text", Value: dashIfEmpty(c.GetCloseDate())},
				{Type: "badge", Value: statusLabel(cycleStatus, l), Variant: statusVariant(c.GetStatus())},
			},
			DataAttrs: map[string]string{
				"name":   name,
				"status": cycleStatus,
			},
			Actions: cycleRowActions(c, l, routes, perms),
		})
	}
	return rows
}

// cycleRowActions builds the View / Open / Close row actions. Open is offered on
// not-yet-closed cycles; Close likewise. Both gate evaluation_cycle:update.
//
// The pyeza table renderer + table-actions.js drive lifecycle POSTs by
// data-action: "activate" reads data-activate-url, "deactivate" reads
// data-deactivate-url — both confirm then POST. We map Open→activate and
// Close→deactivate so the URL is emitted and the JS posts to the lifecycle
// endpoint (testids evaluation_cycle-list-row-action-{open,close} per §F.4).
func cycleRowActions(c *cyclepb.EvaluationCycle, l evaluation_cycle.Labels, routes evaluation_cycle.Routes, perms *types.UserPermissions) []types.TableAction {
	id := c.GetId()
	detailURL := route.ResolveURL(routes.DetailURL, "id", id)
	actions := []types.TableAction{
		{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL, TestID: "evaluation_cycle-list-row-action-view"},
	}
	canUpdate := perms.Can("evaluation_cycle", "update")
	if c.GetStatus() != cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_CLOSED {
		actions = append(actions,
			types.TableAction{
				Type: "activate", Action: "activate", Label: l.Actions.Open,
				URL:            route.ResolveURL(routes.OpenURL, "id", id),
				TestID:         "evaluation_cycle-list-row-action-open",
				ConfirmTitle:   l.Confirm.Open,
				ConfirmMessage: l.Confirm.OpenMessage,
				Disabled:       !canUpdate, DisabledTooltip: l.Errors.PermissionDenied,
			},
			types.TableAction{
				Type: "deactivate", Action: "deactivate", Label: l.Actions.Close,
				URL:            route.ResolveURL(routes.CloseURL, "id", id),
				TestID:         "evaluation_cycle-list-row-action-close",
				ConfirmTitle:   l.Confirm.Close,
				ConfirmMessage: l.Confirm.CloseMessage,
				Disabled:       !canUpdate, DisabledTooltip: l.Errors.PermissionDenied,
			},
		)
	}
	return actions
}

// progressCell renders the per-row "X of Y" text. Uses GetCycleProgress when
// wired; "—" otherwise (the detail page / banner render it precisely).
func progressCell(ctx context.Context, deps *ListViewDeps, cycleID string, l evaluation_cycle.Labels) string {
	if deps.GetCycleProgress == nil {
		return "—"
	}
	p, err := deps.GetCycleProgress(ctx, cycleID)
	if err != nil || p == nil {
		return "—"
	}
	return fmt.Sprintf(l.Banner.ProgressFmt, p.Completed, p.Total)
}

func matchesStatusTab(cycleStatus, tabStatus string) bool {
	switch tabStatus {
	case "all":
		return true
	default:
		return cycleStatus == tabStatus
	}
}

func buildStatusTabs(l evaluation_cycle.Labels, active string, routes evaluation_cycle.Routes) []StatusTab {
	tabs := []struct {
		key   string
		label string
	}{
		{"open", l.Status.Open},
		{"sign_off", l.Status.SignOff},
		{"closed", l.Status.Closed},
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
			TestID: "evaluation_cycle-tab-" + t.key,
		})
	}
	return result
}

func cycleStatusString(s cyclepb.EvaluationCycleStatus) string {
	switch s {
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_OPEN:
		return "open"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_SIGN_OFF:
		return "sign_off"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_CLOSED:
		return "closed"
	default:
		return "open"
	}
}

func statusLabel(status string, l evaluation_cycle.Labels) string {
	switch status {
	case "open":
		return l.Status.Open
	case "sign_off":
		return l.Status.SignOff
	case "closed":
		return l.Status.Closed
	default:
		return l.Status.Open
	}
}

func statusVariant(s cyclepb.EvaluationCycleStatus) string {
	switch s {
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_OPEN:
		return "info"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_SIGN_OFF:
		return "warning"
	case cyclepb.EvaluationCycleStatus_EVALUATION_CYCLE_STATUS_CLOSED:
		return "success"
	default:
		return "default"
	}
}

func periodRange(start, end string) string {
	if start == "" && end == "" {
		return "—"
	}
	return start + " → " + end
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
