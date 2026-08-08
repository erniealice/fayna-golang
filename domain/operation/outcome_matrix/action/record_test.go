package action

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	outcomecriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
)

// ── test fixtures ───────────────────────────────────────────────────────────

const (
	tmplID     = "jt-template-1"
	staffID    = "staff-1"
	colKey     = "tt1:cr1" // "{job_template_task_id}:{criteria_id}"
	criteriaID = "cr1"
	jobTaskID  = "jtinst-1"
	jobPhaseID = "jp-1"
	jobID      = "job-1"
	existingID = "outcome-1"
)

// numericCell builds a one-cell MINE matrix (numeric criterion). recorded=true →
// the cell already has an outcome owned by staffID (update path); recorded=false
// → an empty editable cell (create path).
func numericMatrix(recorded bool) *matrixpb.GetOutcomeMatrixResponse {
	cell := &matrixpb.OutcomeCell{
		JobTaskId:  jobTaskID,
		Editable:   true,
		JobPhaseId: jobPhaseID,
		JobId:      jobID,
	}
	if recorded {
		cell.OutcomeId = existingID
		cell.RecordedBy = staffID
	}
	return &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{{
			JobTemplatePhaseId: "tp1",
			Tasks: []*matrixpb.TaskColumn{{
				JobTemplateTaskId: "tt1",
				Criteria: []*matrixpb.CriterionColumn{{
					ColumnKey: colKey,
					Criteria:  &outcomecriteriapb.OutcomeCriteria{CriteriaType: enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE},
				}},
			}},
		}},
		Rows: []*matrixpb.OutcomeRow{{
			ClientId: "client-1",
			Cells:    map[string]*matrixpb.OutcomeCell{colKey: cell},
		}},
	}
}

type recorder struct {
	createCalls   int
	updateCalls   int
	deleteCalls   int
	order         []string // recompute call order ("phase:<id>", "job:<id>")
	phaseErr      error
	jobRecomputed bool
	jobErr        error
	phaseNil      bool
	jobNil        bool
	readOwner     string             // RecordedBy the ReadTaskOutcome fixture returns (default staffID)
	readCT        enums.CriteriaType // CriteriaType the stored record reports (default NUMERIC_SCORE)
	deleteErr     error

	// recompute-eligibility fixture (wired only when wireElig): eligible + the
	// scheme's in-scope criterion set. Unwired → the action falls back to
	// numeric-type classification.
	wireElig bool
	eligible bool
	inScope  map[string]bool
	eligErr  error
}

func (r *recorder) deps(matrix *matrixpb.GetOutcomeMatrixResponse) *Deps {
	owner := r.readOwner
	if owner == "" {
		owner = staffID
	}
	readCT := r.readCT
	if readCT == enums.CriteriaType_CRITERIA_TYPE_UNSPECIFIED {
		readCT = enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE
	}
	d := &Deps{
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeMatrix: func(context.Context, *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error) {
			return matrix, nil
		},
		ResolveStaff: func(context.Context) (string, error) { return staffID, nil },
		ReadTaskOutcome: func(_ context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error) {
			return &taskoutcomepb.ReadTaskOutcomeResponse{Data: []*taskoutcomepb.TaskOutcome{{
				Id:                req.GetData().GetId(),
				RecordedBy:        owner,
				CriteriaType:      readCT,
				CriteriaVersionId: "cv1",
			}}}, nil
		},
		UpdateTaskOutcome: func(context.Context, *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error) {
			r.updateCalls++
			return &taskoutcomepb.UpdateTaskOutcomeResponse{}, nil
		},
		CreateTaskOutcome: func(context.Context, *taskoutcomepb.CreateTaskOutcomeRequest) (*taskoutcomepb.CreateTaskOutcomeResponse, error) {
			r.createCalls++
			return &taskoutcomepb.CreateTaskOutcomeResponse{Data: []*taskoutcomepb.TaskOutcome{{Id: "new-outcome-9"}}}, nil
		},
		DeleteTaskOutcome: func(_ context.Context, req *taskoutcomepb.DeleteTaskOutcomeRequest) (*taskoutcomepb.DeleteTaskOutcomeResponse, error) {
			r.deleteCalls++
			if r.deleteErr != nil {
				return nil, r.deleteErr
			}
			return &taskoutcomepb.DeleteTaskOutcomeResponse{}, nil
		},
	}
	if !r.phaseNil {
		d.ComputePhaseOutcome = func(_ context.Context, id string) (bool, error) {
			r.order = append(r.order, "phase:"+id)
			return true, r.phaseErr
		}
	}
	if !r.jobNil {
		d.ComputeJobOutcome = func(_ context.Context, id string) (bool, error) {
			r.order = append(r.order, "job:"+id)
			return r.jobRecomputed, r.jobErr
		}
	}
	if r.wireElig {
		d.RecomputeEligibility = func(_ context.Context, _ string) (bool, map[string]bool, error) {
			return r.eligible, r.inScope, r.eligErr
		}
	}
	return d
}

type ackItem struct {
	Key                 string `json:"key"`
	OK                  bool   `json:"ok"`
	OutcomeID           string `json:"outcomeId"`
	NextKey             string `json:"nextKey"`
	Value               string `json:"value"`
	RatingFresh         *bool  `json:"ratingFresh"`
	RatingNotRecomputed string `json:"ratingNotRecomputed"`
	Error               string `json:"error"`
}

func invoke(t *testing.T, d *Deps, form string, perms []string) view.ViewResult {
	t.Helper()
	v := NewRecordAction(d)
	req := httptest.NewRequest("POST", "/action/outcome-matrix/"+tmplID+"/record", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", tmplID)
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions(perms))
	return v.Handle(ctx, &view.ViewContext{Request: req})
}

func cells(t *testing.T, res view.ViewResult) []ackItem {
	t.Helper()
	trig, ok := res.Headers["HX-Trigger"]
	if !ok {
		t.Fatalf("no HX-Trigger header in result: %+v", res.Headers)
	}
	var env struct {
		Result struct {
			Cells []ackItem `json:"cells"`
		} `json:"omcell-result"`
	}
	if err := json.Unmarshal([]byte(trig), &env); err != nil {
		t.Fatalf("HX-Trigger is not the omcell-result envelope (%q): %v", trig, err)
	}
	return env.Result.Cells
}

func byKey(items []ackItem, key string) (ackItem, bool) {
	for _, it := range items {
		if it.Key == key {
			return it, true
		}
	}
	return ackItem{}, false
}

var allPerms = []string{"task_outcome:create", "task_outcome:update", "task_outcome:delete"}

// ── tests ────────────────────────────────────────────────────────────────────

func TestCellMode_SingleUpdate(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=86", allPerms)
	items := cells(t, res)
	if len(items) != 1 {
		t.Fatalf("want 1 ack, got %d: %+v", len(items), items)
	}
	got := items[0]
	if got.Key != "cells."+existingID || !got.OK {
		t.Fatalf("bad ack: %+v", got)
	}
	if got.OutcomeID != existingID {
		t.Errorf("outcomeId: want %q got %q", existingID, got.OutcomeID)
	}
	if got.Value != "86" {
		t.Errorf("normalized value: want 86 got %q", got.Value)
	}
	if got.RatingFresh == nil || !*got.RatingFresh {
		t.Errorf("academic cell + successful recompute → ratingFresh:true, got %v", got.RatingFresh)
	}
	if r.updateCalls != 1 || r.createCalls != 0 {
		t.Errorf("want 1 update 0 create, got %d/%d", r.updateCalls, r.createCalls)
	}
}

func TestCellMode_ClearExisting_Success(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=", allPerms)
	items := cells(t, res)
	if len(items) != 1 {
		t.Fatalf("want 1 ack, got %d: %+v", len(items), items)
	}
	got := items[0]
	if got.Key != "cells."+existingID {
		t.Fatalf("clear delete must target the canonical outcome address: %+v", got)
	}
	if got.NextKey != "new."+jobTaskID+":"+criteriaID {
		t.Fatalf("clear delete must return nextKey=%q, got %q", "new."+jobTaskID+":"+criteriaID, got.NextKey)
	}
	if !got.OK || got.OutcomeID != "" || got.Value != "" {
		t.Fatalf("delete ack must return ok true + empty outcomeId/value: %+v", got)
	}
	if r.deleteCalls != 1 || r.updateCalls != 0 || r.createCalls != 0 {
		t.Errorf("want 1 delete 0 update/create, got %d/%d/%d", r.deleteCalls, r.updateCalls, r.createCalls)
	}
}

func TestCellMode_ClearExisting_WithoutDeletePermission(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=", []string{"task_outcome:create"})
	got := cells(t, res)[0]
	if got.OK {
		t.Fatalf("create-only permission must not clear an existing cell: %+v", got)
	}
	if got.Error != "not_editable" {
		t.Fatalf("expected not_editable for clear without delete permission, got %q", got.Error)
	}
	if r.deleteCalls != 0 {
		t.Errorf("blocked clear must not call DeleteTaskOutcome, got %d", r.deleteCalls)
	}
}

func TestCellMode_ClearExisting_DeleteOnlyPermission(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=", []string{"task_outcome:delete"})
	got := cells(t, res)[0]
	if !got.OK || got.OutcomeID != "" || got.Value != "" {
		t.Fatalf("delete-only permission should still clear existing: %+v", got)
	}
	if r.deleteCalls != 1 {
		t.Errorf("delete-only caller should perform exactly one delete, got %d", r.deleteCalls)
	}
}

func TestCellMode_ClearExisting_ForgedOutcome_Blocked(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells.forged-outcome-x=", allPerms)
	got := cells(t, res)[0]
	if got.OK {
		t.Fatalf("forged outcome clear must be blocked")
	}
	if got.Error == "" {
		t.Fatalf("forged outcome clear must carry a bounded error code: %+v", got)
	}
	if r.deleteCalls != 0 {
		t.Errorf("forged outcome should never call DeleteTaskOutcome, got %d calls", r.deleteCalls)
	}
}

func TestCellMode_ClearExisting_DeleteFailure(t *testing.T) {
	r := &recorder{deleteErr: errors.New("delete failed")}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=", allPerms)
	got := cells(t, res)[0]
	if got.OK {
		t.Fatalf("delete failure must fail the cell")
	}
	if got.NextKey != "" {
		t.Fatalf("delete failure must not emit nextKey")
	}
	if got.Error != "delete_failed" {
		t.Fatalf("want delete_failed on delete error, got %q", got.Error)
	}
	if r.deleteCalls != 1 {
		t.Fatalf("want 1 delete call, got %d", r.deleteCalls)
	}
}

func TestCellMode_ClearExisting_RecomputeClassified(t *testing.T) {
	// A cleared numeric academic cell still drives the same recompute split.
	r := &recorder{wireElig: true, eligible: true, inScope: map[string]bool{criteriaID: true}, jobRecomputed: true}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=", allPerms)
	got := cells(t, res)[0]
	if got.RatingFresh == nil || !*got.RatingFresh {
		t.Fatalf("eligible numeric clear should recompute, got %+v", got)
	}
	if len(r.order) != 2 || r.order[0] != "phase:"+jobPhaseID || r.order[1] != "job:"+jobID {
		t.Errorf("clear should still queue phase then job recompute, got %v", r.order)
	}
}

func TestCellMode_MultiCell_CreateReturnsNewID(t *testing.T) {
	// A matrix with BOTH an existing cell and an empty create cell would need two
	// columns; simpler: create path alone verifies the new-id handshake.
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(false)), "save_mode=cell&new."+jobTaskID+":"+criteriaID+"=7", allPerms)
	items := cells(t, res)
	if len(items) != 1 {
		t.Fatalf("want 1 ack, got %d", len(items))
	}
	got := items[0]
	if !got.OK || got.Key != "new."+jobTaskID+":"+criteriaID {
		t.Fatalf("bad create ack: %+v", got)
	}
	if got.OutcomeID != "new-outcome-9" {
		t.Errorf("MANDATORY new-id handshake: want new-outcome-9, got %q", got.OutcomeID)
	}
	if got.Value != "7" {
		t.Errorf("value: want 7 got %q", got.Value)
	}
	if r.createCalls != 1 {
		t.Errorf("want 1 create, got %d", r.createCalls)
	}
}

func TestCellMode_CreateThenResubmit_Idempotent(t *testing.T) {
	// The client lost the create ack and re-POSTs new.{addr}. The fresh MINE
	// re-derivation now shows an outcome at that address (recorded=true) → the
	// action must UPDATE it (returning the existing id for the rename), never
	// create a duplicate.
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&new."+jobTaskID+":"+criteriaID+"=90", allPerms)
	items := cells(t, res)
	got, ok := byKey(items, "new."+jobTaskID+":"+criteriaID)
	if !ok || !got.OK {
		t.Fatalf("resubmit not acked ok: %+v", items)
	}
	if got.OutcomeID != existingID {
		t.Errorf("idempotent retry must resolve to existing outcome %q, got %q", existingID, got.OutcomeID)
	}
	if r.createCalls != 0 {
		t.Errorf("idempotency breach: create was called %d times (want 0 — duplicate!)", r.createCalls)
	}
	if r.updateCalls != 1 {
		t.Errorf("want the retry to UPDATE (1 call), got %d", r.updateCalls)
	}
}

func TestCellMode_MalformedValue_FailsCellNotBatch(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=not-a-number", allPerms)
	items := cells(t, res)
	if len(items) != 1 {
		t.Fatalf("want 1 ack, got %d", len(items))
	}
	got := items[0]
	if got.OK {
		t.Errorf("malformed numeric must fail the cell, got ok=true")
	}
	if got.Error == "" {
		t.Errorf("failed cell must carry a bounded error code")
	}
}

func TestCellMode_ComputeFailure_RatingFreshFalse(t *testing.T) {
	// The grade persists but the phase recompute fails → ratingFresh:false
	// (stale + retryable), never a failed cell.
	r := &recorder{phaseErr: context.DeadlineExceeded}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=80", allPerms)
	got := cells(t, res)[0]
	if !got.OK {
		t.Fatalf("cell must still report ok:true when only the recompute failed")
	}
	if got.RatingFresh == nil || *got.RatingFresh {
		t.Errorf("compute failure → ratingFresh:false, got %v", got.RatingFresh)
	}
}

func TestCellMode_FrozenJob_NotStale(t *testing.T) {
	// A frozen (authoritative) job_outcome_summary → ComputeJobOutcome returns
	// (false,nil). The grade saved, the pinned rating stands: ratingFresh:true
	// with a ratingNotRecomputed reason, NOT stale.
	r := &recorder{jobRecomputed: false} // phase recomputes ok (true), job skipped
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=80", allPerms)
	got := cells(t, res)[0]
	if got.RatingFresh == nil || !*got.RatingFresh {
		t.Errorf("frozen job must stay fresh (not stale), got %v", got.RatingFresh)
	}
	if got.RatingNotRecomputed != "authoritative_frozen" {
		t.Errorf("want ratingNotRecomputed=authoritative_frozen, got %q", got.RatingNotRecomputed)
	}
}

func TestCellMode_NonAcademic_NoRatingFresh(t *testing.T) {
	// A pass_fail (deportment) cell has no scaled roll-up → recompute is skipped
	// entirely and ratingFresh is omitted.
	matrix := numericMatrix(true)
	matrix.Phases[0].Tasks[0].Criteria[0].Criteria.CriteriaType = enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL
	r := &recorder{readCT: enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL}
	res := invoke(t, r.deps(matrix), "save_mode=cell&cells."+existingID+"=pass", allPerms)
	got := cells(t, res)[0]
	if !got.OK {
		t.Fatalf("pass_fail update should succeed: %+v", got)
	}
	if got.Value != "true" {
		t.Errorf("pass normalized to select value: want true got %q", got.Value)
	}
	if got.RatingFresh != nil {
		t.Errorf("non-academic cell must omit ratingFresh, got %v", *got.RatingFresh)
	}
	if len(r.order) != 0 {
		t.Errorf("non-academic write must not trigger any recompute, got %v", r.order)
	}
}

func TestCellMode_LedgerScheme_NotApplicable(t *testing.T) {
	// A numeric cell whose phase resolves a scheme with NO score scale (a
	// ledger scheme) drives no recompute: it saves normally, acks
	// not-applicable (ratingFresh omitted), and NEVER enqueues a roll-up that
	// would fail loud.
	r := &recorder{wireElig: true, eligible: false}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=21", allPerms)
	got := cells(t, res)[0]
	if !got.OK {
		t.Fatalf("ledger cell must save ok: %+v", got)
	}
	if got.Value != "21" {
		t.Errorf("value: want 21 got %q", got.Value)
	}
	if got.RatingFresh != nil {
		t.Errorf("ineligible numeric cell must omit ratingFresh, got %v", *got.RatingFresh)
	}
	if got.RatingNotRecomputed != "not_applicable" {
		t.Errorf("want ratingNotRecomputed=not_applicable, got %q", got.RatingNotRecomputed)
	}
	if len(r.order) != 0 {
		t.Errorf("ineligible cell must not trigger any recompute, got %v", r.order)
	}
}

func TestCellMode_EligiblePhase_CriterionOutsideGraph_NotApplicable(t *testing.T) {
	// The phase IS score-scaled, but this criterion is not in the scheme's active
	// component graph → the cell still drives no recompute.
	r := &recorder{wireElig: true, eligible: true, inScope: map[string]bool{"some-other-criterion": true}}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=80", allPerms)
	got := cells(t, res)[0]
	if !got.OK {
		t.Fatalf("out-of-graph cell must save ok: %+v", got)
	}
	if got.RatingFresh != nil {
		t.Errorf("out-of-graph criterion must omit ratingFresh, got %v", *got.RatingFresh)
	}
	if got.RatingNotRecomputed != "not_applicable" {
		t.Errorf("want ratingNotRecomputed=not_applicable, got %q", got.RatingNotRecomputed)
	}
	if len(r.order) != 0 {
		t.Errorf("out-of-graph cell must not recompute, got %v", r.order)
	}
}

func TestCellMode_EligiblePhase_InGraph_Recomputes(t *testing.T) {
	// The phase is score-scaled and the criterion IS in the component graph → the
	// cell drives the phase→job recompute and reports ratingFresh:true.
	r := &recorder{wireElig: true, eligible: true, inScope: map[string]bool{criteriaID: true}, jobRecomputed: true}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=80", allPerms)
	got := cells(t, res)[0]
	if got.RatingFresh == nil || !*got.RatingFresh {
		t.Fatalf("eligible in-graph cell must recompute → ratingFresh:true, got %+v", got)
	}
	if len(r.order) != 2 || r.order[0] != "phase:"+jobPhaseID || r.order[1] != "job:"+jobID {
		t.Errorf("want phase then job recompute, got %v", r.order)
	}
}

func TestCellMode_RecomputeOrder_PhaseThenJob(t *testing.T) {
	r := &recorder{jobRecomputed: true}
	invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=80", allPerms)
	if len(r.order) != 2 || r.order[0] != "phase:"+jobPhaseID || r.order[1] != "job:"+jobID {
		t.Fatalf("recompute must be phase THEN job, got %v", r.order)
	}
}

func TestCellMode_NilComputeClosures_RatingFreshFalse(t *testing.T) {
	r := &recorder{phaseNil: true, jobNil: true}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=80", allPerms)
	got := cells(t, res)[0]
	if !got.OK {
		t.Fatalf("nil compute closures must not fail the save")
	}
	if got.RatingFresh == nil || *got.RatingFresh {
		t.Errorf("unwired recompute → ratingFresh:false (fail-safe), got %v", got.RatingFresh)
	}
}

func TestCellMode_IDOR_ForgedOutcome_Blocked(t *testing.T) {
	// An outcome id NOT in the server-derived allow-set is rejected before any
	// use case runs.
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells.forged-outcome-x=80", allPerms)
	got := cells(t, res)[0]
	if got.OK {
		t.Errorf("forged outcome must be blocked")
	}
	if r.updateCalls != 0 {
		t.Errorf("blocked address must never reach UpdateTaskOutcome, got %d calls", r.updateCalls)
	}
}

func TestCellMode_IDOR_OtherOwner_Blocked(t *testing.T) {
	// The address is addressable, but the stored record is owned by another
	// staff member (recorded_by mismatch) → the update-time ownership re-read
	// blocks it. (The allow-set is built from a MINE matrix so this is a
	// defense-in-depth backstop.)
	r := &recorder{readOwner: "someone-else"}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=80", allPerms)
	got := cells(t, res)[0]
	if got.OK {
		t.Errorf("cell owned by another staff must be blocked at update")
	}
	if r.updateCalls != 0 {
		t.Errorf("ownership mismatch must not call UpdateTaskOutcome, got %d", r.updateCalls)
	}
}

func TestLegacyBatch_Unchanged_Success(t *testing.T) {
	// No save_mode → the aggregate formSuccess/formError response (a11y fallback).
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "cells."+existingID+"=80", allPerms)
	trig := res.Headers["HX-Trigger"]
	if trig != `{"formSuccess":true}` {
		t.Fatalf("legacy success response changed: %q", trig)
	}
	if r.updateCalls != 1 {
		t.Errorf("legacy non-cell batch should update editable cell: got %d calls", r.updateCalls)
	}
}

func TestLegacyBatch_ClearExisting_Success(t *testing.T) {
	// Legacy mode also supports clearing an existing cell to delete outcome record.
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "cells."+existingID+"=", allPerms)
	trig := res.Headers["HX-Trigger"]
	if trig != `{"formSuccess":true}` {
		t.Fatalf("legacy clear should still be aggregate success: %q", trig)
	}
	if r.deleteCalls != 1 {
		t.Fatalf("want 1 delete in legacy clear, got %d", r.deleteCalls)
	}
}

func TestLegacyBatch_Unchanged_PartialFailure(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "cells."+existingID+"=80&cells.forged=1", allPerms)
	trig := res.Headers["HX-Trigger"]
	if !strings.Contains(trig, "formError") {
		t.Fatalf("legacy partial-failure must report formError, got %q", trig)
	}
}

// ── criterion value-contract enforcement (20260725 V3) ──────────────────────
//
// The view renders the criterion's min/max/maxlength/option-set as INPUT
// ATTRIBUTES. Those live in the DOM, so they gate nothing: a POST that skips
// the widget persisted whatever parsed. Measured on education1 before the fix:
// an IB-MYP criterion declared 0..8 accepted and stored 76.
//
// The bound is read from the SAME server-derived MINE matrix that grants the
// write, so these tests drive it exactly the way the action does — through the
// matrix fixture, never through a request field.

// boundedMatrix is numericMatrix with a declared score range on the criterion.
func boundedMatrix(recorded bool, min, max int32) *matrixpb.GetOutcomeMatrixResponse {
	m := numericMatrix(recorded)
	oc := m.GetPhases()[0].GetTasks()[0].GetCriteria()[0].GetCriteria()
	oc.MinScore = &min
	oc.MaxScore = &max
	return m
}

func TestBounds_Update_RejectsAboveMax(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(boundedMatrix(true, 0, 8)), "save_mode=cell&cells."+existingID+"=76", allPerms)
	got := cells(t, res)[0]
	if got.OK {
		t.Errorf("76 on a criterion declared 0..8 must be rejected")
	}
	if r.updateCalls != 0 {
		t.Errorf("an out-of-range value must never reach UpdateTaskOutcome, got %d calls", r.updateCalls)
	}
}

func TestBounds_Update_RejectsBelowMin(t *testing.T) {
	r := &recorder{}
	res := invoke(t, r.deps(boundedMatrix(true, 0, 8)), "save_mode=cell&cells."+existingID+"=-1", allPerms)
	if cells(t, res)[0].OK {
		t.Errorf("-1 on a criterion declared 0..8 must be rejected")
	}
	if r.updateCalls != 0 {
		t.Errorf("an out-of-range value must never reach UpdateTaskOutcome, got %d calls", r.updateCalls)
	}
}

func TestBounds_Update_AcceptsInRangeAtBoundary(t *testing.T) {
	// Inclusive on both ends — a criterion declared 0..8 grades 0 and 8.
	for _, v := range []string{"0", "8", "5"} {
		r := &recorder{}
		res := invoke(t, r.deps(boundedMatrix(true, 0, 8)), "save_mode=cell&cells."+existingID+"="+v, allPerms)
		got := cells(t, res)[0]
		if !got.OK {
			t.Errorf("%s is inside 0..8 and must save (err=%q)", v, got.Error)
		}
		if r.updateCalls != 1 {
			t.Errorf("%s: want 1 update call, got %d", v, r.updateCalls)
		}
	}
}

func TestBounds_Create_RejectsAboveMax(t *testing.T) {
	// The create path is the one a blank grade sheet actually uses.
	r := &recorder{}
	res := invoke(t, r.deps(boundedMatrix(false, 0, 8)),
		"save_mode=cell&new."+jobTaskID+":"+criteriaID+"=76", allPerms)
	got := cells(t, res)[0]
	if got.OK {
		t.Errorf("76 on a criterion declared 0..8 must be rejected on create")
	}
	if r.createCalls != 0 {
		t.Errorf("an out-of-range value must never reach CreateTaskOutcome, got %d calls", r.createCalls)
	}
	if got.OutcomeID != "" {
		t.Errorf("a rejected create must not hand the client an id to rename to, got %q", got.OutcomeID)
	}
}

func TestBounds_Unconstrained_CriterionIsUnchanged(t *testing.T) {
	// A criterion that declares no min/max keeps the pre-fix behavior verbatim —
	// the gate is opt-in via the criterion entity, not a new global rule.
	r := &recorder{}
	res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"=1000", allPerms)
	if !cells(t, res)[0].OK {
		t.Errorf("a criterion declaring no range must still accept any parseable score")
	}
}

func TestBounds_RejectsNonFiniteScore(t *testing.T) {
	// NaN/Inf parse as float64 and compare false against every bound, so an
	// unconstrained criterion would otherwise store them.
	for _, v := range []string{"NaN", "Inf", "-Inf"} {
		r := &recorder{}
		res := invoke(t, r.deps(numericMatrix(true)), "save_mode=cell&cells."+existingID+"="+v, allPerms)
		if cells(t, res)[0].OK {
			t.Errorf("%s is not a score and must be rejected", v)
		}
		if r.updateCalls != 0 {
			t.Errorf("%s: must never reach UpdateTaskOutcome, got %d calls", v, r.updateCalls)
		}
	}
}
