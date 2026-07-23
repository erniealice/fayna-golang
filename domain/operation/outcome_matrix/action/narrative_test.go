package action

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
)

// narrativeRec is the narrative-drawer test double. It records the update request
// so a test can assert EXACTLY which columns the save touches (determination_note
// only — never the typed grade value).
type narrativeRec struct {
	updateCalls int
	updateReq   *taskoutcomepb.UpdateTaskOutcomeRequest
	readOwner   string // recorded_by the stored record reports (default staffID)
	readNote    string // determination_note the read returns
}

func (r *narrativeRec) deps(matrix *matrixpb.GetOutcomeMatrixResponse) *NarrativeDeps {
	owner := r.readOwner
	if owner == "" {
		owner = staffID
	}
	note := r.readNote
	return &NarrativeDeps{
		Routes: outcome_matrix.DefaultRoutes(),
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeMatrix: func(context.Context, *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error) {
			return matrix, nil
		},
		ResolveStaff: func(context.Context) (string, error) { return staffID, nil },
		ReadTaskOutcome: func(_ context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error) {
			return &taskoutcomepb.ReadTaskOutcomeResponse{Data: []*taskoutcomepb.TaskOutcome{{
				Id:                req.GetData().GetId(),
				RecordedBy:        owner,
				DeterminationNote: &note,
			}}}, nil
		},
		UpdateTaskOutcome: func(_ context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error) {
			r.updateCalls++
			r.updateReq = req
			return &taskoutcomepb.UpdateTaskOutcomeResponse{}, nil
		},
	}
}

var narrPerms = []string{"task_outcome:read", "task_outcome:update"}

func narrativeInvoke(t *testing.T, d *NarrativeDeps, method, query, body string, perms []string) view.ViewResult {
	t.Helper()
	v := NewNarrativeAction(d)
	target := "/action/outcome-matrix/" + tmplID + "/narrative"
	var req *http.Request
	if method == http.MethodPost {
		req = httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		if query != "" {
			target += "?" + query
		}
		req = httptest.NewRequest(http.MethodGet, target, nil)
	}
	req.SetPathValue("id", tmplID)
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions(perms))
	return v.Handle(ctx, &view.ViewContext{Request: req})
}

func narrativeData(t *testing.T, res view.ViewResult) *NarrativeDrawerData {
	t.Helper()
	if res.Template != "outcome-matrix-narrative-drawer" {
		t.Fatalf("want narrative drawer template, got %q (status %d)", res.Template, res.StatusCode)
	}
	data, ok := res.Data.(*NarrativeDrawerData)
	if !ok {
		t.Fatalf("drawer data is not *NarrativeDrawerData: %T", res.Data)
	}
	return data
}

// ── GET (drawer render) ───────────────────────────────────────────────────────

func TestNarrative_GET_EditableCell_ShowsForm(t *testing.T) {
	r := &narrativeRec{readNote: "Consistent, thoughtful work."}
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodGet, "outcome_id="+existingID, "", narrPerms)
	data := narrativeData(t, res)
	if !data.Editable {
		t.Errorf("owned + editable cell must render the editable drawer")
	}
	if data.Note != "Consistent, thoughtful work." || !data.HasNote {
		t.Errorf("current note not surfaced: %+v", data)
	}
	if data.OutcomeID != existingID {
		t.Errorf("outcome id: want %q got %q", existingID, data.OutcomeID)
	}
}

func TestNarrative_GET_NoUpdatePerm_ReadOnly(t *testing.T) {
	r := &narrativeRec{readNote: "note"}
	// read-only perm: can view, cannot edit → read-only variant.
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodGet, "outcome_id="+existingID, "", []string{"task_outcome:read"})
	if narrativeData(t, res).Editable {
		t.Errorf("a principal without task_outcome:update must get the read-only drawer")
	}
}

func TestNarrative_GET_FrozenCell_ReadOnly(t *testing.T) {
	// The cell is owned but its phase is not editable (frozen/advanced) → the
	// server-derived Editable flag is false → allowedUpdate misses → read-only.
	m := numericMatrix(true)
	m.Rows[0].Cells[colKey].Editable = false
	r := &narrativeRec{readNote: "prior note"}
	res := narrativeInvoke(t, r.deps(m), http.MethodGet, "outcome_id="+existingID, "", narrPerms)
	data := narrativeData(t, res)
	if data.Editable {
		t.Errorf("frozen cell must be read-only")
	}
	if data.Note != "prior note" {
		t.Errorf("read-only drawer must still show the note, got %q", data.Note)
	}
}

func TestNarrative_GET_HardFrozenPhase_ReadOnly(t *testing.T) {
	// The cell is owned and the espyna Editable (ownership) flag is set, but its
	// template phase is HARD-FROZEN (closed schedule / active authoritative final)
	// — the dimension the POST cell-write guard fail-closes on. The shared
	// authority core folds it in from the matrix's per-phase approval roll-up, so
	// the GET must resolve READ-ONLY. This is the F2 GET/POST-parity fix: without
	// the fold-in the GET rendered the editable form while the POST 422'd.
	m := numericMatrix(true)
	m.ApprovalRollups = []*matrixpb.PhaseApprovalRollup{{
		JobTemplatePhaseId: "tp1", // the phase numericMatrix hangs colKey "tt1:cr1" under
		HardFrozen:         true,
	}}
	r := &narrativeRec{readNote: "prior note"}
	res := narrativeInvoke(t, r.deps(m), http.MethodGet, "outcome_id="+existingID, "", narrPerms)
	data := narrativeData(t, res)
	if data.Editable {
		t.Errorf("hard-frozen phase must render the read-only drawer (GET/POST parity), got Editable=true")
	}
	if data.Note != "prior note" {
		t.Errorf("read-only drawer must still surface the note, got %q", data.Note)
	}
}

func TestNarrative_POST_HardFrozenPhase_Blocked(t *testing.T) {
	// POST parity with the GET above: a hard-frozen cell is not in allowedUpdate,
	// so the save denies at the shared core BEFORE the write use case — the SAME
	// verdict the GET renders read-only from (no reliance on the downstream write
	// guard for consistency).
	m := numericMatrix(true)
	m.ApprovalRollups = []*matrixpb.PhaseApprovalRollup{{
		JobTemplatePhaseId: "tp1",
		HardFrozen:         true,
	}}
	r := &narrativeRec{}
	res := narrativeInvoke(t, r.deps(m), http.MethodPost, "", "outcome_id="+existingID+"&note=hi", narrPerms)
	if r.updateCalls != 0 {
		t.Errorf("hard-frozen cell must never reach UpdateTaskOutcome, got %d", r.updateCalls)
	}
	if res.StatusCode == http.StatusOK {
		t.Errorf("hard-frozen save must be rejected, got 200")
	}
}

func TestNarrative_GET_MissingOutcomeID_Denied(t *testing.T) {
	r := &narrativeRec{}
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodGet, "", "", narrPerms)
	if res.Template == "outcome-matrix-narrative-drawer" {
		t.Errorf("a GET with no outcome_id must not render a drawer")
	}
}

// ── POST (save) ───────────────────────────────────────────────────────────────

func TestNarrative_POST_SavesDeterminationNoteOnly(t *testing.T) {
	r := &narrativeRec{}
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodPost, "", "outcome_id="+existingID+"&note=Great+progress", narrPerms)
	if r.updateCalls != 1 {
		t.Fatalf("want 1 update, got %d", r.updateCalls)
	}
	if got := r.updateReq.GetData().GetDeterminationNote(); got != "Great progress" {
		t.Errorf("determination_note: want %q got %q", "Great progress", got)
	}
	// The grade value columns MUST be left untouched (partial update).
	d := r.updateReq.GetData()
	if d.NumericValue != nil || d.TextValue != nil || d.CategoricalValue != nil || d.PassFailValue != nil {
		t.Errorf("narrative save must not set any typed grade value: %+v", d)
	}
	if trig := res.Headers["HX-Trigger"]; !strings.Contains(trig, "formSuccess") {
		t.Errorf("save must signal formSuccess (sheet close), got %q", trig)
	}
}

func TestNarrative_POST_EmptyNote_Clears(t *testing.T) {
	r := &narrativeRec{}
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodPost, "", "outcome_id="+existingID+"&note=", narrPerms)
	_ = res
	if r.updateReq.GetData().DeterminationNote == nil || *r.updateReq.GetData().DeterminationNote != "" {
		t.Errorf("empty submission must explicitly clear the note (empty string set), got %v", r.updateReq.GetData().DeterminationNote)
	}
}

func TestNarrative_POST_ClearButton_Clears(t *testing.T) {
	r := &narrativeRec{}
	// Even with text in the textarea, the explicit Clear button forces empty.
	narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodPost, "", "outcome_id="+existingID+"&note=still+here&clear=1", narrPerms)
	if r.updateReq.GetData().GetDeterminationNote() != "" {
		t.Errorf("clear button must force an empty note, got %q", r.updateReq.GetData().GetDeterminationNote())
	}
}

func TestNarrative_POST_ForgedOutcome_Blocked(t *testing.T) {
	r := &narrativeRec{}
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodPost, "", "outcome_id=forged-x&note=hi", narrPerms)
	if r.updateCalls != 0 {
		t.Errorf("a forged outcome_id must never reach UpdateTaskOutcome, got %d calls", r.updateCalls)
	}
	if res.StatusCode == http.StatusOK {
		t.Errorf("forged save must be rejected, got 200")
	}
}

func TestNarrative_POST_OtherOwner_BlockedAtReread(t *testing.T) {
	// The cell is addressable in the MINE matrix, but the persisted record is
	// owned by another staff (recorded_by mismatch) → the write-time re-read
	// blocks it (defense in depth behind the MINE-derived allow-set).
	r := &narrativeRec{readOwner: "someone-else"}
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodPost, "", "outcome_id="+existingID+"&note=hi", narrPerms)
	if r.updateCalls != 0 {
		t.Errorf("ownership mismatch must not call UpdateTaskOutcome, got %d", r.updateCalls)
	}
	if res.StatusCode == http.StatusOK {
		t.Errorf("other-owner save must be rejected, got 200")
	}
}

func TestNarrative_POST_NoUpdatePerm_Blocked(t *testing.T) {
	r := &narrativeRec{}
	res := narrativeInvoke(t, r.deps(numericMatrix(true)), http.MethodPost, "", "outcome_id="+existingID+"&note=hi", []string{"task_outcome:read"})
	if r.updateCalls != 0 {
		t.Errorf("missing task_outcome:update must block the save, got %d calls", r.updateCalls)
	}
	if res.StatusCode == http.StatusOK {
		t.Errorf("permission-denied save must not return 200")
	}
}
