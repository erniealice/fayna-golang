package list

import (
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	"github.com/erniealice/pyeza-golang/types"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

func ip(s jobphasepb.PhaseApprovalStatus) jobphasepb.PhaseApprovalStatus { return s }

func rollup(phaseID string, status jobphasepb.PhaseApprovalStatus, mixed, hasData, frozen bool, target, blank int32) *matrixpb.PhaseApprovalRollup {
	return &matrixpb.PhaseApprovalRollup{
		JobTemplatePhaseId: phaseID,
		Status:             status,
		Mixed:              mixed,
		HasData:            hasData,
		HardFrozen:         frozen,
		TargetCount:        target,
		BlankRequiredCount: blank,
	}
}

func phaseCol(phaseID, label, taskID string) *matrixpb.PhaseColumn {
	return &matrixpb.PhaseColumn{
		JobTemplatePhaseId: phaseID,
		Label:              label,
		Tasks:              []*matrixpb.TaskColumn{{JobTemplateTaskId: taskID}},
	}
}

const (
	sIP  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS
	sFR  = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW
	sVER = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED
	sPUB = jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED
)

// TestBuildApprovalBar pins the state-gated action affordances + chip variants +
// the D6 blank-count substitution over the full-sheet roll-up.
func TestBuildApprovalBar(t *testing.T) {
	deps := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		Routes: outcome_matrix.DefaultRoutes(),
	}
	perms := types.NewUserPermissions([]string{
		"job_phase:submit", "job_phase:verify", "job_phase:publish", "job_phase:return",
	})
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			phaseCol("pA", "Sem 1", "tA"),
			phaseCol("pB", "Sem 2", "tB"),
			phaseCol("pC", "Sem 3", "tC"),
			phaseCol("pD", "Sem 4", "tD"),
			phaseCol("pE", "Sem 5", "tE"),
			phaseCol("pF", "Sem 6", "tF"),
		},
		ApprovalRollups: []*matrixpb.PhaseApprovalRollup{
			rollup("pA", ip(sIP), false, true, false, 30, 3),  // in_progress → submit only
			rollup("pB", ip(sFR), false, true, false, 30, 0),  // for_review → verify + return
			rollup("pC", ip(sVER), false, true, false, 30, 0), // verified → publish + return
			rollup("pD", ip(sPUB), false, true, false, 30, 0), // published → return (reason required)
			rollup("pE", ip(sIP), true, true, false, 30, 5),   // mixed → return only
			rollup("pF", ip(sVER), false, true, true, 30, 0),  // hard-frozen → no actions
		},
	}

	bar := buildApprovalBar(deps, perms, resp, "tmpl-1")
	if len(bar) != 6 {
		t.Fatalf("expected 6 bar phases, got %d", len(bar))
	}
	by := map[string]ApprovalPhase{}
	for _, p := range bar {
		by[p.PhaseID] = p
	}

	// pA: in_progress
	a := by["pA"]
	if a.ChipVariant != "neutral" {
		t.Errorf("pA chip variant = %q, want neutral", a.ChipVariant)
	}
	if !a.CanSubmit || a.CanVerify || a.CanPublish || a.CanReturn {
		t.Errorf("pA actions submit/verify/publish/return = %v/%v/%v/%v, want true/false/false/false", a.CanSubmit, a.CanVerify, a.CanPublish, a.CanReturn)
	}
	if !strings.Contains(a.SubmitConfirm, "3") {
		t.Errorf("pA submit confirm should carry blank count 3: %q", a.SubmitConfirm)
	}
	if a.SubmitPath != "/action/outcome-matrix/tmpl-1/submit" {
		t.Errorf("pA submit path = %q", a.SubmitPath)
	}

	// pB: for_review
	b := by["pB"]
	if b.ChipVariant != "warning" || !b.CanVerify || b.CanSubmit || b.CanPublish || !b.CanReturn {
		t.Errorf("pB unexpected: variant=%q verify=%v submit=%v publish=%v return=%v", b.ChipVariant, b.CanVerify, b.CanSubmit, b.CanPublish, b.CanReturn)
	}

	// pC: verified
	c := by["pC"]
	if c.ChipVariant != "info" || !c.CanPublish || c.CanVerify || !c.CanReturn {
		t.Errorf("pC unexpected: variant=%q publish=%v verify=%v return=%v", c.ChipVariant, c.CanPublish, c.CanVerify, c.CanReturn)
	}

	// pD: published — return only, reason required
	d := by["pD"]
	if d.ChipVariant != "success" || !d.CanReturn || d.CanPublish || !d.ReturnReasonRequired {
		t.Errorf("pD unexpected: variant=%q return=%v publish=%v reasonReq=%v", d.ChipVariant, d.CanReturn, d.CanPublish, d.ReturnReasonRequired)
	}

	// pE: mixed (lowest in_progress) — submit blocked by mixed, return available
	e := by["pE"]
	if !e.Mixed || e.CanSubmit || !e.CanReturn {
		t.Errorf("pE unexpected: mixed=%v submit=%v return=%v", e.Mixed, e.CanSubmit, e.CanReturn)
	}

	// pF: hard-frozen — no actions at all
	f := by["pF"]
	if !f.HardFrozen || f.CanSubmit || f.CanVerify || f.CanPublish || f.CanReturn {
		t.Errorf("pF hard-frozen exposed an action: submit=%v verify=%v publish=%v return=%v", f.CanSubmit, f.CanVerify, f.CanPublish, f.CanReturn)
	}
	if f.Hint == "" {
		t.Errorf("pF hard-frozen should carry a hint")
	}
}

// TestBuildApprovalBar_NoPermissions confirms the view-layer gate hides every
// action when the principal lacks the verbs (server is still authoritative).
func TestBuildApprovalBar_NoPermissions(t *testing.T) {
	deps := &PageViewDeps{Labels: outcome_matrix.DefaultLabels(), Routes: outcome_matrix.DefaultRoutes()}
	perms := types.NewEmptyUserPermissions()
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases:          []*matrixpb.PhaseColumn{phaseCol("pA", "Sem 1", "tA")},
		ApprovalRollups: []*matrixpb.PhaseApprovalRollup{rollup("pA", sIP, false, true, false, 30, 0)},
	}
	bar := buildApprovalBar(deps, perms, resp, "tmpl-1")
	if len(bar) != 1 {
		t.Fatalf("want 1 phase, got %d", len(bar))
	}
	if bar[0].CanSubmit || bar[0].CanVerify || bar[0].CanPublish || bar[0].CanReturn {
		t.Errorf("no-permission principal should see no actions: %+v", bar[0])
	}
}

// TestPhaseEditableFunc pins the cell-editability render mirror: a cell is
// render-editable only when its phase's roll-up is cleanly IN_PROGRESS and not
// hard-frozen; a locked / mixed / frozen phase forces read-only.
func TestPhaseEditableFunc(t *testing.T) {
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			phaseCol("pA", "Sem 1", "tA"), // in_progress → editable
			phaseCol("pB", "Sem 2", "tB"), // for_review → locked
			phaseCol("pC", "Sem 3", "tC"), // in_progress but hard-frozen → locked
			phaseCol("pE", "Sem 5", "tE"), // mixed → locked
		},
		ApprovalRollups: []*matrixpb.PhaseApprovalRollup{
			rollup("pA", sIP, false, true, false, 1, 0),
			rollup("pB", sFR, false, true, false, 1, 0),
			rollup("pC", sIP, false, true, true, 1, 0),
			rollup("pE", sIP, true, true, false, 1, 0),
		},
	}
	allow := phaseEditableFunc(resp)
	cases := []struct {
		colKey string
		want   bool
	}{
		{"tA:crit1", true},
		{"tB:crit1", false},
		{"tC:crit1", false},
		{"tE:crit1", false},
		{"tUnknown:crit1", true}, // no mapping → permissive (server still guards)
	}
	for _, c := range cases {
		if got := allow(c.colKey); got != c.want {
			t.Errorf("phaseEditable(%q) = %v, want %v", c.colKey, got, c.want)
		}
	}
}
