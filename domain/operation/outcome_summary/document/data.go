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

// templateReferencedAttributeCodes is the LEAK-LAW guard for the v3 template's
// hard-coded {{client_attributes.<code>}} placeholders. buildReportCardData seeds
// the client_attributes map with a blank leaf for EVERY code listed here BEFORE
// overlaying the app-configured DocumentOptions.ClientAttributeCodes values, so a
// referenced placeholder always resolves (blank, never a verbatim leak) even when
// a block consumer omits the code from ClientAttributeCodes. This list MUST stay
// in sync with the {{client_attributes.<code>}} placeholders in gen_template_v3.py
// — add a code here whenever the generator adds a new client_attributes.<code>
// placeholder. Today the v3 artifact references exactly one such code: "lrn".
var templateReferencedAttributeCodes = []string{"lrn"}

// criterionRow is one assessment-criterion row of an item block on the
// block-layout template: the lettered display label ("A - Investigating")
// and the per-phase recomputed MAX marks (phase = job_phase.phase_order).
type criterionRow struct {
	Label  string
	Phase1 string
	Phase2 string
}

// itemRow is one row of the report card (one item / job — e.g. a school
// subject). Generic identifiers; the vertical wording lives in lyngua values
// and the template placeholder keys.
type itemRow struct {
	Name      string
	CritA     string // per-criterion (job_outcome_line sub-strands); "—" when absent
	CritB     string
	CritC     string
	CritD     string
	Total     string // Σ of the per-criterion MAX (/32); "—" when no criteria
	Sem1Band  string // phase_outcome_summary (phase_order 1) — recompute-equivalent
	Sem2Band  string // phase_outcome_summary (phase_order 2) — recompute-equivalent
	YearFinal string // job_outcome_summary.scaled_label — STORED, never recomputed

	// block-layout enrichments — additive; the v1 summary-table template
	// never references these keys.
	ItemTitle string         // display title (rotation pair merged, e.g. "Arts: Visual Arts / Arts: Music"); falls back to Name
	StaffLine string         // "Teacher: X" / "Teachers: X / Y" (label wording from lyngua)
	Criteria  []criterionRow // ordered criterion rows with per-phase marks
	Sem1Total string         // Σ of the phase-1 per-criterion MAX; "" when no marks
	Sem2Total string         // Σ of the phase-2 per-criterion MAX; "" when no marks
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
	Subjects           []itemRow
	// PriceScheduleID is the section's AY anchor (subscription_group.price_schedule_id),
	// threaded to the report-card template-binding resolver. Empty → the resolver
	// returns the workspace-wide fallback binding (or none → embedded template).
	PriceScheduleID string
	// FormationGroups are the Formation-page (DOCX v2) category blocks — the
	// non-academic outcome categories (subject/homeroom deportment) with each
	// strand's frozen authoritative average. Empty when no category filter is
	// configured or the student has no deportment data. See formation.go.
	FormationGroups []formationGroup

	// block-layout enrichments — additive root fields; every one renders
	// blank when its source is unwired/absent (never a placeholder leak).
	SchedulePeriodDisplay string         // "2025-2026" — the price_schedule display period (v2 key ONLY; SchedulePeriod keeps the v1 section-derived value, M4)
	SchedulePeriodSpaced  string         // "2025 - 2026" (cover-page variant of SchedulePeriodDisplay)
	GroupLabel            string         // "Grade 7 - Nickel" (subscription_group display; identity-line variant of GroupLevel+SectionName)
	GroupLead             string         // group-lead staff name (the group-category job's task assignee) — FROZEN v1/v2 "adviser" key (first-job behavior)
	LeadStaffDisplay      string         // block-layout root alias (lead_staff_name_display); populated ONLY when the group category has exactly one job (0/2+ ⇒ blank, so an arbitrary first-job adviser never leaks onto the cover/headers)
	ClientReference       string         // client_attributes.<configured code> value (e.g. LRN)
	ClientAttributes      map[string]any // code → value for every configured ClientAttributeCodes entry (always present, blank when absent)
	PrintedByName         string         // display name of the printing user; falls back to PrintedBy
	PrintedAtLong         string         // "July 13, 2026 08:13 AM" long-form print timestamp
	ItemRatings           []ratingRow    // per-item rating (deportment) table rows (rotation pairs merged)
	GroupRatingPhase1     string         // group (homeroom) rating, phase 1
	GroupRatingPhase2     string         // group (homeroom) rating, phase 2

	// JobIDs is the full set of this card's active jobs (academic + deportment +
	// group), used by the D5 render gate to test whether any feeding sheet is a
	// data-present, workflow-entered, not-fully-published phase. Never rendered.
	JobIDs []string

	// JobCategories is the converged generic block-layout tree: job_category.code
	// → subtree {jobs[], singleton projection}. Assembled by collectCard,
	// flattened by buildReportCardData under the "job_categories" root key, and
	// blank-guarded against the block manifest. Nil on the v1/v2 tiers (the block
	// artifact is the only reader).
	JobCategories map[string]any
}

// ratingRow is one row of the per-item rating (deportment) table on the
// formation page: the merged display title and the per-phase values.
type ratingRow struct {
	Title  string
	Phase1 string
	Phase2 string
}

// buildReportCardData mirrors buildInvoiceData: it flattens the assembled card
// into the doctemplate data map, emitting EVERY key referenced by the template
// as a pre-formatted string (blank/dash, never omitted) so no raw {{..}} leaks
// (engine leaks unresolved placeholders verbatim — G3/G5).
func buildReportCardData(rc reportCard) map[string]any {
	subjects := make([]any, 0, len(rc.Subjects))
	for _, s := range rc.Subjects {
		// v2 criteria loop (crit_*) — FROZEN emission shape.
		criteria := make([]any, 0, len(s.Criteria))
		for _, c := range s.Criteria {
			criteria = append(criteria, map[string]any{
				"crit_label": orBlank(c.Label),
				"crit_sem1":  orBlank(c.Phase1),
				"crit_sem2":  orBlank(c.Phase2),
			})
		}
		title := orBlank(s.ItemTitle)
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
			"teacher_line":  orBlank(s.StaffLine),
			"criteria":      criteria,
			"sem1_total":    orBlank(s.Sem1Total),
			"sem2_total":    orBlank(s.Sem2Total),
		})
	}
	conduct := make([]any, 0, len(rc.ItemRatings))
	for _, r := range rc.ItemRatings {
		conduct = append(conduct, map[string]any{
			"conduct_title": orBlank(r.Title),
			"conduct_sem1":  orBlank(r.Phase1),
			"conduct_sem2":  orBlank(r.Phase2),
		})
	}
	// client_attributes is ALWAYS a map (never nil) that carries a blank leaf for
	// EVERY code the template can reference — enforced HERE at the builder,
	// independent of any consumer's DocumentOptions.ClientAttributeCodes — so a
	// hard-coded {{client_attributes.<code>}} placeholder can never leak (HIGH#2).
	// Seed the template-referenced codes blank first, THEN overlay the resolved
	// per-client values (a real value wins over the seeded blank).
	clientAttributes := map[string]any{}
	for _, code := range templateReferencedAttributeCodes {
		clientAttributes[code] = ""
	}
	for code, val := range rc.ClientAttributes {
		clientAttributes[code] = val
	}
	jobCategories := rc.JobCategories
	if jobCategories == nil {
		jobCategories = map[string]any{}
	}
	data := map[string]any{
		// v1 emissions (FROZEN — the original summary-table template contract).
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
		// v2 block-layout root keys (FROZEN — EMITTED KEY STRINGS unchanged).
		"academic_year_display": orBlank(firstNonEmpty(rc.SchedulePeriodDisplay, rc.SchedulePeriod)),
		"academic_year_spaced":  orBlank(rc.SchedulePeriodSpaced),
		"grade_section":         orBlank(rc.GroupLabel),
		"adviser":               orBlank(rc.GroupLead),
		"client_reference":      orBlank(rc.ClientReference),
		"conduct_rows":          conduct,
		"group_conduct_sem1":    orBlank(rc.GroupRatingPhase1),
		"group_conduct_sem2":    orBlank(rc.GroupRatingPhase2),
		// Converged generic root scalars KEPT by the block contract.
		"price_schedule_name_display":        orBlank(firstNonEmpty(rc.SchedulePeriodDisplay, rc.SchedulePeriod)),
		"price_schedule_name_spaced_display": orBlank(rc.SchedulePeriodSpaced),
		"client_name_display":                orBlank(rc.ClientName),
		"subscription_group_name_display":    orBlank(rc.GroupLabel),
		"lead_staff_name_display":            orBlank(rc.LeadStaffDisplay),
		"client_attributes":                  clientAttributes,
		"printed_by_name":                    orBlank(firstNonEmpty(rc.PrintedByName, rc.PrintedBy)),
		"printed_at_long":                    orBlank(firstNonEmpty(rc.PrintedAtLong, rc.PrintedAt)),
		// Converged generic tree (block artifact reader).
		"job_categories": jobCategories,
	}
	// Blank-guard: seed EVERY manifest scalar path blank and EVERY manifest loop
	// path empty (recursing into real loop items) so no referenced {{leaf}} leaks
	// verbatim. Real values already present are never clobbered.
	applyBlockManifest(data)
	return data
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
	phaseOrder, jobByPhase, _ := fetchPhaseOrders(ctx, d, walkIDs, historical)
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
	conduct := fetchItemRatings(ctx, d, deportJobs, groupJob, historical)
	academicNames := map[string]bool{}
	for _, jid := range jobIDs {
		academicNames[strings.ToLower(cleanSubject(colName(tmplNames, jobTemplate[jid])))] = true
	}
	merged := mergeRotationPairs(conduct, academicNames, fetchInactiveSubjectNames(ctx, d, subID))

	// --- converged block-layout tree fetches (phase codes + strict labels) ---
	// The block tree keys its per-phase leaves by job_template_phase.code, reached
	// from each instance job_phase via template_phase_id. These reads are separate
	// from the academic transcript walk above (which stays scoped to walkIDs) so
	// the deportment phase walk never inflates the criterion roll-up. The strict
	// summary reads carry NO score fallback — the converged raw leaves are STRICT.
	allCatJobIDs := append([]string{}, jobIDs...)
	deportJobCat := map[string]string{}
	deportTemplateIDs := []string{}
	deportTmplSeen := map[string]bool{}
	for _, j := range deportJobs {
		jid, tid := j.GetId(), j.GetJobTemplateId()
		if jid == "" {
			continue
		}
		allCatJobIDs = append(allCatJobIDs, jid)
		deportJobCat[jid] = j.GetJobCategoryId()
		if tid != "" && !deportTmplSeen[tid] {
			deportTmplSeen[tid] = true
			deportTemplateIDs = append(deportTemplateIDs, tid)
		}
	}
	allTemplateIDs := append(append([]string{}, templateIDs...), deportTemplateIDs...)
	templatePhaseCode := fetchTemplatePhaseCodes(ctx, d, allTemplateIDs)
	treeOrder, treeJobByPhase, treePhaseTmpl := fetchPhaseOrders(ctx, d, allCatJobIDs, historical)
	jobOrderCode := map[string]map[int32]string{}
	for pid, jid := range treeJobByPhase {
		code := strings.TrimSpace(templatePhaseCode[treePhaseTmpl[pid]])
		if code == "" {
			continue
		}
		if jobOrderCode[jid] == nil {
			jobOrderCode[jid] = map[int32]string{}
		}
		jobOrderCode[jid][treeOrder[pid]] = code
	}
	treeStrictPhase := fetchPhaseLabelsStrict(ctx, d, allCatJobIDs, treeOrder)
	treeStrictYear := fetchYearLabelsStrict(ctx, d, jobIDs)
	groupCount := 0
	for _, j := range deportJobs {
		if j.GetJobCategoryId() == groupCatID {
			groupCount++
		}
	}

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

	rows := make([]itemRow, 0, len(entries))
	// academicRows pairs each rendered subject row with its job id for the block
	// tree (the itemRow itself carries no id, but the tree needs it for the
	// per-job strict-label and phase-code lookups).
	academicRows := make([]academicTreeRow, 0, len(entries))
	for _, e := range entries {
		sem := semByJob[e.jobID]
		display := cleanSubject(e.name)
		tr := transcripts[e.jobID]
		// crit is derived from the transcript iff this job had ≥1 numeric
		// task_outcome (even an all-zero one); hasMarks distinguishes an all-zero
		// scaffold (present, total="0") from a subject with no task_outcome at all.
		crit, hasMarks := tr.yearCriteria()
		row := itemRow{
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
		// block-layout enrichments (blank-safe when sources are unwired).
		row.ItemTitle = merged.titleFor(display)
		row.StaffLine = staffLine(d.Labels.Student, tr, staffNames)
		row.Criteria, row.Sem1Total, row.Sem2Total = tr.criterionRows(critNames)
		rows = append(rows, row)
		academicRows = append(academicRows, academicTreeRow{jobID: e.jobID, row: row})
	}

	// Formation page (DOCX v2): the deportment-category tables. Additive to the
	// academic subject transcript; empty when the tier has no category filter or the
	// student has no deportment data.
	formationGroups := collectFormationGroups(ctx, d, deportJobs, historical)

	// Per-item rating rows + group rating (block layout).
	itemRatings, groupPhase1, groupPhase2 := buildItemRatings(conduct, merged)
	// Canonical conduct row set (rotation pairs merged, non-enrolled suppressed) —
	// the SAME projection reused for the block-layout subject-deportment .jobs[]
	// loop, so the block card can never regress to split half-rows / placeholders.
	conductRows := deportRows(conduct, merged)

	// Group lead = the modal task assignee on the group-category job. This keeps
	// feeding the FROZEN v1/v2 "adviser" key (its historical first-job behavior).
	groupLead := ""
	if groupJob != nil {
		groupLead = groupLeadName(transcripts[groupJob.GetId()], staffNames)
	}
	// Block-layout root alias (lead_staff_name_display, used on the cover/headers)
	// projects ONLY when the group category has EXACTLY one job — matching the
	// nested singleton rule below. With a corrupt 2+ multiplicity the first-picked
	// groupJob is arbitrary, so its adviser must NOT leak into the block root alias
	// (the nested singleton already blanks). Cardinality is counted (groupCount)
	// before this gate. The frozen v1/v2 "adviser" field is left untouched.
	leadStaffDisplay := ""
	if groupCount == 1 {
		leadStaffDisplay = groupLead
	}

	// Converged generic block-layout tree (job_categories.<code>.jobs[] + the
	// group-category singleton projection). Blank-guarded against the manifest in
	// buildReportCardData. Nil-safe throughout — every unwired source blanks.
	jobCategories := buildJobCategoriesTree(ctx, d, treeInputs{
		cats:         cats,
		academicCat:  catID,
		academic:     academicRows,
		deportJobs:   deportJobCat,
		deportRows:   conductRows,
		jobOrderCode: jobOrderCode,
		strictPhase:  treeStrictPhase,
		groupJob:     groupJob,
		groupCatID:   groupCatID,
		groupCount:   groupCount,
		groupLead:    groupLead,
		historical:   historical,
	}, treeStrictYear)

	name, ay := sectionParts(group.GetName())
	grade, section := gradeSection(name)
	// The v2 display period prefers the price_schedule name (live format
	// "2025-2026", active+inactive two-pass); the v1 `academic_year` key KEEPS
	// the section-derived value — an operator-bound v1 template must not see
	// its data contract change (M4).
	displayPeriod := ay
	if psPeriod := fetchSchedulePeriod(ctx, d, group.GetPriceScheduleId()); psPeriod != "" {
		displayPeriod = psPeriod
	}
	gradeSectionLine := section
	if grade != "" {
		gradeSectionLine = grade + " - " + section
	}
	// D5 render-gate scope: every active job feeding this card (academic +
	// deportment complement + group). A workflow-entered, data-present, not-
	// fully-published phase under ANY of these blocks the render (409).
	gateJobIDs := append([]string{}, jobIDs...)
	for _, dj := range deportJobs {
		if id := dj.GetId(); id != "" {
			gateJobIDs = append(gateJobIDs, id)
		}
	}
	if groupJob != nil && groupJob.GetId() != "" {
		gateJobIDs = append(gateJobIDs, groupJob.GetId())
	}

	rc := &reportCard{
		DocumentHeaderName:    strings.TrimSpace(d.DocumentHeaderName),
		JobIDs:                gateJobIDs,
		SchedulePeriod:        ay,
		SchedulePeriodDisplay: displayPeriod,
		SchedulePeriodSpaced:  strings.Replace(displayPeriod, "-", " - ", 1),
		ClientName:            studentName(ctx, d, clientID),
		GroupLevel:            grade,
		SectionName:           section,
		GroupLabel:            gradeSectionLine,
		LRN:                   "",
		ClientReference:       fetchClientReference(ctx, d, clientID),
		ClientAttributes:      fetchClientAttributes(ctx, d, clientID),
		GroupLead:             groupLead,
		LeadStaffDisplay:      leadStaffDisplay,
		Subjects:              rows,
		PriceScheduleID:       group.GetPriceScheduleId(),
		FormationGroups:       formationGroups,
		ItemRatings:           itemRatings,
		GroupRatingPhase1:     groupPhase1,
		GroupRatingPhase2:     groupPhase2,
		JobCategories:         jobCategories,
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

// fetchPhaseOrders returns three maps keyed by job_phase id: phase_order (for the
// Sem 1 / Sem 2 column mapping), the owning job id (for the per-criterion
// task_outcome roll-up, which walks phase → task → outcome back to the subject),
// and the template_phase id (for the phase-CODE lookup the block-layout tree keys
// its per-phase leaves by, via job_phase.template_phase_id → job_template_phase.code).
func fetchPhaseOrders(ctx context.Context, d *Deps, jobIDs []string, historical bool) (order map[string]int32, jobByPhase, phaseTemplate map[string]string) {
	order = map[string]int32{}
	jobByPhase = map[string]string{}
	phaseTemplate = map[string]string{}
	if d.ListJobPhases == nil || len(jobIDs) == 0 {
		return order, jobByPhase, phaseTemplate
	}
	for start := 0; start < len(jobIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(jobIDs) {
			end = len(jobIDs)
		}
		// The generic adapter defaults to active=true unless an explicit boolean
		// filter is supplied. A past-AY card's job_phase rows are inactive, so a
		// second inactive-admitting pass is added when historical (mirroring
		// fetchJobs / fetchTemplateNames); otherwise every historical phase is
		// silently dropped and the per-phase leaves render blank. The result maps
		// are keyed by phase id, so the active + inactive passes never double-count.
		filterSets := [][]*commonpb.TypedFilter{{listIn("job_id", jobIDs[start:end])}}
		if historical {
			filterSets = append(filterSets, []*commonpb.TypedFilter{listIn("job_id", jobIDs[start:end]), boolEq("active", false)})
		}
		for _, filters := range filterSets {
			resp, err := d.ListJobPhases(ctx, &jobphasepb.ListJobPhasesRequest{
				Filters: &commonpb.FilterRequest{Filters: filters},
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
					if tp := strings.TrimSpace(p.GetTemplatePhaseId()); tp != "" {
						phaseTemplate[id] = tp
					}
				}
			}
		}
	}
	return order, jobByPhase, phaseTemplate
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
			row.Phase1 = fmtNum(v)
			sums[1] += v
			has[1] = true
		}
		if v, ok := t.marks[cid][2]; ok {
			row.Phase2 = fmtNum(v)
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
// other-section grades never accumulate in), resolves the LATEST active outcome
// per (job_task, criterion) (recorded_date DESC, id DESC) and THEN takes the MAX
// across tasks per (job, period, criterion) — D1 layering — and resolves the
// criterion order via
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

	// task_outcome → per (job, period, criterion) mark. D1 LAYERING: canonical
	// cell resolution is the LATEST active outcome per (job_task, criterion)
	// (recorded_date DESC, id DESC — matrix-canonical), and ONLY THEN the MAX
	// across tasks within a (job, period, criterion). This fixes the stale-higher-
	// revision hazard (a newer correction on the same task now supersedes an older
	// higher value) while preserving the locked MAX-across-tasks render contract.
	type cellKey struct{ taskID, critID string }
	type cellAcc struct {
		recorded int64
		id       string
		value    float64
		tmplTask string
	}
	latest := map[cellKey]cellAcc{}
	// jobCritTmplTasks retains EVERY template task that binds a (job, criterion): a
	// criterion can appear on more than one template task. The fold below iterates
	// `latest` in randomized map order, so a single overwrite must NOT decide the
	// criterion's sequence source — collect all candidates and resolve them
	// deterministically after the sequence read (D1 ordering fix).
	jobCritTmplTasks := map[string]map[string]map[string]bool{} // jobID → critID → set(templateTaskID)
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
				tid := t.GetJobTaskId()
				cid := t.GetCriteriaVersionId()
				if taskJob[tid] == "" || cid == "" {
					continue
				}
				k := cellKey{taskID: tid, critID: cid}
				cand := cellAcc{
					recorded: t.GetRecordedDate(),
					id:       t.GetId(),
					value:    t.GetNumericValue(),
					tmplTask: taskTmplTask[tid],
				}
				if cur, ok := latest[k]; !ok ||
					cand.recorded > cur.recorded ||
					(cand.recorded == cur.recorded && cand.id > cur.id) {
					latest[k] = cand
				}
			}
			if len(data) < pageLimit {
				break
			}
		}
	}
	// Fold the latest-per-(task,criterion) cells into per-(job,period,criterion)
	// marks, taking the MAX across tasks. Collect every contributing template task
	// per (job, criterion) for the deterministic sequence resolution below.
	for k, acc := range latest {
		jid := taskJob[k.taskID]
		period := taskPeriod[k.taskID]
		tr := byJob(jid)
		if tr.marks[k.critID] == nil {
			tr.marks[k.critID] = map[int32]float64{}
		}
		if cur, ok := tr.marks[k.critID][period]; !ok || acc.value > cur {
			tr.marks[k.critID][period] = acc.value
		}
		if acc.tmplTask != "" {
			if jobCritTmplTasks[jid] == nil {
				jobCritTmplTasks[jid] = map[string]map[string]bool{}
			}
			if jobCritTmplTasks[jid][k.critID] == nil {
				jobCritTmplTasks[jid][k.critID] = map[string]bool{}
			}
			jobCritTmplTasks[jid][k.critID][acc.tmplTask] = true
			tmplTaskSet[acc.tmplTask] = true
		}
	}

	seq := fetchCriteriaSequence(ctx, d, tmplTaskSet)
	// Resolve each criterion's sequence_order DETERMINISTICALLY: among the template
	// tasks that bind the criterion, pick the lowest sequence_order, tie-broken by
	// the lowest template-task id. The comparison is a total order, so the result is
	// independent of the randomized set-iteration order. Without this a criterion
	// bound by two template tasks with different sequence_order would inherit a
	// randomized order, reshuffling the A–D slots between renders.
	for jid, tr := range out {
		for cid, tts := range jobCritTmplTasks[jid] {
			bestSeq := int32(0)
			bestTT := ""
			for tt := range tts {
				sv, ok := seq[tt][cid]
				if !ok {
					continue
				}
				if bestTT == "" || sv < bestSeq || (sv == bestSeq && tt < bestTT) {
					bestSeq, bestTT = sv, tt
				}
			}
			if bestTT != "" {
				tr.seq[cid] = bestSeq
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
		chunk := ids[start:end]
		resp, err := d.GetStaffListPageData(ctx, &staffpb.GetStaffListPageDataRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{listIn("id", chunk)}},
			// Explicit limit ≥ the chunk size — the adapter's default page (50)
			// would silently drop assignees past the first page.
			Pagination: &commonpb.PaginationRequest{
				Limit:  int32(pageLimit),
				Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: 1}},
			},
		})
		if err != nil {
			log.Printf("report card doc: staff list page data: %v", err)
			continue
		}
		requested := map[string]bool{}
		for _, id := range chunk {
			requested[id] = true
		}
		for _, s := range resp.GetStaffList() {
			id := s.GetId()
			u := s.GetUser()
			// Keep only rows we actually asked for — defense-in-depth against
			// an adapter that ignores the id filter (a first-page-of-everything
			// response must never attribute foreign names).
			if id == "" || u == nil || !requested[id] {
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

// staffLine composes the per-item staff line: one distinct assignee →
// "Teacher: X" (e.g.); two (the rotation pair, phase-1 first) →
// "Teachers: X / Y". Wording comes from the lyngua-backed labels; blank when
// nothing resolves.
func staffLine(labels outcome_summary.PeriodLabels, tr *transcript, names map[string]string) string {
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

// fetchClientAttributes resolves the configured client_attributes.<code> map
// (DocumentOptions.ClientAttributeCodes, e.g. ["lrn","gender"]) for the client.
// Every configured, non-empty, dot-free code is ALWAYS present in the returned
// map (blank when the client has no such attribute) so a
// {{client_attributes.<code>}} placeholder never leaks. Resolution: each code →
// ResolveAttributeIDByCode → ONE client_id-filtered ListClientAttributes pass →
// attribute_id→code→value (first non-empty value wins). Nil-safe: a missing
// closure leaves every configured code blank.
func fetchClientAttributes(ctx context.Context, d *Deps, clientID string) map[string]any {
	out := map[string]any{}
	codeByAttrID := map[string]string{}
	for _, raw := range d.DocOptions.ClientAttributeCodes {
		code := strings.TrimSpace(raw)
		// Codes must be non-empty and dot-free: a dotted code would split into
		// extra dot-path segments in the engine and never address a flat key.
		if code == "" || strings.Contains(code, ".") {
			continue
		}
		out[code] = "" // always-present key (blank until resolved)
		if clientID == "" || d.ResolveAttributeIDByCode == nil || d.ListClientAttributes == nil {
			continue
		}
		attrID, err := d.ResolveAttributeIDByCode(ctx, code)
		if err != nil || strings.TrimSpace(attrID) == "" {
			continue
		}
		codeByAttrID[strings.TrimSpace(attrID)] = code
	}
	if len(codeByAttrID) == 0 || clientID == "" || d.ListClientAttributes == nil {
		return out
	}
	resp, err := d.ListClientAttributes(ctx, &clientattributepb.ListClientAttributesRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("client_id", clientID)}},
	})
	if err != nil {
		log.Printf("report card doc: list client attributes (map): %v", err)
		return out
	}
	for _, ca := range resp.GetData() {
		if !ca.GetActive() {
			continue
		}
		// HIGH#1 client-isolation (defense-in-depth): client_attribute has NO
		// workspace_id column, so its generic list cannot be workspace-scoped — an
		// adapter that ignored the client_id filter (or returned extra rows) could
		// otherwise surface ANOTHER client's attribute value. Re-check the row's own
		// client_id against the requested clientID before accepting its value. The
		// dot-free code guard above still holds.
		if ca.GetClientId() != clientID {
			continue
		}
		code, ok := codeByAttrID[strings.TrimSpace(ca.GetAttributeId())]
		if !ok {
			continue
		}
		if cur, _ := out[code].(string); cur != "" {
			continue
		}
		if v := strings.TrimSpace(ca.GetValue()); v != "" {
			out[code] = v
		}
	}
	return out
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
func isNonEnrolledPlaceholder(row itemRow, hasMarks bool) bool {
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
