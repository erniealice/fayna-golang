package list

import (
	"context"
	"fmt"
	"log"
	"strings"

	job "github.com/erniealice/fayna-golang/domain/operation/job"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// businessTypeEducation is the compose MountContext.BusinessType value that
// swaps this List view's content from the classic per-job table to the
// template-grain delivery summary (20260710 staff-class-list plan; O3 defers
// every other tier).
const businessTypeEducation = "education"

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes       job.Routes
	ListJobs     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	GetInUseIDs  func(ctx context.Context, ids []string) (map[string]bool, error)
	Labels       job.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// BusinessType (compose MountContext.BusinessType) — see
	// businessTypeEducation above.
	BusinessType string

	// Template-grain delivery summary (education tier only). ONE server-side
	// GROUP-BY read (espyna service/operation/job_template_summary) that
	// aggregates + resolver-scopes + resolves every column in the adapter —
	// replacing the former ~76-fetch Go aggregation. Optional/nil-safe: a nil
	// closure renders the education-tier list empty. MatrixDetailURL is the
	// cross-unit outcome_matrix.matrix route ("/outcome-matrix/{id}",
	// id=job_template_id) each summary row links to.
	ListJobTemplateSummaries func(ctx context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error)
	MatrixDetailURL          string

	// Options — app-configured presentation. When Options.Tab.Enabled() is true
	// the list renders a job_category tabstrip above the table (the "/classes"
	// tab-split); with zero-valued options (service-admin) it renders exactly
	// today's flat list — the backward-compat contract.
	Options job.Options

	// Tab-split deps (job_category tabstrip). All optional/nil-safe — a nil
	// closure degrades to the flat list (no tabs). ListJobCategories supplies
	// the tab rows; ListJobTemplates builds the job_template→job_category map the
	// education-tier (template-grain) filter needs (the flat path filters on the
	// job.job_category_id denorm directly — no template join).
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
	ListJobTemplates  func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
}

// PageData holds the data for the job list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig

	// Tab-split extras (job_category tabstrip). Zero on the flat path.
	TabItems  []pyeza.TabItem
	ActiveTab string
	TabsAria  string
}

// NewView creates the job list view. BusinessType "education" swaps the
// content to the template-grain delivery summary (buildDeliverySummaryTable,
// template_summary.go); every other tier keeps the classic per-job table
// (buildJobTable, below) — 20260710 staff-class-list plan, O3.
//
// When the app configures Options.Tab.GroupByField = "job_category"
// (school-admin), the list renders a job_category tabstrip above the table (the
// "/classes" tab-split, mirroring report-cards). With zero-valued options
// (service-admin) it renders the flat list unchanged — the backward-compat
// contract.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		// 2026-05-14 permission-gates P2a: reject direct-URL access.
		if !perms.Can("job", "list") {
			return view.Forbidden("job:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		if deps.Options.Tab.Enabled() {
			return renderTabbed(ctx, deps, viewCtx, status, perms)
		}
		return renderFlat(ctx, deps, viewCtx, status, perms)
	})
}

// renderFlat renders the job list without a tabstrip (today's behavior — the
// service-admin backward-compat path and the education list before the split).
func renderFlat(ctx context.Context, deps *ListViewDeps, viewCtx *view.ViewContext, status string, perms *types.UserPermissions) view.ViewResult {
	var tableConfig *types.TableConfig
	var err error
	if deps.BusinessType == businessTypeEducation {
		tableConfig, err = buildDeliverySummaryTable(ctx, deps, status, perms)
	} else {
		tableConfig, err = buildJobTable(ctx, deps, status, perms)
	}
	if err != nil {
		log.Printf("Failed to list jobs: %v", err)
		return view.Error(fmt.Errorf("failed to load jobs: %w", err))
	}

	pageData := newJobListPageData(deps, viewCtx, status, tableConfig)
	applyJobKBHelp(pageData, viewCtx)
	return view.OK("job-list", pageData)
}

// renderTabbed renders the job list with a job_category tabstrip (the "/classes"
// tab-split). Tabs come from ListJobCategories (active + inactive, ordered by
// sort_order NULLS-LAST → name); ?jc=<id> selects the active tab; the selected
// category filters the table. On the education tier the filter uses a
// job_template→job_category map over the template-grain delivery summary; on the
// flat tier it filters jobs by the job.job_category_id denorm (single-table, no
// template join). Nil-safe: no categories → an empty tabstrip and the unfiltered
// table (never an error).
func renderTabbed(ctx context.Context, deps *ListViewDeps, viewCtx *view.ViewContext, status string, perms *types.UserPermissions) view.ViewResult {
	l := deps.Labels

	cats := listAllJobCategories(ctx, deps)
	sortJobCategories(cats, deps.Options.Tab)

	// ?jc= selects the active tab. Validate it against the fetched set: an empty,
	// stale, foreign, or tampered id falls back to the default category so the
	// tablist never references a non-existent tab (dangling aria state).
	selected := strings.TrimSpace(viewCtx.Request.URL.Query().Get("jc"))
	if selected == "" || !categoryExists(cats, selected) {
		selected = defaultCategory(cats)
	}

	var tableConfig *types.TableConfig
	var counts map[string]int
	var err error
	if deps.BusinessType == businessTypeEducation {
		tableConfig, counts, err = buildDeliverySummaryTableTabbed(ctx, deps, status, selected)
	} else {
		tableConfig, counts, err = buildJobTableTabbed(ctx, deps, status, selected, perms)
	}
	if err != nil {
		log.Printf("Failed to list jobs: %v", err)
		return view.Error(fmt.Errorf("failed to load jobs: %w", err))
	}

	pageData := newJobListPageData(deps, viewCtx, status, tableConfig)
	// Resolve the {status} segment so a tab click preserves the current status
	// (the ListURL pattern is "/courses/list/{status}").
	listURL := route.ResolveURL(deps.Routes.ListURL, "status", status)
	pageData.TabItems = buildJobCategoryTabs(cats, counts, selected, listURL)
	pageData.ActiveTab = tabKey(selected)
	pageData.TabsAria = l.Page.JobCategoryTabsAria
	applyJobKBHelp(pageData, viewCtx)
	return view.OK("job-list", pageData)
}

// newJobListPageData builds the shared PageData shell (title/header/nav) for both
// the flat and tabbed render paths.
func newJobListPageData(deps *ListViewDeps, viewCtx *view.ViewContext, status string, tableConfig *types.TableConfig) *PageData {
	l := deps.Labels
	return &PageData{
		PageData: types.PageData{
			CacheVersion:   viewCtx.CacheVersion,
			Title:          statusPageTitle(l, status),
			CurrentPath:    viewCtx.CurrentPath,
			ActiveNav:      deps.Routes.ActiveNav,
			ActiveSubNav:   jobSubNav(status),
			HeaderTitle:    statusPageTitle(l, status),
			HeaderSubtitle: statusPageCaption(l, status),
			HeaderIcon:     "icon-briefcase",
			CommonLabels:   deps.CommonLabels,
		},
		ContentTemplate: "job-list-content",
		Table:           tableConfig,
	}
}

// applyJobKBHelp attaches the job KB help content when present.
func applyJobKBHelp(pageData *PageData, viewCtx *view.ViewContext) {
	if viewCtx.Translations == nil {
		return
	}
	if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
		if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "job"); kb != nil {
			pageData.HasHelp = true
			pageData.HelpContent = kb.Body
		}
	}
}

// jobListPageLimit is the page size used by fetchScopedJobs' loop — the
// fetchSubscriptionsForPricePlan shape (centymo price_schedule/detail/plan/
// subscriptions.go), capped at the adapter's per-call maximum.
const jobListPageLimit = 100

// fetchScopedJobs pages through ListJobs for ONE status (server-side
// TypedFilter — never client-side), stopping when a batch is short. Row
// scoping (STAFF principals see only their delivery graph) is resolver-level
// (espyna principalscope) — this helper adds no staff_id filter. Shared by
// both the classic per-job table and the education template-grain delivery
// summary.
func fetchScopedJobs(ctx context.Context, deps *ListViewDeps, status string) ([]*jobpb.Job, error) {
	statusValue := jobStatusFilterValue(status)
	var out []*jobpb.Job
	for page := int32(1); ; page++ {
		resp, err := deps.ListJobs(ctx, &jobpb.ListJobsRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{{
					Field: "status",
					FilterType: &commonpb.TypedFilter_StringFilter{
						StringFilter: &commonpb.StringFilter{
							Value:    statusValue,
							Operator: commonpb.StringOperator_STRING_EQUALS,
						},
					},
				}},
			},
			Pagination: &commonpb.PaginationRequest{
				Limit:  jobListPageLimit,
				Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
			},
		})
		if err != nil {
			return out, err
		}
		batch := resp.GetData()
		out = append(out, batch...)
		if int32(len(batch)) < jobListPageLimit {
			return out, nil
		}
	}
}

// jobStatusFilterValue maps a URL status segment ("active", "on-hold", ...)
// to the full JobStatus enum name the job.status column stores (e.g.
// "JOB_STATUS_ACTIVE" — protojson enum serialization, confirmed against
// education1). Unrecognised segments fall back to the active status so the
// filter never silently degrades to "no filter" (every job returned).
func jobStatusFilterValue(status string) string {
	name := "JOB_STATUS_" + strings.ToUpper(strings.ReplaceAll(status, "-", "_"))
	if _, ok := enums.JobStatus_value[name]; ok {
		return name
	}
	return "JOB_STATUS_ACTIVE"
}

// buildJobTable builds the classic per-job TableConfig (professional/general/
// service/... tiers). Fixed 2026-07-11 (20260710 staff-class-list plan build
// spec §3): status is now a server-side TypedFilter (was: fetched with no
// filter, then filtered client-side after truncation) and fetchScopedJobs
// pages through every matching row (was: one adapter-capped page of 100).
func buildJobTable(ctx context.Context, deps *ListViewDeps, status string, perms *types.UserPermissions) (*types.TableConfig, error) {
	jobs, err := fetchScopedJobs(ctx, deps, status)
	if err != nil {
		return nil, err
	}
	return jobTableConfig(ctx, deps, jobs, perms), nil
}

// buildJobTableTabbed is the tabbed per-job path (a non-education app that opts
// into the tab-split). Counts group the whole scoped set by the job.job_category_id
// denorm (single-table — no job→job_template join); the selected category filters
// the rendered rows.
func buildJobTableTabbed(ctx context.Context, deps *ListViewDeps, status, selected string, perms *types.UserPermissions) (*types.TableConfig, map[string]int, error) {
	jobs, err := fetchScopedJobs(ctx, deps, status)
	if err != nil {
		return nil, nil, err
	}
	counts := map[string]int{}
	filtered := make([]*jobpb.Job, 0, len(jobs))
	for _, j := range jobs {
		counts[j.GetJobCategoryId()]++
		if selected == "" || j.GetJobCategoryId() == selected {
			filtered = append(filtered, j)
		}
	}
	return jobTableConfig(ctx, deps, filtered, perms), counts, nil
}

// jobTableConfig builds the classic per-job TableConfig from an already-fetched
// (and possibly category-filtered) job slice.
func jobTableConfig(ctx context.Context, deps *ListViewDeps, jobs []*jobpb.Job, perms *types.UserPermissions) *types.TableConfig {
	// Collect IDs and check which are in use (referenced by dependent tables).
	var inUseIDs map[string]bool
	if deps.GetInUseIDs != nil {
		var itemIDs []string
		for _, j := range jobs {
			itemIDs = append(itemIDs, j.GetId())
		}
		inUseIDs, _ = deps.GetInUseIDs(ctx, itemIDs)
	}

	l := deps.Labels
	columns := jobColumns(l)
	rows := buildTableRows(jobs, l, deps.Routes, inUseIDs, perms)
	types.ApplyColumnStyles(columns, rows)

	tableConfig := &types.TableConfig{
		ID:                   "jobs-table",
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
			Label:           l.Buttons.AddJob,
			ActionURL:       deps.Routes.AddURL,
			Icon:            "icon-plus",
			Disabled:        !perms.Can("job", "create"),
			DisabledTooltip: l.Errors.PermissionDenied,
		},
		BulkActions: &types.BulkActionsConfig{
			Enabled:        true,
			SelectAllLabel: l.BulkActions.SelectAll,
			SelectedLabel:  l.BulkActions.SelectedCount,
			CancelLabel:    l.BulkActions.Cancel,
			Actions: []types.BulkAction{
				{
					Key:              "delete",
					Label:            l.BulkActions.Delete,
					Icon:             "icon-trash-2",
					Variant:          "danger",
					Endpoint:         deps.Routes.BulkDeleteURL,
					ConfirmTitle:     l.BulkActions.BulkDeleteConfirmTitle,
					ConfirmMessage:   l.BulkActions.BulkDeleteConfirmMsg,
					RequiresDataAttr: "deletable",
				},
			},
		},
	}
	types.ApplyTableSettings(tableConfig)
	return tableConfig
}

func jobColumns(l job.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "client", Label: l.Columns.Client},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
		{Key: "created", Label: l.Columns.Created, WidthClass: "col-4xl"},
	}
}

// buildTableRows renders one row per job. status filtering is now applied
// server-side by fetchScopedJobs (TypedFilter) — jobs arrives pre-filtered,
// so no client-side status re-check happens here.
func buildTableRows(jobs []*jobpb.Job, l job.Labels, routes job.Routes, inUseIDs map[string]bool, perms *types.UserPermissions) []types.TableRow {
	rows := []types.TableRow{}
	for _, j := range jobs {
		jobStatus := jobStatusString(j.GetStatus())

		id := j.GetId()
		name := j.GetName()

		// Build client display name
		clientName := ""
		if c := j.GetClient(); c != nil {
			if u := c.GetUser(); u != nil {
				first := u.GetFirstName()
				last := u.GetLastName()
				if first != "" || last != "" {
					clientName = first + " " + last
				}
			}
			if clientName == "" {
				clientName = c.GetName()
			}
		}

		created := j.GetDateCreatedString()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		inUse := inUseIDs[id]
		deleteDisabled := inUse || !perms.Can("job", "delete")
		deleteTooltip := l.Errors.PermissionDenied
		if inUse {
			deleteTooltip = l.Errors.InUse
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: clientName},
				{Type: "badge", Value: jobStatus, Variant: jobStatusVariant(jobStatus)},
				types.DateTimeCell(created, types.DateReadable),
			},
			DataAttrs: map[string]string{
				"name":      name,
				"client":    clientName,
				"status":    jobStatus,
				"created":   created,
				"deletable": boolAttr(!inUse),
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: deleteDisabled, DisabledTooltip: deleteTooltip},
			},
		})
	}
	return rows
}

func boolAttr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func jobStatusString(s enums.JobStatus) string {
	switch s {
	case enums.JobStatus_JOB_STATUS_DRAFT:
		return "draft"
	case enums.JobStatus_JOB_STATUS_PENDING:
		return "pending"
	case enums.JobStatus_JOB_STATUS_PLANNED:
		return "planned"
	case enums.JobStatus_JOB_STATUS_RELEASED:
		return "released"
	case enums.JobStatus_JOB_STATUS_ACTIVE:
		return "active"
	case enums.JobStatus_JOB_STATUS_PAUSED:
		return "paused"
	case enums.JobStatus_JOB_STATUS_COMPLETED:
		return "completed"
	case enums.JobStatus_JOB_STATUS_CLOSED:
		return "closed"
	default:
		return "draft"
	}
}

func jobStatusVariant(status string) string {
	switch status {
	case "draft":
		return "default"
	case "pending":
		return "warning"
	case "planned":
		return "secondary"
	case "released":
		return "success"
	case "active":
		return "success"
	case "paused":
		return "warning"
	case "completed":
		return "info"
	case "closed":
		return "default"
	default:
		return "default"
	}
}

// jobSubNav maps a job status to its sidebar item-key suffix ("jobs-" +
// status), with the "paused" status rewritten to the sidebar's "on-hold" key
// (descriptor.go's jobs-on-hold nav item routes to status=paused).
func jobSubNav(status string) string {
	if status == "paused" {
		return "jobs-on-hold"
	}
	return "jobs-" + status
}

func statusPageTitle(l job.Labels, status string) string {
	switch status {
	case "draft":
		return l.Page.HeadingDraft
	case "planned":
		return l.Page.HeadingPlanned
	case "released":
		return l.Page.HeadingReleased
	case "active":
		return l.Page.HeadingActive
	case "completed":
		return l.Page.HeadingCompleted
	case "closed":
		return l.Page.HeadingClosed
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l job.Labels, status string) string {
	switch status {
	case "draft":
		return l.Page.CaptionDraft
	case "planned":
		return l.Page.CaptionPlanned
	case "released":
		return l.Page.CaptionReleased
	case "active":
		return l.Page.CaptionActive
	case "completed":
		return l.Page.CaptionCompleted
	case "closed":
		return l.Page.CaptionClosed
	default:
		return l.Page.Caption
	}
}
