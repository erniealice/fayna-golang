package detail

import (
	"context"
	"slices"
	"testing"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// Plan 20260927-db-query-performance AC-02 (audit DB-10): the Phases tab must
// ask for the viewed job's phases and tasks instead of the workspace's first
// page and filtering in memory.
func TestLoadPhasesTab_FiltersByJob(t *testing.T) {
	const jobID = "job-a"
	var phaseReq *jobphasepb.ListJobPhasesRequest
	var taskReq *jobtaskpb.ListJobTasksRequest

	deps := &DetailViewDeps{
		ListJobPhases: func(_ context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			phaseReq = req
			return &jobphasepb.ListJobPhasesResponse{Data: []*jobphasepb.JobPhase{
				{Id: "phase-1", JobId: jobID, PhaseOrder: 1},
				{Id: "phase-2", JobId: jobID, PhaseOrder: 2},
				{Id: "phase-x", JobId: "job-b", PhaseOrder: 1}, // defensive: never rendered
			}}, nil
		},
		ListJobTasks: func(_ context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			taskReq = req
			return &jobtaskpb.ListJobTasksResponse{Data: []*jobtaskpb.JobTask{
				{Id: "task-1", JobPhaseId: "phase-1"},
				{Id: "task-2", JobPhaseId: "phase-2"},
				{Id: "task-x", JobPhaseId: "phase-x"},
			}}, nil
		},
	}
	pageData := &PageData{}

	loadPhasesTab(context.Background(), deps, pageData, jobID)

	if phaseReq == nil || len(phaseReq.GetFilters().GetFilters()) != 1 {
		t.Fatalf("phase request filters = %v, want one job_id filter", phaseReq.GetFilters())
	}
	pf := phaseReq.GetFilters().GetFilters()[0]
	if pf.GetField() != "job_id" || pf.GetStringFilter().GetValue() != jobID ||
		pf.GetStringFilter().GetOperator() != commonpb.StringOperator_STRING_EQUALS {
		t.Fatalf("phase filter = %v, want job_id = %q", pf, jobID)
	}

	if taskReq == nil || len(taskReq.GetFilters().GetFilters()) != 1 {
		t.Fatalf("task request filters = %v, want one job_phase_id IN filter", taskReq.GetFilters())
	}
	tf := taskReq.GetFilters().GetFilters()[0]
	if tf.GetField() != "job_phase_id" || tf.GetListFilter().GetOperator() != commonpb.ListOperator_LIST_IN {
		t.Fatalf("task filter = %v, want job_phase_id IN", tf)
	}
	if got := tf.GetListFilter().GetValues(); !slices.Equal(got, []string{"phase-1", "phase-2"}) {
		t.Fatalf("task filter values = %v, want the job's phase ids", got)
	}

	if len(pageData.PhasesList) != 2 {
		t.Fatalf("PhasesList has %d rows, want 2", len(pageData.PhasesList))
	}
	for _, row := range pageData.PhasesList {
		if row.ID == "phase-x" {
			t.Fatal("a foreign job's phase was rendered")
		}
	}
}

func TestLoadPhasesTab_NoPhasesSkipsTaskQuery(t *testing.T) {
	calledTasks := false
	deps := &DetailViewDeps{
		ListJobPhases: func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			return &jobphasepb.ListJobPhasesResponse{}, nil
		},
		ListJobTasks: func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			calledTasks = true
			return &jobtaskpb.ListJobTasksResponse{}, nil
		},
	}
	loadPhasesTab(context.Background(), deps, &PageData{}, "job-a")
	if calledTasks {
		t.Fatal("tasks were listed for a job with no phases")
	}
}
