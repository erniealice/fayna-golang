package action

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// ---------------------------------------------------------------------------
// jobStatusToEnum mapping tests
// ---------------------------------------------------------------------------

func TestJobStatusToEnum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  enums.JobStatus
	}{
		{input: "draft", want: enums.JobStatus_JOB_STATUS_DRAFT},
		{input: "pending", want: enums.JobStatus_JOB_STATUS_PENDING},
		{input: "active", want: enums.JobStatus_JOB_STATUS_ACTIVE},
		{input: "paused", want: enums.JobStatus_JOB_STATUS_PAUSED},
		{input: "completed", want: enums.JobStatus_JOB_STATUS_COMPLETED},
		{input: "closed", want: enums.JobStatus_JOB_STATUS_CLOSED},
		{input: "", want: enums.JobStatus_JOB_STATUS_UNSPECIFIED},
		{input: "unknown", want: enums.JobStatus_JOB_STATUS_UNSPECIFIED},
		{input: "DRAFT", want: enums.JobStatus_JOB_STATUS_UNSPECIFIED}, // case-sensitive
	}

	for _, tt := range tests {
		tt := tt
		t.Run("status_"+tt.input, func(t *testing.T) {
			t.Parallel()
			got := jobStatusToEnum(tt.input)
			if got != tt.want {
				t.Fatalf("jobStatusToEnum(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestJobStatusToEnum_AllEnumsCovered(t *testing.T) {
	t.Parallel()

	// Every non-UNSPECIFIED enum value should be reachable from the switch.
	wantMapped := map[enums.JobStatus]string{
		enums.JobStatus_JOB_STATUS_DRAFT:     "draft",
		enums.JobStatus_JOB_STATUS_PENDING:   "pending",
		enums.JobStatus_JOB_STATUS_ACTIVE:    "active",
		enums.JobStatus_JOB_STATUS_PAUSED:    "paused",
		enums.JobStatus_JOB_STATUS_COMPLETED: "completed",
		enums.JobStatus_JOB_STATUS_CLOSED:    "closed",
	}

	for wantEnum, input := range wantMapped {
		got := jobStatusToEnum(input)
		if got != wantEnum {
			t.Fatalf("jobStatusToEnum(%q) = %v, want %v", input, got, wantEnum)
		}
	}
}

// ---------------------------------------------------------------------------
// Helper: build a minimal Deps with mock functions
// ---------------------------------------------------------------------------

func testLabels() fayna.JobLabels {
	return fayna.JobLabels{
		Errors: fayna.JobErrorLabels{
			PermissionDenied: "permission denied",
			NotFound:         "not found",
		},
	}
}

func testRoutes() fayna.JobRoutes {
	return fayna.DefaultJobRoutes()
}

// ctxWithPerms returns a context that carries the given permission codes.
func ctxWithPerms(perms ...string) context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(perms))
}

// ctxNoPerm returns a context with an empty permission set (denies everything).
func ctxNoPerm() context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(nil))
}

// ---------------------------------------------------------------------------
// NewAddAction tests
// ---------------------------------------------------------------------------

func TestNewAddAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewAddAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/add", nil)
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
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateJob: func(_ context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) {
			createCalled = true
			return &jobpb.CreateJobResponse{
				Data: []*jobpb.Job{{Id: "job-123", Name: req.Data.Name}},
			}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "Test Job")
	form.Set("client_id", "c1")
	form.Set("location_id", "l1")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:create"), viewCtx)

	if !createCalled {
		t.Fatal("CreateJob was not called")
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	// When a new ID is returned, the handler redirects via HX-Redirect.
	if redirect := result.Headers["HX-Redirect"]; redirect == "" {
		t.Fatal("expected HX-Redirect header for successful create with returned ID")
	}
}

func TestNewAddAction_POST_CreateError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateJob: func(_ context.Context, _ *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) {
			return nil, errors.New("db error")
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "Test Job")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:create"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "db error" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "db error")
	}
}

// ---------------------------------------------------------------------------
// NewDeleteAction tests
// ---------------------------------------------------------------------------

func TestNewDeleteAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/delete?id=job-1", nil)
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
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteJob: func(_ context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error) {
			deletedID = req.Data.Id
			return &jobpb.DeleteJobResponse{Success: true}, nil
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/delete?id=job-42", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "job-42" {
		t.Fatalf("DeleteJob called with ID %q, want %q", deletedID, "job-42")
	}
	if trigger := result.Headers["HX-Trigger"]; trigger == "" {
		t.Fatal("expected HX-Trigger header on successful delete")
	}
}

func TestNewDeleteAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/delete", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "ID is required")
	}
}

// ---------------------------------------------------------------------------
// Negative / defensive test cases
// ---------------------------------------------------------------------------

func TestNewAddAction_POST_EmptyName(t *testing.T) {
	t.Parallel()

	var captured *jobpb.CreateJobRequest
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateJob: func(_ context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) {
			captured = req
			return &jobpb.CreateJobResponse{
				Data: []*jobpb.Job{{Id: "job-456", Name: req.Data.Name}},
			}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "")
	form.Set("client_id", "c1")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("job:create"), viewCtx)

	// The handler passes the empty name through to CreateJob without validation.
	// This test documents the current passthrough behavior.
	if captured == nil {
		t.Fatal("CreateJob was not called")
	}
	if captured.Data.Name != "" {
		t.Fatalf("Name = %q, want empty string", captured.Data.Name)
	}
}

func TestNewAddAction_POST_SpecialCharsInName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"html_script_tag", `<script>alert('xss')</script>`},
		{"sql_injection", `'; DROP TABLE jobs; --`},
		{"html_entity", `<img src=x onerror=alert(1)>`},
		{"null_byte", "name\x00with\x00nulls"},
		{"unicode_rtl_override", "job\u202Egnirts"},
		{"very_long_name", strings.Repeat("a", 10000)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var capturedName string
			deps := &Deps{
				Routes: testRoutes(),
				Labels: testLabels(),
				CreateJob: func(_ context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) {
					capturedName = req.Data.Name
					return &jobpb.CreateJobResponse{
						Data: []*jobpb.Job{{Id: "job-789"}},
					}, nil
				},
			}

			v := NewAddAction(deps)

			form := url.Values{}
			form.Set("name", tt.input)

			req := httptest.NewRequest(http.MethodPost, "/action/jobs/add", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			viewCtx := &view.ViewContext{Request: req}

			result := v.Handle(ctxWithPerms("job:create"), viewCtx)

			// Handler should pass the value through to CreateJob without panicking.
			if result.StatusCode != http.StatusOK {
				t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
			}
			if capturedName != tt.input {
				t.Fatalf("CreateJob received Name = %q, want %q", capturedName, tt.input)
			}
		})
	}
}

func TestNewAddAction_POST_MissingAllFields(t *testing.T) {
	t.Parallel()

	var captured *jobpb.CreateJobRequest
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateJob: func(_ context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) {
			captured = req
			return &jobpb.CreateJobResponse{
				Data: []*jobpb.Job{{Id: "job-empty"}},
			}, nil
		},
	}

	v := NewAddAction(deps)

	// POST with empty form body
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/action/jobs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:create"), viewCtx)

	if captured == nil {
		t.Fatal("CreateJob was not called")
	}
	if captured.Data.Name != "" {
		t.Fatalf("Name = %q, want empty string", captured.Data.Name)
	}
	// Should still succeed (handler doesn't validate required fields).
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestNewAddAction_POST_NoResponseData(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateJob: func(_ context.Context, _ *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error) {
			// Return success but with empty Data slice (no ID).
			return &jobpb.CreateJobResponse{Data: nil}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("name", "Test Job")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:create"), viewCtx)

	// When no ID is returned, the handler falls through to HTMXSuccess.
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	// Should NOT have HX-Redirect since no new ID.
	if redirect := result.Headers["HX-Redirect"]; redirect != "" {
		t.Fatalf("expected no HX-Redirect when response has no data, got %q", redirect)
	}
}

func TestNewAddAction_GET_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewAddAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/jobs/add", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "permission denied" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "permission denied")
	}
}

// ---------------------------------------------------------------------------
// NewEditAction negative tests
// ---------------------------------------------------------------------------

func TestNewEditAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewEditAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/jobs/edit/job-1", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewEditAction_GET_ReadError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		ReadJob: func(_ context.Context, _ *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error) {
			return nil, errors.New("not found in database")
		},
	}

	v := NewEditAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/jobs/edit/job-1", nil)
	req.SetPathValue("id", "job-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "not found" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "not found")
	}
}

func TestNewEditAction_GET_EmptyReadData(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		ReadJob: func(_ context.Context, _ *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error) {
			return &jobpb.ReadJobResponse{Data: []*jobpb.Job{}}, nil
		},
	}

	v := NewEditAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/jobs/edit/job-1", nil)
	req.SetPathValue("id", "job-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "not found" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "not found")
	}
}

func TestNewEditAction_POST_EmptyID(t *testing.T) {
	t.Parallel()

	var capturedID string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateJob: func(_ context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) {
			capturedID = req.Data.Id
			return &jobpb.UpdateJobResponse{}, nil
		},
	}

	v := NewEditAction(deps)

	form := url.Values{}
	form.Set("name", "Updated Job")

	// No "id" path value set — simulates empty ID.
	req := httptest.NewRequest(http.MethodPost, "/action/jobs/edit/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("job:update"), viewCtx)

	// Documents that the handler passes the empty ID to UpdateJob.
	if capturedID != "" {
		t.Fatalf("expected empty ID to be passed through, got %q", capturedID)
	}
}

func TestNewEditAction_POST_UpdateError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateJob: func(_ context.Context, _ *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) {
			return nil, errors.New("update failed: conflict")
		},
	}

	v := NewEditAction(deps)

	form := url.Values{}
	form.Set("name", "Updated Job")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/edit/job-1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "job-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "update failed: conflict" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "update failed: conflict")
	}
}

// ---------------------------------------------------------------------------
// NewDeleteAction negative tests (additional)
// ---------------------------------------------------------------------------

func TestNewDeleteAction_EmptyIDInQueryAndForm(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewDeleteAction(deps)

	// Both query param and form body have empty "id".
	form := url.Values{}
	form.Set("id", "")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/delete?id=", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "ID is required")
	}
}

func TestNewDeleteAction_DeleteError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteJob: func(_ context.Context, _ *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error) {
			return nil, errors.New("foreign key constraint violation")
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/delete?id=job-99", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "foreign key constraint violation" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "foreign key constraint violation")
	}
}

func TestNewDeleteAction_IDFromFormBody(t *testing.T) {
	t.Parallel()

	var deletedID string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteJob: func(_ context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error) {
			deletedID = req.Data.Id
			return &jobpb.DeleteJobResponse{Success: true}, nil
		},
	}

	v := NewDeleteAction(deps)

	// No query param, ID from form body only.
	form := url.Values{}
	form.Set("id", "job-from-form")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "job-from-form" {
		t.Fatalf("DeleteJob called with ID %q, want %q", deletedID, "job-from-form")
	}
}

// ---------------------------------------------------------------------------
// NewSetStatusAction negative tests
// ---------------------------------------------------------------------------

func TestNewSetStatusAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/set-status?id=job-1&status=active", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewSetStatusAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/set-status?status=active", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "ID is required")
	}
}

func TestNewSetStatusAction_MissingStatus(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/set-status?id=job-1", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Status is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Status is required")
	}
}

func TestNewSetStatusAction_InvalidStatus(t *testing.T) {
	t.Parallel()

	var capturedStatus enums.JobStatus
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateJob: func(_ context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) {
			capturedStatus = req.Data.Status
			return &jobpb.UpdateJobResponse{}, nil
		},
	}

	v := NewSetStatusAction(deps)

	// Invalid status string maps to JOB_STATUS_UNSPECIFIED via jobStatusToEnum.
	req := httptest.NewRequest(http.MethodPost, "/action/jobs/set-status?id=job-1&status=BOGUS", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if capturedStatus != enums.JobStatus_JOB_STATUS_UNSPECIFIED {
		t.Fatalf("Status = %v, want JOB_STATUS_UNSPECIFIED", capturedStatus)
	}
}

func TestNewSetStatusAction_UpdateError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateJob: func(_ context.Context, _ *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) {
			return nil, errors.New("invalid state transition")
		},
	}

	v := NewSetStatusAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/set-status?id=job-1&status=active", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "invalid state transition" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "invalid state transition")
	}
}

func TestNewSetStatusAction_StatusFromForm(t *testing.T) {
	t.Parallel()

	var capturedStatus enums.JobStatus
	var capturedID string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateJob: func(_ context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) {
			capturedID = req.Data.Id
			capturedStatus = req.Data.Status
			return &jobpb.UpdateJobResponse{}, nil
		},
	}

	v := NewSetStatusAction(deps)

	// No query params — ID and status from form body.
	form := url.Values{}
	form.Set("id", "job-form")
	form.Set("status", "paused")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if capturedID != "job-form" {
		t.Fatalf("ID = %q, want %q", capturedID, "job-form")
	}
	if capturedStatus != enums.JobStatus_JOB_STATUS_PAUSED {
		t.Fatalf("Status = %v, want JOB_STATUS_PAUSED", capturedStatus)
	}
}

// ---------------------------------------------------------------------------
// NewBulkDeleteAction negative tests
// ---------------------------------------------------------------------------

func TestNewBulkDeleteAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "job-1")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-delete", strings.NewReader(form.Encode()))
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
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

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
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteJob: func(_ context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error) {
			deletedIDs = append(deletedIDs, req.Data.Id)
			return &jobpb.DeleteJobResponse{Success: true}, nil
		},
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "job-1")
	form.Add("id", "job-1") // duplicate
	form.Add("id", "job-2")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	// Handler does not deduplicate — it calls DeleteJob for each ID including duplicates.
	if len(deletedIDs) != 3 {
		t.Fatalf("DeleteJob called %d times, want 3 (duplicates not deduplicated)", len(deletedIDs))
	}
}

func TestNewBulkDeleteAction_PartialFailure(t *testing.T) {
	t.Parallel()

	var deletedIDs []string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteJob: func(_ context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error) {
			if req.Data.Id == "job-bad" {
				return nil, errors.New("cannot delete")
			}
			deletedIDs = append(deletedIDs, req.Data.Id)
			return &jobpb.DeleteJobResponse{Success: true}, nil
		},
	}

	v := NewBulkDeleteAction(deps)

	form := url.Values{}
	form.Add("id", "job-1")
	form.Add("id", "job-bad")
	form.Add("id", "job-2")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:delete"), viewCtx)

	// Handler continues past errors and still returns success.
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d (handler ignores individual errors)", result.StatusCode, http.StatusOK)
	}
	if len(deletedIDs) != 2 {
		t.Fatalf("Successfully deleted %d jobs, want 2", len(deletedIDs))
	}
}

// ---------------------------------------------------------------------------
// NewBulkSetStatusAction negative tests
// ---------------------------------------------------------------------------

func TestNewBulkSetStatusAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "job-1")
	form.Set("target_status", "active")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-set-status", strings.NewReader(form.Encode()))
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
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Set("target_status", "active")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

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
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "job-1")
	form.Add("id", "job-2")
	// target_status omitted

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Target status is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Target status is required")
	}
}

func TestNewBulkSetStatusAction_InvalidTargetStatus(t *testing.T) {
	t.Parallel()

	var capturedStatuses []enums.JobStatus
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateJob: func(_ context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) {
			capturedStatuses = append(capturedStatuses, req.Data.Status)
			return &jobpb.UpdateJobResponse{}, nil
		},
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "job-1")
	form.Add("id", "job-2")
	form.Set("target_status", "NONEXISTENT")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	// All statuses should be UNSPECIFIED since the input is invalid.
	for i, s := range capturedStatuses {
		if s != enums.JobStatus_JOB_STATUS_UNSPECIFIED {
			t.Fatalf("capturedStatuses[%d] = %v, want JOB_STATUS_UNSPECIFIED", i, s)
		}
	}
}

func TestNewBulkSetStatusAction_PartialFailure(t *testing.T) {
	t.Parallel()

	var successIDs []string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateJob: func(_ context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error) {
			if req.Data.Id == "job-fail" {
				return nil, errors.New("update rejected")
			}
			successIDs = append(successIDs, req.Data.Id)
			return &jobpb.UpdateJobResponse{}, nil
		},
	}

	v := NewBulkSetStatusAction(deps)

	form := url.Values{}
	form.Add("id", "job-1")
	form.Add("id", "job-fail")
	form.Add("id", "job-2")
	form.Set("target_status", "active")

	req := httptest.NewRequest(http.MethodPost, "/action/jobs/bulk-set-status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("job:update"), viewCtx)

	// Handler continues past errors and returns success.
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if len(successIDs) != 2 {
		t.Fatalf("Successfully updated %d jobs, want 2", len(successIDs))
	}
}

// ---------------------------------------------------------------------------
// jobStatusToEnum adversarial inputs
// ---------------------------------------------------------------------------

func TestJobStatusToEnum_Adversarial(t *testing.T) {
	t.Parallel()

	adversarial := []string{
		"<script>alert(1)</script>",
		"'; DROP TABLE jobs; --",
		"\x00",
		strings.Repeat("x", 100000),
		" draft", // leading space
		"draft ", // trailing space
		"Draft",
		"ACTIVE",
	}

	for _, input := range adversarial {
		got := jobStatusToEnum(input)
		if got != enums.JobStatus_JOB_STATUS_UNSPECIFIED {
			t.Fatalf("jobStatusToEnum(%q) = %v, want JOB_STATUS_UNSPECIFIED", input, got)
		}
	}
}
