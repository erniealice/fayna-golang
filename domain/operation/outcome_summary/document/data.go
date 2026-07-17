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
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

const (
	pageLimit = 100
	maxPages  = 100
	dash      = "—"
)

// criterionRow is one assessment-criterion row of a subject block on the v2
// (block-layout) template: the lettered display label ("A - Investigating")
// and the per-period recomputed MAX marks.
type criterionRow struct {
	Label string
	Sem1  string
	Sem2  string
}

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

	// v2 (block-layout) enrichments — additive; the v1 summary-table template
	// never references these keys.
	Title       string         // display title (rotation pair merged, e.g. "Arts: Visual Arts / Arts: Music"); falls back to Name
	TeacherLine string         // "Teacher: X" / "Teachers: X / Y" (label wording from lyngua)
	Criteria    []criterionRow // ordered criterion rows with per-period marks
	Sem1Total   string         // Σ of the period-1 per-criterion MAX; "" when no marks
	Sem2Total   string         // Σ of the period-2 per-criterion MAX; "" when no marks
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
	// FormationGroups are the Formation-page (DOCX v2) category blocks — the
	// non-academic outcome categories (subject/homeroom deportment) with each
	// strand's frozen authoritative average. Empty when no category filter is
	// configured or the student has no deportment data. See formation.go.
	FormationGroups []formationGroup

	// v2 (block-layout) enrichments — additive root fields; every one renders
	// blank when its source is unwired/absent (never a placeholder leak).
	SchedulePeriodSpaced string          // "2025 - 2026" (cover-page variant of SchedulePeriod)
	GradeSection         string          // "Grade 7 - Nickel" (identity-line variant of GroupLevel+SectionName)
	Adviser              string          // group-lead staff name (the group-category job's task assignee)
	ClientReference      string          // client_attributes.<configured code> value (e.g. LRN)
	PrintedByName        string          // display name of the printing user; falls back to PrintedBy
	PrintedAtLong        string          // "July 13, 2026 08:13 AM" long-form print timestamp
	ConductRows          []conductRow    // per-subject conduct table rows (rotation pairs merged)
	GroupConductSem1     string          // group (homeroom) conduct, period 1
	GroupConductSem2     string          // group (homeroom) conduct, period 2
}

// conductRow is one row of the per-subject conduct (deportment) table on the
// formation page: the merged display title and the per-period values.
type conductRow struct {
	Title string
	Sem1  string
	Sem2  string
}

// buildReportCardData mirrors buildInvoiceData: it flattens the assembled card
// into the doctemplate data map, emitting EVERY key referenced by the template
// as a pre-formatted string (blank/dash, never omitted) so no raw {{..}} leaks
// (engine leaks unresolved placeholders verbatim — G3/G5).
func buildReportCardData(rc reportCard) map[string]any {
	subjects := make([]any, 0, len(rc.Subjects))
	for _, s := range rc.Subjects {
		criteria := make([]any, 0, len(s.Criteria))
		for _, c := range s.Criteria {
			criteria = append(criteria, map[string]any{
				"crit_label": orBlank(c.Label),
				"crit_sem1":  orBlank(c.Sem1),
				"crit_sem2":  orBlank(c.Sem2),
			})
		}
		title := orBlank(s.Title)
		if title == "" {
			title = orBlank(s.Name)
		}
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
			// v2 block-layout keys (additive; unreferenced by the v1 template).
			"subject_title": title,
			"teacher_line":  orBlank(s.TeacherLine),
			"criteria":      criteria,
			"sem1_total":    orBlank(s.Sem1Total),
			"sem2_total":    orBlank(s.Sem2Total),
		})
	}
	conduct := make([]any, 0, len(rc.ConductRows))
	for _, r := range rc.ConductRows {
		conduct = append(conduct, map[string]any{
			"conduct_title": orBlank(r.Title),
			"conduct_sem1":  orBlank(r.Sem1),
			"conduct_sem2":  orBlank(r.Sem2),
		})
	}
	return map[string]any{
		"school_name":      orBlank(rc.DocumentHeaderName),
		"academic_year":    orBlank(rc.SchedulePeriod),
		"student_name":     orBlank(rc.ClientName),
		"grade_level":      orBlank(rc.GroupLevel),
		"section_name":     orBlank(rc.SectionName),
		"lrn":              orDash(rc.LRN),
		"printed_by":       orBlank(rc.PrintedBy),
		"printed_at":       orBlank(rc.PrintedAt),
		"subjects":         subjects,
		"formation_groups": formationData(rc.FormationGroups),
		// v2 block-layout root keys (additive).
		"academic_year_spaced": orBlank(rc.SchedulePeriodSpaced),
		"grade_section":        orBlank(rc.GradeSection),
		"adviser":              orBlank(rc.Adviser),
		"client_reference":     orBlank(rc.ClientReference),
		"printed_by_name":      orBlank(firstNonEmpty(rc.PrintedByName, rc.PrintedBy)),
		"printed_at_long":      orBlank(firstNonEmpty(rc.PrintedAtLong, rc.PrintedAt)),
		"conduct_rows":         conduct,
		"group_conduct_sem1":   orBlank(rc.GroupConductSem1),
		"group_conduct_sem2":   orBlank(rc.GroupConductSem2),
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
	cats := fetchCategories(ctx, d)
	groupCatID := categoryIDByCode(cats, d.DocOptions.GroupCategoryFilter)
	jobTemplate := map[string]string{}
	jobIDs := make([]string, 0, len(jobs))
	templateIDs := []string{}
	tmplSeen := map[string]bool{}
	// deportJobs are the NON-academic categorized jobs (the deportment complement)
	// that feed the Formation page (DOCX v2). They ride the same enrollment
	// subscription as the academic jobs; the academic loop drops them, the
	// Formation builder keeps them. Only populated when a real category filter is
	// configured (KeepJobInCategory returns true for every job when catID=="", so
	// the !Keep branch never fires on the unfiltered service-admin tier).
	var deportJobs []*jobpb.Job
	// groupJob is the single job of the configured GROUP category
	// (DocumentOptions.GroupCategoryFilter, e.g. homeroom): its task assignee is
	// the document's group lead ("Adviser") and its per-phase summaries the group
	// conduct row. Nil when unconfigured or absent.
	var groupJob *jobpb.Job
	for _, j := range jobs {
		if !catOK {
			continue
		}
		if !outcome_summary.KeepJobInCategory(catID, j.GetJobCategoryId()) {
			if jc := j.GetJobCategoryId(); jc != "" && jc != catID {
				deportJobs = append(deportJobs, j)
				if groupJob == nil && groupCatID != "" && jc == groupCatID {
					groupJob = j
				}
			}
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
	// The transcript walk covers the academic jobs plus the group-category job:
	// the group job contributes no subject row, but its graded tasks carry the
	// group-lead assignee (the document "Adviser").
	walkIDs := jobIDs
	if groupJob != nil && groupJob.GetId() != "" {
		walkIDs = append(append([]string{}, jobIDs...), groupJob.GetId())
	}
	phaseOrder, jobByPhase := fetchPhaseOrders(ctx, d, walkIDs)
	semByJob := fetchSemesterLabels(ctx, d, walkIDs, phaseOrder)
	yearByJob := fetchYearLabels(ctx, d, jobIDs)
	// Per-criterion marks + per-period teacher assignment from task_outcome /
	// job_task, scoped to THIS card's jobs (fixes cross-AY accumulation). Keyed
	// by job id; MAX-within-criterion per period; A/B/C/D ordered by
	// template_task_criteria.sequence_order.
	transcripts := fetchTranscripts(ctx, d, jobByPhase, phaseOrder)

	// Display enrichment lookups (all optional/nil-safe → blank fields).
	critNames := fetchCriterionNames(ctx, d, transcripts)
	staffNames := fetchStaffNames(ctx, d, transcripts)

	// Rotation-pair merge (G1): a canonical subject (e.g. "Arts") whose
	// conduct-category strands ("Arts: Visual Arts", "Arts: Music") identify the
	// per-period variants renders under the merged pair title. The conduct table
	// reuses the same pairing.
	conduct := fetchConduct(ctx, d, deportJobs, groupJob, historical)
	academicNames := map[string]bool{}
	for _, jid := range jobIDs {
		academicNames[strings.ToLower(cleanSubject(colName(tmplNames, jobTemplate[jid])))] = true
	}
	merged := mergeRotationPairs(conduct, academicNames, fetchInactiveSubjectNames(ctx, d, subID))

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
		tr := transcripts[e.jobID]
		// crit is derived from the transcript iff this job had ≥1 numeric
		// task_outcome (even an all-zero one); hasMarks distinguishes an all-zero
		// scaffold (present, total="0") from a subject with no task_outcome at all.
		crit, hasMarks := tr.yearCriteria()
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
		// v2 block-layout enrichments (blank-safe when sources are unwired).
		row.Title = merged.titleFor(display)
		row.TeacherLine = teacherLine(d.Labels.Student, tr, staffNames)
		row.Criteria, row.Sem1Total, row.Sem2Total = tr.criterionRows(critNames)
		rows = append(rows, row)
	}

	// Formation page (DOCX v2): the deportment-category tables. Additive to the
	// academic subject transcript; empty when the tier has no category filter or the
	// student has no deportment data.
	formationGroups := collectFormationGroups(ctx, d, deportJobs, historical)

	// Per-subject conduct rows + group conduct (v2 block layout).
	conductRows, groupSem1, groupSem2 := buildConductRows(conduct, merged)

	// Group lead ("Adviser") = the modal task assignee on the group-category job.
	adviser := ""
	if groupJob != nil {
		adviser = groupLeadName(transcripts[groupJob.GetId()], staffNames)
	}

	name, ay := sectionParts(group.GetName())
	grade, section := gradeSection(name)
	// Prefer the price_schedule display name for the schedule period (live
	// format "2025-2026"); the section-name suffix ("2025-26") stays the
	// fallback.
	if psPeriod := fetchSchedulePeriod(ctx, d, group.GetPriceScheduleId()); psPeriod != "" {
		ay = psPeriod
	}
	gradeSectionLine := section
	if grade != "" {
		gradeSectionLine = grade + " - " + section
	}
	rc := &reportCard{
		DocumentHeaderName:   strings.TrimSpace(d.DocumentHeaderName),
		SchedulePeriod:       ay,
		SchedulePeriodSpaced: strings.Replace(ay, "-", " - ", 1),
		ClientName:           studentName(ctx, d, clientID),
		GroupLevel:           grade,
		SectionName:          section,
		GradeSection:         gradeSectionLine,
		LRN:                  "",
		ClientReference:      fetchClientReference(ctx, d, clientID),
		Adviser:              adviser,
		Subjects:             rows,
		PriceScheduleID:      group.GetPriceScheduleId(),
		FormationGroups:      formationGroups,
		ConductRows:          conductRows,
		GroupConductSem1:     groupSem1,
		GroupConductSem2:     groupSem2,
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

// transcript is one job's full graded-leaf walk result: per-period MAX marks
// per criterion (period = job_phase.phase_order; 0 collects order-less phases
// so the year collapse never loses a mark), the criterion presentation order,
// and the per-period task-assignee tally (the teacher signal).
type transcript struct {
	marks    map[string]map[int32]float64 // critID → phaseOrder → MAX numeric_value
	seq      map[string]int32             // critID → template_task_criteria.sequence_order
	teachers map[int32]map[string]int     // phaseOrder → assignee (staff id) → active-task count
}

const noSeq = int32(1) << 30

// orderedCrits returns the job's criterion ids in presentation order
// (sequence_order ASC, id ASC — the grade sheet's A/B/C/D order).
func (t *transcript) orderedCrits() []string {
	if t == nil || len(t.marks) == 0 {
		return nil
	}
	ids := make([]string, 0, len(t.marks))
	for cid := range t.marks {
		ids = append(ids, cid)
	}
	sort.SliceStable(ids, func(i, j int) bool {
		si, sj := t.seqOf(ids[i]), t.seqOf(ids[j])
		if si != sj {
			return si < sj
		}
		return ids[i] < ids[j]
	})
	return ids
}

func (t *transcript) seqOf(cid string) int32 {
	if t == nil || t.seq == nil {
		return noSeq
	}
	if s, ok := t.seq[cid]; ok {
		return s
	}
	return noSeq
}

// yearCriteria collapses the per-period marks into the v1 A-D + total shape
// (MAX across the year per criterion — identical semantics to the original
// year-collapsed fetch). ok reports whether the job had any numeric mark.
func (t *transcript) yearCriteria() (criteria, bool) {
	var c criteria
	if t == nil || len(t.marks) == 0 {
		return c, false
	}
	ids := t.orderedCrits()
	slots := []*string{&c.a, &c.b, &c.c, &c.d}
	sum := 0.0
	for i, cid := range ids {
		best := 0.0
		first := true
		for _, v := range t.marks[cid] {
			if first || v > best {
				best, first = v, false
			}
		}
		sum += best
		if i < len(slots) {
			*slots[i] = fmtNum(best)
		}
	}
	c.total = fmtNum(sum)
	return c, true
}

// criterionRows builds the v2 block-layout criterion rows ("A - Investigating"
// + per-period marks) and the per-period criteria totals. A period with no
// marks stays blank ("").
func (t *transcript) criterionRows(names map[string]string) ([]criterionRow, string, string) {
	if t == nil || len(t.marks) == 0 {
		return nil, "", ""
	}
	ids := t.orderedCrits()
	rows := make([]criterionRow, 0, len(ids))
	sums := map[int32]float64{}
	has := map[int32]bool{}
	for i, cid := range ids {
		display := strings.TrimSpace(names[cid])
		if display == "" {
			display = cid
		}
		label := display
		if i < 26 {
			label = string(rune('A'+i)) + " - " + display
		}
		row := criterionRow{Label: label}
		if v, ok := t.marks[cid][1]; ok {
			row.Sem1 = fmtNum(v)
			sums[1] += v
			has[1] = true
		}
		if v, ok := t.marks[cid][2]; ok {
			row.Sem2 = fmtNum(v)
			sums[2] += v
			has[2] = true
		}
		rows = append(rows, row)
	}
	sem1, sem2 := "", ""
	if has[1] {
		sem1 = fmtNum(sums[1])
	}
	if has[2] {
		sem2 = fmtNum(sums[2])
	}
	return rows, sem1, sem2
}

// fetchTranscripts loads the per-criterion marks AND the per-period task
// assignees for the card's jobs from job_task / task_outcome — the
// authoritative per-criterion leaf (job_outcome_line on education1 is
// per-subject only). It walks the grade-sheet join path job_phase → job_task →
// task_outcome, SCOPED to this card's phases/jobs (so a student's prior-AY /
// other-section grades never accumulate in), groups by (job, period, criterion)
// taking the numeric MAX per period, and resolves the criterion order via
// template_task_criteria.sequence_order (falling back to a stable criteria-id
// order when that source is unavailable). Keyed by job id. Fully nil-safe — a
// missing closure yields an empty map and the criterion columns render blank.
func fetchTranscripts(ctx context.Context, d *Deps, jobByPhase map[string]string, phaseOrder map[string]int32) map[string]*transcript {
	out := map[string]*transcript{}
	if d.ListJobTasks == nil || d.ListTaskOutcomes == nil || len(jobByPhase) == 0 {
		return out
	}

	phaseIDs := make([]string, 0, len(jobByPhase))
	for pid := range jobByPhase {
		phaseIDs = append(phaseIDs, pid)
	}

	byJob := func(jid string) *transcript {
		t := out[jid]
		if t == nil {
			t = &transcript{
				marks:    map[string]map[int32]float64{},
				seq:      map[string]int32{},
				teachers: map[int32]map[string]int{},
			}
			out[jid] = t
		}
		return t
	}

	// job_task: task id → owning job (via phase), period (via phase order),
	// template_task id (for the sequence_order lookup), and assignee tally.
	taskJob := map[string]string{}
	taskPeriod := map[string]int32{}
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
				period := phaseOrder[tk.GetJobPhaseId()]
				taskJob[id] = jid
				taskPeriod[id] = period
				taskIDs = append(taskIDs, id)
				if tt := tk.GetTemplateTaskId(); tt != "" {
					taskTmplTask[id] = tt
				}
				if who := strings.TrimSpace(tk.GetAssignedTo()); who != "" && period > 0 {
					t := byJob(jid)
					if t.teachers[period] == nil {
						t.teachers[period] = map[string]int{}
					}
					t.teachers[period][who]++
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

	// task_outcome: MAX numeric_value per (job, period, criterion). Also
	// remember which template_task each (job, criterion) came from, for the
	// sequence lookup.
	jobCritTmplTask := map[string]map[string]string{} // jobID → critID → templateTaskID
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
				period := taskPeriod[t.GetJobTaskId()]
				tr := byJob(jid)
				if tr.marks[cid] == nil {
					tr.marks[cid] = map[int32]float64{}
				}
				if cur, ok := tr.marks[cid][period]; !ok || v > cur {
					tr.marks[cid][period] = v
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
	for jid, tr := range out {
		for cid, tt := range jobCritTmplTask[jid] {
			if smap, ok := seq[tt]; ok {
				if sv, ok := smap[cid]; ok {
					tr.seq[cid] = sv
				}
			}
		}
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

// fetchCriterionNames resolves the display strings for every criterion id in
// the transcripts via outcome_criteria. The DESCRIPTION is preferred (it
// carries the operator's exact rendered casing, e.g. "Knowing and
// understanding"); the NAME is the fallback. Nil-safe — a missing closure
// yields an empty map and labels fall back to the raw id.
func fetchCriterionNames(ctx context.Context, d *Deps, transcripts map[string]*transcript) map[string]string {
	out := map[string]string{}
	if d.ListOutcomeCriterias == nil {
		return out
	}
	seen := map[string]bool{}
	ids := []string{}
	for _, tr := range transcripts {
		if tr == nil {
			continue
		}
		for cid := range tr.marks {
			if cid != "" && !seen[cid] {
				seen[cid] = true
				ids = append(ids, cid)
			}
		}
	}
	if len(ids) == 0 {
		return out
	}
	sort.Strings(ids)
	for start := 0; start < len(ids); start += pageLimit {
		end := start + pageLimit
		if end > len(ids) {
			end = len(ids)
		}
		resp, err := d.ListOutcomeCriterias(ctx, &criteriapb.ListOutcomeCriteriasRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("id", ids[start:end])}},
		})
		if err != nil {
			log.Printf("report card doc: list outcome criteria: %v", err)
			continue
		}
		for _, c := range resp.GetData() {
			id := c.GetId()
			if id == "" {
				continue
			}
			if desc := strings.TrimSpace(c.GetDescription()); desc != "" {
				out[id] = desc
			} else if name := strings.TrimSpace(c.GetName()); name != "" {
				out[id] = name
			}
		}
	}
	return out
}

// fetchStaffNames resolves the task assignees (staff ids) captured in the
// transcripts to display names via the User-hydrating staff read
// (GetStaffListPageData — the bare ListStaffs never populates Staff.User).
// Nil-safe — a missing closure yields an empty map and staff lines stay blank.
func fetchStaffNames(ctx context.Context, d *Deps, transcripts map[string]*transcript) map[string]string {
	out := map[string]string{}
	if d.GetStaffListPageData == nil {
		return out
	}
	seen := map[string]bool{}
	ids := []string{}
	for _, tr := range transcripts {
		if tr == nil {
			continue
		}
		for _, tally := range tr.teachers {
			for sid := range tally {
				if sid != "" && !seen[sid] {
					seen[sid] = true
					ids = append(ids, sid)
				}
			}
		}
	}
	if len(ids) == 0 {
		return out
	}
	sort.Strings(ids)
	for start := 0; start < len(ids); start += pageLimit {
		end := start + pageLimit
		if end > len(ids) {
			end = len(ids)
		}
		resp, err := d.GetStaffListPageData(ctx, &staffpb.GetStaffListPageDataRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("id", ids[start:end])}},
		})
		if err != nil {
			log.Printf("report card doc: staff list page data: %v", err)
			continue
		}
		for _, s := range resp.GetStaffList() {
			id := s.GetId()
			u := s.GetUser()
			if id == "" || u == nil {
				continue
			}
			name := strings.TrimSpace(strings.TrimSpace(u.GetFirstName()) + " " + strings.TrimSpace(u.GetLastName()))
			if name != "" {
				out[id] = name
			}
		}
	}
	return out
}

// topAssignee picks the modal assignee from a per-period tally. Ties prefer an
// assignee different from avoid (the other period's pick — the rotation-pair
// signal), then the lexicographically smallest id for determinism.
func topAssignee(tally map[string]int, avoid string) string {
	best, bestCount := "", -1
	for sid, n := range tally {
		switch {
		case n > bestCount:
			best, bestCount = sid, n
		case n == bestCount:
			bPref, sPref := best != avoid, sid != avoid
			if sPref && !bPref {
				best = sid
			} else if sPref == bPref && sid < best {
				best = sid
			}
		}
	}
	return best
}

// teacherLine composes the per-subject staff line: one distinct assignee →
// "Teacher: X"; two (the rotation pair, period-1 first) → "Teachers: X / Y".
// Wording comes from the lyngua-backed labels; blank when nothing resolves.
func teacherLine(labels outcome_summary.PeriodLabels, tr *transcript, names map[string]string) string {
	if tr == nil {
		return ""
	}
	s1 := topAssignee(tr.teachers[1], "")
	s2 := topAssignee(tr.teachers[2], s1)
	n1, n2 := strings.TrimSpace(names[s1]), strings.TrimSpace(names[s2])
	if n1 == n2 {
		n2 = ""
	}
	single := strings.TrimSpace(labels.StaffLabel)
	plural := strings.TrimSpace(labels.StaffPluralLabel)
	switch {
	case n1 != "" && n2 != "":
		return plural + " " + n1 + " / " + n2
	case n1 != "":
		return single + " " + n1
	case n2 != "":
		return single + " " + n2
	}
	return ""
}

// groupLeadName resolves the group-category job's modal assignee (across every
// period) to a display name — the document's "Adviser" line.
func groupLeadName(tr *transcript, names map[string]string) string {
	if tr == nil {
		return ""
	}
	merged := map[string]int{}
	for _, tally := range tr.teachers {
		for sid, n := range tally {
			merged[sid] += n
		}
	}
	return strings.TrimSpace(names[topAssignee(merged, "")])
}

// fetchSchedulePeriod resolves the section's price_schedule display period —
// the schedule name with any non-numeric prefix dropped ("AY 2025-2026" →
// "2025-2026"). Blank when the schedule is unresolvable (callers keep the
// section-name fallback).
func fetchSchedulePeriod(ctx context.Context, d *Deps, priceScheduleID string) string {
	if d.ListPriceSchedules == nil || priceScheduleID == "" {
		return ""
	}
	// List-and-match: the schedule list is tiny and the workspace-bound list is
	// the closure every sibling surface already uses. TWO passes — the default
	// (active) list and an explicit inactive list — because a CLOSED historical
	// schedule row is active=false (the fetchSection precedent).
	requests := []*priceschedulepb.ListPriceSchedulesRequest{
		{},
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{boolEq("active", false)}}},
	}
	for _, req := range requests {
		resp, err := d.ListPriceSchedules(ctx, req)
		if err != nil {
			log.Printf("report card doc: list price schedules: %v", err)
			continue
		}
		for _, ps := range resp.GetData() {
			if ps.GetId() != priceScheduleID {
				continue
			}
			name := strings.TrimSpace(ps.GetName())
			if idx := strings.IndexFunc(name, func(r rune) bool { return r >= '0' && r <= '9' }); idx >= 0 {
				return strings.TrimSpace(name[idx:])
			}
			return name
		}
	}
	return ""
}

// fetchClientReference resolves the configured client-reference attribute
// (DocumentOptions.ClientReferenceAttributeCode, e.g. "lrn") for the client.
// Blank when unconfigured, unwired, or absent.
func fetchClientReference(ctx context.Context, d *Deps, clientID string) string {
	code := strings.TrimSpace(d.DocOptions.ClientReferenceAttributeCode)
	if code == "" || clientID == "" || d.ListClientAttributes == nil || d.ResolveAttributeIDByCode == nil {
		return ""
	}
	attrID, err := d.ResolveAttributeIDByCode(ctx, code)
	if err != nil || strings.TrimSpace(attrID) == "" {
		return ""
	}
	resp, err := d.ListClientAttributes(ctx, &clientattributepb.ListClientAttributesRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
			stringEq("client_id", clientID), stringEq("attribute_id", attrID),
		}},
	})
	if err != nil {
		log.Printf("report card doc: list client attributes: %v", err)
		return ""
	}
	for _, ca := range resp.GetData() {
		if !ca.GetActive() {
			continue
		}
		if v := strings.TrimSpace(ca.GetValue()); v != "" {
			return v
		}
	}
	return ""
}

// categoryIDByCode maps a job_category CODE to its id via the already-fetched
// category metadata. Empty when the code is blank or unknown.
func categoryIDByCode(cats map[string]catInfo, code string) string {
	want := strings.ToLower(strings.TrimSpace(code))
	if want == "" {
		return ""
	}
	for id, c := range cats {
		if strings.ToLower(strings.TrimSpace(c.code)) == want {
			return id
		}
	}
	return ""
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
