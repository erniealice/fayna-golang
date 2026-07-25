package action

// authz.go — the SHARED authority core for outcome-matrix writes (N-1d LOCKED).
//
// Both write verbs on the sheet — the value record action (record.go) and the
// per-cell narrative drawer (narrative.go) — gate through resolveCellAuthority so
// they pass a BYTE-IDENTICAL grader/phase/frozen check by construction. There is
// exactly one gate; neither verb re-implements it, so they can never drift.
//
// The single source of truth is the acting principal's MINE-scoped matrix,
// re-derived server-side per request. The core folds every axis the gate needs
// into ONE editability verdict (allowedUpdate / allowedCreate):
//   - grader-ownership: MINE scope only surfaces cells the acting staff records,
//     so a cell in this matrix is one the principal owns (recorded_by == acting).
//     The espyna OutcomeCell.Editable flag encodes this ownership axis.
//   - frozen/finalized: a hard-frozen (finalized / closed-schedule) phase is
//     folded in HERE from the matrix's per-template-phase approval roll-up
//     (HardFrozen) — the Editable flag alone does NOT carry it, so the core must
//     apply it, or the drawer GET would render editable a cell the write use case
//     fail-closes on (the F2 GET/POST parity bug). Both write verbs and the drawer
//     GET now consult this single frozen-aware verdict, so they cannot drift.
// The POST body's outcome_id / job_task_id keys are attacker-controlled and are
// NEVER trusted as a scope — only an address the server's own MINE matrix
// contains is addressable (a forged id outside the principal's roster can never
// reach a create/update use case, and can never recover a capability the view
// omitted).

import (
	"context"
	"log"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	outcomecriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// cellBounds is one criterion's declared VALUE contract — the same enforcement
// entity the view renders as the input's min/max/step/maxlength/option-set
// attributes (list/page.go buildCellInput). Those attributes are a client-side
// convenience only: a POST reaches the record action with whatever the caller
// typed, and before this the action's typed parse accepted any parseable float,
// so a criterion declared 0..8 happily persisted 76.
//
// It is read from the SAME server-derived MINE matrix that grants the write, so
// the bound can never be supplied (or widened) by the request.
//
// Zero value = unconstrained: nil min/max, zero maxTextLen and a nil allow-list
// all mean "this criterion declares no limit on that axis". A criterion that
// declares nothing therefore behaves exactly as before.
type cellBounds struct {
	min        *float64
	max        *float64
	maxTextLen int             // 0 ⇒ unbounded
	allowed    map[string]bool // categorical allow-list; empty ⇒ unconstrained
}

// boundsFromCriteria projects the criterion entity onto the value contract.
// Mirrors list/page.go buildCellInput field-for-field so the server enforces
// exactly what the client advertises — no second, divergent rule set.
func boundsFromCriteria(oc *outcomecriteriapb.OutcomeCriteria) cellBounds {
	var b cellBounds
	if oc == nil {
		return b
	}
	if oc.MinScore != nil {
		v := float64(oc.GetMinScore())
		b.min = &v
	}
	if oc.MaxScore != nil {
		v := float64(oc.GetMaxScore())
		b.max = &v
	}
	if oc.MaxTextLength != nil {
		b.maxTextLen = int(oc.GetMaxTextLength())
	}
	if opts := oc.GetAllowedDeterminations(); len(opts) > 0 {
		b.allowed = make(map[string]bool, len(opts))
		for _, o := range opts {
			b.allowed[o] = true
		}
	}
	return b
}

// authorityDeps is the minimal seam resolveCellAuthority needs: the acting-staff
// resolver and the MINE-scoped matrix re-derivation. Both the record Deps and the
// narrative Deps carry these exact closure signatures, so the two actions resolve
// their authority through the identical code (no parallel implementation).
type authorityDeps struct {
	ResolveStaff     func(ctx context.Context) (string, error)
	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)
}

// cellAuthority is the acting principal's server-derived write authority over one
// outcome matrix. The maps are keyed only by SERVER identity (never a POST value)
// and contain only cells the MINE re-derivation marked editable, so a lookup that
// misses is a denial.
type cellAuthority struct {
	// actingStaff is the resolved staff_id; always non-empty (resolveCellAuthority
	// fails closed when it cannot resolve one).
	actingStaff string
	// matrix is the re-derived MINE matrix (column tree + rows) — callers may read
	// it for cell context (criterion / row identity), never for authority.
	matrix *matrixpb.GetOutcomeMatrixResponse

	typeByColKey  map[string]enums.CriteriaType // column_key → criteria type
	allowedCreate map[string]enums.CriteriaType // "{job_task_id}:{criteria_id}" → type
	allowedUpdate map[string]bool               // outcome_id → updatable (owned + editable)
	byOutcome     map[string]srvCell            // outcome_id → server cell (recompute keys)
	byCreateAddr  map[string]srvCell            // "{job_task_id}:{criteria_id}" → server cell
}

// resolveCellAuthority resolves the acting staff, re-derives their MINE-scoped
// matrix for templateID, and builds the owned/editable cell maps. ok is false
// (the caller MUST deny) when the staff cannot be resolved, the template id is
// empty, the matrix closure is unwired, or the re-derivation fails — every one a
// fail-closed denial, never a partial/authority-bypassing result.
//
// This is the extracted security preamble the value record action ran inline;
// record.go and narrative.go now both call it, which is exactly what makes their
// gating identical rather than a hand-copied twin that could rot.
func resolveCellAuthority(ctx context.Context, deps authorityDeps, templateID string) (*cellAuthority, bool) {
	var actingStaff string
	if deps.ResolveStaff != nil {
		actingStaff, _ = deps.ResolveStaff(ctx)
	}
	if actingStaff == "" {
		return nil, false
	}

	if templateID == "" || deps.GetOutcomeMatrix == nil {
		return nil, false
	}
	matrix, err := deps.GetOutcomeMatrix(ctx, &matrixpb.GetOutcomeMatrixRequest{
		JobTemplateId: templateID,
		Scope:         matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE,
	})
	if err != nil || matrix == nil {
		log.Printf("[outcome-matrix] scope re-derivation failed for template %s: %v", templateID, err)
		return nil, false
	}

	// The column tree carries each criterion's enforcement entity — the
	// criteria_type there (not the value's shape) decides which typed column a
	// create writes, so a crafted value can never land in the wrong column. The
	// same pass indexes each leaf column_key → its owning job_template_phase_id so
	// a cell (addressed by column_key) can be tested against the hard-frozen set
	// below.
	typeByColKey := make(map[string]enums.CriteriaType)
	phaseByColKey := make(map[string]string)      // column_key → job_template_phase_id
	boundsByColKey := make(map[string]cellBounds) // column_key → value contract
	for _, phase := range matrix.GetPhases() {
		for _, task := range phase.GetTasks() {
			for _, crit := range task.GetCriteria() {
				typeByColKey[crit.GetColumnKey()] = crit.GetCriteria().GetCriteriaType()
				phaseByColKey[crit.GetColumnKey()] = phase.GetJobTemplatePhaseId()
				boundsByColKey[crit.GetColumnKey()] = boundsFromCriteria(crit.GetCriteria())
			}
		}
	}

	// Hard-frozen dimension (N-1d parity fix, F2). The espyna OutcomeCell.Editable
	// flag encodes grader-ownership + instance existence but NOT the hard-frozen
	// (closed-schedule / active-authoritative-final) state the cell-write use case
	// fail-closes on — so before this fold-in the drawer GET resolved a frozen
	// cell EDITABLE while the matching POST 422'd at the write guard, GET and POST
	// disagreeing. The matrix response already carries the truthful per-template-
	// phase roll-up (workspace-scoped, independent of MINE/ALL), so the shared core
	// consults the SAME frozen dimension both verbs need: a cell under a hard-frozen
	// template phase never enters allowedUpdate/allowedCreate, so the drawer GET
	// renders read-only AND the record/narrative POST denies at the core, by
	// construction. Phases with no roll-up entry (mock / non-postgres builds)
	// default NOT frozen — record.go's fixtures carry no roll-ups, so they are
	// unaffected and behave byte-identically.
	frozenPhase := make(map[string]bool)
	for _, ru := range matrix.GetApprovalRollups() {
		if ru.GetHardFrozen() {
			frozenPhase[ru.GetJobTemplatePhaseId()] = true
		}
	}

	allowedCreate := make(map[string]enums.CriteriaType) // "{job_task_id}:{criteria_id}" → criteria type
	allowedUpdate := make(map[string]bool)               // outcome_id → updatable
	byOutcome := make(map[string]srvCell)                // outcome_id → server cell (recompute keys)
	byCreateAddr := make(map[string]srvCell)             // "{job_task_id}:{criteria_id}" → server cell
	for _, row := range matrix.GetRows() {
		// Cells are keyed by column_key "{job_template_task_id}:{criteria_id}";
		// the criteria id half addresses the create, paired with the cell's own
		// job_task instance id.
		for colKey, cell := range row.GetCells() {
			if !cell.GetEditable() {
				continue
			}
			// Folded-in frozen verdict: a cell under a hard-frozen template phase is
			// never editable, even when the espyna Editable (ownership) flag is set.
			// This is the ONE frozen rule the drawer GET, the narrative POST, and the
			// value-record POST all consult — the GET can no longer render editable a
			// cell the POST would fail-close on.
			if frozenPhase[phaseByColKey[colKey]] {
				continue
			}
			_, criteriaID, ok := splitCellAddr(colKey)
			if !ok {
				continue
			}
			sc := srvCell{
				outcomeID:  cell.GetOutcomeId(),
				jobTaskID:  cell.GetJobTaskId(),
				criteriaID: criteriaID,
				ct:         typeByColKey[colKey],
				bounds:     boundsByColKey[colKey],
				jobPhaseID: cell.GetJobPhaseId(),
				jobID:      cell.GetJobId(),
			}
			if cell.GetOutcomeId() != "" {
				allowedUpdate[cell.GetOutcomeId()] = true
				byOutcome[cell.GetOutcomeId()] = sc
				// A recorded cell's create-address maps back to its outcome so a
				// lost-ack new.* retry resolves to an UPDATE, never a dup.
				if cell.GetJobTaskId() != "" {
					byCreateAddr[cell.GetJobTaskId()+":"+criteriaID] = sc
				}
				continue
			}
			if cell.GetJobTaskId() != "" {
				allowedCreate[cell.GetJobTaskId()+":"+criteriaID] = typeByColKey[colKey]
				byCreateAddr[cell.GetJobTaskId()+":"+criteriaID] = sc
			}
		}
	}

	return &cellAuthority{
		actingStaff:   actingStaff,
		matrix:        matrix,
		typeByColKey:  typeByColKey,
		allowedCreate: allowedCreate,
		allowedUpdate: allowedUpdate,
		byOutcome:     byOutcome,
		byCreateAddr:  byCreateAddr,
	}, true
}
