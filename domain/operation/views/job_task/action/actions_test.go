package action

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	operation "github.com/erniealice/fayna-golang/domain/operation"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// ---------------------------------------------------------------------------
// taskStatusToEnum mapping tests
// ---------------------------------------------------------------------------

func TestTaskStatusToEnum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  jobtaskpb.TaskStatus
	}{
		{input: "pending", want: jobtaskpb.TaskStatus_TASK_STATUS_PENDING},
		{input: "in_progress", want: jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS},
		{input: "completed", want: jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED},
		{input: "skipped", want: jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED},
		{input: "hold", want: jobtaskpb.TaskStatus_TASK_STATUS_HOLD},
		{input: "rework", want: jobtaskpb.TaskStatus_TASK_STATUS_REWORK},
		{input: "TASK_STATUS_PENDING", want: jobtaskpb.TaskStatus_TASK_STATUS_PENDING},
		{input: "TASK_STATUS_IN_PROGRESS", want: jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS},
		{input: "TASK_STATUS_COMPLETED", want: jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED},
		{input: "TASK_STATUS_SKIPPED", want: jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED},
		{input: "TASK_STATUS_HOLD", want: jobtaskpb.TaskStatus_TASK_STATUS_HOLD},
		{input: "TASK_STATUS_REWORK", want: jobtaskpb.TaskStatus_TASK_STATUS_REWORK},
		{input: "", want: jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED},
		{input: "unknown", want: jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED},
		{input: "PENDING", want: jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED},   // case-sensitive
		{input: "Completed", want: jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED}, // case-sensitive
	}

	for _, tt := range tests {
		tt := tt
		t.Run("status_"+tt.input, func(t *testing.T) {
			t.Parallel()
			got := taskStatusToEnum(tt.input)
			if got != tt.want {
				t.Fatalf("taskStatusToEnum(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTaskStatusToEnum_AllEnumsCovered(t *testing.T) {
	t.Parallel()

	// Every non-UNSPECIFIED enum value should be reachable from the switch.
	wantMapped := map[jobtaskpb.TaskStatus]string{
		jobtaskpb.TaskStatus_TASK_STATUS_PENDING:     "pending",
		jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS: "in_progress",
		jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED:   "completed",
		jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED:     "skipped",
		jobtaskpb.TaskStatus_TASK_STATUS_HOLD:        "hold",
		jobtaskpb.TaskStatus_TASK_STATUS_REWORK:      "rework",
	}

	for wantEnum, input := range wantMapped {
		got := taskStatusToEnum(input)
		if got != wantEnum {
			t.Fatalf("taskStatusToEnum(%q) = %v, want %v", input, got, wantEnum)
		}
	}
}

// ---------------------------------------------------------------------------
// taskStatusToEnum adversarial inputs
// ---------------------------------------------------------------------------

func TestTaskStatusToEnum_Adversarial(t *testing.T) {
	t.Parallel()

	adversarial := []string{
		"<script>alert(1)</script>",
		"'; DROP TABLE job_task; --",
		"\x00",
		strings.Repeat("x", 100000),
		" pending", // leading space
		"pending ", // trailing space
		"Pending",
		"IN_PROGRESS",
		"COMPLETED",
		"task_status_pending",
	}

	for _, input := range adversarial {
		got := taskStatusToEnum(input)
		if got != jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED {
			t.Fatalf("taskStatusToEnum(%q) = %v, want TASK_STATUS_UNSPECIFIED", input, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Helper: build minimal Deps with mock functions
// ---------------------------------------------------------------------------

func testTaskLabels() operation.JobTaskLabels {
	return operation.JobTaskLabels{
		Errors: operation.JobTaskErrorLabels{
			PermissionDenied: "permission denied",
			NotFound:         "not found",
			IDRequired:       "Task ID is required",
		},
	}
}

func testTaskRoutes() operation.JobTaskRoutes {
	return operation.DefaultJobTaskRoutes()
}

// ctxWithPerms returns a context that carries the given permission codes.
func ctxWithPerms(perms ...string) context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(perms))
}

// ctxNoPerm returns a context with an empty permission set (denies everything).
func ctxNoPerm() context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(nil))
}

// makeTaskReadResp builds a minimal ReadJobTaskResponse with the given task values.
func makeTaskReadResp(id, phaseID, name string) *jobtaskpb.ReadJobTaskResponse {
	return &jobtaskpb.ReadJobTaskResponse{
		Data: []*jobtaskpb.JobTask{
			{Id: id, JobPhaseId: phaseID, Name: name},
		},
	}
}

// ---------------------------------------------------------------------------
// NewAddAction tests
// ---------------------------------------------------------------------------

func TestNewAddAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewAddAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/add", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "permission denied" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "permission denied")
	}
}

func TestNewAddAction_POST_Success(t *testing.T) {
	t.Parallel()

	createCalled := false
	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		CreateJobTask: func(_ context.Context, req *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error) {
			createCalled = true
			return &jobtaskpb.CreateJobTaskResponse{
				Data: []*jobtaskpb.JobTask{{Id: "task-123", Name: req.Data.Name}},
			}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("job_phase_id", "phase-1")
	form.Set("name", "Task One")
	form.Set("status", "TASK_STATUS_PENDING")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:create"), viewCtx)

	if !createCalled {
		t.Fatal("CreateJobTask was not called")
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if redirect := result.Headers["HX-Redirect"]; redirect == "" {
		t.Fatal("expected HX-Redirect header for successful create with returned ID")
	}
}

func TestNewAddAction_POST_CreateError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		CreateJobTask: func(_ context.Context, _ *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error) {
			return nil, errors.New("db error")
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "Task One")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:create"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "db error" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "db error")
	}
}

func TestNewAddAction_POST_NoResponseData(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		CreateJobTask: func(_ context.Context, _ *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error) {
			return &jobtaskpb.CreateJobTaskResponse{Data: nil}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "Task One")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:create"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if redirect := result.Headers["HX-Redirect"]; redirect != "" {
		t.Fatalf("expected no HX-Redirect when response has no data, got %q", redirect)
	}
}

// ---------------------------------------------------------------------------
// NewDeleteAction tests
// ---------------------------------------------------------------------------

func TestNewDeleteAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/delete?id=task-1", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewDeleteAction_Success(t *testing.T) {
	t.Parallel()

	deletedID := ""
	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		DeleteJobTask: func(_ context.Context, req *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error) {
			deletedID = req.Data.Id
			return &jobtaskpb.DeleteJobTaskResponse{Success: true}, nil
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/delete?id=task-42", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "task-42" {
		t.Fatalf("DeleteJobTask called with ID %q, want %q", deletedID, "task-42")
	}
	if trigger := result.Headers["HX-Trigger"]; trigger == "" {
		t.Fatal("expected HX-Trigger header on successful delete")
	}
}

func TestNewDeleteAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/delete", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Task ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Task ID is required")
	}
}

func TestNewDeleteAction_IDFromFormBody(t *testing.T) {
	t.Parallel()

	var deletedID string
	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		DeleteJobTask: func(_ context.Context, req *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error) {
			deletedID = req.Data.Id
			return &jobtaskpb.DeleteJobTaskResponse{Success: true}, nil
		},
	}

	v := NewDeleteAction(deps)

	form := url.Values{}
	form.Set("id", "task-from-form")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "task-from-form" {
		t.Fatalf("DeleteJobTask called with ID %q, want %q", deletedID, "task-from-form")
	}
}

func TestNewDeleteAction_DeleteError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		DeleteJobTask: func(_ context.Context, _ *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error) {
			return nil, errors.New("foreign key constraint violation")
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/delete?id=task-99", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "foreign key constraint violation" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "foreign key constraint violation")
	}
}

// ---------------------------------------------------------------------------
// NewBulkDeleteAction tests
// ---------------------------------------------------------------------------

func TestNewBulkDeleteAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "task-1")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewBulkDeleteAction_NoIDs(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "No IDs provided" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "No IDs provided")
	}
}

func TestNewBulkDeleteAction_PartialFailure(t *testing.T) {
	t.Parallel()

	var deletedIDs []string
	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		DeleteJobTask: func(_ context.Context, req *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error) {
			if req.Data.Id == "task-bad" {
				return nil, errors.New("cannot delete")
			}
			deletedIDs = append(deletedIDs, req.Data.Id)
			return &jobtaskpb.DeleteJobTaskResponse{Success: true}, nil
		},
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "task-1")
	form.Add("id", "task-bad")
	form.Add("id", "task-2")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d (handler ignores individual errors)", result.StatusCode, http.StatusOK)
	}
	if len(deletedIDs) != 2 {
		t.Fatalf("Successfully deleted %d tasks, want 2", len(deletedIDs))
	}
}

// ---------------------------------------------------------------------------
// NewSetStatusAction tests
// ---------------------------------------------------------------------------

func TestNewSetStatusAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/set-status?id=task-1&status=in_progress", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewSetStatusAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/set-status?status=in_progress", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Task ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Task ID is required")
	}
}

func TestNewSetStatusAction_InvalidStatus(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/set-status?id=task-1&status=BOGUS", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d (invalid status should be rejected)", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Invalid task status" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Invalid task status")
	}
}

func TestNewSetStatusAction_Success(t *testing.T) {
	t.Parallel()

	var capturedStatus jobtaskpb.TaskStatus
	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		ReadJobTask: func(_ context.Context, req *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error) {
			return makeTaskReadResp(req.Data.Id, "phase-1", "Task One"), nil
		},
		UpdateJobTask: func(_ context.Context, req *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error) {
			capturedStatus = req.Data.Status
			return &jobtaskpb.UpdateJobTaskResponse{}, nil
		},
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/set-status?id=task-1&status=completed", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:update"), viewCtx)

	if result.StatusCode != http.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusNoContent)
	}
	if capturedStatus != jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED {
		t.Fatalf("Status = %v, want TASK_STATUS_COMPLETED", capturedStatus)
	}
	if redirect := result.Headers["HX-Redirect"]; redirect == "" {
		t.Fatal("expected HX-Redirect header on successful set-status")
	}
}

// ---------------------------------------------------------------------------
// NewBulkSetStatusAction tests
// ---------------------------------------------------------------------------

func TestNewBulkSetStatusAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "task-1")
	form.Set("target_status", "in_progress")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewBulkSetStatusAction_MissingTargetStatus(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "task-1")
	form.Add("id", "task-2")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Target status is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Target status is required")
	}
}

func TestNewBulkSetStatusAction_Success(t *testing.T) {
	t.Parallel()

	var updatedIDs []string
	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
		ReadJobTask: func(_ context.Context, req *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error) {
			return makeTaskReadResp(req.Data.Id, "phase-1", "Task"), nil
		},
		UpdateJobTask: func(_ context.Context, req *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error) {
			updatedIDs = append(updatedIDs, req.Data.Id)
			return &jobtaskpb.UpdateJobTaskResponse{}, nil
		},
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "task-1")
	form.Add("id", "task-2")
	form.Set("target_status", "TASK_STATUS_COMPLETED")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:update"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if len(updatedIDs) != 2 {
		t.Fatalf("UpdateJobTask called %d times, want 2", len(updatedIDs))
	}
}

func TestNewBulkSetStatusAction_NoIDs(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testTaskRoutes(),
		Labels: testTaskLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Set("target_status", "in_progress")

	req := httptest.NewRequest(http.MethodPost, "/action/job-task/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_task:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "No IDs provided" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "No IDs provided")
	}
}
