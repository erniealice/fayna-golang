package action

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	job_phase "github.com/erniealice/fayna-golang/domain/operation/job_phase"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
)

// ---------------------------------------------------------------------------
// phaseStatusToEnum mapping tests
// ---------------------------------------------------------------------------

func TestPhaseStatusToEnum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  jobphasepb.PhaseStatus
	}{
		{input: "pending", want: jobphasepb.PhaseStatus_PHASE_STATUS_PENDING},
		{input: "active", want: jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE},
		{input: "completed", want: jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED},
		{input: "PHASE_STATUS_PENDING", want: jobphasepb.PhaseStatus_PHASE_STATUS_PENDING},
		{input: "PHASE_STATUS_ACTIVE", want: jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE},
		{input: "PHASE_STATUS_COMPLETED", want: jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED},
		{input: "", want: jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED},
		{input: "unknown", want: jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED},
		{input: "PENDING", want: jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED}, // case-sensitive
		{input: "Active", want: jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED},  // case-sensitive
	}

	for _, tt := range tests {
		tt := tt
		t.Run("status_"+tt.input, func(t *testing.T) {
			t.Parallel()
			got := phaseStatusToEnum(tt.input)
			if got != tt.want {
				t.Fatalf("phaseStatusToEnum(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPhaseStatusToEnum_AllEnumsCovered(t *testing.T) {
	t.Parallel()

	// Every non-UNSPECIFIED enum value should be reachable from the switch.
	wantMapped := map[jobphasepb.PhaseStatus]string{
		jobphasepb.PhaseStatus_PHASE_STATUS_PENDING:   "pending",
		jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE:    "active",
		jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED: "completed",
	}

	for wantEnum, input := range wantMapped {
		got := phaseStatusToEnum(input)
		if got != wantEnum {
			t.Fatalf("phaseStatusToEnum(%q) = %v, want %v", input, got, wantEnum)
		}
	}
}

// ---------------------------------------------------------------------------
// phaseStatusToEnum adversarial inputs
// ---------------------------------------------------------------------------

func TestPhaseStatusToEnum_Adversarial(t *testing.T) {
	t.Parallel()

	adversarial := []string{
		"<script>alert(1)</script>",
		"'; DROP TABLE job_phase; --",
		"\x00",
		strings.Repeat("x", 100000),
		" pending", // leading space
		"pending ", // trailing space
		"Pending",
		"ACTIVE",
		"COMPLETED",
		"phase_status_pending",
	}

	for _, input := range adversarial {
		got := phaseStatusToEnum(input)
		if got != jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED {
			t.Fatalf("phaseStatusToEnum(%q) = %v, want PHASE_STATUS_UNSPECIFIED", input, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Helper: build minimal Deps with mock functions
// ---------------------------------------------------------------------------

func testPhaseLabels() job_phase.Labels {
	return job_phase.Labels{
		Errors: job_phase.ErrorLabels{
			PermissionDenied: "permission denied",
			NotFound:         "not found",
			IDRequired:       "Phase ID is required",
		},
	}
}

func testPhaseRoutes() job_phase.Routes {
	return job_phase.DefaultRoutes()
}

// ctxWithPerms returns a context that carries the given permission codes.
func ctxWithPerms(perms ...string) context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(perms))
}

// ctxNoPerm returns a context with an empty permission set (denies everything).
func ctxNoPerm() context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(nil))
}

// makePhaseReadResp builds a minimal ReadJobPhaseResponse with the given phase values.
func makePhaseReadResp(id, jobID, name string) *jobphasepb.ReadJobPhaseResponse {
	return &jobphasepb.ReadJobPhaseResponse{
		Data: []*jobphasepb.JobPhase{
			{Id: id, JobId: jobID, Name: name},
		},
	}
}

// ---------------------------------------------------------------------------
// NewAddAction tests
// ---------------------------------------------------------------------------

func TestNewAddAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewAddAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/add", nil)
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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		CreateJobPhase: func(_ context.Context, req *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error) {
			createCalled = true
			return &jobphasepb.CreateJobPhaseResponse{
				Data: []*jobphasepb.JobPhase{{Id: "phase-123", Name: req.Data.Name}},
			}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("job_id", "job-1")
	form.Set("name", "Phase One")
	form.Set("status", "PHASE_STATUS_PENDING")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:create"), viewCtx)

	if !createCalled {
		t.Fatal("CreateJobPhase was not called")
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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		CreateJobPhase: func(_ context.Context, _ *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error) {
			return nil, errors.New("db error")
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "Phase One")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:create"), viewCtx)

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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		CreateJobPhase: func(_ context.Context, _ *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error) {
			// Return success but with empty Data slice (no ID).
			return &jobphasepb.CreateJobPhaseResponse{Data: nil}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "Phase One")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:create"), viewCtx)

	// When no ID is returned, falls through to HTMXSuccess.
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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/delete?id=phase-1", nil)
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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		DeleteJobPhase: func(_ context.Context, req *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error) {
			deletedID = req.Data.Id
			return &jobphasepb.DeleteJobPhaseResponse{Success: true}, nil
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/delete?id=phase-42", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "phase-42" {
		t.Fatalf("DeleteJobPhase called with ID %q, want %q", deletedID, "phase-42")
	}
	if trigger := result.Headers["HX-Trigger"]; trigger == "" {
		t.Fatal("expected HX-Trigger header on successful delete")
	}
}

func TestNewDeleteAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/delete", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Phase ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Phase ID is required")
	}
}

func TestNewDeleteAction_IDFromFormBody(t *testing.T) {
	t.Parallel()

	var deletedID string
	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		DeleteJobPhase: func(_ context.Context, req *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error) {
			deletedID = req.Data.Id
			return &jobphasepb.DeleteJobPhaseResponse{Success: true}, nil
		},
	}

	v := NewDeleteAction(deps)

	form := url.Values{}
	form.Set("id", "phase-from-form")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "phase-from-form" {
		t.Fatalf("DeleteJobPhase called with ID %q, want %q", deletedID, "phase-from-form")
	}
}

func TestNewDeleteAction_DeleteError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		DeleteJobPhase: func(_ context.Context, _ *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error) {
			return nil, errors.New("foreign key constraint violation")
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/delete?id=phase-99", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:delete"), viewCtx)

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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "phase-1")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-delete", strings.NewReader(form.Encode()))
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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "No IDs provided" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "No IDs provided")
	}
}

func TestNewBulkDeleteAction_DuplicateIDs(t *testing.T) {
	t.Parallel()

	var deletedIDs []string
	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		DeleteJobPhase: func(_ context.Context, req *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error) {
			deletedIDs = append(deletedIDs, req.Data.Id)
			return &jobphasepb.DeleteJobPhaseResponse{Success: true}, nil
		},
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "phase-1")
	form.Add("id", "phase-1") // duplicate
	form.Add("id", "phase-2")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	// Handler does not deduplicate — calls DeleteJobPhase for each ID including duplicates.
	if len(deletedIDs) != 3 {
		t.Fatalf("DeleteJobPhase called %d times, want 3 (duplicates not deduplicated)", len(deletedIDs))
	}
}

func TestNewBulkDeleteAction_PartialFailure(t *testing.T) {
	t.Parallel()

	var deletedIDs []string
	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		DeleteJobPhase: func(_ context.Context, req *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error) {
			if req.Data.Id == "phase-bad" {
				return nil, errors.New("cannot delete")
			}
			deletedIDs = append(deletedIDs, req.Data.Id)
			return &jobphasepb.DeleteJobPhaseResponse{Success: true}, nil
		},
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "phase-1")
	form.Add("id", "phase-bad")
	form.Add("id", "phase-2")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:delete"), viewCtx)

	// Handler continues past errors and still returns success.
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d (handler ignores individual errors)", result.StatusCode, http.StatusOK)
	}
	if len(deletedIDs) != 2 {
		t.Fatalf("Successfully deleted %d phases, want 2", len(deletedIDs))
	}
}

// ---------------------------------------------------------------------------
// NewSetStatusAction tests
// ---------------------------------------------------------------------------

func TestNewSetStatusAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/set-status?id=phase-1&status=active", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewSetStatusAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/set-status?status=active", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Phase ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Phase ID is required")
	}
}

func TestNewSetStatusAction_MissingStatus(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/set-status?id=phase-1", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Status is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Status is required")
	}
}

func TestNewSetStatusAction_InvalidStatus(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewSetStatusAction(deps)

	// Invalid status string maps to PHASE_STATUS_UNSPECIFIED, which is rejected.
	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/set-status?id=phase-1&status=BOGUS", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d (invalid status should be rejected)", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Invalid phase status" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Invalid phase status")
	}
}

func TestNewSetStatusAction_UpdateError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		ReadJobPhase: func(_ context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error) {
			return makePhaseReadResp(req.Data.Id, "job-1", "Phase One"), nil
		},
		UpdateJobPhase: func(_ context.Context, _ *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error) {
			return nil, errors.New("invalid state transition")
		},
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/set-status?id=phase-1&status=active", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "invalid state transition" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "invalid state transition")
	}
}

func TestNewSetStatusAction_Success(t *testing.T) {
	t.Parallel()

	var capturedStatus jobphasepb.PhaseStatus
	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		ReadJobPhase: func(_ context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error) {
			return makePhaseReadResp(req.Data.Id, "job-1", "Phase One"), nil
		},
		UpdateJobPhase: func(_ context.Context, req *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error) {
			capturedStatus = req.Data.Status
			return &jobphasepb.UpdateJobPhaseResponse{}, nil
		},
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/set-status?id=phase-1&status=completed", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	// Returns 204 with HX-Redirect on success.
	if result.StatusCode != http.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusNoContent)
	}
	if capturedStatus != jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED {
		t.Fatalf("Status = %v, want PHASE_STATUS_COMPLETED", capturedStatus)
	}
	if redirect := result.Headers["HX-Redirect"]; redirect == "" {
		t.Fatal("expected HX-Redirect header on successful set-status")
	}
}

func TestNewSetStatusAction_StatusFromForm(t *testing.T) {
	t.Parallel()

	var capturedID string
	var capturedStatus jobphasepb.PhaseStatus
	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		ReadJobPhase: func(_ context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error) {
			return makePhaseReadResp(req.Data.Id, "job-1", "Phase One"), nil
		},
		UpdateJobPhase: func(_ context.Context, req *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error) {
			capturedID = req.Data.Id
			capturedStatus = req.Data.Status
			return &jobphasepb.UpdateJobPhaseResponse{}, nil
		},
	}

	v := NewSetStatusAction(deps)

	// No query params — ID and status from form body.
	form := url.Values{}
	form.Set("id", "phase-form")
	form.Set("status", "PHASE_STATUS_ACTIVE")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	if result.StatusCode != http.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusNoContent)
	}
	if capturedID != "phase-form" {
		t.Fatalf("ID = %q, want %q", capturedID, "phase-form")
	}
	if capturedStatus != jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE {
		t.Fatalf("Status = %v, want PHASE_STATUS_ACTIVE", capturedStatus)
	}
}

// ---------------------------------------------------------------------------
// NewBulkSetStatusAction tests
// ---------------------------------------------------------------------------

func TestNewBulkSetStatusAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "phase-1")
	form.Set("target_status", "active")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewBulkSetStatusAction_NoIDs(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Set("target_status", "active")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "No IDs provided" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "No IDs provided")
	}
}

func TestNewBulkSetStatusAction_MissingTargetStatus(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "phase-1")
	form.Add("id", "phase-2")
	// target_status omitted

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

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
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		ReadJobPhase: func(_ context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error) {
			return makePhaseReadResp(req.Data.Id, "job-1", "Phase"), nil
		},
		UpdateJobPhase: func(_ context.Context, req *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error) {
			updatedIDs = append(updatedIDs, req.Data.Id)
			return &jobphasepb.UpdateJobPhaseResponse{}, nil
		},
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "phase-1")
	form.Add("id", "phase-2")
	form.Set("target_status", "PHASE_STATUS_COMPLETED")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if len(updatedIDs) != 2 {
		t.Fatalf("UpdateJobPhase called %d times, want 2", len(updatedIDs))
	}
}

func TestNewBulkSetStatusAction_PartialFailure(t *testing.T) {
	t.Parallel()

	var successIDs []string
	deps := &Deps{
		Routes: testPhaseRoutes(),
		Labels: testPhaseLabels(),
		ReadJobPhase: func(_ context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error) {
			return makePhaseReadResp(req.Data.Id, "job-1", "Phase"), nil
		},
		UpdateJobPhase: func(_ context.Context, req *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error) {
			if req.Data.Id == "phase-fail" {
				return nil, errors.New("update rejected")
			}
			successIDs = append(successIDs, req.Data.Id)
			return &jobphasepb.UpdateJobPhaseResponse{}, nil
		},
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "phase-1")
	form.Add("id", "phase-fail")
	form.Add("id", "phase-2")
	form.Set("target_status", "active")

	req := httptest.NewRequest(http.MethodPost, "/action/job-phase/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job_phase:update"), viewCtx)

	// Handler continues past errors and returns success.
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d (handler ignores individual errors)", result.StatusCode, http.StatusOK)
	}
	if len(successIDs) != 2 {
		t.Fatalf("Successfully updated %d phases, want 2", len(successIDs))
	}
}
