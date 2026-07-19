package document

import (
	"context"
	"errors"
	"testing"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

func strptr(s string) *string { return &s }

// gateDeps builds a Deps with canned job_phase / job_task / task_outcome closures
// (same set returned for every query — used for singleton, null-template-phase cases).
func gateDeps(phases []*jobphasepb.JobPhase, tasks []*jobtaskpb.JobTask, outcomes []*taskoutcomepb.TaskOutcome) *Deps {
	return &Deps{
		ListJobPhases: func(_ context.Context, _ *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			return &jobphasepb.ListJobPhasesResponse{Data: phases, Success: true}, nil
		},
		ListJobTasks: func(_ context.Context, _ *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			return &jobtaskpb.ListJobTasksResponse{Data: tasks, Success: true}, nil
		},
		ListTaskOutcomes: func(_ context.Context, _ *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
			return &taskoutcomepb.ListTaskOutcomesResponse{Data: outcomes, Success: true}, nil
		},
	}
}

// fullSDeps distinguishes the by-job_id card read (cardPhases) from the
// by-template_phase_id full-sheet read (sheetPhases) so the full-S aggregate can be
// exercised (codex §B4).
func fullSDeps(cardPhases, sheetPhases []*jobphasepb.JobPhase, tasks []*jobtaskpb.JobTask, outcomes []*taskoutcomepb.TaskOutcome) *Deps {
	return &Deps{
		ListJobPhases: func(_ context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			field := ""
			if f := req.GetFilters().GetFilters(); len(f) > 0 {
				field = f[0].GetField()
			}
			if field == "template_phase_id" {
				return &jobphasepb.ListJobPhasesResponse{Data: sheetPhases, Success: true}, nil
			}
			return &jobphasepb.ListJobPhasesResponse{Data: cardPhases, Success: true}, nil
		},
		ListJobTasks: func(_ context.Context, _ *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			return &jobtaskpb.ListJobTasksResponse{Data: tasks, Success: true}, nil
		},
		ListTaskOutcomes: func(_ context.Context, _ *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
			return &taskoutcomepb.ListTaskOutcomesResponse{Data: outcomes, Success: true}, nil
		},
	}
}

func phase(id string, status jobphasepb.PhaseApprovalStatus, returnedBy string) *jobphasepb.JobPhase {
	p := &jobphasepb.JobPhase{Id: id, Active: true, ApprovalStatus: status}
	if returnedBy != "" {
		p.ReturnedBy = strptr(returnedBy)
	}
	return p
}

func sheetPhase(id, tp string, status jobphasepb.PhaseApprovalStatus, returnedBy string) *jobphasepb.JobPhase {
	p := phase(id, status, returnedBy)
	p.TemplatePhaseId = strptr(tp)
	return p
}

// TestReportRenderStatus pins the D5 render-gate predicate (singleton / null
// template-phase cases — evaluated as their own sheet).
func TestReportRenderStatus(t *testing.T) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		fr  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		pub = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}

	cases := []struct {
		name     string
		phases   []*jobphasepb.JobPhase
		tasks    []*jobtaskpb.JobTask
		outcomes []*taskoutcomepb.TaskOutcome
		want     bool
	}{
		{"never_workflowed_backfill_renders", []*jobphasepb.JobPhase{phase("jp-1", ip, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, false},
		{"for_review_with_data_blocks", []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, true},
		{"published_renders", []*jobphasepb.JobPhase{phase("jp-1", pub, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, false},
		{"returned_with_data_blocks", []*jobphasepb.JobPhase{phase("jp-1", ip, "user-x")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome}, true},
		{"for_review_no_data_renders", []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{}, false},
		{"for_review_no_tasks_renders", []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, []*jobtaskpb.JobTask{}, []*taskoutcomepb.TaskOutcome{outcome}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := gateDeps(c.phases, c.tasks, c.outcomes)
			got, err := reportRenderStatus(context.Background(), d, []string{"job-1"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("reportRenderStatus = %v, want %v", got, c.want)
			}
		})
	}
}

// TestReportRenderStatus_FullSheet exercises the full-S, template-phase-grain
// aggregate (codex §B4): the card's own phase alone would render, but a sibling
// member of the same approval sheet forces a block.
func TestReportRenderStatus_FullSheet(t *testing.T) {
	const (
		ip  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		fr  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		pub = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	task := &jobtaskpb.JobTask{Id: "jt-1", Active: true}
	outcome := &taskoutcomepb.TaskOutcome{Id: "to-1", Active: true}

	// (2) published member with data + a late pristine IN_PROGRESS member → block.
	// The card only holds the PUBLISHED phase (would render alone), but the sheet has
	// a pristine sibling, so BOOL_AND(published) is false → block.
	t.Run("published_plus_pristine_member_blocks", func(t *testing.T) {
		card := []*jobphasepb.JobPhase{sheetPhase("jp-pub", "sheet-1", pub, "")}
		sheet := []*jobphasepb.JobPhase{
			sheetPhase("jp-pub", "sheet-1", pub, ""),
			sheetPhase("jp-pristine", "sheet-1", ip, ""),
		}
		d := fullSDeps(card, sheet, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome})
		got, err := reportRenderStatus(context.Background(), d, []string{"job-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("full-S with a pristine sibling member must block, got render")
		}
	})

	// (3) workflow-entered member and data-bearing member differ.
	t.Run("entered_and_data_on_different_members_blocks", func(t *testing.T) {
		card := []*jobphasepb.JobPhase{sheetPhase("jp-a", "sheet-1", ip, "")} // card's own is pristine
		sheet := []*jobphasepb.JobPhase{
			sheetPhase("jp-a", "sheet-1", ip, ""), // data-bearing (below), not entered
			sheetPhase("jp-b", "sheet-1", fr, ""), // entered, no data of its own
		}
		d := fullSDeps(card, sheet, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome})
		got, err := reportRenderStatus(context.Background(), d, []string{"job-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("entered + data on different members must block, got render")
		}
	})

	// fully published sheet renders.
	t.Run("all_published_renders", func(t *testing.T) {
		card := []*jobphasepb.JobPhase{sheetPhase("jp-a", "sheet-1", pub, "")}
		sheet := []*jobphasepb.JobPhase{
			sheetPhase("jp-a", "sheet-1", pub, ""),
			sheetPhase("jp-b", "sheet-1", pub, ""),
		}
		d := fullSDeps(card, sheet, []*jobtaskpb.JobTask{task}, []*taskoutcomepb.TaskOutcome{outcome})
		got, err := reportRenderStatus(context.Background(), d, []string{"job-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("fully published sheet must render, got blocked")
		}
	})
}

// TestReportRenderStatus_FailClosed confirms the gate FAILS CLOSED (returns an
// error → 503) on nil deps or any list error / permission denial (codex §B4).
func TestReportRenderStatus_FailClosed(t *testing.T) {
	t.Run("nil_job_phase_dep_errors", func(t *testing.T) {
		if _, err := reportRenderStatus(context.Background(), &Deps{}, []string{"job-1"}); err == nil {
			t.Errorf("nil ListJobPhases must fail closed (error), got nil")
		}
	})

	t.Run("empty_job_set_renders", func(t *testing.T) {
		blocked, err := reportRenderStatus(context.Background(), gateDeps(nil, nil, nil), nil)
		if err != nil || blocked {
			t.Errorf("empty job set must render with no error, got blocked=%v err=%v", blocked, err)
		}
	})

	t.Run("job_phase_list_error_fails_closed", func(t *testing.T) {
		d := &Deps{
			ListJobPhases: func(_ context.Context, _ *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
				return nil, errors.New("permission denied: job_phase:list")
			},
		}
		if _, err := reportRenderStatus(context.Background(), d, []string{"job-1"}); err == nil {
			t.Errorf("a list error / permission denial must fail closed (error), got nil")
		}
	})

	t.Run("task_outcome_list_error_fails_closed", func(t *testing.T) {
		fr := jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		d := &Deps{
			ListJobPhases: func(_ context.Context, _ *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
				return &jobphasepb.ListJobPhasesResponse{Data: []*jobphasepb.JobPhase{phase("jp-1", fr, "")}, Success: true}, nil
			},
			ListJobTasks: func(_ context.Context, _ *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
				return &jobtaskpb.ListJobTasksResponse{Data: []*jobtaskpb.JobTask{{Id: "jt-1", Active: true}}, Success: true}, nil
			},
			ListTaskOutcomes: func(_ context.Context, _ *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
				return nil, errors.New("permission denied: task_outcome:list")
			},
		}
		if _, err := reportRenderStatus(context.Background(), d, []string{"job-1"}); err == nil {
			t.Errorf("a task_outcome list error must fail closed (error), got nil")
		}
	})
}

// TestPhaseWorkflowEntered pins the workflow-entered predicate.
func TestPhaseWorkflowEntered(t *testing.T) {
	const (
		unspec = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_UNSPECIFIED
		ip     = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
		fr     = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
		pub    = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
	)
	cases := []struct {
		name string
		p    *jobphasepb.JobPhase
		want bool
	}{
		{"pristine_in_progress", &jobphasepb.JobPhase{ApprovalStatus: ip}, false},
		{"unspecified_null_audit", &jobphasepb.JobPhase{ApprovalStatus: unspec}, false},
		{"for_review", &jobphasepb.JobPhase{ApprovalStatus: fr}, true},
		{"published", &jobphasepb.JobPhase{ApprovalStatus: pub}, true},
		{"in_progress_but_submitted", &jobphasepb.JobPhase{ApprovalStatus: ip, SubmittedBy: strptr("u")}, true},
		{"in_progress_but_returned", &jobphasepb.JobPhase{ApprovalStatus: ip, ReturnedBy: strptr("u")}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := phaseWorkflowEntered(c.p); got != c.want {
				t.Errorf("phaseWorkflowEntered = %v, want %v", got, c.want)
			}
		})
	}
}
