package outcome_summary

import (
	"context"
	"testing"

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

	ev := FetchJobMarkEvidence(context.Background(), phases, tasks, outcomes,
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
	ev := FetchJobMarkEvidence(context.Background(), phases, tasks, outcomes, []string{"job1"})
	if got, ok := ev["job1"]; ok || got.HasMarks {
		t.Errorf("inactive/nil-numeric outcomes must carry no evidence, got %+v ok=%v", got, ok)
	}
}

// TestFetchJobMarkEvidence_NilSafe: any nil closure (a tier that never wired the
// walk, e.g. service-admin) yields an empty map — nothing is blanked downstream.
func TestFetchJobMarkEvidence_NilSafe(t *testing.T) {
	phases := phasesFn(&jobphasepb.JobPhase{Id: "ph1", JobId: "job1", Active: true})
	tasks := tasksFn(&jobtaskpb.JobTask{Id: "tk1", JobPhaseId: "ph1", Active: true})

	if ev := FetchJobMarkEvidence(context.Background(), nil, tasks, outcomesFn(), []string{"job1"}); len(ev) != 0 {
		t.Errorf("nil listJobPhases must yield empty evidence, got %v", ev)
	}
	if ev := FetchJobMarkEvidence(context.Background(), phases, nil, outcomesFn(), []string{"job1"}); len(ev) != 0 {
		t.Errorf("nil listJobTasks must yield empty evidence, got %v", ev)
	}
	if ev := FetchJobMarkEvidence(context.Background(), phases, tasks, nil, []string{"job1"}); len(ev) != 0 {
		t.Errorf("nil listTaskOutcomes must yield empty evidence, got %v", ev)
	}
	if ev := FetchJobMarkEvidence(context.Background(), phases, tasks, outcomesFn(), nil); len(ev) != 0 {
		t.Errorf("empty jobIDs must yield empty evidence, got %v", ev)
	}
}
