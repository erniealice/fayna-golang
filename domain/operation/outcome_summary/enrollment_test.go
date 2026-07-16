package outcome_summary

import (
	"context"
	"errors"
	"fmt"
	sortpkg "sort"
	"testing"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// fptr returns a *float64 for the optional task_outcome numeric_value.
func fptr(v float64) *float64 { return &v }

// TestIsNonEnrolledCell is the enrollment-suppression invariant (owed B1/T2
// test): a phantom (untaken-elective all-zero scaffold floored to "1") is a
// placeholder → BLANK; a genuinely-enrolled subject scored a real 0 or 1 has a
// positive task mark → SHOWN; a historical import (no task rows but a stored
// band) → KEPT. NEVER blank a real grade.
func TestIsNonEnrolledCell(t *testing.T) {
	phantom := EnrollmentEvidence{HasMarks: true, HasPositiveMark: false} // all-zero scaffold
	enrolled := EnrollmentEvidence{HasMarks: true, HasPositiveMark: true} // ≥1 positive mark
	noTasks := EnrollmentEvidence{HasMarks: false, HasPositiveMark: false}

	cases := []struct {
		name string
		ev   EnrollmentEvidence
		band []string
		want bool // true = placeholder (blank)
	}{
		// PHANTOM → BLANK: an all-zero scaffold whose year-final floored to "1".
		{"phantom floor-1 single band", phantom, []string{"1"}, true},
		{"phantom floor-1 all three bands", phantom, []string{"1", "1", "1"}, true},
		{"phantom floor-0", phantom, []string{"0"}, true},

		// REAL grade → SHOWN: positive mark evidence keeps it, even at the floor.
		{"real 1 with positive mark", enrolled, []string{"1"}, false},
		{"real 0 with positive mark", enrolled, []string{"0"}, false},
		{"real 1 S1-only (empty S2) with positive mark", enrolled, []string{"1", "5", ""}, false},

		// A real (>1) stored band keeps the cell regardless of mark evidence.
		{"non-floor band >1 keeps even without marks", noTasks, []string{"7"}, false},
		{"non-floor band >1 keeps a scaffold", phantom, []string{"6"}, false},

		// HISTORICAL import: no task rows, a stored floor band → KEPT (the case a
		// naive "no positive mark → blank" would wrongly erase).
		{"historical import floor-1 no tasks", noTasks, []string{"1"}, false},
		{"historical import floor-0 no tasks", noTasks, []string{"0"}, false},

		// Fully-blank cell (no summary at all) → placeholder.
		{"fully blank no summary", noTasks, []string{""}, true},
		{"fully blank no bands", noTasks, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsNonEnrolledCell(c.ev, c.band...); got != c.want {
				t.Fatalf("IsNonEnrolledCell(%+v, %v) = %v, want %v", c.ev, c.band, got, c.want)
			}
		})
	}
}

// TestNumGreaterThan pins the shared numeric-band test.
func TestNumGreaterThan(t *testing.T) {
	cases := []struct {
		s    string
		n    float64
		want bool
	}{
		{"7", 1, true},
		{"1", 1, false},
		{"0", 1, false},
		{"2", 0, true},
		{"0", 0, false},
		{"", 0, false},
		{"—", 1, false},
		{" 6 ", 1, true},
	}
	for _, c := range cases {
		if got := NumGreaterThan(c.s, c.n); got != c.want {
			t.Errorf("NumGreaterThan(%q, %v) = %v, want %v", c.s, c.n, got, c.want)
		}
	}
}

// --- FetchJobMarkEvidence walk ------------------------------------------------

func phasesFn(phases ...*jobphasepb.JobPhase) func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
	return func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
		return &jobphasepb.ListJobPhasesResponse{Data: phases}, nil
	}
}

func tasksFn(tasks ...*jobtaskpb.JobTask) func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
	return func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
		return &jobtaskpb.ListJobTasksResponse{Data: tasks}, nil
	}
}

func outcomesFn(outcomes ...*taskoutcomepb.TaskOutcome) func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
	return func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		return &taskoutcomepb.ListTaskOutcomesResponse{Data: outcomes}, nil
	}
}

// TestFetchJobMarkEvidence exercises the job_phase → job_task → task_outcome
// walk: a positive mark → HasPositiveMark; an all-zero scaffold → HasMarks but
// not positive; a subject with no task_outcome rows → absent (HasMarks=false).
func TestFetchJobMarkEvidence(t *testing.T) {
	phases := phasesFn(
		&jobphasepb.JobPhase{Id: "phPos", JobId: "jobPos", Active: true},
		&jobphasepb.JobPhase{Id: "phZero", JobId: "jobZero", Active: true},
		&jobphasepb.JobPhase{Id: "phNone", JobId: "jobNone", Active: true},
	)
	tasks := tasksFn(
		&jobtaskpb.JobTask{Id: "tkPos", JobPhaseId: "phPos", Active: true},
		&jobtaskpb.JobTask{Id: "tkZero", JobPhaseId: "phZero", Active: true},
		&jobtaskpb.JobTask{Id: "tkNone", JobPhaseId: "phNone", Active: true},
	)
	outcomes := outcomesFn(
		&taskoutcomepb.TaskOutcome{JobTaskId: "tkPos", NumericValue: fptr(5), Active: true},
		&taskoutcomepb.TaskOutcome{JobTaskId: "tkZero", NumericValue: fptr(0), Active: true},
		// tkNone: no task_outcome rows at all (historical / untaken-with-no-scaffold).
	)

	ev, _ := FetchJobMarkEvidence(context.Background(), phases, tasks, outcomes,
		[]string{"jobPos", "jobZero", "jobNone"})

	if got := ev["jobPos"]; !got.HasMarks || !got.HasPositiveMark {
		t.Errorf("jobPos = %+v, want HasMarks && HasPositiveMark", got)
	}
	if got := ev["jobZero"]; !got.HasMarks || got.HasPositiveMark {
		t.Errorf("jobZero (all-zero scaffold) = %+v, want HasMarks && !HasPositiveMark", got)
	}
	if got, ok := ev["jobNone"]; ok || got.HasMarks {
		t.Errorf("jobNone (no task_outcome) = %+v ok=%v, want absent / zero evidence", got, ok)
	}
}

// TestFetchJobMarkEvidence_InactiveAndNilSkipped: an inactive task_outcome or a
// nil numeric_value carries no evidence.
func TestFetchJobMarkEvidence_InactiveAndNilSkipped(t *testing.T) {
	phases := phasesFn(&jobphasepb.JobPhase{Id: "ph1", JobId: "job1", Active: true})
	tasks := tasksFn(&jobtaskpb.JobTask{Id: "tk1", JobPhaseId: "ph1", Active: true})
	outcomes := outcomesFn(
		&taskoutcomepb.TaskOutcome{JobTaskId: "tk1", NumericValue: fptr(9), Active: false}, // inactive
		&taskoutcomepb.TaskOutcome{JobTaskId: "tk1", NumericValue: nil, Active: true},      // no numeric
	)
	ev, _ := FetchJobMarkEvidence(context.Background(), phases, tasks, outcomes, []string{"job1"})
	if got, ok := ev["job1"]; ok || got.HasMarks {
		t.Errorf("inactive/nil-numeric outcomes must carry no evidence, got %+v ok=%v", got, ok)
	}
}

// TestFetchJobMarkEvidence_ErrorFailsClosed guards the fail-closed evidence
// contract (codex #3): a read error mid-walk must abort with (nil, err) — never
// a partial map. A full first page of all-zero marks (forcing a second page)
// followed by an errored second page must NOT yield HasMarks=true for job1; the
// caller relies on nil evidence to keep the grade instead of blanking a real one.
func TestFetchJobMarkEvidence_ErrorFailsClosed(t *testing.T) {
	phases := phasesFn(&jobphasepb.JobPhase{Id: "ph1", JobId: "job1", Active: true})
	tasks := tasksFn(&jobtaskpb.JobTask{Id: "tk1", JobPhaseId: "ph1", Active: true})

	full := make([]*taskoutcomepb.TaskOutcome, markEvidencePageLimit)
	for i := range full {
		full[i] = &taskoutcomepb.TaskOutcome{JobTaskId: "tk1", NumericValue: fptr(0), Active: true}
	}
	call := 0
	outcomes := func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		call++
		if call == 1 {
			return &taskoutcomepb.ListTaskOutcomesResponse{Data: full}, nil // full page → a second page is requested
		}
		return nil, errors.New("transient db error")
	}

	ev, err := FetchJobMarkEvidence(context.Background(), phases, tasks, outcomes, []string{"job1"})
	if err == nil {
		t.Fatal("want a non-nil error when a page read fails, got nil")
	}
	if ev != nil {
		t.Errorf("want nil evidence on error (fail-closed), got %+v", ev)
	}
}

// --- pagination-stability contract -------------------------------------------

// inSet extracts the LIST_IN filter values as a membership set.
func inSet(f *commonpb.FilterRequest) map[string]bool {
	set := map[string]bool{}
	for _, tf := range f.GetFilters() {
		for _, v := range tf.GetListFilter().GetValues() {
			set[v] = true
		}
	}
	return set
}

// pageParams reads the 1-based page + limit from an offset pagination request.
func pageParams(p *commonpb.PaginationRequest) (page, limit int) {
	page, limit = 1, int(p.GetLimit())
	if pg := int(p.GetOffset().GetPage()); pg > 0 {
		page = pg
	}
	return page, limit
}

// idSortPage mimics the base PostgresOperations.List: when the request carries a
// unique id-sort it orders deterministically by id and slices the offset page —
// so paging is stable and no row is dropped or duplicated across pages.
func idSortPage[T interface{ GetId() string }](rows []T, sort *commonpb.SortRequest, page, limit int) []T {
	if sort != nil && len(sort.Fields) > 0 && sort.Fields[0].Field == "id" {
		ordered := append([]T(nil), rows...)
		sortpkg.Slice(ordered, func(i, j int) bool { return ordered[i].GetId() < ordered[j].GetId() })
		rows = ordered
	}
	start := (page - 1) * limit
	if limit <= 0 || start >= len(rows) {
		return nil
	}
	end := start + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

// TestFetchJobMarkEvidence_PaginationStable is the regression guard for the
// unstable-OFFSET under-blank: 50 jobs × 3 phases = 150 phases (and matching
// tasks/outcomes) inside ONE IN-chunk force a genuine multi-page walk (limit
// 100). It asserts (1) every paged request across all three levels carries the
// unique id-sort (without which OFFSET paging over tied date_created silently
// drops whole jobs), and (2) all 50 jobs' evidence survives the page seam.
func TestFetchJobMarkEvidence_PaginationStable(t *testing.T) {
	const nJobs = 50
	var (
		jobIDs   []string
		phases   []*jobphasepb.JobPhase
		tasks    []*jobtaskpb.JobTask
		outcomes []*taskoutcomepb.TaskOutcome
	)
	for j := 0; j < nJobs; j++ {
		jid := fmt.Sprintf("job%03d", j)
		jobIDs = append(jobIDs, jid)
		for p := 0; p < 3; p++ { // 3 phases/job → 150 rows/level → 2 pages
			pid := fmt.Sprintf("ph%03d-%d", j, p)
			tid := fmt.Sprintf("tk%03d-%d", j, p)
			phases = append(phases, &jobphasepb.JobPhase{Id: pid, JobId: jid, Active: true})
			tasks = append(tasks, &jobtaskpb.JobTask{Id: tid, JobPhaseId: pid, Active: true})
			outcomes = append(outcomes, &taskoutcomepb.TaskOutcome{Id: fmt.Sprintf("oc%03d-%d", j, p), JobTaskId: tid, NumericValue: fptr(5), Active: true})
		}
	}

	var sortFields []string // the Sort field seen on every paged request
	record := func(s *commonpb.SortRequest) {
		if s == nil || len(s.Fields) == 0 {
			sortFields = append(sortFields, "")
			return
		}
		sortFields = append(sortFields, s.Fields[0].Field)
	}

	listPhases := func(_ context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
		record(req.GetSort())
		in := inSet(req.GetFilters())
		var matched []*jobphasepb.JobPhase
		for _, x := range phases {
			if in[x.GetJobId()] {
				matched = append(matched, x)
			}
		}
		page, limit := pageParams(req.GetPagination())
		return &jobphasepb.ListJobPhasesResponse{Data: idSortPage(matched, req.GetSort(), page, limit)}, nil
	}
	listTasks := func(_ context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
		record(req.GetSort())
		in := inSet(req.GetFilters())
		var matched []*jobtaskpb.JobTask
		for _, x := range tasks {
			if in[x.GetJobPhaseId()] {
				matched = append(matched, x)
			}
		}
		page, limit := pageParams(req.GetPagination())
		return &jobtaskpb.ListJobTasksResponse{Data: idSortPage(matched, req.GetSort(), page, limit)}, nil
	}
	listOutcomes := func(_ context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		record(req.GetSort())
		in := inSet(req.GetFilters())
		var matched []*taskoutcomepb.TaskOutcome
		for _, x := range outcomes {
			if in[x.GetJobTaskId()] {
				matched = append(matched, x)
			}
		}
		page, limit := pageParams(req.GetPagination())
		return &taskoutcomepb.ListTaskOutcomesResponse{Data: idSortPage(matched, req.GetSort(), page, limit)}, nil
	}

	ev, _ := FetchJobMarkEvidence(context.Background(), listPhases, listTasks, listOutcomes, jobIDs)

	// (1) every paged request carried the unique id-sort.
	if len(sortFields) == 0 {
		t.Fatal("no list requests recorded")
	}
	sawMultiPage := false
	for _, f := range sortFields {
		if f != "id" {
			t.Fatalf("a paged request omitted the id-sort (fields seen: %v)", sortFields)
		}
	}
	// A multi-page walk means > 3 calls total (each level pages twice for 150 rows).
	if len(sortFields) > 3 {
		sawMultiPage = true
	}
	if !sawMultiPage {
		t.Fatalf("expected a multi-page walk (>3 list calls), got %d", len(sortFields))
	}

	// (2) all 50 jobs' positive-mark evidence survived the page seam.
	if len(ev) != nJobs {
		t.Fatalf("evidence map has %d jobs, want %d (a job was dropped across pages)", len(ev), nJobs)
	}
	for j := 0; j < nJobs; j++ {
		jid := fmt.Sprintf("job%03d", j)
		if got := ev[jid]; !got.HasMarks || !got.HasPositiveMark {
			t.Fatalf("job %s = %+v, want HasMarks && HasPositiveMark (paging dropped its rows)", jid, got)
		}
	}
}

// TestFetchJobMarkEvidence_NilSafe: any nil closure (a tier that never wired the
// walk, e.g. service-admin) yields an empty map — nothing is blanked downstream.
func TestFetchJobMarkEvidence_NilSafe(t *testing.T) {
	phases := phasesFn(&jobphasepb.JobPhase{Id: "ph1", JobId: "job1", Active: true})
	tasks := tasksFn(&jobtaskpb.JobTask{Id: "tk1", JobPhaseId: "ph1", Active: true})

	if ev, _ := FetchJobMarkEvidence(context.Background(), nil, tasks, outcomesFn(), []string{"job1"}); len(ev) != 0 {
		t.Errorf("nil listJobPhases must yield empty evidence, got %v", ev)
	}
	if ev, _ := FetchJobMarkEvidence(context.Background(), phases, nil, outcomesFn(), []string{"job1"}); len(ev) != 0 {
		t.Errorf("nil listJobTasks must yield empty evidence, got %v", ev)
	}
	if ev, _ := FetchJobMarkEvidence(context.Background(), phases, tasks, nil, []string{"job1"}); len(ev) != 0 {
		t.Errorf("nil listTaskOutcomes must yield empty evidence, got %v", ev)
	}
	if ev, _ := FetchJobMarkEvidence(context.Background(), phases, tasks, outcomesFn(), nil); len(ev) != 0 {
		t.Errorf("empty jobIDs must yield empty evidence, got %v", ev)
	}
}
