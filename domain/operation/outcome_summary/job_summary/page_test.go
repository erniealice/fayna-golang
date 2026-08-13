package job_summary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
)

// FAY-7: the job outcome summary detail page (reached from the section grid's
// /jobs/detail/{id}/summary link) had no Layer-3 read gate. These pin the
// fail-closed / granted-principal behavior for the job_outcome_summary:list
// list ∧ read capability — the same gate used by the sibling outcome_summary views (list,
// client_card, section, template_settings).

func jobSummaryFn(s *jobsumpb.JobOutcomeSummary) func(context.Context, *jobsumpb.GetJobOutcomeSummaryByJobRequest) (*jobsumpb.GetJobOutcomeSummaryByJobResponse, error) {
	return func(context.Context, *jobsumpb.GetJobOutcomeSummaryByJobRequest) (*jobsumpb.GetJobOutcomeSummaryByJobResponse, error) {
		return &jobsumpb.GetJobOutcomeSummaryByJobResponse{JobOutcomeSummary: s}, nil
	}
}

func newJobSummaryRequest(id string) *http.Request {
	req := httptest.NewRequest("GET", "/jobs/detail/"+id+"/summary", nil)
	req.SetPathValue("id", id)
	return req
}

// TestNewView_Ungranted_Forbidden: a principal lacking job_outcome_summary:list
// must be rejected BEFORE the GetJobOutcomeSummaryByJob read (fail-closed).
func TestNewView_Ungranted_Forbidden(t *testing.T) {
	called := false
	deps := &Deps{
		GetJobOutcomeSummaryByJob: func(context.Context, *jobsumpb.GetJobOutcomeSummaryByJobRequest) (*jobsumpb.GetJobOutcomeSummaryByJobResponse, error) {
			called = true
			return &jobsumpb.GetJobOutcomeSummaryByJobResponse{}, nil
		},
	}
	v := NewView(deps)

	req := newJobSummaryRequest("job-1")
	ctx := view.WithUserPermissions(req.Context(), types.NewEmptyUserPermissions())
	res := v.Handle(ctx, &view.ViewContext{Request: req, CurrentPath: req.URL.Path, CacheVersion: "test"})

	if res.Template != "forbidden" {
		t.Fatalf("Template = %q, want %q", res.Template, "forbidden")
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, http.StatusForbidden)
	}
	if called {
		t.Fatal("GetJobOutcomeSummaryByJob must not be called for an ungranted principal (fail-closed BEFORE the read)")
	}
}

// TestNewView_Granted_Renders: a principal holding job_outcome_summary:list
// renders the page normally.
func TestNewView_Granted_Renders(t *testing.T) {
	summary := &jobsumpb.JobOutcomeSummary{Id: "sum-1", JobId: "job-1", Active: true}
	deps := &Deps{GetJobOutcomeSummaryByJob: jobSummaryFn(summary)}
	v := NewView(deps)

	req := newJobSummaryRequest("job-1")
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_outcome_summary:list", "job_outcome_summary:read"}))
	res := v.Handle(ctx, &view.ViewContext{Request: req, CurrentPath: req.URL.Path, CacheVersion: "test"})

	if res.Template != "job-outcome-summary" {
		t.Fatalf("Template = %q, want %q", res.Template, "job-outcome-summary")
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d (via view.OK)", res.StatusCode, http.StatusOK)
	}
	pd, ok := res.Data.(*PageData)
	if !ok {
		t.Fatalf("Data type = %T, want *PageData", res.Data)
	}
	if pd.Summary["id"] != "sum-1" {
		t.Fatalf("Summary[id] = %v, want %q", pd.Summary["id"], "sum-1")
	}
}

func TestNewView_LegacyDetailRequiresListAndReadBeforeDependency(t *testing.T) {
	for _, tc := range []struct {
		name  string
		codes []string
		want  bool
	}{
		{name: "list only", codes: []string{"job_outcome_summary:list"}},
		{name: "read only", codes: []string{"job_outcome_summary:read"}},
		{name: "list and read", codes: []string{"job_outcome_summary:list", "job_outcome_summary:read"}, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			deps := &Deps{GetJobOutcomeSummaryByJob: func(context.Context, *jobsumpb.GetJobOutcomeSummaryByJobRequest) (*jobsumpb.GetJobOutcomeSummaryByJobResponse, error) {
				called = true
				return &jobsumpb.GetJobOutcomeSummaryByJobResponse{JobOutcomeSummary: &jobsumpb.JobOutcomeSummary{Id: "sum-1", JobId: "job-1"}}, nil
			}}
			req := newJobSummaryRequest("job-1")
			res := NewView(deps).Handle(view.WithUserPermissions(req.Context(), types.NewUserPermissions(tc.codes)), &view.ViewContext{Request: req, CurrentPath: req.URL.Path, CacheVersion: "test"})
			if tc.want {
				if res.StatusCode != http.StatusOK || !called {
					t.Fatalf("status/called=%d/%v, want 200/true", res.StatusCode, called)
				}
				return
			}
			if res.StatusCode != http.StatusForbidden || called {
				t.Fatalf("status/called=%d/%v, want 403/false", res.StatusCode, called)
			}
		})
	}
}
