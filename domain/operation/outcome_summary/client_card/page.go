// Package client_card renders view-3 of the report-cards surface: one
// student's per-subject transcript within a section. Rows = the student's
// subjects (jobs, subject-name order); columns are GROUPED by grading phase —
// Semester 1 [Progress | Final], Semester 2 [Progress | Final], plus the Year
// Final. Each Final cell is the semester's grade (phase_outcome_summary
// scaled_label, the IB 1-7 band label); Year Final is the job_outcome_summary
// grade. Progress columns render blank ("—") — education1 stores only the two
// semester composites, no progress-period rating.
//
// Security (mirrors the section grid): the section id is EXISTS-gated against
// the session workspace (the workspace-aware ListSubscriptionGroups returns a
// foreign group as no-rows → fail-closed), AND the client must be an active (or
// frozen-historical) MEMBER of that section — a foreign client id, or a client
// from another section, resolves to no subscription → fail-closed. This closes
// the IDOR axis. Every read is workspace-bound at the espyna adapter; the job
// set is staff-narrowed.
package client_card

import (
	"context"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

// pageLimit / maxPages: the job set here is ONE student's subjects (≤ ~15), far
// below the adapter default cap — but page the job read explicitly anyway (the
// fetchSectionJobs lesson: an uncapped read silently truncates), bounded so the
// loop always halts.
const pageLimit = 100
const maxPages = 100

// Deps holds the student-card view dependencies. All list closures are
// workspace-bound at the espyna adapter; the view composes no client_id/
// workspace filter of its own beyond the section + membership gates.
type Deps struct {
	Routes       outcome_summary.Routes
	Labels       outcome_summary.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// CategoryFilter (a job_category code, e.g. "academic") + ListJobCategories
	// gate the subject set to that category — same-origin deportment jobs are
	// dropped (gate H2). Empty code or nil closure → no filter. Resolved once per
	// request via outcome_summary.ResolveCategoryID.
	CategoryFilter    string
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)

	// BandByCategory + IncludeAllCategories (R9 W-A6) — the DEDICATED client-card
	// banding policy (Options.ClientCard). BandByCategory groups the subject rows
	// into native job-category TableRowGroup bands; IncludeAllCategories LIFTS the
	// card's H2 academic-only filter FOR BANDING so deportment subjects render
	// under their own band (the report-card document + section grid keep H2 —
	// separate fetches). Both zero (service-admin / unset) → today's flat card,
	// byte-identical (module contract). The lift and the band-title read together
	// issue EXACTLY ONE job_category read (statement count unchanged).
	BandByCategory       bool
	IncludeAllCategories bool

	ListSubscriptionGroups        func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListSubscriptionGroupMembers  func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListJobs                      func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListJobTemplates              func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ListClients                   func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
	ListJobOutcomeSummarys        func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)
	ListPhaseOutcomeSummarysByJob func(ctx context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error)
	ListJobPhases                 func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)

	// Non-enrolled-placeholder evidence walk (job_phase → job_task →
	// task_outcome). ListJobPhases (above) is reused. Optional/nil-safe: when
	// ListJobTasks or ListTaskOutcomes is nil the evidence map is empty and no
	// grade cell is blanked (the flat surface stays byte-identical).
	ListJobTasks     func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	ListTaskOutcomes func(ctx context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error)
}

// PageData is the student-card page data.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	NotComputed     bool
	Banner          string
}

// NewView creates the per-student report-card view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "list") {
			return view.Forbidden("job_outcome_summary:list")
		}

		sectionID := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		clientID := strings.TrimSpace(viewCtx.Request.PathValue("client_id"))
		if sectionID == "" || clientID == "" {
			return view.Forbidden("job_outcome_summary:list")
		}

		group := fetchSection(ctx, deps, sectionID)
		if group == nil {
			return view.Forbidden("job_outcome_summary:list")
		}
		historical := !group.GetActive()

		// IDOR gate: the client must belong to THIS section (its subscription in
		// the group's member set). No membership → fail-closed.
		subID := memberSubscription(ctx, deps, sectionID, clientID, historical)
		if subID == "" {
			return view.Forbidden("job_outcome_summary:list")
		}

		name := studentName(ctx, deps, clientID)
		table := buildTable(ctx, deps, subID, historical)
		return okPage(viewCtx, deps, group, clientID, name, table)
	})
}

// fetchSection returns the section (workspace-scoped EXISTS gate) or nil. The
// generic List defaults to active=true, so a second explicit active=false read
// covers a frozen (inactive) historical section.
func fetchSection(ctx context.Context, deps *Deps, sectionID string) *subscriptiongrouppb.SubscriptionGroup {
	if deps.ListSubscriptionGroups == nil {
		return nil
	}
	requests := []*subscriptiongrouppb.ListSubscriptionGroupsRequest{
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("id", sectionID)}}},
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
			stringEq("id", sectionID),
			boolEq("active", false),
		}}},
	}
	for _, req := range requests {
		resp, err := deps.ListSubscriptionGroups(ctx, req)
		if err != nil {
			log.Printf("student card: list subscription group by id: %v", err)
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

// memberSubscription returns the client's subscription_id within this section
// (active members, or the frozen full roster in historical mode), or "" when the
// client is not a member — the IDOR fail-closed signal.
func memberSubscription(ctx context.Context, deps *Deps, sectionID, clientID string, historical bool) string {
	if deps.ListSubscriptionGroupMembers == nil {
		return ""
	}
	requests := []*subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("subscription_group_id", sectionID)}}},
	}
	if historical {
		requests = append(requests, &subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
				stringEq("subscription_group_id", sectionID),
				boolEq("active", false),
			}},
		})
	}
	for _, req := range requests {
		resp, err := deps.ListSubscriptionGroupMembers(ctx, req)
		if err != nil {
			log.Printf("student card: list members: %v", err)
			continue
		}
		for _, m := range resp.GetData() {
			if !historical && !m.GetActive() {
				continue
			}
			if m.GetClientId() == clientID && m.GetSubscriptionId() != "" {
				return m.GetSubscriptionId()
			}
		}
	}
	return ""
}

// buildTable assembles the student's subject × phase grid. Returns nil when the
// student has no jobs or no computed grades yet (→ empty-state banner).
func buildTable(ctx context.Context, deps *Deps, subID string, historical bool) *types.TableConfig {
	l := deps.Labels

	jobs := fetchJobs(ctx, deps, subID, historical)
	if len(jobs) == 0 {
		return nil
	}

	// Card-specific category inclusion policy (R9 W-A6). Two mutually exclusive
	// modes, each issuing EXACTLY ONE job_category read (statement count
	// unchanged):
	//   - liftH2 == false (default / service-admin): the H2 academic-only filter
	//     as today — ResolveCategoryID (a code-filtered read) drops same-origin
	//     deportment jobs. catID "" with catOK=true (no filter) keeps every job;
	//     catOK=false (a configured filter that could not resolve) fails CLOSED —
	//     the card drops every job.
	//   - liftH2 == true (banding + IncludeAllCategories): H2 is LIFTED for the
	//     card's OWN table so deportment subjects render under their own band; the
	//     ONE read instead fetches ALL active categories for band titles/order
	//     (fetchBandCategories). The document download + section grid keep H2 —
	//     they run separate job fetches with their own CategoryFilter.
	liftH2 := deps.BandByCategory && deps.IncludeAllCategories
	var catID string
	catOK := true
	var bandCats []*jobcategorypb.JobCategory
	if liftH2 {
		bandCats = fetchBandCategories(ctx, deps)
	} else {
		catID, catOK = outcome_summary.ResolveCategoryID(ctx, deps.ListJobCategories, deps.CategoryFilter)
	}

	jobTemplate := map[string]string{} // job_id -> template_id
	jobIDs := make([]string, 0, len(jobs))
	templateIDs := []string{}
	tmplSeen := map[string]bool{}
	for _, j := range jobs {
		if !liftH2 && (!catOK || !outcome_summary.KeepJobInCategory(catID, j.GetJobCategoryId())) {
			continue
		}
		jid, tid := j.GetId(), j.GetJobTemplateId()
		if jid == "" || tid == "" {
			continue
		}
		jobTemplate[jid] = tid
		jobIDs = append(jobIDs, jid)
		if !tmplSeen[tid] {
			tmplSeen[tid] = true
			templateIDs = append(templateIDs, tid)
		}
	}

	tmplNames, tmplCategory := fetchTemplateNames(ctx, deps, templateIDs, historical)
	phaseOrder := fetchPhaseOrders(ctx, deps, jobIDs)              // job_phase_id -> phase_order
	semByJob := fetchSemesterLabels(ctx, deps, jobIDs, phaseOrder) // job_id -> {order -> label}
	yearByJob := fetchYearLabels(ctx, deps, jobIDs)                // job_id -> year-final label

	if len(semByJob) == 0 && len(yearByJob) == 0 {
		return nil
	}

	// Non-enrolled-placeholder evidence (job_phase → job_task → task_outcome):
	// an untaken-elective scaffold rides in with an all-zero task set and bands
	// floored to "1"; its grade cells must render BLANK (matching prod), not the
	// floor. A genuinely enrolled subject — even one scored a real 0/1 — carries
	// a positive task mark and is kept. Nil-safe: unwired closures → empty map →
	// nothing blanked. Fail-closed: on a read error keep every grade (blank
	// nothing) rather than risk blanking a real one from incomplete evidence.
	evByJob, err := outcome_summary.FetchJobMarkEvidence(ctx, deps.ListJobPhases, deps.ListJobTasks, deps.ListTaskOutcomes, jobIDs)
	if err != nil {
		log.Printf("outcome summary client card: enrollment evidence unavailable, keeping all grades: %v", err)
		evByJob = nil
	}

	empty := l.Section.RatingEmpty
	if empty == "" {
		empty = "—"
	}

	// One row per subject (job), subject-name ASC (prod's canonical order).
	type entry struct{ jobID, name string }
	entries := make([]entry, 0, len(jobIDs))
	for _, jid := range jobIDs {
		entries = append(entries, entry{jid, colName(tmplNames, jobTemplate[jid])})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := strings.ToLower(entries[i].name), strings.ToLower(entries[j].name)
		if a == b {
			return entries[i].jobID < entries[j].jobID
		}
		return a < b
	})

	rows := make([]types.TableRow, 0, len(entries))
	for _, e := range entries {
		sem := semByJob[e.jobID]
		s1f := ratingCell(sem[1], empty)            // Sem 1 Final
		s2f := ratingCell(sem[2], empty)            // Sem 2 Final
		yf := ratingCell(yearByJob[e.jobID], empty) // Year Final
		// Blank the grade cells of a non-enrolled placeholder subject (an
		// untaken-elective all-zero scaffold whose bands floored to "1"): render
		// "" (matching prod/MMIS), NOT the floor and NOT the "—" no-data marker.
		// The subject-name cell stays so the row set is stable. A genuinely
		// enrolled subject — even one scored a real 0/1 — has a positive task
		// mark and keeps its grades.
		if outcome_summary.IsNonEnrolledCell(evByJob[e.jobID], yearByJob[e.jobID], sem[1], sem[2]) {
			s1f, s2f, yf = blankCell(), blankCell(), blankCell()
		}
		cells := []types.TableCell{
			{Value: e.name},
			ratingCell("", empty), // Sem 1 Progress — no data
			s1f,
			ratingCell("", empty), // Sem 2 Progress — no data
			s2f,
			yf,
		}
		rows = append(rows, types.TableRow{
			ID:        e.jobID,
			DataAttrs: map[string]string{"testid": "rc-subject-" + short(e.jobID)},
			Cells:     cells,
		})
	}

	cfg := &types.TableConfig{
		ID:              "report-cards-student",
		ColumnGroups:    buildColumnGroups(l),
		NameColumnLabel: l.Student.SubjectColumn,
		ShowSearch:      true,
		ShowColumns:     true,
		ShowDensity:     true,
		ShowExport:      true,
		ShowEntries:     true,
		Labels:          deps.TableLabels,
		Caption:         l.Student.Title,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Section.NotComputedBanner,
		},
	}
	// Job-category bands (R9 W-A6): when configured AND ≥2 distinct effective
	// categories are present, group the (already subject-name-sorted) rows into
	// native TableRowGroup bands. The flat `rows` slice is built identically
	// above, so the degrade path (<2 categories) and the unbanded path stay
	// byte-equal to today. No BulkActions is set, so the band-header bulk-select
	// controls render inert (CSS-hidden — table.css:3427/3456).
	if groups := bandSubjectRows(l, rows, jobTemplate, tmplCategory, bandCats); groups != nil {
		cfg.Groups = groups
	} else {
		cfg.Rows = rows
	}
	return cfg
}

// buildColumnGroups builds the nested semester headers: Semester 1 [Progress |
// Final] · Semester 2 [Progress | Final] · Year [Final]. The Subject column is
// the auto-generated first column (NameColumnLabel); NoSort because the grouped
// grid is server-composed.
func buildColumnGroups(l outcome_summary.Labels) []types.ColumnGroup {
	prog := l.Student.ProgressColumn
	fin := l.Student.FinalColumn
	return []types.ColumnGroup{
		{Label: l.Student.Period1, Columns: []types.TableColumn{
			{Key: "s1p", Label: prog, Align: "center", NoSort: true, MinWidth: "5rem"},
			{Key: "s1f", Label: fin, Align: "center", NoSort: true, MinWidth: "5rem"},
		}},
		{Label: l.Student.Period2, Columns: []types.TableColumn{
			{Key: "s2p", Label: prog, Align: "center", NoSort: true, MinWidth: "5rem"},
			{Key: "s2f", Label: fin, Align: "center", NoSort: true, MinWidth: "5rem"},
		}},
		{Label: l.Student.YearColumn, Columns: []types.TableColumn{
			{Key: "yf", Label: fin, Align: "center", NoSort: true, MinWidth: "5rem"},
		}},
	}
}

// fetchJobs returns the student's subscription-origin subject jobs
// (staff-narrowed at the adapter), paged explicitly. Historical mode also
// accepts inactive jobs.
func fetchJobs(ctx context.Context, deps *Deps, subID string, historical bool) []*jobpb.Job {
	var out []*jobpb.Job
	if deps.ListJobs == nil || subID == "" {
		return out
	}
	seen := map[string]bool{}
	filterSets := [][]*commonpb.TypedFilter{
		{stringEq("origin_id", subID)},
	}
	if historical {
		filterSets = append(filterSets, []*commonpb.TypedFilter{
			stringEq("origin_id", subID),
			boolEq("active", false),
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
				log.Printf("student card: list jobs (page %d): %v", page, err)
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
	return out
}

// fetchTemplateNames resolves job_template_id -> name AND job_template_id ->
// job_category_id (proto field 32, the authoritative CURRENT category FK per
// §3.0), chunked; historical mode also reads inactive templates (frozen-year
// subjects). The category map backs the job-category bands (R9 W-A6) — the read
// already returns the FK, so capturing it is free (no extra statement). Callers
// that do not band simply ignore the second return.
func fetchTemplateNames(ctx context.Context, deps *Deps, templateIDs []string, historical bool) (names, categories map[string]string) {
	names = map[string]string{}
	categories = map[string]string{}
	if deps.ListJobTemplates == nil || len(templateIDs) == 0 {
		return names, categories
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
				boolEq("active", false),
			})
		}
		for _, filters := range filterSets {
			resp, err := deps.ListJobTemplates(ctx, &jobtemplatepb.ListJobTemplatesRequest{
				Filters: &commonpb.FilterRequest{Filters: filters},
			})
			if err != nil {
				log.Printf("student card: list job templates: %v", err)
				continue
			}
			for _, t := range resp.GetData() {
				if id := t.GetId(); id != "" {
					names[id] = t.GetName()
					categories[id] = t.GetJobCategoryId()
				}
			}
		}
	}
	return names, categories
}

// fetchBandCategories issues the ONE all-category read (replacing the
// code-filtered ResolveCategoryID read when H2 is lifted, so the client-card
// statement count is unchanged) and returns the ACTIVE job_category rows sorted
// by the category sort contract (sort_order NULLS LAST → name → id) for band
// titles + order. A nil closure or read error → nil (banding degrades to flat).
func fetchBandCategories(ctx context.Context, deps *Deps) []*jobcategorypb.JobCategory {
	if deps.ListJobCategories == nil {
		return nil
	}
	resp, err := deps.ListJobCategories(ctx, &jobcategorypb.ListJobCategoriesRequest{})
	if err != nil {
		log.Printf("student card: list job categories for bands: %v", err)
		return nil
	}
	var cats []*jobcategorypb.JobCategory
	for _, c := range resp.GetData() {
		if c.GetActive() && c.GetId() != "" {
			cats = append(cats, c)
		}
	}
	sortBandCategories(cats)
	return cats
}

// sortBandCategories orders the card's bands by job_category.sort_order ASC
// (NULLs LAST), then name ASC, then id — the category primitive's own sort
// contract (mirrors the landing's sortLandingCategories, plan §3.3). TabOptions
// governs only the price_schedule tabstrip and never this axis.
func sortBandCategories(cats []*jobcategorypb.JobCategory) {
	sort.SliceStable(cats, func(i, j int) bool {
		a, b := cats[i], cats[j]
		ao, aok := bandOrderOf(a)
		bo, bok := bandOrderOf(b)
		if aok != bok {
			return aok // NULLS LAST
		}
		if aok && bok && ao != bo {
			return ao < bo
		}
		an, bn := strings.ToLower(a.GetName()), strings.ToLower(b.GetName())
		if an == bn {
			return a.GetId() < b.GetId()
		}
		return an < bn
	})
}

// bandOrderOf returns the category's sort_order and whether it is set (NULL =
// not set → sorts last).
func bandOrderOf(c *jobcategorypb.JobCategory) (int32, bool) {
	if c != nil && c.SortOrder != nil {
		return c.GetSortOrder(), true
	}
	return 0, false
}

// bandSubjectRows groups the subject rows into native job-category
// TableRowGroup bands, or returns nil to keep the flat row set. Banding applies
// only when the card is configured for it (cats non-empty ⇒ liftH2 was on) AND
// ≥2 distinct effective bands are present (the degrade contract). Effective
// category per row = job_template.job_category_id (current-FK authority, §3.0);
// a NULL or out-of-corpus (stale/inactive/foreign) category folds into a single
// trailing Uncategorized band — never dropped, never duplicated. Row order
// within a band is preserved (subject-name ASC, as built by the caller); band
// order = the pre-sorted `cats` contract, Uncategorized last.
func bandSubjectRows(l outcome_summary.Labels, rows []types.TableRow, jobTemplate, tmplCategory map[string]string, cats []*jobcategorypb.JobCategory) []types.TableRowGroup {
	if len(cats) == 0 || len(rows) == 0 {
		return nil
	}
	corpus := make(map[string]bool, len(cats))
	for _, c := range cats {
		corpus[c.GetId()] = true
	}
	buckets := map[string][]types.TableRow{} // "" = Uncategorized bucket
	for _, r := range rows {
		eff := tmplCategory[jobTemplate[r.ID]]
		if eff == "" || !corpus[eff] {
			eff = "" // NULL/foreign → Uncategorized
		}
		buckets[eff] = append(buckets[eff], r)
	}
	groups := make([]types.TableRowGroup, 0, len(cats)+1)
	for _, c := range cats {
		rs := buckets[c.GetId()]
		if len(rs) == 0 {
			continue
		}
		key := bandKey(c)
		groups = append(groups, types.TableRowGroup{
			ID:        "rc-band-" + key,
			Title:     c.GetName(),
			Rows:      rs,
			DataAttrs: map[string]string{"testid": "rc-band-" + key},
		})
	}
	if un := buckets[""]; len(un) > 0 {
		title := strings.TrimSpace(l.Student.UncategorizedBand)
		if title == "" {
			title = "Uncategorized"
		}
		groups = append(groups, types.TableRowGroup{
			ID:        "rc-band-uncategorized",
			Title:     title,
			Rows:      un,
			DataAttrs: map[string]string{"testid": "rc-band-uncategorized"},
		})
	}
	if len(groups) < 2 {
		return nil // degrade: <2 distinct bands → flat rows
	}
	return groups
}

// bandKey returns a stable, HTML-id-safe band suffix: the category CODE (a
// unique snake_case job_category.code, e.g. "academic") when present, else the
// collision-resistant id tail.
func bandKey(c *jobcategorypb.JobCategory) string {
	if code := strings.TrimSpace(c.GetCode()); code != "" {
		return code
	}
	return short(c.GetId())
}

// fetchPhaseOrders maps job_phase_id -> phase_order for the student's jobs, so
// each phase_outcome_summary can be placed in its semester column (1 -> Sem 1,
// 2 -> Sem 2). job_phase rows are per-job (their ids are unique), but the ORDER
// is the stable cross-job alignment key.
func fetchPhaseOrders(ctx context.Context, deps *Deps, jobIDs []string) map[string]int32 {
	out := map[string]int32{}
	if deps.ListJobPhases == nil || len(jobIDs) == 0 {
		return out
	}
	for start := 0; start < len(jobIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(jobIDs) {
			end = len(jobIDs)
		}
		resp, err := deps.ListJobPhases(ctx, &jobphasepb.ListJobPhasesRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_id", jobIDs[start:end])}},
		})
		if err != nil {
			log.Printf("student card: list job phases: %v", err)
			continue
		}
		for _, p := range resp.GetData() {
			if id := p.GetId(); id != "" {
				out[id] = p.GetPhaseOrder()
			}
		}
	}
	return out
}

// fetchSemesterLabels returns job_id -> (phase_order -> semester grade). Reads
// each job's phase summaries (ListByJob now projects scaled_label). Falls back
// to a formatted scaled_score when the label is empty.
func fetchSemesterLabels(ctx context.Context, deps *Deps, jobIDs []string, phaseOrder map[string]int32) map[string]map[int32]string {
	out := map[string]map[int32]string{}
	if deps.ListPhaseOutcomeSummarysByJob == nil {
		return out
	}
	for _, jid := range jobIDs {
		resp, err := deps.ListPhaseOutcomeSummarysByJob(ctx, &phasesumpb.ListPhaseOutcomeSummarysByJobRequest{JobId: jid})
		if err != nil {
			log.Printf("student card: list phase summaries by job: %v", err)
			continue
		}
		for _, s := range resp.GetPhaseOutcomeSummarys() {
			if !s.GetActive() {
				continue
			}
			ord := phaseOrder[s.GetJobPhaseId()]
			if ord == 0 {
				continue
			}
			label := strings.TrimSpace(s.GetScaledLabel())
			if label == "" && s.ScaledScore != nil {
				label = strconv.FormatFloat(s.GetScaledScore(), 'f', -1, 64)
			}
			if label == "" {
				continue
			}
			if out[jid] == nil {
				out[jid] = map[int32]string{}
			}
			out[jid][ord] = label
		}
	}
	return out
}

// fetchYearLabels returns job_id -> year-final grade (job_outcome_summary
// scaled_label, falling back to scaled_score), chunked by job_id.
func fetchYearLabels(ctx context.Context, deps *Deps, jobIDs []string) map[string]string {
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
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_id", jobIDs[start:end])}},
		})
		if err != nil {
			log.Printf("student card: list job outcome summaries: %v", err)
			continue
		}
		for _, s := range resp.GetData() {
			if !s.GetActive() || s.GetJobId() == "" {
				continue
			}
			if lbl := strings.TrimSpace(s.GetScaledLabel()); lbl != "" {
				out[s.GetJobId()] = lbl
			} else if s.ScaledScore != nil {
				out[s.GetJobId()] = strconv.FormatFloat(s.GetScaledScore(), 'f', -1, 64)
			}
		}
	}
	return out
}

// studentName resolves the client's display name (class-list form
// "Last, First" when both parts resolve).
func studentName(ctx context.Context, deps *Deps, clientID string) string {
	if deps.ListClients == nil || clientID == "" {
		return clientID
	}
	resp, err := deps.ListClients(ctx, &clientpb.ListClientsRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("id", clientID)}},
	})
	if err != nil {
		log.Printf("student card: list client: %v", err)
		return clientID
	}
	for _, c := range resp.GetData() {
		if c.GetId() != clientID {
			continue
		}
		last := strings.TrimSpace(c.GetLastName())
		first := strings.TrimSpace(c.GetFirstName())
		if last != "" && first != "" {
			return last + ", " + first
		}
		if n := strings.TrimSpace(c.GetName()); n != "" {
			return n
		}
		if u := c.GetUser(); u != nil {
			if n := strings.TrimSpace(u.GetLastName() + ", " + u.GetFirstName()); n != ", " {
				return n
			}
		}
	}
	return clientID
}

// okPage assembles the PageData. Header: breadcrumb = the section name (links
// back to the section grid), title = the student name, caption = the student
// subtitle label. Mirrors the section view's okPage header shape.
func okPage(viewCtx *view.ViewContext, deps *Deps, group *subscriptiongrouppb.SubscriptionGroup, clientID, name string, table *types.TableConfig) view.ViewResult {
	l := deps.Labels
	// PDF download affordance — only when there are computed grades (a blank card
	// would 404 at the endpoint). ClientDocumentURL resolved for this section +
	// client, "?format=pdf" (string-concat query, the SectionExportURL idiom). The
	// template renders it as a plain download anchor (Content-Disposition handles
	// the save — no fetch/blob JS).
	var downloadURL string
	if table != nil && deps.Routes.ClientDocumentURL != "" {
		downloadURL = route.ResolveURL(deps.Routes.ClientDocumentURL, "id", group.GetId(), "client_id", clientID) + "?format=pdf"
	}
	downloadLabel := l.Student.DownloadAction
	if strings.TrimSpace(downloadLabel) == "" {
		downloadLabel = "Download PDF"
	}
	// Download-PDF lives in the table's toolbar primary-action slot (Q-R9-9),
	// reclaiming the standalone band's row + margin. Download=true makes the
	// toolbar anchor emit download + hx-boost="false" so the body-boosted app
	// does not intercept the file download as an HTMX navigation (mirrors the old
	// band's attributes). Only attached when a URL exists (no URL → no primary
	// action), preserving the former {{if .DocumentDownloadURL}} gate; a blank
	// card (table == nil ⇒ downloadURL == "") shows no action.
	if downloadURL != "" {
		table.PrimaryAction = &types.PrimaryAction{
			Label:    downloadLabel,
			Href:     downloadURL,
			Download: true,
			TestID:   "rc-download-pdf",
		}
	}
	pd := &PageData{
		PageData: types.PageData{
			CacheVersion:        viewCtx.CacheVersion,
			Title:               name,
			CurrentPath:         viewCtx.CurrentPath,
			ActiveNav:           deps.Routes.ActiveNav,
			ActiveSubNav:        "report-cards",
			HeaderBreadcrumb:    group.GetName(),
			HeaderBreadcrumbURL: route.ResolveURL(deps.Routes.SectionURL, "id", group.GetId()),
			HeaderTitle:         name,
			HeaderSubtitle:      l.Student.Subtitle,
			HeaderIcon:          "icon-award",
			CommonLabels:        deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-student-content",
		Table:           table,
		NotComputed:     table == nil,
		Banner:          l.Section.NotComputedBanner,
	}
	return view.OK("outcome-summary-student", pd)
}

// --- small helpers -------------------------------------------------------

func ratingCell(label, empty string) types.TableCell {
	if strings.TrimSpace(label) == "" {
		return types.TableCell{Value: empty, Align: "center"}
	}
	return types.TableCell{Value: label, Align: "center"}
}

// blankCell is a truly-empty grade cell ("" — not the "—" no-data marker) for a
// non-enrolled placeholder subject. Distinct from ratingCell("", empty), which
// renders "—".
func blankCell() types.TableCell {
	return types.TableCell{Value: "", Align: "center"}
}

func stringEq(field, value string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_StringFilter{
			StringFilter: &commonpb.StringFilter{Value: value, Operator: commonpb.StringOperator_STRING_EQUALS},
		},
	}
}

func boolEq(field string, v bool) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field:      field,
		FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: v}},
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

func colName(names map[string]string, id string) string {
	if n := strings.TrimSpace(names[id]); n != "" {
		return n
	}
	return id
}

// short returns the uuidv7 random tail for a stable, collision-resistant testid
// suffix (see the section view's short() note).
func short(id string) string {
	if len(id) > 8 {
		return id[len(id)-8:]
	}
	return id
}
