package list

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// listIn builds a LIST_IN TypedFilter (shared shape with the section view).
func listIn(field string, values []string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_ListFilter{
			ListFilter: &commonpb.ListFilter{Values: values, Operator: commonpb.ListOperator_LIST_IN},
		},
	}
}

// maxLandingPages bounds the historical-count offset loop independently of the
// adapter's short-final-page termination (defense against an adapter that
// ignores OFFSET). 100 × 100 = 10k rows, far above any real section.
const maxLandingPages = 100

// boolEq builds a BooleanFilter TypedFilter.
func boolEq(field string, v bool) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field:      field,
		FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: v}},
	}
}

// stringEq builds a STRING_EQUALS TypedFilter.
func stringEq(field, value string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_StringFilter{
			StringFilter: &commonpb.StringFilter{Value: value, Operator: commonpb.StringOperator_STRING_EQUALS},
		},
	}
}

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes                 outcome_summary.Routes
	ListJobOutcomeSummarys func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)
	Labels                 outcome_summary.Labels
	CommonLabels           pyeza.CommonLabels
	TableLabels            types.TableLabels

	// Options — app-configured presentation. When Options.List.SubscriptionGroups()
	// is true the view renders the tabbed section landing; otherwise it renders
	// the flat job_outcome_summary table (today's behavior, unchanged — the
	// backward-compat contract for service-admin's zero-valued options).
	Options outcome_summary.Options

	// Landing deps (view-1 tabbed section list). All optional/nil-safe.
	ListPriceSchedules       func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	ListSubscriptionGroups   func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListJobTemplateSummaries func(ctx context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error)

	// Historical-count fallback deps: the grouped ListJobTemplateSummaries
	// aggregate binds active rows only, so a historical (inactive) schedule's
	// sections come back with no counts — these two reads fill them in for
	// the rendered tab. Optional/nil-safe (missing → counts stay blank).
	ListSubscriptionGroupMembers func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListJobs                     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
}

// PageData holds the data for the outcome summary list page (flat or landing).
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig

	// Landing extras (view-1). Zero on the flat path.
	Landing   bool
	TabItems  []pyeza.TabItem
	ActiveTab string
	TabsAria  string
}

// NewView creates the outcome summary list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a + P2b: page previously had no
		// perms lookup. Reject direct-URL access without
		// job_outcome_summary:list (catalog entity name for this report).
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "list") {
			return view.Forbidden("job_outcome_summary:list")
		}

		// Backward-compat (phases.md P3): the tabbed section landing mounts ONLY
		// when the app configures List.Entity = "subscription_group". With
		// zero-valued options (service-admin) this renders EXACTLY today's flat
		// job_outcome_summary table. Any other List.Entity value is logged and
		// falls through to the flat list (Q-LIST-5 fail-safe).
		if entity := strings.TrimSpace(deps.Options.List.Entity); entity != "" && !deps.Options.List.SubscriptionGroups() {
			log.Printf("outcome summary list: unsupported List.Entity %q — rendering the flat list", entity)
		}
		if deps.Options.List.SubscriptionGroups() {
			return renderLanding(ctx, deps, viewCtx)
		}
		return renderFlat(ctx, deps, viewCtx)
	})
}

// renderFlat is the original flat job_outcome_summary table (unchanged).
func renderFlat(ctx context.Context, deps *ListViewDeps, viewCtx *view.ViewContext) view.ViewResult {
	resp, err := deps.ListJobOutcomeSummarys(ctx, &jobsumpb.ListJobOutcomeSummarysRequest{})
	if err != nil {
		log.Printf("Failed to list outcome summaries: %v", err)
		return view.Error(fmt.Errorf("failed to load outcome summaries: %w", err))
	}

	l := deps.Labels
	columns := summaryColumns(l)
	rows := buildTableRows(resp.GetData(), l, deps.Routes)
	types.ApplyColumnStyles(columns, rows)

	tableConfig := &types.TableConfig{
		ID:                   "outcome-summary-table",
		Columns:              columns,
		Rows:                 rows,
		ShowSearch:           true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "job",
		DefaultSortDirection: "desc",
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}

	// List page highlights "report-cards" in sidebar, not "jobs"
	activeSubNav := "report-cards"

	pageData := &PageData{
		PageData: types.PageData{
			CacheVersion:   viewCtx.CacheVersion,
			Title:          l.Page.JobHeading,
			CurrentPath:    viewCtx.CurrentPath,
			ActiveNav:      deps.Routes.ActiveNav,
			ActiveSubNav:   activeSubNav,
			HeaderTitle:    l.Page.JobHeading,
			HeaderSubtitle: l.Page.JobCaption,
			HeaderIcon:     "icon-award",
			CommonLabels:   deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-list-content",
		Table:           tableConfig,
	}

	return view.OK("outcome-summary-list", pageData)
}

// renderLanding renders view-1: a tabstrip of price_schedules (one tab per row,
// incl. inactive — Q-TAB-1) over a table of the selected schedule's
// subscription_groups (sections) with student/subject counts and a per-row view
// action into the section grid (view-2). All reads are workspace-bound at the
// espyna adapter (dbOps.List is workspace-aware); the landing composes no raw
// client_id/workspace filter.
func renderLanding(ctx context.Context, deps *ListViewDeps, viewCtx *view.ViewContext) view.ViewResult {
	l := deps.Labels

	// 1. price_schedules (ALL rows, incl. inactive — Q-TAB-1), ordered per the
	//    Tab options (sort_order NULLS LAST, name ASC — Q-SORT-3). The generic
	//    List defaults to active=true unless an explicit `active` BooleanFilter
	//    is supplied (core/operations.go), and a single boolean can't match
	//    both — so fetch active (default) + inactive (explicit active=false)
	//    and merge, giving the inactive AY its historical-browsing tab.
	schedules := listAllSchedules(ctx, deps)
	sortSchedules(schedules, deps.Options.Tab)

	// 2. selected tab: ?ps= wins; default = the first ACTIVE schedule in sort
	//    order (falling back to the first schedule when none is active). "Active"
	//    is a generic price_schedule field — this reconciles Q-TAB-1's "first by
	//    sort_order" with plan §3.2's "(active AY)" so the current period leads by
	//    default while inactive periods remain reachable via ?ps=.
	selected := strings.TrimSpace(viewCtx.Request.URL.Query().Get("ps"))
	if selected == "" {
		selected = defaultSchedule(schedules)
	}

	// 3. subscription_groups, ACTIVE + INACTIVE — workspace-scoped at the
	//    adapter. Historical schedules (Q-TAB-1's inactive tabs) hold only
	//    inactive groups, so the active-only default would render their tab
	//    with count 0 and an empty table; the same two-call merge as
	//    listAllSchedules keeps the historical view correct.
	groups := listAllGroups(ctx, deps)

	// 4. counts via ONE grouped read (no N+1): subjects = distinct templates per
	//    section, students = the section's largest per-subject cohort. The
	//    aggregate covers ACTIVE rows only, so historical sections then fill
	//    their counts from direct member/job reads (selected tab only).
	subjectCount, studentCount := sectionCounts(ctx, deps)
	fillHistoricalCounts(ctx, deps, groups, selected, subjectCount, studentCount)

	// 5. tabs (one per schedule; Count = active sections under it).
	tabs := buildTabs(schedules, groups, selected, l, deps.Routes)

	// 6. section rows for the selected tab, name ASC.
	rows := buildSectionRows(groups, selected, subjectCount, studentCount, l, deps.Routes)

	tableConfig := &types.TableConfig{
		ID:                   "report-cards-sections",
		Columns:              landingColumns(l),
		Rows:                 rows,
		ShowSearch:           true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowExport:           true,
		ShowEntries:          true,
		ShowActions:          true,
		DefaultSortColumn:    "section",
		DefaultSortDirection: "asc",
		Labels:               deps.TableLabels,
		Caption:              l.Landing.Title,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}
	types.ApplyColumnStyles(tableConfig.Columns, tableConfig.Rows)

	pageData := &PageData{
		PageData: types.PageData{
			CacheVersion:   viewCtx.CacheVersion,
			Title:          l.Landing.Title,
			CurrentPath:    viewCtx.CurrentPath,
			ActiveNav:      deps.Routes.ActiveNav,
			ActiveSubNav:   "report-cards",
			HeaderTitle:    l.Landing.Title,
			HeaderSubtitle: l.Landing.Subtitle,
			HeaderIcon:     "icon-award",
			CommonLabels:   deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-landing-content",
		Table:           tableConfig,
		Landing:         true,
		TabItems:        tabs,
		ActiveTab:       tabKey(selected),
		TabsAria:        l.Landing.TabsAriaLabel,
	}

	return view.OK("outcome-summary-list", pageData)
}

// listAllSchedules returns every price_schedule (active + inactive) so both AY
// tabs render (Q-TAB-1). The generic List defaults to active=true unless an
// explicit `active` BooleanFilter is present, and one boolean value can't span
// both — so we make one default (active) call plus one active=false call and
// concatenate. Nil-safe: no closure → empty slice → no tabs.
func listAllSchedules(ctx context.Context, deps *ListViewDeps) []*priceschedulepb.PriceSchedule {
	if deps.ListPriceSchedules == nil {
		return nil
	}
	out := make([]*priceschedulepb.PriceSchedule, 0, 4)
	if resp, err := deps.ListPriceSchedules(ctx, &priceschedulepb.ListPriceSchedulesRequest{}); err != nil {
		log.Printf("report cards landing: list active price schedules: %v", err)
	} else {
		out = append(out, resp.GetData()...)
	}
	if resp, err := deps.ListPriceSchedules(ctx, &priceschedulepb.ListPriceSchedulesRequest{
		Filters: &commonpb.FilterRequest{
			Filters: []*commonpb.TypedFilter{{
				Field:      "active",
				FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}},
			}},
		},
	}); err != nil {
		log.Printf("report cards landing: list inactive price schedules: %v", err)
	} else {
		out = append(out, resp.GetData()...)
	}
	return out
}

// listAllGroups returns every subscription_group (active + inactive) so
// historical schedules' sections stay visible. Same default-active +
// explicit-inactive merge as listAllSchedules. Nil-safe → empty slice.
func listAllGroups(ctx context.Context, deps *ListViewDeps) []*subscriptiongrouppb.SubscriptionGroup {
	if deps.ListSubscriptionGroups == nil {
		return nil
	}
	out := make([]*subscriptiongrouppb.SubscriptionGroup, 0, 16)
	if resp, err := deps.ListSubscriptionGroups(ctx, &subscriptiongrouppb.ListSubscriptionGroupsRequest{}); err != nil {
		log.Printf("report cards landing: list active subscription groups: %v", err)
	} else {
		out = append(out, resp.GetData()...)
	}
	if resp, err := deps.ListSubscriptionGroups(ctx, &subscriptiongrouppb.ListSubscriptionGroupsRequest{
		Filters: &commonpb.FilterRequest{
			Filters: []*commonpb.TypedFilter{{
				Field:      "active",
				FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}},
			}},
		},
	}); err != nil {
		log.Printf("report cards landing: list inactive subscription groups: %v", err)
	} else {
		out = append(out, resp.GetData()...)
	}
	return out
}

// sortSchedules orders price_schedules per the Tab options: when SortByOrder is
// set, sort_order ASC with NULLs LAST then name ASC (Q-SORT-3); otherwise name
// ASC. A "desc" direction reverses the primary key. Stable.
func sortSchedules(schedules []*priceschedulepb.PriceSchedule, opts outcome_summary.TabOptions) {
	desc := opts.Direction() == "desc"
	byOrder := opts.SortByOrder()
	sort.SliceStable(schedules, func(i, j int) bool {
		a, b := schedules[i], schedules[j]
		if byOrder {
			ai, aok := orderOf(a)
			bi, bok := orderOf(b)
			if aok != bok {
				// NULLS LAST regardless of direction.
				return aok
			}
			if aok && bok && ai != bi {
				if desc {
					return ai > bi
				}
				return ai < bi
			}
		}
		an := strings.ToLower(a.GetName())
		bn := strings.ToLower(b.GetName())
		if an == bn {
			return a.GetId() < b.GetId()
		}
		if desc && byOrder {
			// name stays the ascending tiebreak even under desc sort_order.
			return an < bn
		}
		if desc {
			return an > bn
		}
		return an < bn
	})
}

// defaultSchedule returns the id of the first ACTIVE schedule in the (already
// sorted) slice, falling back to the first schedule when none is active, or ""
// when the slice is empty.
func defaultSchedule(schedules []*priceschedulepb.PriceSchedule) string {
	for _, s := range schedules {
		if s.GetActive() {
			return s.GetId()
		}
	}
	if len(schedules) > 0 {
		return schedules[0].GetId()
	}
	return ""
}

// orderOf returns the schedule's sort_order and whether it is set (NULL = not
// set → sorts last).
func orderOf(s *priceschedulepb.PriceSchedule) (int32, bool) {
	if s != nil && s.SortOrder != nil {
		return s.GetSortOrder(), true
	}
	return 0, false
}

// sectionCounts derives per-section (subscription_group) subject + student
// counts from ONE ListJobTemplateSummaries grouped read (staff-scoped at the
// adapter): subjects = distinct job_template rows for the group; students = the
// group's largest per-subject job_count (its roster size for a section where
// every student shares ≥1 subject). Nil-safe → empty maps (blank counts).
func sectionCounts(ctx context.Context, deps *ListViewDeps) (subjects map[string]int, students map[string]int) {
	subjects = map[string]int{}
	students = map[string]int{}
	if deps.ListJobTemplateSummaries == nil {
		return
	}
	resp, err := deps.ListJobTemplateSummaries(ctx, &summarypb.ListJobTemplateSummariesRequest{})
	if err != nil {
		log.Printf("report cards landing: list job template summaries: %v", err)
		return
	}
	seen := map[string]map[string]bool{}
	for _, s := range resp.GetSummaries() {
		gid := s.GetSubscriptionGroupId()
		if gid == "" {
			continue
		}
		if seen[gid] == nil {
			seen[gid] = map[string]bool{}
		}
		if tid := s.GetJobTemplateId(); tid != "" && !seen[gid][tid] {
			seen[gid][tid] = true
			subjects[gid]++
		}
		if jc := int(s.GetJobCount()); jc > students[gid] {
			students[gid] = jc
		}
	}
	return
}

// fillHistoricalCounts fills Students/Subjects counts for the selected tab's
// INACTIVE sections (their rows are invisible to the active-bound
// ListJobTemplateSummaries aggregate): students = distinct member clients
// (active + inactive members — the frozen roster), subjects = distinct
// job_templates over the members' jobs (active + inactive). Bulk reads, no
// N+1: one member read (+inactive merge) and chunked job reads across the
// tab's groups. Nil-safe; missing closures leave counts blank.
func fillHistoricalCounts(
	ctx context.Context,
	deps *ListViewDeps,
	groups []*subscriptiongrouppb.SubscriptionGroup,
	selected string,
	subjectCount, studentCount map[string]int,
) {
	if deps.ListSubscriptionGroupMembers == nil {
		return
	}
	var groupIDs []string
	for _, g := range groups {
		if g.GetActive() || (selected != "" && g.GetPriceScheduleId() != selected) {
			continue
		}
		if studentCount[g.GetId()] == 0 || subjectCount[g.GetId()] == 0 {
			groupIDs = append(groupIDs, g.GetId())
		}
	}
	if len(groupIDs) == 0 {
		return
	}

	// members (both liveness states — the listAllSchedules merge pattern).
	// One call PER GROUP so each result set stays under the adapter's default
	// list cap (a bulk multi-group read truncates and undercounts).
	subToGroup := map[string]string{}               // subscription_id -> group_id
	clientsPerGroup := map[string]map[string]bool{} // group_id -> distinct clients
	for _, gid := range groupIDs {
		requests := []*subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
			{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("subscription_group_id", gid)}}},
			{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
				stringEq("subscription_group_id", gid),
				boolEq("active", false),
			}}},
		}
		for _, req := range requests {
			resp, err := deps.ListSubscriptionGroupMembers(ctx, req)
			if err != nil {
				log.Printf("report cards landing: historical member counts: %v", err)
				continue
			}
			for _, m := range resp.GetData() {
				sid, cid := m.GetSubscriptionId(), m.GetClientId()
				if sid == "" || cid == "" {
					continue
				}
				subToGroup[sid] = gid
				if clientsPerGroup[gid] == nil {
					clientsPerGroup[gid] = map[string]bool{}
				}
				clientsPerGroup[gid][cid] = true
			}
		}
	}
	for gid, clients := range clientsPerGroup {
		if studentCount[gid] == 0 {
			studentCount[gid] = len(clients)
		}
	}

	// subjects: distinct templates over the members' jobs (both liveness
	// states), chunked by origin_id PER GROUP (one section's job set is the
	// scale the section grid already reads in one call).
	if deps.ListJobs == nil || len(subToGroup) == 0 {
		return
	}
	subsPerGroup := map[string][]string{}
	for sid, gid := range subToGroup {
		subsPerGroup[gid] = append(subsPerGroup[gid], sid)
	}
	const chunk = 100
	for gid, subIDs := range subsPerGroup {
		tmpls := map[string]bool{}
		for start := 0; start < len(subIDs); start += chunk {
			end := start + chunk
			if end > len(subIDs) {
				end = len(subIDs)
			}
			jobReqs := []*jobpb.ListJobsRequest{
				{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("origin_id", subIDs[start:end])}}},
				{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
					listIn("origin_id", subIDs[start:end]),
					boolEq("active", false),
				}}},
			}
			for _, req := range jobReqs {
				// Page explicitly — a section's job set exceeds the adapters'
				// default row caps (the fetchSectionJobs lesson). maxPages
				// bounds the loop even if the adapter ignored OFFSET.
				for page := int32(1); page <= maxLandingPages; page++ {
					req.Pagination = &commonpb.PaginationRequest{
						Limit:  int32(chunk),
						Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
					}
					resp, err := deps.ListJobs(ctx, req)
					if err != nil {
						log.Printf("report cards landing: historical subject counts (page %d): %v", page, err)
						break
					}
					for _, j := range resp.GetData() {
						if tid := j.GetJobTemplateId(); tid != "" {
							tmpls[tid] = true
						}
					}
					if len(resp.GetData()) < chunk {
						break
					}
				}
			}
		}
		if subjectCount[gid] == 0 {
			subjectCount[gid] = len(tmpls)
		}
	}
}

// buildTabs builds one TabItem per schedule; Count = its sections (active or
// not — a historical schedule's sections are all inactive but still count);
// the tab whose id == selected is marked active. Href carries ?ps=<id>.
func buildTabs(
	schedules []*priceschedulepb.PriceSchedule,
	groups []*subscriptiongrouppb.SubscriptionGroup,
	selected string,
	l outcome_summary.Labels,
	routes outcome_summary.Routes,
) []pyeza.TabItem {
	counts := map[string]int{}
	for _, g := range groups {
		counts[g.GetPriceScheduleId()]++
	}
	tabs := make([]pyeza.TabItem, 0, len(schedules))
	for _, s := range schedules {
		label := s.GetName()
		if !s.GetActive() && l.Landing.InactiveSuffix != "" {
			label = label + " " + l.Landing.InactiveSuffix
		}
		tabs = append(tabs, pyeza.TabItem{
			Key:   tabKey(s.GetId()),
			Label: label,
			Href:  routes.ListURL + "?ps=" + s.GetId(),
			Count: counts[s.GetId()],
		})
	}
	return tabs
}

// buildSectionRows builds the section table rows for the selected schedule
// (inactive groups included — the historical view), name ASC. Each row links
// (view action) into the per-section grid.
func buildSectionRows(
	groups []*subscriptiongrouppb.SubscriptionGroup,
	selected string,
	subjectCount, studentCount map[string]int,
	l outcome_summary.Labels,
	routes outcome_summary.Routes,
) []types.TableRow {
	var filtered []*subscriptiongrouppb.SubscriptionGroup
	for _, g := range groups {
		if selected != "" && g.GetPriceScheduleId() != selected {
			continue
		}
		filtered = append(filtered, g)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		a := strings.ToLower(filtered[i].GetName())
		b := strings.ToLower(filtered[j].GetName())
		if a == b {
			return filtered[i].GetId() < filtered[j].GetId()
		}
		return a < b
	})

	rows := make([]types.TableRow, 0, len(filtered))
	for _, g := range filtered {
		gid := g.GetId()
		rows = append(rows, types.TableRow{
			ID:        gid,
			DataAttrs: map[string]string{"testid": "rc-section-" + short(gid)},
			Cells: []types.TableCell{
				{Value: g.GetName()},
				{Value: fmt.Sprintf("%d", studentCount[gid])},
				{Value: fmt.Sprintf("%d", subjectCount[gid])},
			},
			Actions: []types.TableAction{
				{
					Type:   "view",
					Label:  l.Landing.ViewAction,
					Href:   route.ResolveURL(routes.SectionURL, "id", gid),
					TestID: "rc-view-" + short(gid),
				},
				{
					// The download JS appends ?id=<row id> (this section's own
					// id) — the export handler ignores it and serves the full
					// section grid.
					Type:   "download",
					Label:  l.Landing.DownloadAction,
					Action: "download",
					URL:    route.ResolveURL(routes.SectionExportURL, "id", gid),
					TestID: "rc-download-" + short(gid),
				},
			},
		})
	}
	return rows
}

func landingColumns(l outcome_summary.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "section", Label: l.Landing.GroupColumn, MinWidth: "12.5rem"},
		{Key: "students", Label: l.Landing.MembersColumn, MinWidth: "6.25rem", Align: "right"},
		{Key: "subjects", Label: l.Landing.TemplatesColumn, MinWidth: "6.25rem", Align: "right"},
	}
}

// tabKey builds a stable, greppable tab key from a price_schedule id.
func tabKey(psID string) string {
	if psID == "" {
		return ""
	}
	return "rc-tab-" + short(psID)
}

// short truncates an opaque id for a stable key/testid suffix.
// short returns a stable, collision-resistant slug for keys/testids. It takes
// the LAST 8 chars (the uuidv7 random tail), NOT the first 8 — every education1
// entity is a uuidv7 minted in one workspace/batch, so they share the ~13-char
// timestamp PREFIX (all "019ecb8e-…"): a first-8 slug collides across every
// row/cell/tab, which silently marked all tabstrip items active (aria-selected
// key collision) and made testids non-unique. The random tail is unique.
func short(id string) string {
	if len(id) > 8 {
		return id[len(id)-8:]
	}
	return id
}

func summaryColumns(l outcome_summary.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "job", Label: l.Columns.Job, MinWidth: "8.75rem"},
		{Key: "determination", Label: l.Detail.OverallDetermination, MinWidth: "7.5rem"},
		{Key: "score", Label: l.Detail.Score, MinWidth: "5rem"},
		{Key: "scoring_method", Label: l.Detail.ScoringMethod, MinWidth: "7.5rem"},
		{Key: "total", Label: l.Detail.TotalCriteria, MinWidth: "3.75rem"},
		{Key: "pass", Label: l.Detail.PassCount, MinWidth: "3.75rem"},
		{Key: "fail", Label: l.Detail.FailCount, MinWidth: "3.75rem"},
		{Key: "issued_by", Label: l.Detail.IssuedBy, MinWidth: "6.25rem"},
	}
}

func buildTableRows(summaries []*jobsumpb.JobOutcomeSummary, l outcome_summary.Labels, routes outcome_summary.Routes) []types.TableRow {
	var rows []types.TableRow
	for _, s := range summaries {
		determination := overallDeterminationString(s.GetOverallDetermination())
		variant := overallDeterminationVariant(s.GetOverallDetermination())

		row := types.TableRow{
			ID: s.GetId(),
			Cells: []types.TableCell{
				{Value: s.GetJobId()},
				{Type: "badge", Value: determination, Variant: variant},
				{Value: fmt.Sprintf("%.2f", s.GetSummaryScore())},
				{Value: s.GetScoringMethod().String()},
				{Value: fmt.Sprintf("%d", s.GetTotalCriteriaCount())},
				{Value: fmt.Sprintf("%d", s.GetPassCount())},
				{Value: fmt.Sprintf("%d", s.GetFailCount())},
				{Value: s.GetIssuedBy()},
			},
			Actions: []types.TableAction{
				{Type: "view", Label: "View Summary", Href: strings.NewReplacer("{id}", s.GetJobId()).Replace(routes.JobSummaryURL)},
			},
		}
		rows = append(rows, row)
	}
	return rows
}

func overallDeterminationString(d enums.OverallDetermination) string {
	switch d {
	case enums.OverallDetermination_OVERALL_DETERMINATION_ACCEPTED:
		return "Accepted"
	case enums.OverallDetermination_OVERALL_DETERMINATION_CONDITIONALLY_ACCEPTED:
		return "Conditional"
	case enums.OverallDetermination_OVERALL_DETERMINATION_REJECTED:
		return "Rejected"
	case enums.OverallDetermination_OVERALL_DETERMINATION_IN_PROGRESS:
		return "In Progress"
	case enums.OverallDetermination_OVERALL_DETERMINATION_DEFERRED:
		return "Deferred"
	default:
		return "Unspecified"
	}
}

func overallDeterminationVariant(d enums.OverallDetermination) string {
	switch d {
	case enums.OverallDetermination_OVERALL_DETERMINATION_ACCEPTED:
		return "success"
	case enums.OverallDetermination_OVERALL_DETERMINATION_CONDITIONALLY_ACCEPTED:
		return "warning"
	case enums.OverallDetermination_OVERALL_DETERMINATION_REJECTED:
		return "danger"
	case enums.OverallDetermination_OVERALL_DETERMINATION_IN_PROGRESS:
		return "info"
	case enums.OverallDetermination_OVERALL_DETERMINATION_DEFERRED:
		return "default"
	default:
		return "default"
	}
}
