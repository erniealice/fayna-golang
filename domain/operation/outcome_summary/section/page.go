// Package section renders view-2 of the report-cards surface: a per-section
// (subscription_group) grid of students × subjects, each cell = the student's
// year-final rating (job_outcome_summary.scaled_label) for that subject,
// linking out to the existing per-job summary. It is a READ-ONLY reporting
// table (types.TableConfig + Groups gender bands) — the editing surface is the
// grade sheet (outcome_matrix).
//
// Security (Q-SEC-7): the section id is EXISTS-gated against the session
// workspace (the workspace-aware ListSubscriptionGroups adapter returns a
// foreign-workspace group as no-rows → fail-closed). Every read is
// workspace-bound at the espyna adapter; the row/column set is derived from the
// section's JOBS (ListJobs), which the adapter narrows to the acting STAFF
// principal's reachable jobs — so a teacher sees only their students' cards.
package section

import (
	"context"
	"html"
	"log"
	"sort"
	"strconv"
	"strings"

	texttemplate "html/template"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	subscriptiongroupworkspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_workspace_user"
)

// actionsColumnKey is the column Key of the frozen per-row action cell (view
// student card + CSV download). It is a UI control, not report data, so the CSV
// export skips it (header + each row's cell). Declared once here (T8) and shared
// by buildColumns + the export handler so a rename can never desync the two —
// a mismatch would leak raw HTML action anchors into every CSV row.
const actionsColumnKey = "rc-actions"

// pageLimit chunks ListFilter(IN) id sets so each call's result set stays under
// the adapter's default cap (the fetchClientNames pattern).
const pageLimit = 100

// maxPages bounds every offset page-loop independently of the adapter's own
// termination (which relies on a short final page). A section's job set is
// ≤ roster×subjects (≈300–500); this ceiling (100 pages × 100 rows = 10k) is
// far above any real section yet guarantees the loop halts even if a
// misbehaving adapter ignored OFFSET and returned a full page forever.
const maxPages = 100

// downloadIcon is the inline SVG for the per-row CSV download button. The
// section grid renders the download as the frozen SECOND column (an HTML cell),
// not a trailing actions cell, so it needs the icon markup inline (mirrors
// pyeza's icon-download) rather than via a {{template}} call.
const downloadIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>`

// viewIcon is the inline SVG (mirrors pyeza's icon-eye) for the per-row "view
// student card" action rendered in the frozen actions column.
const viewIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/></svg>`

// Deps holds the section-grid view dependencies. All list closures are
// workspace-bound at the espyna adapter; the view composes no client_id/
// workspace filter of its own.
type Deps struct {
	Routes       outcome_summary.Routes
	Labels       outcome_summary.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Options — app-configured row presentation (bands + sort) + CategoryFilter
	// (a job_category code, e.g. "academic"). Zero value → flat rows (no bands),
	// name sort, and no category filter.
	Options outcome_summary.Options

	// ListJobCategories resolves Options.CategoryFilter to its id so same-origin
	// deportment jobs are dropped from the academic grid (gate H2). Optional/
	// nil-safe (nil closure or empty code → no filter).
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)

	ListSubscriptionGroups       func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListSubscriptionGroupMembers func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListJobs                     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListJobTemplates             func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ListClients                  func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
	ListJobOutcomeSummarys       func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)
	ListClientAttributes         func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ResolveAttributeIDByCode     func(ctx context.Context, code string) (string, error)

	// Non-enrolled-placeholder evidence walk (job_phase → job_task →
	// task_outcome). Optional/nil-safe: when any is nil (a tier that never wired
	// the walk, e.g. service-admin) FetchJobMarkEvidence yields empty evidence and
	// no cell is blanked — the flat surface is byte-identical.
	ListJobPhases    func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)
	ListJobTasks     func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	ListTaskOutcomes func(ctx context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error)

	// Header-caption deps: the group's servicing grants
	// (subscription_group_workspace_user) + the User-hydrating workspace-member
	// list that resolves their display names. Optional/nil-safe: missing →
	// the caption falls back to the lyngua'd detail-link label.
	ListSubscriptionGroupWorkspaceUsers func(ctx context.Context, req *subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersRequest) (*subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersResponse, error)
	ListWorkspaceUsers                  func(ctx context.Context, req *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error)
}

// PageData is the section-grid page data.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	NotComputed     bool
	Banner          string
}

// student is one row's resolved identity.
type student struct {
	clientID  string
	name      string
	lastName  string
	firstName string
}

// listName renders the class-list name form "{last_name}, {first_name}"
// (prod's report-card roster format), falling back to the plain display name
// when either part is missing.
func (s student) listName() string {
	if s.lastName != "" && s.firstName != "" {
		return s.lastName + ", " + s.firstName
	}
	return s.name
}

// NewView creates the per-section report-card grid view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "list") {
			return view.Forbidden("job_outcome_summary:list")
		}

		sectionID := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		if sectionID == "" {
			return view.Forbidden("job_outcome_summary:list")
		}

		group, table := buildSectionTable(ctx, deps, sectionID)
		if group == nil {
			return view.Forbidden("job_outcome_summary:list")
		}
		grantHolders := fetchGrantHolderNames(ctx, deps, sectionID)
		if table == nil {
			// Empty-state: no computed summaries for this section → banner, not
			// a blank grid (D.4 do-not-ship-blank).
			return okPage(viewCtx, deps, group, grantHolders, nil, deps.Labels.Section.NotComputedBanner)
		}
		return okPage(viewCtx, deps, group, grantHolders, table, "")
	})
}

// fetchGrantHolderNames resolves the group's OWNER servicing-grant holders
// (subscription_group_workspace_user rows with is_owner — the group's
// lead(s); co-owners possible) to display names, name-ASC. Non-owner grants
// are deliberately excluded from the caption. The name hop rides the
// User-hydrating workspace-member list (one call — it returns the whole
// workspace's members, so build a map). Nil-safe: missing closures, no owner
// grants, or unresolvable names → nil (caption falls back to the lyngua'd
// detail-link label).
func fetchGrantHolderNames(ctx context.Context, deps *Deps, sectionID string) []string {
	if deps.ListSubscriptionGroupWorkspaceUsers == nil || deps.ListWorkspaceUsers == nil {
		return nil
	}
	resp, err := deps.ListSubscriptionGroupWorkspaceUsers(ctx, &subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersRequest{
		Filters: &commonpb.FilterRequest{
			Filters: []*commonpb.TypedFilter{stringEq("subscription_group_id", sectionID)},
		},
	})
	if err != nil {
		log.Printf("report cards section: list group workspace users: %v", err)
		return nil
	}
	owners := map[string]bool{} // workspace_user_id of is_owner grants
	for _, g := range resp.GetData() {
		if g.GetActive() && g.GetIsOwner() && g.GetWorkspaceUserId() != "" {
			owners[g.GetWorkspaceUserId()] = true
		}
	}
	if len(owners) == 0 {
		return nil
	}
	wuResp, err := deps.ListWorkspaceUsers(ctx, &workspaceuserpb.ListWorkspaceUsersRequest{})
	if err != nil {
		log.Printf("report cards section: list workspace users: %v", err)
		return nil
	}
	var names []string
	for _, wu := range wuResp.GetData() {
		if !owners[wu.GetId()] {
			continue
		}
		if u := wu.GetUser(); u != nil {
			if name := strings.TrimSpace(u.GetFirstName() + " " + u.GetLastName()); name != "" {
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	return names
}

// buildSectionTable assembles the per-section grid: the workspace-gated group
// plus a fully-ordered TableConfig (bands + rows + cells). Shared by the HTML
// view and the CSV export handler so both render the identical grid.
// Returns (nil, nil) when the group fails the workspace EXISTS gate and
// (group, nil) when the section has no computed summaries yet.
func buildSectionTable(ctx context.Context, deps *Deps, sectionID string) (*subscriptiongrouppb.SubscriptionGroup, *types.TableConfig) {
	// EXISTS gate: the group must belong to the session workspace. The
	// ListSubscriptionGroups adapter is workspace-scoped, so a foreign or
	// missing id returns no rows → fail-closed (no leak).
	group := fetchSection(ctx, deps, sectionID)
	if group == nil {
		return nil, nil
	}

	l := deps.Labels

	// Historical mode: an inactive group is a FROZEN past section — its
	// members (and possibly jobs) are inactive rows, so the liveness filters
	// below relax to render the roster as it stood.
	historical := !group.GetActive()

	// members(section) → subscription_id → client_id.
	subToClient := fetchMembers(ctx, deps, sectionID, historical)

	// jobs(origin_id IN subs, active, SUBSCRIPTION) — staff-narrowed at the
	// adapter. Rows + columns are derived from THIS set (Q-SEC-7).
	jobs := fetchSectionJobs(ctx, deps, keysOf(subToClient), historical)

	// H2: keep only the configured category's subjects (academic), dropping
	// same-origin deportment jobs that would otherwise render as columns. catID
	// "" with catOK=true (no filter) keeps every job; catOK=false (a configured
	// filter that could not be resolved) fails CLOSED — the grid drops every job.
	catID, catOK := outcome_summary.ResolveCategoryID(ctx, deps.ListJobCategories, deps.Options.CategoryFilter)

	// Build the (client, template) → job map + distinct sets.
	cellJob := map[string]string{} // clientID+"\x00"+templateID -> jobID
	clientIDs := []string{}
	clientSeen := map[string]bool{}
	templateIDs := []string{}
	tmplSeen := map[string]bool{}
	jobIDs := []string{}
	for _, j := range jobs {
		if !catOK || !outcome_summary.KeepJobInCategory(catID, j.GetJobCategoryId()) {
			continue
		}
		jobID := j.GetId()
		tid := j.GetJobTemplateId()
		clientID := subToClient[j.GetOriginId()]
		if clientID == "" {
			clientID = j.GetClientId()
		}
		if clientID == "" || tid == "" || jobID == "" {
			continue
		}
		if !clientSeen[clientID] {
			clientSeen[clientID] = true
			clientIDs = append(clientIDs, clientID)
		}
		if !tmplSeen[tid] {
			tmplSeen[tid] = true
			templateIDs = append(templateIDs, tid)
		}
		cellJob[clientID+"\x00"+tid] = jobID
		jobIDs = append(jobIDs, jobID)
	}

	// summaries(job_id IN jobs) → jobID → scaled label.
	labelByJob := fetchSummaryLabels(ctx, deps, jobIDs, l)
	if len(labelByJob) == 0 {
		return group, nil
	}

	// Non-enrolled-placeholder evidence: one bulk job_phase → job_task →
	// task_outcome walk keyed by the section's jobs. An untaken-elective
	// scaffold rides in with an all-zero task set and a floored ("1") year-final;
	// its cell must render BLANK (matching prod), not the floor. A genuinely
	// enrolled subject — even one scored a real 0/1 — carries a positive task
	// mark and is kept. Nil-safe: unwired closures → empty map → nothing blanked.
	// Fail-closed: on a read error the evidence is incomplete, so we keep every
	// grade (blank nothing) rather than risk blanking a real one.
	evByJob, err := outcome_summary.FetchJobMarkEvidence(ctx, deps.ListJobPhases, deps.ListJobTasks, deps.ListTaskOutcomes, jobIDs)
	if err != nil {
		log.Printf("outcome summary section: enrollment evidence unavailable, keeping all grades: %v", err)
		evByJob = nil
	}

	// client display names + last_name (for the row sort).
	students := fetchStudents(ctx, deps, clientIDs)

	// gender (or configured) attribute values for bands.
	attrValues := fetchAttributeValues(ctx, deps, clientIDs)

	// template names (columns), name ASC.
	tmplNames := fetchTemplateNames(ctx, deps, templateIDs, historical)
	columns := buildColumns(templateIDs, tmplNames, l)

	table := &types.TableConfig{
		ID:          "report-cards-grid",
		Columns:     columns,
		ShowSearch:  true,
		ShowColumns: true,
		ShowDensity: true,
		ShowExport:  true,
		ShowEntries: true,
		// The per-row download is now the frozen SECOND column, not a trailing
		// actions cell — so no trailing actions column.
		ShowActions: false,
		// Freeze the first two columns (student + download) while the subject
		// columns scroll horizontally (pyeza generic, ID-agnostic).
		TableClass:  "data-table-freeze2",
		Labels:      deps.TableLabels,
		Caption:     l.Section.Title,
		FixedLayout: false,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Section.NotComputedBanner,
		},
	}

	rows := buildRows(students, orderedColumnIDs(columns), cellJob, labelByJob, evByJob, sectionID, deps.Routes, l)
	applyRowPresentation(table, rows, students, attrValues, deps.Options)
	numberRows(table)
	types.ApplyColumnStyles(table.Columns, allRows(table))
	return group, table
}

// numberRows prefixes each student cell with its sequence number in final
// presentation order — CONTINUOUS across bands (prod's class-list numbering:
// male 1..N, female N+1..M), applied after banding/sorting so the numbers
// reflect what renders. The CSV export shares the cell value, so exports
// carry the same "{n} {last}, {first}" text.
func numberRows(table *types.TableConfig) {
	n := 0
	number := func(rows []types.TableRow) {
		for i := range rows {
			n++
			if len(rows[i].Cells) > 0 {
				rows[i].Cells[0].Value = strconv.Itoa(n) + " " + rows[i].Cells[0].Value
			}
		}
	}
	if len(table.Groups) > 0 {
		for gi := range table.Groups {
			number(table.Groups[gi].Rows)
		}
		return
	}
	number(table.Rows)
}

// okPage assembles the PageData (grid or empty-state banner). Header shape:
// breadcrumb "<list title> > <group name>" (crumb links back to the landing)
// with a caption below linking to the listed group's own detail page
// (GroupDetailURL — /sections/detail/{id} on education). The caption text is
// the group's servicing-grant holder names (comma-separated — a group can
// carry several subscription_group_workspace_user rows), falling back to the
// lyngua'd DetailLink label when no grants resolve.
func okPage(viewCtx *view.ViewContext, deps *Deps, group *subscriptiongrouppb.SubscriptionGroup, grantHolders []string, table *types.TableConfig, banner string) view.ViewResult {
	l := deps.Labels
	caption := strings.Join(grantHolders, ", ")
	if caption == "" {
		caption = l.Section.DetailLink
	}
	pd := &PageData{
		PageData: types.PageData{
			CacheVersion:        viewCtx.CacheVersion,
			Title:               l.Section.Title,
			CurrentPath:         viewCtx.CurrentPath,
			ActiveNav:           deps.Routes.ActiveNav,
			ActiveSubNav:        "report-cards",
			HeaderBreadcrumb:    l.Section.Title,
			HeaderBreadcrumbURL: deps.Routes.ListURL,
			HeaderTitle:         group.GetName(),
			HeaderSubtitle:      caption,
			HeaderSubtitleURL:   route.ResolveURL(deps.Routes.GroupDetailURL, "id", group.GetId()),
			HeaderIcon:          "icon-award",
			CommonLabels:        deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-section-content",
		Table:           table,
		NotComputed:     table == nil,
		Banner:          banner,
	}
	return view.OK("outcome-summary-section", pd)
}

// fetchSection returns the section (workspace-scoped EXISTS gate) or nil.
// Historical (inactive) groups resolve too — the generic List defaults to
// active=true, so a second explicit active=false read covers them (the
// listAllSchedules pattern). The workspace scope applies to both reads.
func fetchSection(ctx context.Context, deps *Deps, sectionID string) *subscriptiongrouppb.SubscriptionGroup {
	if deps.ListSubscriptionGroups == nil {
		return nil
	}
	requests := []*subscriptiongrouppb.ListSubscriptionGroupsRequest{
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("id", sectionID)}}},
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
			stringEq("id", sectionID),
			{Field: "active", FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}}},
		}}},
	}
	for _, req := range requests {
		resp, err := deps.ListSubscriptionGroups(ctx, req)
		if err != nil {
			log.Printf("report cards section: list subscription group by id: %v", err)
			continue
		}
		for _, g := range resp.GetData() {
			if g.GetId() == sectionID {
				return g
			}
		}
	}
	return nil
}

// fetchMembers returns subscription_id → client_id for the section's members
// (active only — or the frozen full roster in historical mode). The generic
// List defaults to active=true rows, so historical mode adds an explicit
// active=false read (the listAllSchedules merge pattern).
func fetchMembers(ctx context.Context, deps *Deps, sectionID string, historical bool) map[string]string {
	out := map[string]string{}
	if deps.ListSubscriptionGroupMembers == nil {
		return out
	}
	requests := []*subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("subscription_group_id", sectionID)}}},
	}
	if historical {
		requests = append(requests, &subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
				stringEq("subscription_group_id", sectionID),
				{Field: "active", FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}}},
			}},
		})
	}
	for _, req := range requests {
		resp, err := deps.ListSubscriptionGroupMembers(ctx, req)
		if err != nil {
			log.Printf("report cards section: list members: %v", err)
			continue
		}
		for _, m := range resp.GetData() {
			if !historical && !m.GetActive() {
				continue
			}
			if sid, cid := m.GetSubscriptionId(), m.GetClientId(); sid != "" && cid != "" {
				out[sid] = cid
			}
		}
	}
	return out
}

// fetchSectionJobs returns the section's active subscription-origin jobs
// (staff-narrowed at the adapter), chunked by origin_id. Historical mode also
// accepts inactive jobs (a frozen past section's jobs may be retired) via an
// extra explicit active=false read per chunk.
//
// Every read PAGES explicitly (Limit 100 + offset pages) until exhausted: a
// section's job set (roster × subjects ≈ 300) far exceeds the adapters'
// default row caps, and an uncapped single call silently truncates — the
// grid then renders the missing cells as "—" with no error.
func fetchSectionJobs(ctx context.Context, deps *Deps, subIDs []string, historical bool) []*jobpb.Job {
	var out []*jobpb.Job
	if deps.ListJobs == nil || len(subIDs) == 0 {
		return out
	}
	seen := map[string]bool{}
	for start := 0; start < len(subIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(subIDs) {
			end = len(subIDs)
		}
		filterSets := [][]*commonpb.TypedFilter{
			{listIn("origin_id", subIDs[start:end])},
		}
		if historical {
			filterSets = append(filterSets, []*commonpb.TypedFilter{
				listIn("origin_id", subIDs[start:end]),
				{Field: "active", FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}}},
			})
		}
		for _, filters := range filterSets {
			for page := int32(1); page <= maxPages; page++ {
				resp, err := deps.ListJobs(ctx, &jobpb.ListJobsRequest{
					Filters: &commonpb.FilterRequest{Filters: filters},
					Pagination: &commonpb.PaginationRequest{
						Limit:  int32(pageLimit),
						Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
					},
				})
				if err != nil {
					log.Printf("report cards section: list jobs (page %d): %v", page, err)
					break
				}
				for _, j := range resp.GetData() {
					if seen[j.GetId()] {
						continue
					}
					seen[j.GetId()] = true
					if !historical && !j.GetActive() {
						continue
					}
					if j.GetOriginType() == enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION {
						out = append(out, j)
					}
				}
				if len(resp.GetData()) < pageLimit {
					break
				}
			}
		}
	}
	return out
}

// fetchSummaryLabels returns jobID → year-final label (scaled_label, falling
// back to a formatted scaled_score), chunked by job_id.
func fetchSummaryLabels(ctx context.Context, deps *Deps, jobIDs []string, l outcome_summary.Labels) map[string]string {
	out := map[string]string{}
	if deps.ListJobOutcomeSummarys == nil || len(jobIDs) == 0 {
		return out
	}
	for start := 0; start < len(jobIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(jobIDs) {
			end = len(jobIDs)
		}
		resp, err := deps.ListJobOutcomeSummarys(ctx, &jobsumpb.ListJobOutcomeSummarysRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{listIn("job_id", jobIDs[start:end])},
			},
		})
		if err != nil {
			log.Printf("report cards section: list job outcome summaries: %v", err)
			continue
		}
		for _, s := range resp.GetData() {
			if !s.GetActive() {
				continue
			}
			jid := s.GetJobId()
			if jid == "" {
				continue
			}
			if lbl := strings.TrimSpace(s.GetScaledLabel()); lbl != "" {
				out[jid] = lbl
			} else if s.ScaledScore != nil {
				out[jid] = strconv.FormatFloat(s.GetScaledScore(), 'f', -1, 64)
			}
		}
	}
	return out
}

// fetchStudents resolves client_id → display name + last_name, chunked.
func fetchStudents(ctx context.Context, deps *Deps, clientIDs []string) map[string]student {
	out := map[string]student{}
	if deps.ListClients == nil || len(clientIDs) == 0 {
		return out
	}
	for start := 0; start < len(clientIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(clientIDs) {
			end = len(clientIDs)
		}
		resp, err := deps.ListClients(ctx, &clientpb.ListClientsRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{listIn("id", clientIDs[start:end])},
			},
		})
		if err != nil {
			log.Printf("report cards section: list clients: %v", err)
			continue
		}
		for _, c := range resp.GetData() {
			id := c.GetId()
			if id == "" {
				continue
			}
			out[id] = student{clientID: id, name: clientDisplayName(c), lastName: clientLastName(c), firstName: clientFirstName(c)}
		}
	}
	return out
}

// fetchAttributeValues resolves each Options-referenced attribute code to its id
// then loads the roster clients' values — code → (client_id → value). Mirrors
// the outcome_matrix hydration. Nil-safe.
func fetchAttributeValues(ctx context.Context, deps *Deps, clientIDs []string) map[string]map[string]string {
	codes := deps.Options.AttributeCodes()
	if len(codes) == 0 || deps.ListClientAttributes == nil || deps.ResolveAttributeIDByCode == nil || len(clientIDs) == 0 {
		return nil
	}
	out := make(map[string]map[string]string, len(codes))
	for _, code := range codes {
		attrID, err := deps.ResolveAttributeIDByCode(ctx, code)
		if err != nil || attrID == "" {
			log.Printf("report cards section: attribute code %q did not resolve (bands ignored for it): %v", code, err)
			continue
		}
		vals := map[string]string{}
		for start := 0; start < len(clientIDs); start += pageLimit {
			end := start + pageLimit
			if end > len(clientIDs) {
				end = len(clientIDs)
			}
			resp, err := deps.ListClientAttributes(ctx, &clientattributepb.ListClientAttributesRequest{
				Filters: &commonpb.FilterRequest{
					Filters: []*commonpb.TypedFilter{
						stringEq("attribute_id", attrID),
						listIn("client_id", clientIDs[start:end]),
					},
				},
			})
			if err != nil {
				log.Printf("report cards section: list client attributes for %q: %v", code, err)
				continue
			}
			for _, ca := range resp.GetData() {
				if cid, v := ca.GetClientId(), strings.TrimSpace(ca.GetValue()); cid != "" && v != "" {
					vals[cid] = v
				}
			}
		}
		out[code] = vals
	}
	return out
}

// fetchTemplateNames resolves job_template_id → name, chunked. Historical
// mode also reads inactive templates (a frozen year's subjects are retired
// from active curricula) — without it the columns fall back to raw ids.
func fetchTemplateNames(ctx context.Context, deps *Deps, templateIDs []string, historical bool) map[string]string {
	out := map[string]string{}
	if deps.ListJobTemplates == nil || len(templateIDs) == 0 {
		return out
	}
	for start := 0; start < len(templateIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(templateIDs) {
			end = len(templateIDs)
		}
		filterSets := [][]*commonpb.TypedFilter{
			{listIn("id", templateIDs[start:end])},
		}
		if historical {
			filterSets = append(filterSets, []*commonpb.TypedFilter{
				listIn("id", templateIDs[start:end]),
				{Field: "active", FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}}},
			})
		}
		for _, filters := range filterSets {
			resp, err := deps.ListJobTemplates(ctx, &jobtemplatepb.ListJobTemplatesRequest{
				Filters: &commonpb.FilterRequest{Filters: filters},
			})
			if err != nil {
				log.Printf("report cards section: list job templates: %v", err)
				continue
			}
			for _, t := range resp.GetData() {
				if id := t.GetId(); id != "" {
					out[id] = t.GetName()
				}
			}
		}
	}
	return out
}

// buildColumns builds the first (student) column + one column per template,
// ordered by template name ASC (prod's canonical subject order).
func buildColumns(templateIDs []string, names map[string]string, l outcome_summary.Labels) []types.TableColumn {
	ordered := append([]string(nil), templateIDs...)
	sort.SliceStable(ordered, func(i, j int) bool {
		a := strings.ToLower(colName(names, ordered[i]))
		b := strings.ToLower(colName(names, ordered[j]))
		if a == b {
			return ordered[i] < ordered[j]
		}
		return a < b
	})
	cols := make([]types.TableColumn, 0, len(ordered)+2)
	// NoSort: header sorting operates on a single tbody; the banded grid's
	// order is server-composed (Options.Row), so the control would be inert.
	// The student column is the first FROZEN column (TableClass
	// "data-table-freeze2", set in buildSectionTable). Its width must equal the
	// CSS --freeze2-c2 default (14rem) so the second frozen column's sticky left
	// offset lines up.
	cols = append(cols, types.TableColumn{Key: "student", Label: l.Section.ClientColumn, Width: "14rem", MinWidth: "14rem", NoSort: true})
	// Second frozen column: the per-row actions (view student card + CSV
	// download). Blank header mirrors the prod report card's action column.
	// Excluded from the CSV export by key (export.go skips "rc-actions").
	cols = append(cols, types.TableColumn{Key: actionsColumnKey, Label: "", Width: "5rem", MinWidth: "5rem", Align: "center", NoSort: true})
	for _, tid := range ordered {
		cols = append(cols, types.TableColumn{Key: "tmpl-" + tid, Label: colName(names, tid), MinWidth: "6.25rem", Align: "center", NoSort: true})
	}
	return cols
}

// orderedColumnIDs returns the template ids in the same order as buildColumns
// laid out the subject columns (skipping the leading student column).
func orderedColumnIDs(cols []types.TableColumn) []string {
	var ids []string
	for _, c := range cols {
		if tid, ok := strings.CutPrefix(c.Key, "tmpl-"); ok {
			ids = append(ids, tid)
		}
	}
	return ids
}

// buildRows builds one row per student: student-name cell + a rating cell per
// subject column (linking to the per-job summary; "—" when no summary) + a
// per-row report-card PDF download action (the per-client document endpoint,
// ?format=pdf). CSVValue carries the raw rating so the client-side table
// export and the section CSV endpoint emit clean text.
func buildRows(
	students map[string]student,
	templateIDs []string,
	cellJob, labelByJob map[string]string,
	evByJob map[string]outcome_summary.EnrollmentEvidence,
	sectionID string,
	routes outcome_summary.Routes,
	l outcome_summary.Labels,
) []types.TableRow {
	empty := l.Section.RatingEmpty
	if empty == "" {
		empty = "—"
	}
	rows := make([]types.TableRow, 0, len(students))
	for clientID, st := range students {
		cells := make([]types.TableCell, 0, len(templateIDs)+2)
		cells = append(cells, types.TableCell{Value: st.listName()})
		// Frozen 2nd column: per-row actions — VIEW this student's report card
		// (boosted nav → view-3), then DOWNLOAD the rendered card as a PDF (the
		// per-client document endpoint, ?format=pdf — same idiom as the client
		// card's header button). The view link is a normal boosted anchor
		// (hx-push-url); the download is hx-boost="false" + download so the
		// boosted body doesn't AJAX-swap the attachment response.
		studentURL := route.ResolveURL(routes.ClientCardURL, "id", sectionID, "client_id", clientID)
		docURL := route.ResolveURL(routes.ClientDocumentURL, "id", sectionID, "client_id", clientID) + "?format=pdf"
		viewAnchor := `<a href="` + html.EscapeString(studentURL) + `" class="action-btn view" title="` + html.EscapeString(l.Student.ViewAction) + `" aria-label="` + html.EscapeString(l.Student.ViewAction) + `" data-testid="rc-view-` + short(clientID) + `" hx-push-url="true">` + viewIcon + `</a>`
		dlAnchor := `<a href="` + html.EscapeString(docURL) + `" class="action-btn download" title="` + html.EscapeString(l.Student.DownloadAction) + `" aria-label="` + html.EscapeString(l.Student.DownloadAction) + `" data-testid="rc-download-` + short(clientID) + `" hx-boost="false" download>` + downloadIcon + `</a>`
		cells = append(cells, types.TableCell{Type: "html", HTML: texttemplate.HTML(`<div class="action-buttons">` + viewAnchor + dlAnchor + `</div>`)})
		for _, tid := range templateIDs {
			testid := html.EscapeString("rc-cell-" + short(clientID) + "-" + short(tid))
			jobID := cellJob[clientID+"\x00"+tid]
			label := labelByJob[jobID]
			switch {
			case jobID != "" && label != "" && outcome_summary.IsNonEnrolledCell(evByJob[jobID], label):
				// Non-enrolled placeholder (untaken-elective all-zero scaffold
				// whose year-final floored to "1"): render the cell truly BLANK
				// ("" — matching prod/MMIS), NOT the floor and NOT the "—" no-data
				// marker. Empty Value/CSVValue → the CSV column is blank too.
				span := `<span class="rc-cell-empty" data-testid="` + testid + `"></span>`
				cells = append(cells, types.TableCell{Type: "html", HTML: texttemplate.HTML(span)})
			case jobID != "" && label != "":
				url := route.ResolveURL(routes.JobSummaryURL, "id", jobID)
				anchor := `<a href="` + html.EscapeString(url) + `" class="table-link" data-testid="` + testid + `" hx-push-url="true">` + html.EscapeString(label) + `</a>`
				cells = append(cells, types.TableCell{Type: "html", HTML: texttemplate.HTML(anchor), CSVValue: label})
			default:
				span := `<span class="rc-cell-empty" data-testid="` + testid + `">` + html.EscapeString(empty) + `</span>`
				cells = append(cells, types.TableCell{Type: "html", HTML: texttemplate.HTML(span), CSVValue: empty})
			}
		}
		rows = append(rows, types.TableRow{
			ID:        clientID,
			DataAttrs: map[string]string{"testid": "rc-row-" + short(clientID)},
			Cells:     cells,
		})
	}
	return rows
}

// applyRowPresentation applies the configured sort + gender bands. When a
// group-by attribute is configured the rows are partitioned into
// TableRowGroup bands — ordered by Row.GroupValueOrder when set (listed
// values lead, in list order), then value-ascending, no-value band last —
// otherwise the flat rows are sorted in place.
func applyRowPresentation(
	table *types.TableConfig,
	rows []types.TableRow,
	students map[string]student,
	attrValues map[string]map[string]string,
	opts outcome_summary.Options,
) {
	sortRows(rows, students, opts.Row)

	groupCode, ok := outcome_summary.ClientAttributeCode(opts.Row.GroupByField)
	var vals map[string]string
	if ok {
		vals = attrValues[groupCode]
	}
	if !ok || len(vals) == 0 {
		table.Rows = rows
		return
	}

	// Partition into value bands (rows keep their sorted order within a band).
	order := []string{}
	seen := map[string]bool{}
	buckets := map[string][]types.TableRow{}
	for _, r := range rows {
		v := vals[r.ID]
		if !seen[v] {
			seen[v] = true
			order = append(order, v)
		}
		buckets[v] = append(buckets[v], r)
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if (a == "") != (b == "") {
			return b == "" // no-value band last
		}
		ra, aListed := opts.Row.GroupValueRank(a)
		rb, bListed := opts.Row.GroupValueRank(b)
		if aListed != bListed {
			return aListed // configured values lead
		}
		if aListed && bListed && ra != rb {
			return ra < rb
		}
		return strings.ToLower(a) < strings.ToLower(b)
	})
	groups := make([]types.TableRowGroup, 0, len(order))
	for _, v := range order {
		title := v
		if title == "" {
			title = "—"
		}
		groups = append(groups, types.TableRowGroup{
			ID:        "rc-band-" + slug(v),
			Title:     title,
			Rows:      buckets[v],
			DataAttrs: map[string]string{"testid": "rc-band-" + slug(v)},
		})
	}
	table.Groups = groups
}

// sortRows orders rows by the configured client-column SortField (only
// "last_name" is implemented today), direction-aware, stable; ties fall back to
// display name then id.
func sortRows(rows []types.TableRow, students map[string]student, opts outcome_summary.RowOptions) {
	desc := opts.Direction() == "desc"
	field := strings.TrimSpace(opts.SortField)
	sort.SliceStable(rows, func(i, j int) bool {
		si, sj := students[rows[i].ID], students[rows[j].ID]
		var a, b string
		switch field {
		case "last_name":
			a, b = strings.ToLower(si.lastName), strings.ToLower(sj.lastName)
		default:
			a, b = strings.ToLower(si.name), strings.ToLower(sj.name)
		}
		if a == b {
			an, bn := strings.ToLower(si.name), strings.ToLower(sj.name)
			if an == bn {
				return rows[i].ID < rows[j].ID
			}
			return an < bn
		}
		// values present sort before empties regardless of direction.
		if (a == "") != (b == "") {
			return a != ""
		}
		if desc {
			return a > b
		}
		return a < b
	})
}

// allRows flattens the table's rows/groups for ApplyColumnStyles.
func allRows(table *types.TableConfig) []types.TableRow {
	if table == nil {
		return nil
	}
	if len(table.Groups) == 0 {
		return table.Rows
	}
	var out []types.TableRow
	for _, g := range table.Groups {
		out = append(out, g.Rows...)
	}
	return out
}

// --- small helpers -------------------------------------------------------

func stringEq(field, value string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_StringFilter{
			StringFilter: &commonpb.StringFilter{Value: value, Operator: commonpb.StringOperator_STRING_EQUALS},
		},
	}
}

func listIn(field string, values []string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_ListFilter{
			ListFilter: &commonpb.ListFilter{Values: values, Operator: commonpb.ListOperator_LIST_IN},
		},
	}
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func colName(names map[string]string, id string) string {
	if n := strings.TrimSpace(names[id]); n != "" {
		return n
	}
	return id
}

// clientDisplayName mirrors the fayna client-name pattern (client's own name
// column first, then embedded User first+last, then id).
func clientDisplayName(c *clientpb.Client) string {
	if name := strings.TrimSpace(c.GetName()); name != "" {
		return name
	}
	if fn := strings.TrimSpace(c.GetFirstName() + " " + c.GetLastName()); fn != "" {
		return fn
	}
	if u := c.GetUser(); u != nil {
		if name := strings.TrimSpace(u.GetFirstName() + " " + u.GetLastName()); name != "" {
			return name
		}
	}
	return c.GetId()
}

// clientLastName prefers the client's own last_name column, then the embedded
// User's last name.
func clientLastName(c *clientpb.Client) string {
	if ln := strings.TrimSpace(c.GetLastName()); ln != "" {
		return ln
	}
	if u := c.GetUser(); u != nil {
		if ln := strings.TrimSpace(u.GetLastName()); ln != "" {
			return ln
		}
	}
	return ""
}

// clientFirstName mirrors clientLastName for the first-name column.
func clientFirstName(c *clientpb.Client) string {
	if fn := strings.TrimSpace(c.GetFirstName()); fn != "" {
		return fn
	}
	if u := c.GetUser(); u != nil {
		if fn := strings.TrimSpace(u.GetFirstName()); fn != "" {
			return fn
		}
	}
	return ""
}

// short returns a stable, collision-resistant slug (the uuidv7 random TAIL,
// not the shared timestamp prefix — see the list view's short() note). Keeps
// rc-row / rc-cell testids unique across a section's students and subjects.
func short(id string) string {
	if len(id) > 8 {
		return id[len(id)-8:]
	}
	return id
}

func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "none"
	}
	return b.String()
}
