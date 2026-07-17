package document

// tree.go — the converged generic job_categories tree (the block-layout data
// contract) plus the manifest-driven blank-guard.
//
// The tree is a proto-grounded map keyed by job_category.code:
//
//	job_categories.<category_code>
//	  jobs[]                                    # one item per job in the category
//	    job_template_name_display
//	    staff_line_display
//	    job_outcome_summary_scaled_label        # STRICT raw label (no score fallback)
//	    job_template_phases.<phase_code>
//	      phase_outcome_summary_scaled_label    # STRICT raw label
//	      task_outcome_numeric_value_total_derived
//	    outcome_criteria[]
//	      outcome_criteria_label_display
//	      job_template_phases.<phase_code>
//	        task_outcome_numeric_value_max_derived
//	  # Singleton projection (root scalars) — emitted ONLY when the configured
//	  # group category has exactly one job for this client; zero/multiple leaves
//	  # the whole subtree blank (manifest-seeded) plus one bounded log line.
//	  lead_staff_name_display
//	  job_template_phases.<phase_code>
//	    phase_outcome_summary_scaled_label
//	    job_template_tasks.<task_code>
//	      task_outcomes.<criteria_code>
//	        numeric_value
//	  task_outcomes.<criteria_code>
//	    numeric_value_total_derived
//
// Every phase/task/criterion code is DATA — derived from the workspace's own
// job_template_phase.code / job_template_task.code / outcome_criteria.code — never
// a literal in this file. The engine leaks unresolved leaves verbatim, so the
// manifest blank-guard (applyBlockManifest) seeds every referenced path before the
// real values overlay.

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// --- manifest blank-guard -------------------------------------------------

// manifestNode is one loop scope in the block manifest: the item-relative scalar
// paths the artifact references and its own nested loops. The root has the same
// shape (its scalars are absolute paths).
type manifestNode struct {
	Scalars []string                `json:"scalars"`
	Loops   map[string]manifestNode `json:"loops"`
}

var (
	blockManifestOnce   sync.Once
	blockManifestParsed manifestNode
)

// loadBlockManifest parses the embedded block manifest exactly once. On a parse
// failure it logs and leaves an empty manifest (the blank-guard degrades to a
// no-op rather than panicking a render).
func loadBlockManifest() manifestNode {
	blockManifestOnce.Do(func() {
		raw := ManifestBlock()
		if len(raw) == 0 {
			return
		}
		if err := json.Unmarshal(raw, &blockManifestParsed); err != nil {
			log.Printf("report card doc: block manifest parse: %v", err)
			blockManifestParsed = manifestNode{}
		}
	})
	return blockManifestParsed
}

// applyBlockManifest seeds EVERY manifest-referenced path in data with a blank
// leaf (scalars) or an empty list (loops), building nested maps as needed, then
// recurses into each real loop item so per-item scalars are seeded too. Real
// values already present are never clobbered — seeding only fills what is absent.
// This is the LEAK-LAW guard: the engine emits any unresolved {{leaf}} verbatim,
// so every path the artifact can reference must resolve (blank) even when the
// data omits it.
func applyBlockManifest(data map[string]any) {
	conformToManifest(data, loadBlockManifest())
}

func conformToManifest(m map[string]any, node manifestNode) {
	for _, path := range node.Scalars {
		ensureBlankLeaf(m, strings.Split(path, "."))
	}
	for loopPath, child := range node.Loops {
		list := ensureList(m, strings.Split(loopPath, "."))
		for _, it := range list {
			if im, ok := it.(map[string]any); ok {
				conformToManifest(im, child)
			}
		}
	}
}

// ensureBlankLeaf walks segs from m, creating intermediate maps, and sets a blank
// string leaf iff the final key is absent (never overwrites a real value).
func ensureBlankLeaf(m map[string]any, segs []string) {
	if len(segs) == 0 {
		return
	}
	if len(segs) == 1 {
		if _, ok := m[segs[0]]; !ok {
			m[segs[0]] = ""
		}
		return
	}
	child, ok := m[segs[0]].(map[string]any)
	if !ok {
		child = map[string]any{}
		m[segs[0]] = child
	}
	ensureBlankLeaf(child, segs[1:])
}

// ensureList walks segs from m (creating intermediate maps) and guarantees the
// final key holds a []any, seeding an empty list when absent. Returns the list so
// callers can recurse into real items.
func ensureList(m map[string]any, segs []string) []any {
	if len(segs) == 0 {
		return nil
	}
	cur := m
	for _, s := range segs[:len(segs)-1] {
		child, ok := cur[s].(map[string]any)
		if !ok {
			child = map[string]any{}
			cur[s] = child
		}
		cur = child
	}
	last := segs[len(segs)-1]
	lst, ok := cur[last].([]any)
	if !ok {
		lst = []any{}
		cur[last] = lst
	}
	return lst
}

// --- strict summary reads (NEW tree leaves; no score fallback) ------------

// fetchYearLabelsStrict returns job id → job_outcome_summary.scaled_label with NO
// numeric-score fallback (the converged year-final leaf is STRICT: blank when the
// stored label is empty). Distinct from fetchYearLabels, which keeps the v1/v2
// display fallback. Nil-safe.
func fetchYearLabelsStrict(ctx context.Context, d *Deps, jobIDs []string) map[string]string {
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
			log.Printf("report card doc: list job outcome summaries (strict): %v", err)
			continue
		}
		for _, s := range resp.GetData() {
			if !s.GetActive() || s.GetJobId() == "" {
				continue
			}
			if lbl := strings.TrimSpace(s.GetScaledLabel()); lbl != "" {
				out[s.GetJobId()] = lbl
			}
		}
	}
	return out
}

// fetchPhaseLabelsStrict returns job id → phase_order → phase_outcome_summary
// .scaled_label with NO numeric-score fallback (the converged phase leaf is
// STRICT). Distinct from fetchSemesterLabels, which keeps the v1/v2 fallback.
// Nil-safe.
func fetchPhaseLabelsStrict(ctx context.Context, d *Deps, jobIDs []string, phaseOrder map[string]int32) map[string]map[int32]string {
	out := map[string]map[int32]string{}
	if d.ListPhaseOutcomeSummarysByJob == nil {
		return out
	}
	for _, jid := range jobIDs {
		resp, err := d.ListPhaseOutcomeSummarysByJob(ctx, &phasesumpb.ListPhaseOutcomeSummarysByJobRequest{JobId: jid})
		if err != nil {
			log.Printf("report card doc: list phase summaries by job (strict): %v", err)
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
			lbl := strings.TrimSpace(s.GetScaledLabel())
			if lbl == "" {
				continue
			}
			if out[jid] == nil {
				out[jid] = map[int32]string{}
			}
			out[jid][ord] = lbl
		}
	}
	return out
}

// strictPhaseLabel is a nil-safe read of the strict phase-label map.
func strictPhaseLabel(m map[string]map[int32]string, jobID string, order int32) string {
	if m == nil {
		return ""
	}
	if byOrder, ok := m[jobID]; ok {
		return byOrder[order]
	}
	return ""
}

// fetchTemplatePhaseCodes returns job_template_phase.id → code for the given
// template ids, via the ownership-safe ListByJobTemplate read (which projects the
// code column, Wave A). The document keys its tree by phase CODE, reached from
// each instance job_phase through job_phase.template_phase_id. Nil-safe: a missing
// closure yields an empty map and every phase key resolves blank.
func fetchTemplatePhaseCodes(ctx context.Context, d *Deps, templateIDs []string) map[string]string {
	out := map[string]string{}
	if d.ListJobTemplatePhasesByTemplate == nil || len(templateIDs) == 0 {
		return out
	}
	seen := map[string]bool{}
	for _, tid := range templateIDs {
		if tid == "" || seen[tid] {
			continue
		}
		seen[tid] = true
		resp, err := d.ListJobTemplatePhasesByTemplate(ctx, &jobtemplatephasepb.ListByJobTemplateRequest{JobTemplateId: tid})
		if err != nil {
			log.Printf("report card doc: list template phases by template: %v", err)
			continue
		}
		for _, p := range resp.GetJobTemplatePhases() {
			id := p.GetId()
			code := strings.TrimSpace(p.GetCode())
			if id != "" && code != "" {
				out[id] = code
			}
		}
	}
	return out
}

// --- the tree builder -----------------------------------------------------

// treeInputs is the assembled fetch bundle the tree builder consumes. Everything
// here is derived by collectCard from the membership-gated job walk — no ids come
// from the request, workspace scope is applied in the adapters.
type treeInputs struct {
	cats        map[string]catInfo // category id → {name, order, code}
	academicCat string             // the resolved academic (CategoryFilter) category id

	academic []academicTreeRow // one per rendered academic subject (post-suppression)

	deportJobs map[string]string // deport job id → job_category id (category resolution)
	deportRows []deportRow       // canonical conduct rows (rotation pairs merged, non-enrolled suppressed)

	jobOrderCode map[string]map[int32]string // job id → phase_order → phase_code
	strictPhase  map[string]map[int32]string // job id → phase_order → strict phase label

	groupJob   *jobpb.Job // the configured group-category job (nil when absent)
	groupCatID string     // the group category id
	groupCount int        // number of jobs in the group category (singleton rule)
	groupLead  string     // resolved group-lead ("Adviser") display name
}

// academicTreeRow pairs a rendered academic subject row with its job id (needed
// for the per-job strict label + phase-code lookups the itemRow itself does not
// carry).
type academicTreeRow struct {
	jobID string
	row   itemRow
}

// buildJobCategoriesTree assembles the whole job_categories subtree (map keyed by
// category code). Emitted under the "job_categories" root key by
// buildReportCardData and blank-guarded by the manifest afterwards.
func buildJobCategoriesTree(ctx context.Context, d *Deps, in treeInputs, strictYear map[string]string) map[string]any {
	tree := map[string]any{}
	catMap := func(code string) map[string]any {
		if code == "" {
			return nil
		}
		m, ok := tree[code].(map[string]any)
		if !ok {
			m = map[string]any{}
			tree[code] = m
		}
		return m
	}

	// Academic category .jobs[] — rich items reusing the rendered subject rows.
	academicCode := strings.TrimSpace(in.cats[in.academicCat].code)
	if academicCode != "" && len(in.academic) > 0 {
		list := make([]any, 0, len(in.academic))
		for _, a := range in.academic {
			list = append(list, academicJobTreeItem(a.row, in.jobOrderCode[a.jobID], in.strictPhase[a.jobID], strictYear[a.jobID]))
		}
		if m := catMap(academicCode); m != nil {
			m["jobs"] = list
		}
	}

	// Non-academic categories .jobs[] — the CANONICAL conduct projection: rotation
	// pairs merged into one row (each strand contributing its own period's value),
	// non-enrolled placeholder strands suppressed. This is the SAME row set as the
	// v2 conduct table (deportRows), so the two surfaces never drift into split
	// half-rows or resurrected placeholder rows. Grouped by the period-1 strand's
	// job_category code; per-period leaves keyed by phase code with STRICT labels.
	byCatCode := map[string][]any{}
	for _, dr := range in.deportRows {
		code := strings.TrimSpace(in.cats[in.deportJobs[dr.sem1Job]].code)
		if code == "" {
			continue
		}
		byCatCode[code] = append(byCatCode[code], deportRowTreeItem(dr, in.jobOrderCode, in.strictPhase))
	}
	for code, list := range byCatCode {
		if m := catMap(code); m != nil {
			m["jobs"] = list
		}
	}

	// Singleton projection for the configured group category — exactly-one-job
	// rule (Q-T4): zero or 2+ jobs leaves the subtree blank (manifest-seeded) plus
	// one bounded diagnostic.
	groupCode := strings.TrimSpace(in.cats[in.groupCatID].code)
	if groupCode != "" {
		switch {
		case in.groupCount == 1 && in.groupJob != nil:
			singleton := buildSingletonProjection(ctx, d, in.groupJob, in.groupLead,
				in.jobOrderCode[in.groupJob.GetId()], in.strictPhase[in.groupJob.GetId()])
			if m := catMap(groupCode); m != nil {
				for k, v := range singleton {
					m[k] = v
				}
			}
		default:
			log.Printf("report card doc: singleton category %q has %d job(s) (want exactly 1); projection left blank", groupCode, in.groupCount)
		}
	}

	return tree
}

// academicJobTreeItem builds one .jobs[] item for the academic category from a
// rendered subject row. Per-phase leaves are keyed by phase CODE (orderCode maps
// phase_order → code); the year-final and phase labels are STRICT. Missing leaves
// are left absent — the manifest blank-guard seeds them.
func academicJobTreeItem(row itemRow, orderCode map[int32]string, strictPhase map[int32]string, strictYear string) map[string]any {
	name := orBlank(row.ItemTitle)
	if name == "" {
		name = orBlank(row.Name)
	}
	phases := map[string]any{}
	setPhase := func(order int32, total string) {
		code := codeForOrder(orderCode, order)
		if code == "" {
			return
		}
		phases[code] = map[string]any{
			"phase_outcome_summary_scaled_label":       phaseLabelAt(strictPhase, order),
			"task_outcome_numeric_value_total_derived": orBlank(total),
		}
	}
	setPhase(1, row.Sem1Total)
	setPhase(2, row.Sem2Total)

	criteria := make([]any, 0, len(row.Criteria))
	for _, c := range row.Criteria {
		critPhases := map[string]any{}
		setCritPhase := func(order int32, mark string) {
			code := codeForOrder(orderCode, order)
			if code == "" {
				return
			}
			critPhases[code] = map[string]any{"task_outcome_numeric_value_max_derived": orBlank(mark)}
		}
		setCritPhase(1, c.Phase1)
		setCritPhase(2, c.Phase2)
		criteria = append(criteria, map[string]any{
			"outcome_criteria_label_display": orBlank(c.Label),
			"job_template_phases":            critPhases,
		})
	}

	return map[string]any{
		"job_template_name_display":        name,
		"staff_line_display":               orBlank(row.StaffLine),
		"job_outcome_summary_scaled_label": orBlank(strictYear),
		"job_template_phases":              phases,
		"outcome_criteria":                 criteria,
	}
}

// deportRowTreeItem builds one non-academic .jobs[] item from a canonical conduct
// row: the merged/solo display title plus STRICT per-phase labels keyed by phase
// code. A rotation pair reads period 1 from sem1Job and period 2 from sem2Job
// (each strand's own phase code + its own strict summary); a solo strand reads
// both periods from the one job. A period whose enrollment gate is off
// (showSemN=false) is left absent — the manifest blank-guard seeds it.
func deportRowTreeItem(dr deportRow, jobOrderCode map[string]map[int32]string, strictPhase map[string]map[int32]string) map[string]any {
	phases := map[string]any{}
	setPhase := func(jobID string, order int32, show bool) {
		if !show || jobID == "" {
			return
		}
		code := codeForOrder(jobOrderCode[jobID], order)
		if code == "" {
			return
		}
		phases[code] = map[string]any{
			"phase_outcome_summary_scaled_label": phaseLabelAt(strictPhase[jobID], order),
		}
	}
	setPhase(dr.sem1Job, 1, dr.showSem1)
	setPhase(dr.sem2Job, 2, dr.showSem2)
	return map[string]any{
		"job_template_name_display": orBlank(dr.title),
		"job_template_phases":       phases,
	}
}

// buildSingletonProjection builds the group-category singleton root scalars: the
// lead-staff alias, per-phase strict labels, the per-(task,criterion) coded cells
// (attendance grid), and the per-criterion totals (Σ across phases of RECORDED
// values only — absence stays blank, a recorded 0 renders "0").
func buildSingletonProjection(ctx context.Context, d *Deps, groupJob *jobpb.Job, groupLead string, orderCode map[int32]string, strictPhase map[int32]string) map[string]any {
	out := map[string]any{
		"lead_staff_name_display": orBlank(groupLead),
	}

	phases := map[string]any{}
	phaseMap := func(code string) map[string]any {
		if code == "" {
			return nil
		}
		m, ok := phases[code].(map[string]any)
		if !ok {
			m = map[string]any{}
			phases[code] = m
		}
		return m
	}
	// Per-phase strict labels (keyed by phase code).
	for _, order := range []int32{1, 2} {
		code := codeForOrder(orderCode, order)
		if m := phaseMap(code); m != nil {
			m["phase_outcome_summary_scaled_label"] = phaseLabelAt(strictPhase, order)
		}
	}

	// Coded cells + per-criterion totals from the ownership-joined latest-cell read.
	totals := map[string]float64{}
	hasTotal := map[string]bool{}
	if d.ListCodedTaskOutcomeValuesByJob != nil && groupJob.GetId() != "" {
		resp, err := d.ListCodedTaskOutcomeValuesByJob(ctx, &taskoutcomepb.ListCodedTaskOutcomeValuesByJobRequest{
			JobIds: []string{groupJob.GetId()},
		})
		if err != nil {
			log.Printf("report card doc: list coded task outcome values: %v", err)
		} else {
			for _, v := range resp.GetValues() {
				pc := strings.TrimSpace(v.GetPhaseCode())
				tc := strings.TrimSpace(v.GetTaskCode())
				cc := strings.TrimSpace(v.GetCriteriaCode())
				// Skip rows with any empty code — they cannot address a cell path.
				if pc == "" || tc == "" || cc == "" {
					continue
				}
				pm := phaseMap(pc)
				tasks, ok := pm["job_template_tasks"].(map[string]any)
				if !ok {
					tasks = map[string]any{}
					pm["job_template_tasks"] = tasks
				}
				task, ok := tasks[tc].(map[string]any)
				if !ok {
					task = map[string]any{}
					tasks[tc] = task
				}
				outcomes, ok := task["task_outcomes"].(map[string]any)
				if !ok {
					outcomes = map[string]any{}
					task["task_outcomes"] = outcomes
				}
				cell, ok := outcomes[cc].(map[string]any)
				if !ok {
					cell = map[string]any{}
					outcomes[cc] = cell
				}
				cell["numeric_value"] = codedCellValue(v)
				// Totals: sum only RECORDED values (nil = no outcome ≠ 0).
				if v.NumericValue != nil {
					totals[cc] += v.GetNumericValue()
					hasTotal[cc] = true
				}
			}
		}
	}
	out["job_template_phases"] = phases

	taskOutcomes := map[string]any{}
	for cc := range hasTotal {
		taskOutcomes[cc] = map[string]any{
			"numeric_value_total_derived": fmtNum(totals[cc]),
		}
	}
	out["task_outcomes"] = taskOutcomes
	return out
}

// codedCellValue formats a coded cell: blank when no outcome was recorded (nil),
// the numeric value otherwise (a recorded 0 renders "0", integers without a
// trailing decimal).
func codedCellValue(v *taskoutcomepb.CodedTaskOutcomeValue) string {
	if v == nil || v.NumericValue == nil {
		return ""
	}
	return fmtNum(v.GetNumericValue())
}

// codeForOrder is a nil-safe read of the phase_order → phase_code map.
func codeForOrder(orderCode map[int32]string, order int32) string {
	if orderCode == nil {
		return ""
	}
	return strings.TrimSpace(orderCode[order])
}

// phaseLabelAt is a nil-safe read of a phase_order → label map.
func phaseLabelAt(byOrder map[int32]string, order int32) string {
	if byOrder == nil {
		return ""
	}
	return orBlank(byOrder[order])
}

