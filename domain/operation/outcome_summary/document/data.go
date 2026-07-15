package document

import (
	"context"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

const (
	pageLimit = 100
	maxPages  = 100
	dash      = "—"
)

// subjectRow is one row of the report card (one subject / job).
type subjectRow struct {
	Name      string
	CritA     string // per-criterion (job_outcome_line sub-strands); "—" when absent
	CritB     string
	CritC     string
	CritD     string
	Total     string // Σ of the per-criterion MAX (/32); "—" when no criteria
	Sem1Band  string // phase_outcome_summary (phase_order 1) — recompute-equivalent
	Sem2Band  string // phase_outcome_summary (phase_order 2) — recompute-equivalent
	YearFinal string // job_outcome_summary.scaled_label — STORED, never recomputed
}

// reportCard is one client's assembled card (meta + ordered subjects). Generic
// field identifiers; the vertical rendered wording lives in lyngua values and in
// the .docx template's placeholder keys (school_name/academic_year/… — the
// operator template contract, unchanged).
type reportCard struct {
	DocumentHeaderName string
	SchedulePeriod     string
	ClientName         string
	GroupLevel         string
	SectionName        string
	LRN                string
	PrintedBy          string
	PrintedAt          string
	Subjects           []subjectRow
	// PriceScheduleID is the section's AY anchor (subscription_group.price_schedule_id),
	// threaded to the report-card template-binding resolver. Empty → the resolver
	// returns the workspace-wide fallback binding (or none → embedded template).
	PriceScheduleID string
}

// buildReportCardData mirrors buildInvoiceData: it flattens the assembled card
// into the doctemplate data map, emitting EVERY key referenced by the template
// as a pre-formatted string (blank/dash, never omitted) so no raw {{..}} leaks
// (engine leaks unresolved placeholders verbatim — G3/G5).
func buildReportCardData(rc reportCard) map[string]any {
	subjects := make([]any, 0, len(rc.Subjects))
	for _, s := range rc.Subjects {
		subjects = append(subjects, map[string]any{
			"subject_name":   orBlank(s.Name),
			"crit_a":         orDash(s.CritA),
			"crit_b":         orDash(s.CritB),
			"crit_c":         orDash(s.CritC),
			"crit_d":         orDash(s.CritD),
			"criteria_total": orDash(s.Total),
			"sem1_band":      orDash(s.Sem1Band),
			"sem2_band":      orDash(s.Sem2Band),
			"myp_overall":    orDash(s.YearFinal),
		})
	}
	return map[string]any{
		"school_name":   orBlank(rc.DocumentHeaderName),
		"academic_year": orBlank(rc.SchedulePeriod),
		"student_name":  orBlank(rc.ClientName),
		"grade_level":   orBlank(rc.GroupLevel),
		"section_name":  orBlank(rc.SectionName),
		"lrn":           orDash(rc.LRN),
		"printed_by":    orBlank(rc.PrintedBy),
		"printed_at":    orBlank(rc.PrintedAt),
		"subjects":      subjects,
	}
}

// collectCard assembles one client's report card by mirroring the view-3
// client_card fetch (section EXISTS gate → membership IDOR gate → jobs →
// phase/year summaries) and ADDING the job_outcome_line per-criterion fetch.
// Returns (nil,false) on any gate failure (fail-closed; the handler maps to
// 403/404 without leaking which check failed).
func collectCard(ctx context.Context, d *Deps, sectionID, clientID string) (*reportCard, bool) {
	group := fetchSection(ctx, d, sectionID)
	if group == nil {
		return nil, false
	}
	historical := !group.GetActive()

	subID := memberSubscription(ctx, d, sectionID, clientID, historical)
	if subID == "" {
		return nil, false
	}

	jobs := fetchJobs(ctx, d, subID, historical)
	// H2: keep only the configured category's subjects (academic), dropping
	// same-origin deportment jobs. catID "" with catOK=true (no filter) keeps
	// every job; catOK=false (a configured filter that could not be resolved)
	// fails CLOSED — the document drops every job (empty transcript, not a leak).
	catID, catOK := outcome_summary.ResolveCategoryID(ctx, d.ListJobCategories, d.CategoryFilter)
	jobTemplate := map[string]string{}
	jobIDs := make([]string, 0, len(jobs))
	templateIDs := []string{}
	tmplSeen := map[string]bool{}
	for _, j := range jobs {
		if !catOK || !outcome_summary.KeepJobInCategory(catID, j.GetJobCategoryId()) {
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

	tmplNames := fetchTemplateNames(ctx, d, templateIDs, historical)
	phaseOrder, jobByPhase := fetchPhaseOrders(ctx, d, jobIDs)
	semByJob := fetchSemesterLabels(ctx, d, jobIDs, phaseOrder)
	yearByJob := fetchYearLabels(ctx, d, jobIDs)
	// Per-criterion marks from task_outcome, scoped to THIS card's jobs (fixes
	// cross-AY accumulation). Keyed by job id, MAX-within-criterion, A/B/C/D
	// ordered by template_task_criteria.sequence_order.
	critByJob := fetchCriteriaByJob(ctx, d, jobByPhase)

	// One row per subject (job), subject-name ASC — prod's canonical order.
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

	rows := make([]subjectRow, 0, len(entries))
	for _, e := range entries {
		sem := semByJob[e.jobID]
		display := cleanSubject(e.name)
		// crit is present in critByJob iff this job had ≥1 numeric task_outcome
		// (even an all-zero one); hasMarks distinguishes an all-zero scaffold
		// (present, total="0") from a subject with no task_outcome at all.
		crit, hasMarks := critByJob[e.jobID]
		row := subjectRow{
			Name:     display,
			CritA:    crit.a,
			CritB:    crit.b,
			CritC:    crit.c,
			CritD:    crit.d,
			Total:    crit.total,
			Sem1Band: sem[1],
			Sem2Band: sem[2],
			// Year-final / MYP Overall is the STORED job_outcome_summary only —
			// blank when absent (no recompute, no leaf substitution).
			YearFinal: yearByJob[e.jobID],
		}
		// Suppress NON-ENROLLED placeholder subjects (mirrors the grade-loader
		// T8 suppressed_zero_grade rule at the DOCX layer). A subject the student
		// is not taking — a parallel-language track (Korean / the other of an
		// English/Filipino pair) — rides in as an all-zero scaffold: every
		// criterion is 0 and the year-final is only the transmute-of-zero floor
		// ("0"/"1"). MMIS publishes no row for it. See
		// docs/plan/20260714-job-category-primitive/compare-report.md
		// §"Language placeholders". A genuinely-enrolled subject is never dropped:
		// it either carries a positive per-criterion mark (confirmed on
		// education1 — every all-zero-mark job is a non-enrolled track; a real 0,
		// e.g. a percentage subject whose IB band reads "0", still has a positive
		// task_outcome) or, for historical imports, has no task_outcome but a real
		// stored year-final.
		if isNonEnrolledPlaceholder(row, hasMarks) {
			continue
		}
		rows = append(rows, row)
	}

	name, ay := sectionParts(group.GetName())
	grade, section := gradeSection(name)
	rc := &reportCard{
		DocumentHeaderName: strings.TrimSpace(d.DocumentHeaderName),
		SchedulePeriod:     ay,
		ClientName:         studentName(ctx, d, clientID),
		GroupLevel:         grade,
		SectionName:        section,
		LRN:                "",
		Subjects:           rows,
		PriceScheduleID:    group.GetPriceScheduleId(),
	}
	return rc, true
}

// --- fetch helpers (mirror client_card) ----------------------------------

func fetchSection(ctx context.Context, d *Deps, sectionID string) *subscriptiongrouppb.SubscriptionGroup {
	if d.ListSubscriptionGroups == nil {
		return nil
	}
	requests := []*subscriptiongrouppb.ListSubscriptionGroupsRequest{
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("id", sectionID)}}},
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("id", sectionID), boolEq("active", false)}}},
	}
	for _, req := range requests {
		resp, err := d.ListSubscriptionGroups(ctx, req)
		if err != nil {
			log.Printf("report card doc: list subscription group: %v", err)
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

func memberSubscription(ctx context.Context, d *Deps, sectionID, clientID string, historical bool) string {
	if d.ListSubscriptionGroupMembers == nil {
		return ""
	}
	requests := []*subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("subscription_group_id", sectionID)}}},
	}
	if historical {
		requests = append(requests, &subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
				stringEq("subscription_group_id", sectionID), boolEq("active", false),
			}},
		})
	}
	for _, req := range requests {
		resp, err := d.ListSubscriptionGroupMembers(ctx, req)
		if err != nil {
			log.Printf("report card doc: list members: %v", err)
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

func fetchJobs(ctx context.Context, d *Deps, subID string, historical bool) []*jobpb.Job {
	var out []*jobpb.Job
	if d.ListJobs == nil || subID == "" {
		return out
	}
	seen := map[string]bool{}
	filterSets := [][]*commonpb.TypedFilter{{stringEq("origin_id", subID)}}
	if historical {
		filterSets = append(filterSets, []*commonpb.TypedFilter{stringEq("origin_id", subID), boolEq("active", false)})
	}
	for _, filters := range filterSets {
		for page := int32(1); page <= maxPages; page++ {
			resp, err := d.ListJobs(ctx, &jobpb.ListJobsRequest{
				Filters: &commonpb.FilterRequest{Filters: filters},
				Pagination: &commonpb.PaginationRequest{
					Limit:  int32(pageLimit),
					Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
				},
			})
			if err != nil {
				log.Printf("report card doc: list jobs (page %d): %v", page, err)
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

func fetchTemplateNames(ctx context.Context, d *Deps, templateIDs []string, historical bool) map[string]string {
	out := map[string]string{}
	if d.ListJobTemplates == nil || len(templateIDs) == 0 {
		return out
	}
	for start := 0; start < len(templateIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(templateIDs) {
			end = len(templateIDs)
		}
		filterSets := [][]*commonpb.TypedFilter{{listIn("id", templateIDs[start:end])}}
		if historical {
			filterSets = append(filterSets, []*commonpb.TypedFilter{listIn("id", templateIDs[start:end]), boolEq("active", false)})
		}
		for _, filters := range filterSets {
			resp, err := d.ListJobTemplates(ctx, &jobtemplatepb.ListJobTemplatesRequest{
				Filters: &commonpb.FilterRequest{Filters: filters},
			})
			if err != nil {
				log.Printf("report card doc: list job templates: %v", err)
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

// fetchPhaseOrders returns two maps keyed by job_phase id: phase_order (for the
// Sem 1 / Sem 2 column mapping) and the owning job id (for the per-criterion
// task_outcome roll-up, which walks phase → task → outcome back to the subject).
func fetchPhaseOrders(ctx context.Context, d *Deps, jobIDs []string) (order map[string]int32, jobByPhase map[string]string) {
	order = map[string]int32{}
	jobByPhase = map[string]string{}
	if d.ListJobPhases == nil || len(jobIDs) == 0 {
		return order, jobByPhase
	}
	for start := 0; start < len(jobIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(jobIDs) {
			end = len(jobIDs)
		}
		resp, err := d.ListJobPhases(ctx, &jobphasepb.ListJobPhasesRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_id", jobIDs[start:end])}},
		})
		if err != nil {
			log.Printf("report card doc: list job phases: %v", err)
			continue
		}
		for _, p := range resp.GetData() {
			if id := p.GetId(); id != "" {
				order[id] = p.GetPhaseOrder()
				if jid := p.GetJobId(); jid != "" {
					jobByPhase[id] = jid
				}
			}
		}
	}
	return order, jobByPhase
}

func fetchSemesterLabels(ctx context.Context, d *Deps, jobIDs []string, phaseOrder map[string]int32) map[string]map[int32]string {
	out := map[string]map[int32]string{}
	if d.ListPhaseOutcomeSummarysByJob == nil {
		return out
	}
	for _, jid := range jobIDs {
		resp, err := d.ListPhaseOutcomeSummarysByJob(ctx, &phasesumpb.ListPhaseOutcomeSummarysByJobRequest{JobId: jid})
		if err != nil {
			log.Printf("report card doc: list phase summaries by job: %v", err)
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

func fetchYearLabels(ctx context.Context, d *Deps, jobIDs []string) map[string]string {
	out := map[string]string{}
	if d.ListJobOutcomeSummarys == nil || len(jobIDs) == 0 {
		return out
	}
	for start := 0; start < len(jobIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(jobIDs) {
			end = len(jobIDs)
		}
		resp, err := d.ListJobOutcomeSummarys(ctx, &jobsumpb.ListJobOutcomeSummarysRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_id", jobIDs[start:end])}},
		})
		if err != nil {
			log.Printf("report card doc: list job outcome summaries: %v", err)
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

// criteria holds one subject's per-criterion breakdown: the MAX-within-criterion
// marks in A/B/C/D order and their Σ (the /32 criteria total). Blank slots stay
// "" and render as a dash.
type criteria struct {
	a, b, c, d, total string
}

// fetchCriteriaByJob loads the per-criterion marks for the card's jobs from
// task_outcome — the authoritative per-criterion leaf (job_outcome_line on
// education1 is per-subject only). It walks the grade-sheet join path
// job_phase → job_task → task_outcome, SCOPED to this card's phases/jobs (so a
// student's prior-AY / other-section grades never accumulate in), groups by
// (job, criterion) taking the numeric MAX across the year, orders the criteria
// A/B/C/D via template_task_criteria.sequence_order (falling back to a stable
// criteria-id order when that source is unavailable), and sums the MAXes as the
// /32 total. Keyed by job id. Fully nil-safe — a missing closure yields an empty
// map and the criterion columns render as dashes.
func fetchCriteriaByJob(ctx context.Context, d *Deps, jobByPhase map[string]string) map[string]criteria {
	out := map[string]criteria{}
	if d.ListJobTasks == nil || d.ListTaskOutcomes == nil || len(jobByPhase) == 0 {
		return out
	}

	phaseIDs := make([]string, 0, len(jobByPhase))
	for pid := range jobByPhase {
		phaseIDs = append(phaseIDs, pid)
	}

	// job_task: task id → owning job (via phase) + template_task id (for the
	// sequence_order lookup).
	taskJob := map[string]string{}
	taskTmplTask := map[string]string{}
	taskIDs := make([]string, 0, len(phaseIDs))
	for start := 0; start < len(phaseIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(phaseIDs) {
			end = len(phaseIDs)
		}
		for page := int32(1); page <= maxPages; page++ {
			resp, err := d.ListJobTasks(ctx, &jobtaskpb.ListJobTasksRequest{
				Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_phase_id", phaseIDs[start:end])}},
				Pagination: &commonpb.PaginationRequest{
					Limit:  int32(pageLimit),
					Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
				},
			})
			if err != nil {
				log.Printf("report card doc: list job tasks: %v", err)
				break
			}
			data := resp.GetData()
			for _, tk := range data {
				id := tk.GetId()
				jid := jobByPhase[tk.GetJobPhaseId()]
				if id == "" || jid == "" || !tk.GetActive() {
					continue
				}
				if _, seen := taskJob[id]; seen {
					continue
				}
				taskJob[id] = jid
				taskIDs = append(taskIDs, id)
				if tt := tk.GetTemplateTaskId(); tt != "" {
					taskTmplTask[id] = tt
				}
			}
			if len(data) < pageLimit {
				break
			}
		}
	}
	if len(taskIDs) == 0 {
		return out
	}

	// task_outcome: MAX numeric_value per (job, criterion). Also remember which
	// template_task each (job, criterion) came from, for the sequence lookup.
	critMax := map[string]map[string]float64{}        // jobID → criteriaVersionID → MAX
	jobCritTmplTask := map[string]map[string]string{} // jobID → criteriaVersionID → templateTaskID
	tmplTaskSet := map[string]bool{}
	for start := 0; start < len(taskIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(taskIDs) {
			end = len(taskIDs)
		}
		chunk := taskIDs[start:end]
		for page := int32(1); page <= maxPages; page++ {
			resp, err := d.ListTaskOutcomes(ctx, &taskoutcomepb.ListTaskOutcomesRequest{
				Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_task_id", chunk)}},
				Pagination: &commonpb.PaginationRequest{
					Limit:  int32(pageLimit),
					Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
				},
			})
			if err != nil {
				log.Printf("report card doc: list task outcomes: %v", err)
				break
			}
			data := resp.GetData()
			for _, t := range data {
				if !t.GetActive() || t.NumericValue == nil {
					continue
				}
				jid := taskJob[t.GetJobTaskId()]
				cid := t.GetCriteriaVersionId()
				if jid == "" || cid == "" {
					continue
				}
				v := t.GetNumericValue()
				if critMax[jid] == nil {
					critMax[jid] = map[string]float64{}
				}
				if cur, ok := critMax[jid][cid]; !ok || v > cur {
					critMax[jid][cid] = v
				}
				if tt := taskTmplTask[t.GetJobTaskId()]; tt != "" {
					if jobCritTmplTask[jid] == nil {
						jobCritTmplTask[jid] = map[string]string{}
					}
					jobCritTmplTask[jid][cid] = tt
					tmplTaskSet[tt] = true
				}
			}
			if len(data) < pageLimit {
				break
			}
		}
	}

	seq := fetchCriteriaSequence(ctx, d, tmplTaskSet)

	// Assemble each subject: order its criteria by sequence_order (criteria with
	// no sequence sort last, tie-broken by id for determinism), slot the first
	// four into A/B/C/D, and Σ all MAXes as the /32 total.
	const noSeq = int32(1) << 30
	for jid, cmax := range critMax {
		type cv struct {
			seq int32
			id  string
			val float64
		}
		items := make([]cv, 0, len(cmax))
		for cid, val := range cmax {
			s := noSeq
			if tt := jobCritTmplTask[jid][cid]; tt != "" {
				if smap, ok := seq[tt]; ok {
					if sv, ok := smap[cid]; ok {
						s = sv
					}
				}
			}
			items = append(items, cv{s, cid, val})
		}
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].seq != items[j].seq {
				return items[i].seq < items[j].seq
			}
			return items[i].id < items[j].id
		})
		var c criteria
		slots := []*string{&c.a, &c.b, &c.c, &c.d}
		sum := 0.0
		for i, it := range items {
			sum += it.val
			if i < len(slots) {
				*slots[i] = fmtNum(it.val)
			}
		}
		if len(items) > 0 {
			c.total = fmtNum(sum)
		}
		out[jid] = c
	}
	return out
}

// fetchCriteriaSequence loads template_task_criteria.sequence_order for the given
// job_template_task ids, returning templateTaskID → criteriaVersionID → order.
// This is the A/B/C/D column ordering the grade sheet uses.
func fetchCriteriaSequence(ctx context.Context, d *Deps, tmplTaskSet map[string]bool) map[string]map[string]int32 {
	seq := map[string]map[string]int32{}
	if d.ListTemplateTaskCriterias == nil || len(tmplTaskSet) == 0 {
		return seq
	}
	tmplTaskIDs := make([]string, 0, len(tmplTaskSet))
	for tt := range tmplTaskSet {
		tmplTaskIDs = append(tmplTaskIDs, tt)
	}
	for start := 0; start < len(tmplTaskIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(tmplTaskIDs) {
			end = len(tmplTaskIDs)
		}
		chunk := tmplTaskIDs[start:end]
		for page := int32(1); page <= maxPages; page++ {
			resp, ok := listTemplateTaskCriteriasSafe(ctx, d, &ttcpb.ListTemplateTaskCriteriasRequest{
				Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("job_template_task_id", chunk)}},
				Pagination: &commonpb.PaginationRequest{
					Limit:  int32(pageLimit),
					Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
				},
			})
			if !ok {
				// Sequence source unavailable — bail with whatever ordering we
				// have; callers fall back to a stable order (see below).
				return seq
			}
			data := resp.GetData()
			for _, cr := range data {
				if !cr.GetActive() {
					continue
				}
				tt := cr.GetJobTemplateTaskId()
				cid := cr.GetOutcomeCriteriaId()
				if tt == "" || cid == "" {
					continue
				}
				if seq[tt] == nil {
					seq[tt] = map[string]int32{}
				}
				seq[tt][cid] = cr.GetSequenceOrder()
			}
			if len(data) < pageLimit {
				break
			}
		}
	}
	return seq
}

// listTemplateTaskCriteriasSafe calls the sequence-order source defensively.
// The A/B/C/D ordering is optional enrichment: on containers where the
// template_task_criteria repository is not wired, the underlying use case
// dereferences a nil repository and panics *after* its transaction has already
// rolled back. Recovering here is therefore safe (no partial DB state) and lets
// the criterion columns fall back to a stable order rather than failing the
// whole download. Returns ok=false on any panic or error.
func listTemplateTaskCriteriasSafe(ctx context.Context, d *Deps, req *ttcpb.ListTemplateTaskCriteriasRequest) (resp *ttcpb.ListTemplateTaskCriteriasResponse, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("report card doc: template_task_criteria unavailable, using stable criterion order: %v", r)
			resp, ok = nil, false
		}
	}()
	r, err := d.ListTemplateTaskCriterias(ctx, req)
	if err != nil {
		log.Printf("report card doc: list template task criteria: %v", err)
		return nil, false
	}
	return r, true
}

// isNonEnrolledPlaceholder reports whether a subject row is a non-enrolled
// placeholder that must not render — the DOCX mirror of the grade-loader T8
// suppressed_zero_grade rule. A row is a placeholder when it carries NO positive
// grade evidence:
//
//   - no per-criterion mark is > 0 (the authoritative discriminator: on
//     education1 every all-zero-mark job is a non-enrolled parallel track; a
//     real subject — including a percentage subject whose IB year-final reads
//     "0" — always has ≥1 positive task_outcome), AND
//   - no REAL year-final or semester band (> 1; the transmute-of-zero floor is
//     "0"/"1" and is not evidence of enrollment), AND
//   - it either HAS task_outcome marks (the all-zero active scaffold) OR has no
//     summary at all (a fully-blank row).
//
// A genuinely-enrolled subject with a real 0 keeps rendering: it has a positive
// mark somewhere, a real (>1) stored final, or — for historical imports — no
// task_outcome but a real year-final (hasMarks=false + a summary present).
func isNonEnrolledPlaceholder(row subjectRow, hasMarks bool) bool {
	// Delegate to the shared enrollment predicate so the grid/card surfaces and
	// the DOCX apply ONE definition of "non-enrolled placeholder". HasPositiveMark
	// is derived from this row's already-fetched per-criterion marks (the DOCX
	// fetches marks per-criterion via fetchCriteriaByJob rather than the generic
	// existence walk, so it computes the positive-mark signal here from the row).
	ev := outcome_summary.EnrollmentEvidence{
		HasMarks: hasMarks,
		HasPositiveMark: outcome_summary.NumGreaterThan(row.CritA, 0) ||
			outcome_summary.NumGreaterThan(row.CritB, 0) ||
			outcome_summary.NumGreaterThan(row.CritC, 0) ||
			outcome_summary.NumGreaterThan(row.CritD, 0) ||
			outcome_summary.NumGreaterThan(row.Total, 0),
	}
	return outcome_summary.IsNonEnrolledCell(ev, row.YearFinal, row.Sem1Band, row.Sem2Band)
}

// --- small helpers --------------------------------------------------------

var ayRe = regexp.MustCompile(`\(\s*AY\s*([^)]+?)\s*\)`)
var gradeRe = regexp.MustCompile(`^(Grade\s+\S+)\s+(.+)$`)
var subjSuffixRe = regexp.MustCompile(`\s*(?:—|–|-)\s*AY\s.*$`)

// sectionParts splits "Grade 9 Gold (AY 2025-26)" → ("Grade 9 Gold", "2025-26").
func sectionParts(full string) (name, ay string) {
	full = strings.TrimSpace(full)
	if m := ayRe.FindStringSubmatch(full); len(m) == 2 {
		ay = strings.TrimSpace(m[1])
	}
	name = strings.TrimSpace(ayRe.ReplaceAllString(full, ""))
	return name, ay
}

// gradeSection splits "Grade 9 Gold" → ("Grade 9", "Gold").
func gradeSection(name string) (grade, section string) {
	if m := gradeRe.FindStringSubmatch(name); len(m) == 3 {
		return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	}
	return "", name
}

// cleanSubject strips the trailing " — AY 2025-2026" from a subject label.
func cleanSubject(name string) string {
	return strings.TrimSpace(subjSuffixRe.ReplaceAllString(name, ""))
}

func studentName(ctx context.Context, d *Deps, clientID string) string {
	if d.ListClients == nil || clientID == "" {
		return clientID
	}
	resp, err := d.ListClients(ctx, &clientpb.ListClientsRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("id", clientID)}},
	})
	if err != nil {
		log.Printf("report card doc: list client: %v", err)
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
	}
	return clientID
}

func colName(names map[string]string, id string) string {
	if n := strings.TrimSpace(names[id]); n != "" {
		return n
	}
	return id
}

func fmtNum(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func orBlank(s string) string { return strings.TrimSpace(s) }

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return dash
	}
	return s
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
