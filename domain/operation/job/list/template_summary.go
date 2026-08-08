package list

// template_summary.go — the education-tier ("Classes") template-grain
// delivery summary that replaces the per-job table on the job List view.
//
// docs/plan/20260710-staff-class-list/tasks.md S6 (LOCKED, O5): the view's
// data assembly is ONE server-side aggregate call — espyna's
// service/operation/job_template_summary ListJobTemplateSummaries — which does
// the GROUP-BY, resolver-scoping, and every column resolution
// (group/deliverer/schedule/product) in a single query. This replaces the
// former ~76-fetch Go aggregation (page-loop jobs + seats + group/plan/staff
// lookups). One row per job_template that has >=1 (resolver-scoped) job for the
// URL {status} segment. Columns: template name, delivery group name, deliverer
// (staff of record), item count (DISTINCT scoped jobs), schedule name. Row link
// -> outcome_matrix.matrix ("/outcome-matrix/{id}", id=job_template_id).
//
// Row scoping is entirely resolver-level (espyna principalscope inside the
// adapter — STAFF principals see only their reachable jobs, the seat tier
// widens that to the class grain; non-staff see all). This file passes NO
// staff_id request filter anywhere.
//
// NAMING: every identifier here is generic (JobTemplate / SubscriptionGroup /
// Staff — the real espyna entity names). Education vocabulary
// ("Classes"/"Section"/"Teacher"/"Students"/"Academic Year") enters ONLY via
// lyngua (packages/lyngua/translations/en/education/job.json) — never as a Go
// identifier, filename, or default label here.

import (
	"context"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

	espynahttp "github.com/erniealice/espyna-golang/contrib/http"
	"github.com/erniealice/espyna-golang/shared/tableparams"
	job "github.com/erniealice/fayna-golang/domain/operation/job"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"

	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// templateSummaryRow is one job_template's aggregated delivery-summary row.
type templateSummaryRow struct {
	TemplateID string
	// GroupID is the row's subscription_group_id. The delivery aggregate is
	// already at (template x group) grain — it was simply being discarded here,
	// which is why three "Arts — Grade 10" rows (Palladium/Platinum/Tantalum,
	// 29/28/30 students) all linked to the same template-scoped sheet.
	GroupID       string
	TemplateName  string
	GroupName     string
	DelivererName string
	ItemCount     int
	ScheduleName  string

	// hideItemCount blanks the item-count cell instead of rendering ItemCount.
	// Set only on template-grain rows (a job_category the delivery aggregate
	// doesn't cover) where the count is genuinely unknown — an aggregate row
	// leaves this false, so its cell is byte-for-byte unchanged.
	hideItemCount bool
}

// buildDeliverySummaryTable builds the template-grain TableConfig for the
// education tier from the single ListJobTemplateSummaries call. The row/column
// building, labels, links (outcome-matrix per row), and empty state are
// unchanged from the pre-S6 view-side compose; only the data-assembly layer
// moved server-side.
// buildDeliverySummaryTable is the server-page education path. The service owns
// selected-category filtering, fallback selection, counts, ordering, and paging;
// the view only maps its response into Pyeza's table contract.
func buildDeliverySummaryTable(ctx context.Context, deps *ListViewDeps, status string, p tableparams.TableQueryParams, selected string, includeTemplateFallback bool) (*types.TableConfig, map[string]int, error) {
	p = boundTemplateSummaryParams(p)
	resp, err := listTemplateSummaries(ctx, deps, status, p, selected, includeTemplateFallback)
	if err != nil {
		return nil, nil, err
	}
	return templateSummaryTableConfig(deps, responseTemplateSummaryRows(resp), p, status, selected, resp), responseCategoryCounts(resp), nil
}

// templateSummaryTableConfig builds the template-grain TableConfig from an
// already-fetched (and possibly category-filtered) summary-row slice.
func templateSummaryTableConfig(deps *ListViewDeps, rows []templateSummaryRow, p tableparams.TableQueryParams, status, selected string, resp *summarypb.ListJobTemplateSummariesResponse) *types.TableConfig {
	l := deps.Labels
	columns := templateSummaryColumns(l)
	tableRows := make([]types.TableRow, 0, len(rows))
	for _, r := range rows {
		// Prefer the section-scoped URL so each row addresses its OWN section.
		// Falls back to the template URL when either the route or the row's group
		// id is absent (service-admin, or a template-grain row the delivery
		// aggregate does not cover).
		// Row-link grain is a DEPLOYMENT choice (Options.RowLink), not something
		// this file decides from the data. Both URLs are valid surfaces: a
		// workspace whose templates each serve one group gains nothing from the
		// extra segment. Falls back to template grain when the app has not opted
		// in, when the route is unconfigured, or when a row carries no group id
		// (the template-grain fallback rows) — never a half-resolved path.
		matrixURL := route.ResolveURL(deps.MatrixDetailURL, "id", r.TemplateID)
		if deps.Options.RowLinkScopedByGroup() && deps.MatrixGroupDetailURL != "" && r.GroupID != "" {
			matrixURL = route.ResolveURL(deps.MatrixGroupDetailURL, "id", r.TemplateID, "group_id", r.GroupID)
		}
		// Aggregate rows show their DISTINCT-job count; template-grain rows (a
		// category the aggregate doesn't cover) have no count and render blank.
		itemValue := strconv.Itoa(r.ItemCount)
		if r.hideItemCount {
			itemValue = ""
		}
		tableRows = append(tableRows, types.TableRow{
			ID:   r.TemplateID,
			Href: matrixURL,
			Cells: []types.TableCell{
				{Type: "text", Value: r.TemplateName},
				{Type: "text", Value: r.GroupName},
				{Type: "text", Value: r.DelivererName},
				{Type: "number", Value: itemValue},
				{Type: "text", Value: r.ScheduleName},
			},
			DataAttrs: map[string]string{
				"name":      r.TemplateName,
				"group":     r.GroupName,
				"deliverer": r.DelivererName,
				"schedule":  r.ScheduleName,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: matrixURL},
			},
		})
	}
	types.ApplyColumnStyles(columns, tableRows)

	// data-refresh-url / data-pagination-url MUST point at the table-only endpoint
	// (Routes.TableURL, list.NewTableView) so HTMX swaps just the table-card
	// partial. Pointing at the full ListURL re-renders the whole page (app-shell +
	// tabstrip) and the JS full-card-swap nests it inside the card. Falls back to
	// ListURL only when TableURL is unset. Mirrors centymo subscription list.
	refreshURL := route.ResolveURL(deps.Routes.ListURL, "status", status)
	if deps.Routes.TableURL != "" {
		refreshURL = route.ResolveURL(deps.Routes.TableURL, "status", status)
	}
	sp := &types.ServerPagination{
		Enabled:       true,
		Mode:          "offset",
		CurrentPage:   int(resp.GetPagination().GetCurrentPage()),
		PageSize:      p.PageSize,
		TotalRows:     int(resp.GetPagination().GetTotalItems()),
		TotalPages:    int(resp.GetPagination().GetTotalPages()),
		SearchQuery:   p.Search,
		SortColumn:    p.SortColumn,
		SortDirection: p.SortDir,
		FiltersJSON:   p.FiltersRaw,
		PaginationURL: summaryPaginationURL(refreshURL, selected),
	}
	sp.BuildDisplay()

	tableConfig := &types.TableConfig{
		ID:                   "job-template-summary-table",
		RefreshURL:           refreshURL,
		Columns:              columns,
		Rows:                 tableRows,
		ShowSearch:           true,
		ShowActions:          true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "group",
		DefaultSortDirection: "asc",
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
		ServerPagination: sp,
	}
	types.ApplyTableSettings(tableConfig)
	return tableConfig
}

// templateSummaryColumns declares the template-grain columns. Go defaults
// (job.Labels.Columns.{Group,Deliverer,Items,Schedule}) stay generic
// ("Group"/"Deliverer"/"Items"/"Schedule"); education overrides them to
// "Section"/"Teacher"/"Students"/"Academic Year" via lyngua.
func templateSummaryColumns(l job.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "group", Label: l.Columns.Group},
		{Key: "deliverer", Label: l.Columns.Deliverer},
		{Key: "items", Label: l.Columns.Items},
		{Key: "schedule", Label: l.Columns.Schedule},
	}
}

// listTemplateSummaries issues the ONE server-side aggregate call (espyna
// service/operation/job_template_summary) for the {status} segment and maps
// each returned JobTemplateSummary to a view row. All aggregation, resolver-
// scoping, status filtering, and group/deliverer/schedule/product resolution
// happen server-side in a single GROUP-BY query. The server's response ordering
// is retained verbatim so pagination and sorting remain consistent.
func listTemplateSummaries(ctx context.Context, deps *ListViewDeps, status string, p tableparams.TableQueryParams, selected string, includeTemplateFallback bool) (*summarypb.ListJobTemplateSummariesResponse, error) {
	if deps.ListJobTemplateSummaries == nil {
		return &summarypb.ListJobTemplateSummariesResponse{}, nil
	}
	listParams := espynahttp.ToListParams(p, templateSummarySearchFields)
	// ToListParams appends the generic entity `id` as a stable tie-breaker. A
	// summary row has no singular id: its adapter owns a complete composite
	// identity (group, template, schedule, output product), so forwarding the
	// synthetic field would either be rejected or weaken that grain contract.
	if fields := listParams.Sort.GetFields(); len(fields) > 0 && fields[len(fields)-1].GetField() == "id" {
		listParams.Sort.Fields = fields[:len(fields)-1]
	}
	req := &summarypb.ListJobTemplateSummariesRequest{
		Status:     jobStatusFilterValue(status),
		Pagination: listParams.Pagination,
		Search:     listParams.Search,
		Sort:       listParams.Sort,
	}
	if selected != "" {
		req.JobCategoryId = &selected
	}
	if includeTemplateFallback {
		req.IncludeTemplateFallback = &includeTemplateFallback
	}
	resp, err := deps.ListJobTemplateSummaries(ctx, req)
	if err != nil {
		log.Printf("Failed to list job template summaries: %v", err)
		return nil, err
	}
	return resp, nil
}

var templateSummarySearchFields = []string{"name", "group", "deliverer", "items", "schedule"}

const maxTemplateSummaryPageSize = 100

func boundTemplateSummaryParams(p tableparams.TableQueryParams) tableparams.TableQueryParams {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 25
	} else if p.PageSize > maxTemplateSummaryPageSize {
		p.PageSize = maxTemplateSummaryPageSize
	}
	return p
}

func responseTemplateSummaryRows(resp *summarypb.ListJobTemplateSummariesResponse) []templateSummaryRow {
	summaries := resp.GetSummaries()
	rows := make([]templateSummaryRow, 0, len(summaries))
	for _, s := range summaries {
		rows = append(rows, templateSummaryRow{
			TemplateID:    s.GetJobTemplateId(),
			GroupID:       s.GetSubscriptionGroupId(),
			TemplateName:  s.GetJobTemplateName(),
			GroupName:     s.GetSubscriptionGroupName(),
			DelivererName: joinDelivererNames(s.GetDeliverers()),
			ItemCount:     int(s.GetJobCount()),
			ScheduleName:  s.GetPriceScheduleName(),
			hideItemCount: s.GetTemplateGrainFallback(),
		})
	}
	return rows
}

func responseCategoryCounts(resp *summarypb.ListJobTemplateSummariesResponse) map[string]int {
	counts := make(map[string]int, len(resp.GetJobCategoryCounts()))
	for _, count := range resp.GetJobCategoryCounts() {
		counts[count.GetJobCategoryId()] = int(count.GetSummaryCount())
	}
	return counts
}

func summaryPaginationURL(listURL, selected string) string {
	if selected == "" {
		return listURL
	}
	u, err := url.Parse(listURL)
	if err != nil {
		return listURL
	}
	q := u.Query()
	q.Set("jc", selected)
	u.RawQuery = q.Encode()
	return u.String()
}

// joinDelivererNames renders a template's deliverer column. A template can have
// MORE THAN ONE deliverer (a merged deliverable delivered by several staff — e.g.
// a Section's two rotation-strand Teachers); their names render comma-joined
// ("A. Purisima, D. Cabornay"). The names are sorted for a STABLE display order
// even when a non-postgres provider returns the deliverers unordered (the postgres
// adapter already emits them in staff_name order; sorting here is defensive, the
// same discipline as the row re-sort above). Blank names are dropped.
func joinDelivererNames(deliverers []*summarypb.Deliverer) string {
	names := make([]string, 0, len(deliverers))
	for _, d := range deliverers {
		if n := d.GetStaffName(); n != "" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
