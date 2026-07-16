package phase_summary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
)

// FAY-7: the phase outcome summary detail page (reached from the section
// grid's /jobs/detail/{id}/phase/{phase_id}/summary link) had no Layer-3 read
// gate. These pin the fail-closed / granted-principal behavior for the
// job_outcome_summary:list verb — the same gate used by the sibling
// outcome_summary views (list, client_card, section, template_settings) and
// its job_summary sibling.

func phaseSummaryFn(s *phasesumpb.PhaseOutcomeSummary) func(context.Context, *phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error) {
	return func(context.Context, *phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error) {
		return &phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse{PhaseOutcomeSummary: s}, nil
	}
}

func newPhaseSummaryRequest(id, phaseID string) *http.Request {
	req := httptest.NewRequest("GET", "/jobs/detail/"+id+"/phase/"+phaseID+"/summary", nil)
	req.SetPathValue("id", id)
	req.SetPathValue("phase_id", phaseID)
	return req
}

// TestNewView_Ungranted_Forbidden: a principal lacking job_outcome_summary:list
// must be rejected BEFORE the GetPhaseOutcomeSummaryByJobPhase read (fail-closed).
func TestNewView_Ungranted_Forbidden(t *testing.T) {
	called := false
	deps := &Deps{
		GetPhaseOutcomeSummaryByJobPhase: func(context.Context, *phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error) {
			called = true
			return &phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse{}, nil
		},
	}
	v := NewView(deps)

	req := newPhaseSummaryRequest("job-1", "phase-1")
	ctx := view.WithUserPermissions(req.Context(), types.NewEmptyUserPermissions())
	res := v.Handle(ctx, &view.ViewContext{Request: req, CurrentPath: req.URL.Path, CacheVersion: "test"})

	if res.Template != "forbidden" {
		t.Fatalf("Template = %q, want %q", res.Template, "forbidden")
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, http.StatusForbidden)
	}
	if called {
		t.Fatal("GetPhaseOutcomeSummaryByJobPhase must not be called for an ungranted principal (fail-closed BEFORE the read)")
	}
}

// TestNewView_Granted_Renders: a principal holding job_outcome_summary:list
// renders the page normally.
func TestNewView_Granted_Renders(t *testing.T) {
	summary := &phasesumpb.PhaseOutcomeSummary{Id: "psum-1", JobId: "job-1", JobPhaseId: "phase-1", Active: true}
	deps := &Deps{GetPhaseOutcomeSummaryByJobPhase: phaseSummaryFn(summary)}
	v := NewView(deps)

	req := newPhaseSummaryRequest("job-1", "phase-1")
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_outcome_summary:list"}))
	res := v.Handle(ctx, &view.ViewContext{Request: req, CurrentPath: req.URL.Path, CacheVersion: "test"})

	if res.Template != "phase-outcome-summary" {
		t.Fatalf("Template = %q, want %q", res.Template, "phase-outcome-summary")
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d (via view.OK)", res.StatusCode, http.StatusOK)
	}
	pd, ok := res.Data.(*PageData)
	if !ok {
		t.Fatalf("Data type = %T, want *PageData", res.Data)
	}
	if pd.Summary["id"] != "psum-1" {
		t.Fatalf("Summary[id] = %v, want %q", pd.Summary["id"], "psum-1")
	}
}
