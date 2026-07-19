package document

import (
	"context"
	"strings"
	"testing"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

func f64p(f float64) *float64 { return &f }
func i64p(i int64) *int64     { return &i }

// resolvePath walks a dotted path through nested map[string]any. Returns the leaf
// value and whether every hop resolved (a missing hop ⇒ ok=false).
func resolvePath(m map[string]any, path string) (any, bool) {
	segs := strings.Split(path, ".")
	var cur any = m
	for _, s := range segs {
		mp, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := mp[s]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func mustBlank(t *testing.T, m map[string]any, path string) {
	t.Helper()
	v, ok := resolvePath(m, path)
	if !ok {
		t.Fatalf("manifest path %q did not resolve (would leak verbatim)", path)
	}
	if s, isStr := v.(string); !isStr || s != "" {
		t.Fatalf("manifest path %q = %#v, want blank string", path, v)
	}
}

func codedFn(values ...*taskoutcomepb.CodedTaskOutcomeValue) func(context.Context, *taskoutcomepb.ListCodedTaskOutcomeValuesByJobRequest) (*taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse, error) {
	return func(context.Context, *taskoutcomepb.ListCodedTaskOutcomeValuesByJobRequest) (*taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse, error) {
		return &taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse{Values: values}, nil
	}
}

// --- manifest blank-seed --------------------------------------------------

// An empty card still resolves EVERY manifest-referenced path to a blank leaf (or
// empty loop) — the engine leaks any unresolved {{leaf}} verbatim, so this is the
// leak-law guard.
func TestBuildReportCardData_ManifestBlankSeed(t *testing.T) {
	data := buildReportCardData(reportCard{})

	// A deep singleton scalar path with no data resolves to "" (not absent).
	mustBlank(t, data, "job_categories.homeroom_deportment.job_template_phases.s1.job_template_tasks.m07.task_outcomes.days_present.numeric_value")
	mustBlank(t, data, "job_categories.homeroom_deportment.job_template_phases.s2.phase_outcome_summary_scaled_label")
	mustBlank(t, data, "job_categories.homeroom_deportment.task_outcomes.times_tardy.numeric_value_total_derived")
	mustBlank(t, data, "client_attributes.lrn")
	mustBlank(t, data, "lead_staff_name_display")

	// Every top-level manifest loop resolves to an (empty) []any, never a leak.
	for _, p := range []string{"job_categories.academic.jobs", "job_categories.subject_deportment.jobs"} {
		v, ok := resolvePath(data, p)
		if !ok {
			t.Fatalf("loop path %q did not resolve", p)
		}
		lst, isList := v.([]any)
		if !isList || len(lst) != 0 {
			t.Fatalf("loop path %q = %#v, want empty []any", p, v)
		}
	}
}

// A real overlay wins over the seed, and manifest paths the data omits still get
// blanked INSIDE a real loop item (per-item conform).
func TestBuildReportCardData_ConformSeedsInsideLoopItems(t *testing.T) {
	rc := reportCard{
		JobCategories: map[string]any{
			"academic": map[string]any{
				"jobs": []any{
					// A deliberately-sparse item: only the name is set. Every other
					// manifest item-scalar must be seeded blank by the conform pass.
					map[string]any{"job_template_name_display": "Mathematics"},
				},
			},
		},
	}
	data := buildReportCardData(rc)
	jobs, _ := resolvePath(data, "job_categories.academic.jobs")
	item := jobs.([]any)[0].(map[string]any)
	if item["job_template_name_display"] != "Mathematics" {
		t.Fatalf("real overlay lost: %#v", item["job_template_name_display"])
	}
	// Item-relative manifest scalar with no data → blank, present.
	mustBlank(t, item, "staff_line_display")
	mustBlank(t, item, "job_template_phases.s1.task_outcome_numeric_value_total_derived")
	// The nested outcome_criteria loop is seeded to an empty list.
	oc, ok := item["outcome_criteria"].([]any)
	if !ok || len(oc) != 0 {
		t.Fatalf("nested loop outcome_criteria = %#v, want empty []any", item["outcome_criteria"])
	}
}

// --- tree shape -----------------------------------------------------------

func TestBuildJobCategoriesTree_Shape(t *testing.T) {
	d := &Deps{
		ListCodedTaskOutcomeValuesByJob: codedFn(
			&taskoutcomepb.CodedTaskOutcomeValue{PhaseCode: "s1", TaskCode: "m07", CriteriaCode: "days_present", NumericValue: f64p(18)},
			&taskoutcomepb.CodedTaskOutcomeValue{PhaseCode: "s1", TaskCode: "m07", CriteriaCode: "school_days", NumericValue: f64p(20)},
			&taskoutcomepb.CodedTaskOutcomeValue{PhaseCode: "s2", TaskCode: "m01", CriteriaCode: "days_present"}, // nil → no outcome
		),
	}
	in := treeInputs{
		cats: map[string]catInfo{
			"cat-acad": {name: "Academic", code: "academic"},
			"cat-subj": {name: "Subject Deportment", code: "subject_deportment"},
			"cat-home": {name: "Homeroom Deportment", code: "homeroom_deportment"},
		},
		academicCat: "cat-acad",
		academic: []academicTreeRow{
			{jobID: "jA", row: itemRow{
				Name: "Mathematics", StaffLine: "Teacher: X",
				Sem1Total: "20", Sem2Total: "18",
				Criteria: []criterionRow{{Label: "A - Investigating", Phase1: "7", Phase2: "4"}},
			}},
		},
		deportJobs: map[string]string{"jS": "cat-subj"},
		deportRows: []deportRow{
			{title: "Arts: Music", sem1Job: "jS", sem2Job: "jS", showSem1: true, showSem2: true},
		},
		jobOrderCode: map[string]map[int32]string{
			"jA": {1: "s1", 2: "s2"},
			"jS": {1: "s1", 2: "s2"},
			"jH": {1: "s1", 2: "s2"},
		},
		strictPhase: map[string]map[int32]string{
			"jA": {1: "6", 2: "7"},
			"jS": {1: "90", 2: "88"},
			"jH": {1: "A", 2: "B"},
		},
		groupJob:   &jobpb.Job{Id: "jH"},
		groupCatID: "cat-home",
		groupCount: 1,
		groupLead:  "Adviser Y",
	}
	tree := buildJobCategoriesTree(context.Background(), d, in, map[string]string{"jA": "7"})

	// Academic .jobs[] item.
	if v, _ := resolvePath(tree, "academic.jobs"); len(v.([]any)) != 1 {
		t.Fatalf("academic jobs = %#v", v)
	}
	item := tree["academic"].(map[string]any)["jobs"].([]any)[0].(map[string]any)
	assertLeaf(t, item, "job_template_name_display", "Mathematics")
	assertLeaf(t, item, "staff_line_display", "Teacher: X")
	assertLeaf(t, item, "job_outcome_summary_scaled_label", "7") // STRICT
	assertLeaf(t, item, "job_template_phases.s1.phase_outcome_summary_scaled_label", "6")
	assertLeaf(t, item, "job_template_phases.s1.task_outcome_numeric_value_total_derived", "20")
	assertLeaf(t, item, "job_template_phases.s2.task_outcome_numeric_value_total_derived", "18")
	crit := item["outcome_criteria"].([]any)[0].(map[string]any)
	assertLeaf(t, crit, "outcome_criteria_label_display", "A - Investigating")
	assertLeaf(t, crit, "job_template_phases.s1.task_outcome_numeric_value_max_derived", "7")
	assertLeaf(t, crit, "job_template_phases.s2.task_outcome_numeric_value_max_derived", "4")

	// Subject-deportment .jobs[] item — name + strict phase labels only.
	sub := tree["subject_deportment"].(map[string]any)["jobs"].([]any)[0].(map[string]any)
	assertLeaf(t, sub, "job_template_name_display", "Arts: Music")
	assertLeaf(t, sub, "job_template_phases.s1.phase_outcome_summary_scaled_label", "90")

	// Homeroom singleton projection (root scalars under the category).
	home := tree["homeroom_deportment"].(map[string]any)
	assertLeaf(t, home, "lead_staff_name_display", "Adviser Y")
	assertLeaf(t, home, "job_template_phases.s1.phase_outcome_summary_scaled_label", "A")
	assertLeaf(t, home, "job_template_phases.s1.job_template_tasks.m07.task_outcomes.days_present.numeric_value", "18")
	assertLeaf(t, home, "job_template_phases.s1.job_template_tasks.m07.task_outcomes.school_days.numeric_value", "20")
	assertLeaf(t, home, "task_outcomes.days_present.numeric_value_total_derived", "18")
	assertLeaf(t, home, "task_outcomes.school_days.numeric_value_total_derived", "20")
}

// --- singleton rule (0 / 1 / 2 jobs) --------------------------------------

func TestBuildJobCategoriesTree_SingletonRule(t *testing.T) {
	base := func(count int, gj *jobpb.Job) treeInputs {
		return treeInputs{
			cats:         map[string]catInfo{"cat-home": {name: "Homeroom", code: "homeroom_deportment"}},
			jobOrderCode: map[string]map[int32]string{"jH": {1: "s1", 2: "s2"}},
			strictPhase:  map[string]map[int32]string{"jH": {1: "A"}},
			groupJob:     gj,
			groupCatID:   "cat-home",
			groupCount:   count,
			groupLead:    "Adviser Y",
		}
	}
	d := &Deps{ListCodedTaskOutcomeValuesByJob: codedFn()}

	// Exactly one → singleton projected.
	one := buildJobCategoriesTree(context.Background(), d, base(1, &jobpb.Job{Id: "jH"}), nil)
	if v, ok := resolvePath(one, "homeroom_deportment.lead_staff_name_display"); !ok || v != "Adviser Y" {
		t.Fatalf("one-job singleton must project lead, got %#v ok=%v", v, ok)
	}

	// Zero → whole subtree blank (no singleton scalar emitted).
	zero := buildJobCategoriesTree(context.Background(), d, base(0, nil), nil)
	if _, ok := resolvePath(zero, "homeroom_deportment.lead_staff_name_display"); ok {
		t.Fatalf("zero-job category must NOT project a singleton")
	}

	// Two → whole subtree blank (corrupt multiplicity is never first-wins).
	two := buildJobCategoriesTree(context.Background(), d, base(2, &jobpb.Job{Id: "jH"}), nil)
	if _, ok := resolvePath(two, "homeroom_deportment.lead_staff_name_display"); ok {
		t.Fatalf("two-job category must NOT project a singleton (blank + log)")
	}
}

// --- absence vs recorded zero ---------------------------------------------

func TestBuildSingletonProjection_AbsenceVsZero(t *testing.T) {
	d := &Deps{
		ListCodedTaskOutcomeValuesByJob: codedFn(
			&taskoutcomepb.CodedTaskOutcomeValue{PhaseCode: "s1", TaskCode: "m07", CriteriaCode: "days_present"},                         // nil → blank
			&taskoutcomepb.CodedTaskOutcomeValue{PhaseCode: "s1", TaskCode: "m08", CriteriaCode: "days_present", NumericValue: f64p(0)},  // recorded 0 → "0"
			&taskoutcomepb.CodedTaskOutcomeValue{PhaseCode: "s1", TaskCode: "m09", CriteriaCode: "days_present", NumericValue: f64p(18)}, // 18
			&taskoutcomepb.CodedTaskOutcomeValue{PhaseCode: "s1", TaskCode: "m07", CriteriaCode: "times_tardy"},                          // nil only → blank total
		),
	}
	out := buildSingletonProjection(context.Background(), d, &jobpb.Job{Id: "jH"}, "Adviser",
		map[int32]string{1: "s1", 2: "s2"}, map[int32]string{1: "A"}, false)

	assertLeaf(t, out, "job_template_phases.s1.job_template_tasks.m07.task_outcomes.days_present.numeric_value", "")  // absence → blank
	assertLeaf(t, out, "job_template_phases.s1.job_template_tasks.m08.task_outcomes.days_present.numeric_value", "0") // recorded 0
	assertLeaf(t, out, "job_template_phases.s1.job_template_tasks.m09.task_outcomes.days_present.numeric_value", "18")
	// Total = Σ recorded only (0 + 18 = 18); the nil m07 does not count.
	assertLeaf(t, out, "task_outcomes.days_present.numeric_value_total_derived", "18")
	// times_tardy has only a nil cell → NO total emitted (blank when zero contributors).
	if _, ok := resolvePath(out, "task_outcomes.times_tardy.numeric_value_total_derived"); ok {
		t.Fatalf("times_tardy total must stay blank (no recorded contributor), got a value")
	}
}

// --- Q6 historical-AY coded-path routing ----------------------------------

// A past-AY (historical) card must resolve its coded/attendance cells through the
// inactive-admitting historical reader, NOT the active-only live reader. The live
// reader here returns EMPTY (as it would against a past card's inactive ancestry);
// only the historical reader carries values. historical=true must therefore route
// to the historical reader and populate the grid — proving the fix un-blanks past
// cards without touching the live path.
func TestBuildSingletonProjection_HistoricalRouting(t *testing.T) {
	liveCalled, histCalled := false, false
	d := &Deps{
		ListCodedTaskOutcomeValuesByJob: func(context.Context, *taskoutcomepb.ListCodedTaskOutcomeValuesByJobRequest) (*taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse, error) {
			liveCalled = true
			return &taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse{}, nil // live path is blank for a past card
		},
		ListCodedTaskOutcomeValuesByJobHistorical: func(context.Context, *taskoutcomepb.ListCodedTaskOutcomeValuesByJobRequest) (*taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse, error) {
			histCalled = true
			return &taskoutcomepb.ListCodedTaskOutcomeValuesByJobResponse{Values: []*taskoutcomepb.CodedTaskOutcomeValue{
				{PhaseCode: "s1", TaskCode: "m07", CriteriaCode: "days_present", NumericValue: f64p(16)},
				{PhaseCode: "s1", TaskCode: "m07", CriteriaCode: "school_days", NumericValue: f64p(20)},
			}}, nil
		},
	}

	hist := buildSingletonProjection(context.Background(), d, &jobpb.Job{Id: "jH"}, "Adviser",
		map[int32]string{1: "s1", 2: "s2"}, map[int32]string{1: "A"}, true)
	if !histCalled || liveCalled {
		t.Fatalf("historical=true must call ONLY the historical reader (histCalled=%v liveCalled=%v)", histCalled, liveCalled)
	}
	assertLeaf(t, hist, "job_template_phases.s1.job_template_tasks.m07.task_outcomes.days_present.numeric_value", "16")
	assertLeaf(t, hist, "job_template_phases.s1.job_template_tasks.m07.task_outcomes.school_days.numeric_value", "20")
	assertLeaf(t, hist, "task_outcomes.days_present.numeric_value_total_derived", "16")

	// And the mirror: historical=false must use ONLY the live reader (no regression
	// of the current-AY card onto the historical path).
	liveCalled, histCalled = false, false
	_ = buildSingletonProjection(context.Background(), d, &jobpb.Job{Id: "jH"}, "Adviser",
		map[int32]string{1: "s1", 2: "s2"}, map[int32]string{1: "A"}, false)
	if !liveCalled || histCalled {
		t.Fatalf("historical=false must call ONLY the live reader (liveCalled=%v histCalled=%v)", liveCalled, histCalled)
	}
}

// --- strict labels vs v1/v2 fallback --------------------------------------

// The NEW tree year/phase leaves read ScaledLabel STRICTLY (blank when empty even
// if ScaledScore is set), while the FROZEN v1/v2 helpers keep the score fallback.
func TestStrictLabelsVsFallback(t *testing.T) {
	ctx := context.Background()
	d := &Deps{
		ListJobOutcomeSummarys: summariesFn(
			&jobsumpb.JobOutcomeSummary{JobId: "jA", Active: true, ScaledScore: f64p(3)}, // label empty, score set
		),
		ListPhaseOutcomeSummarysByJob: phaseSummariesFn(map[string][]*phasesumpb.PhaseOutcomeSummary{
			"jA": {{JobPhaseId: "p1", Active: true, ScaledScore: f64p(5)}}, // label empty, score set
		}),
	}
	phaseOrder := map[string]int32{"p1": 1}

	// v1/v2 fallback helpers substitute the score.
	if got := fetchYearLabels(ctx, d, []string{"jA"}); got["jA"] != "3" {
		t.Fatalf("fetchYearLabels (fallback) = %q, want 3", got["jA"])
	}
	if got := fetchSemesterLabels(ctx, d, []string{"jA"}, phaseOrder); got["jA"][1] != "5" {
		t.Fatalf("fetchSemesterLabels (fallback) = %q, want 5", got["jA"][1])
	}

	// STRICT tree helpers leave the leaf blank (absent).
	if got := fetchYearLabelsStrict(ctx, d, []string{"jA"}); got["jA"] != "" {
		t.Fatalf("fetchYearLabelsStrict = %q, want blank", got["jA"])
	}
	if got := fetchPhaseLabelsStrict(ctx, d, []string{"jA"}, phaseOrder); got["jA"] != nil && got["jA"][1] != "" {
		t.Fatalf("fetchPhaseLabelsStrict = %q, want blank", got["jA"][1])
	}
}

// --- D1 layering ----------------------------------------------------------

// Latest active outcome per (job_task, criterion) FIRST (recorded_date DESC, id
// DESC), THEN MAX across tasks. A newer LOWER revision on the same task supersedes
// an older higher one; two tasks on one criterion take the MAX.
func TestFetchTranscripts_D1Layering(t *testing.T) {
	d := &Deps{
		ListJobTasks: jobTasksFn(
			&jobtaskpb.JobTask{Id: "t1", JobPhaseId: "ph1", Active: true},
			&jobtaskpb.JobTask{Id: "t2", JobPhaseId: "ph1", Active: true},
		),
		ListTaskOutcomes: taskOutcomesFn(
			// c1 across two tasks: t1 latest 3, t2 latest 7 → MAX 7.
			&taskoutcomepb.TaskOutcome{Id: "a1", JobTaskId: "t1", CriteriaVersionId: "c1", NumericValue: f64p(3), RecordedDate: i64p(100), Active: true},
			&taskoutcomepb.TaskOutcome{Id: "a2", JobTaskId: "t2", CriteriaVersionId: "c1", NumericValue: f64p(7), RecordedDate: i64p(100), Active: true},
			// c2 on t1 only: older HIGHER (8) superseded by newer LOWER (2).
			&taskoutcomepb.TaskOutcome{Id: "b1", JobTaskId: "t1", CriteriaVersionId: "c2", NumericValue: f64p(8), RecordedDate: i64p(100), Active: true},
			&taskoutcomepb.TaskOutcome{Id: "b2", JobTaskId: "t1", CriteriaVersionId: "c2", NumericValue: f64p(2), RecordedDate: i64p(200), Active: true},
			// c3 on t1 only: same recorded_date, higher id wins (4 over 9).
			&taskoutcomepb.TaskOutcome{Id: "aa", JobTaskId: "t1", CriteriaVersionId: "c3", NumericValue: f64p(9), RecordedDate: i64p(300), Active: true},
			&taskoutcomepb.TaskOutcome{Id: "bb", JobTaskId: "t1", CriteriaVersionId: "c3", NumericValue: f64p(4), RecordedDate: i64p(300), Active: true},
		),
	}
	out := fetchTranscripts(context.Background(), d, map[string]string{"ph1": "jA"}, map[string]int32{"ph1": 1})
	tr := out["jA"]
	if tr == nil {
		t.Fatalf("no transcript built")
	}
	if got := tr.marks["c1"][1]; got != 7 {
		t.Fatalf("c1 MAX-across-tasks = %v, want 7", got)
	}
	if got := tr.marks["c2"][1]; got != 2 {
		t.Fatalf("c2 newer-lower-wins = %v, want 2 (not stale 8)", got)
	}
	if got := tr.marks["c3"][1]; got != 4 {
		t.Fatalf("c3 id-tiebreak = %v, want 4 (higher id)", got)
	}
}

// --- D1 criterion-sequence determinism ------------------------------------

// A criterion bound by TWO template tasks with DIFFERENT sequence_order must
// resolve to a single, stable order regardless of the randomized map-iteration
// order of the latest-cell fold. Lowest sequence_order wins (tie-break: lowest
// template-task id). Run repeatedly: the resolved order — and thus the A–D slot —
// must never flip between renders.
func TestFetchTranscripts_D1CriterionSequenceDeterministic(t *testing.T) {
	d := &Deps{
		ListJobTasks: jobTasksFn(
			&jobtaskpb.JobTask{Id: "jt-hi", JobPhaseId: "ph1", Active: true, TemplateTaskId: strp("tt-hi")},
			&jobtaskpb.JobTask{Id: "jt-lo", JobPhaseId: "ph1", Active: true, TemplateTaskId: strp("tt-lo")},
		),
		ListTaskOutcomes: taskOutcomesFn(
			// c1 recorded on BOTH tasks (so both template tasks bind it). MAX = 6.
			&taskoutcomepb.TaskOutcome{Id: "o1", JobTaskId: "jt-hi", CriteriaVersionId: "c1", NumericValue: f64p(5), RecordedDate: i64p(100), Active: true},
			&taskoutcomepb.TaskOutcome{Id: "o2", JobTaskId: "jt-lo", CriteriaVersionId: "c1", NumericValue: f64p(6), RecordedDate: i64p(100), Active: true},
			// c2 bound only by tt-lo (seq 2).
			&taskoutcomepb.TaskOutcome{Id: "o3", JobTaskId: "jt-lo", CriteriaVersionId: "c2", NumericValue: f64p(7), RecordedDate: i64p(100), Active: true},
		),
		ListTemplateTaskCriterias: ttcFn(
			// c1: tt-lo binds it at seq 1, tt-hi at seq 4 → the lower (1) must win.
			&ttcpb.TemplateTaskCriteria{JobTemplateTaskId: "tt-lo", OutcomeCriteriaId: "c1", SequenceOrder: 1, Active: true},
			&ttcpb.TemplateTaskCriteria{JobTemplateTaskId: "tt-hi", OutcomeCriteriaId: "c1", SequenceOrder: 4, Active: true},
			&ttcpb.TemplateTaskCriteria{JobTemplateTaskId: "tt-lo", OutcomeCriteriaId: "c2", SequenceOrder: 2, Active: true},
		),
	}
	for i := 0; i < 25; i++ {
		out := fetchTranscripts(context.Background(), d, map[string]string{"ph1": "jA"}, map[string]int32{"ph1": 1})
		tr := out["jA"]
		if tr == nil {
			t.Fatalf("no transcript built")
		}
		if got := tr.seqOf("c1"); got != 1 {
			t.Fatalf("iter %d: c1 seq = %d, want 1 (lowest across binding template tasks)", i, got)
		}
		if got := tr.seqOf("c2"); got != 2 {
			t.Fatalf("iter %d: c2 seq = %d, want 2", i, got)
		}
		if crits := tr.orderedCrits(); len(crits) != 2 || crits[0] != "c1" || crits[1] != "c2" {
			t.Fatalf("iter %d: ordered crits = %v, want [c1 c2]", i, crits)
		}
	}
}

// --- deportment canonical projection (rotation pairs + non-enrolled) -------

// The block-layout subject-deportment .jobs[] loop must render the SAME row set
// as the v2 conduct table: rotation pairs merged into ONE row (each strand's own
// period value, keyed by phase code) and non-enrolled placeholder strands
// suppressed — never split half-rows or resurrected placeholders. This pins the
// Baliling shape (merged Arts row, no Korean placeholder) through the block tree.
func TestBuildJobCategoriesTree_DeportRotationPairAndNonEnrolled(t *testing.T) {
	c := &ratingContext{
		nameOf: map[string]string{
			"jva":  "Arts: Visual Arts",
			"jmu":  "Arts: Music",
			"jkor": "Language and Literature: Korean",  // non-enrolled placeholder
			"jeng": "Language and Literature: English", // enrolled solo (prefix is NOT an academic subject → no merge)
		},
		pos: map[string]map[int32]string{
			"jva":  {1: "100"},
			"jmu":  {2: "98"},
			"jeng": {1: "100", 2: "100"},
		},
		avg: map[string]string{
			"jva": "100", "jmu": "98", "jeng": "100",
			"jkor": "0", // transmute-of-zero floor → suppressed
		},
	}
	c.strandJobs = jobsFromNames(c.nameOf)
	merged := mergeRotationPairs(c, map[string]bool{"arts": true}, nil)
	rows := deportRows(c, merged)

	in := treeInputs{
		cats:       map[string]catInfo{"cat-subj": {name: "Subject Deportment", code: "subject_deportment"}},
		deportJobs: map[string]string{"jva": "cat-subj", "jmu": "cat-subj", "jkor": "cat-subj", "jeng": "cat-subj"},
		deportRows: rows,
		jobOrderCode: map[string]map[int32]string{
			"jva":  {1: "s1", 2: "s2"},
			"jmu":  {1: "s1", 2: "s2"},
			"jeng": {1: "s1", 2: "s2"},
		},
		strictPhase: map[string]map[int32]string{
			"jva":  {1: "100"},
			"jmu":  {2: "98"},
			"jeng": {1: "100", 2: "100"},
		},
	}
	tree := buildJobCategoriesTree(context.Background(), &Deps{}, in, nil)

	jobsAny, ok := resolvePath(tree, "subject_deportment.jobs")
	if !ok {
		t.Fatalf("subject_deportment.jobs did not resolve")
	}
	jobs := jobsAny.([]any)
	if len(jobs) != 2 {
		var titles []string
		for _, j := range jobs {
			titles = append(titles, j.(map[string]any)["job_template_name_display"].(string))
		}
		t.Fatalf("subject_deportment jobs = %d %v, want 2 (merged Arts + solo English; Korean suppressed)", len(jobs), titles)
	}
	for _, j := range jobs {
		if title := j.(map[string]any)["job_template_name_display"].(string); title == "Language and Literature: Korean" {
			t.Fatalf("non-enrolled Korean placeholder must NOT render")
		}
	}
	// Sorted by title: merged Arts row first (each strand contributes its period).
	arts := jobs[0].(map[string]any)
	assertLeaf(t, arts, "job_template_name_display", "Arts: Visual Arts / Arts: Music")
	assertLeaf(t, arts, "job_template_phases.s1.phase_outcome_summary_scaled_label", "100")
	assertLeaf(t, arts, "job_template_phases.s2.phase_outcome_summary_scaled_label", "98")
	// Solo enrolled English row.
	eng := jobs[1].(map[string]any)
	assertLeaf(t, eng, "job_template_name_display", "Language and Literature: English")
	assertLeaf(t, eng, "job_template_phases.s1.phase_outcome_summary_scaled_label", "100")
	assertLeaf(t, eng, "job_template_phases.s2.phase_outcome_summary_scaled_label", "100")
}

// --- shared mock helpers --------------------------------------------------

func ttcFn(rows ...*ttcpb.TemplateTaskCriteria) func(context.Context, *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error) {
	return func(context.Context, *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error) {
		return &ttcpb.ListTemplateTaskCriteriasResponse{Data: rows}, nil
	}
}

func assertLeaf(t *testing.T, m map[string]any, path, want string) {
	t.Helper()
	v, ok := resolvePath(m, path)
	if !ok {
		t.Fatalf("path %q did not resolve", path)
	}
	if v != want {
		t.Fatalf("path %q = %#v, want %q", path, v, want)
	}
}

func jobTasksFn(tasks ...*jobtaskpb.JobTask) func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
	return func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
		return &jobtaskpb.ListJobTasksResponse{Data: tasks}, nil
	}
}

func taskOutcomesFn(outcomes ...*taskoutcomepb.TaskOutcome) func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
	return func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		return &taskoutcomepb.ListTaskOutcomesResponse{Data: outcomes}, nil
	}
}

func phaseSummariesFn(byJob map[string][]*phasesumpb.PhaseOutcomeSummary) func(context.Context, *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error) {
	return func(_ context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error) {
		return &phasesumpb.ListPhaseOutcomeSummarysByJobResponse{PhaseOutcomeSummarys: byJob[req.GetJobId()]}, nil
	}
}
