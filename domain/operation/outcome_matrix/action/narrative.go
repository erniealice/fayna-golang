package action

// narrative.go — the per-cell NARRATIVE drawer (N-1 LOCKED 2026-07-23).
//
// One dedicated route, two verbs, ONE authority core:
//   - GET  NarrativeURL?outcome_id=…  → render the drawer form. Editability is
//     resolved SERVER-SIDE (the shared authority core); the client never decides.
//     An editable cell (owned + phase IN_PROGRESS + not frozen) gets a textarea +
//     Save + Clear; any other cell (other-grader / frozen / finalized) gets a
//     read-only view of the note.
//   - POST NarrativeURL               → save the cell's determination_note through
//     the task_outcome:update use case (NEVER raw SQL). An empty submission clears
//     the note. The write is gated by the SAME resolveCellAuthority the value
//     record action uses, then re-guarded on recorded_by ownership at write time —
//     byte-identical to record.go's cell-edit gate, so a forged POST can never
//     recover a capability the view omitted.
//
// This is NOT an extension of the value record protocol: a narrative is type-less,
// multi-line prose and drives no scaled-summary recompute, so it never enters the
// typed-value grammar / micro-batch ack machinery. The determination_note is the
// ONE canonical narrative field for all years (f14); text_value (f9) is inert
// provenance and is never read here.

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// NarrativeDeps holds the dependencies for the narrative drawer view. It reuses
// the SAME ResolveStaff + GetOutcomeMatrix closures the record action gates on
// (so authority is byte-identical) plus the task_outcome read/update use cases the
// module already wires. All writes route through UpdateTaskOutcome (never raw
// SQL). Optional/nil-safe: a missing closure fails the drawer closed, never panics.
type NarrativeDeps struct {
	Routes outcome_matrix.Routes
	Labels outcome_matrix.Labels

	// GetOutcomeMatrix re-derives the acting principal's MINE-scoped matrix — the
	// single authority on which cells are editable (see authz.go).
	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)
	// ResolveStaff maps the acting session user → staff_id ("" == fail-closed).
	ResolveStaff func(ctx context.Context) (string, error)

	// ReadTaskOutcome loads the current note (GET display) + backs the write-time
	// recorded_by ownership re-read (POST). UpdateTaskOutcome persists the note.
	ReadTaskOutcome   func(ctx context.Context, req *taskoutcomepb.ReadTaskOutcomeRequest) (*taskoutcomepb.ReadTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest) (*taskoutcomepb.UpdateTaskOutcomeResponse, error)

	// ResolveCellRatingDescriptions and UpdateTaskOutcomeIfUnchanged back the
	// descriptor list (narrative_selection.go). Same closures the record action
	// uses. Unwired: no list on GET, denied on POST.
	ResolveCellRatingDescriptions func(ctx context.Context, req *matrixpb.ResolveCellRatingDescriptionsRequest) (*matrixpb.ResolveCellRatingDescriptionsResponse, error)
	UpdateTaskOutcomeIfUnchanged  func(ctx context.Context, req *taskoutcomepb.UpdateTaskOutcomeRequest, expected *taskoutcomepb.TaskOutcome) (*taskoutcomepb.UpdateTaskOutcomeResponse, bool, error)
}

// NarrativeDrawerData is the template-facing shape for outcome-matrix-narrative-drawer.
// CommonLabels / Nonce / WorkspaceID are injected by the ViewAdapter via reflection
// (WorkspaceID backs the action_workspace_guard signature on the POST form).
type NarrativeDrawerData struct {
	FormAction string // resolved NarrativeURL ({id} filled) — the POST target
	OutcomeID  string // the target task_outcome id (hidden form field)
	Editable   bool   // server-resolved: textarea+Save+Clear vs read-only view
	Note       string // current determination_note
	HasNote    bool   // Note != "" (drives the read-only empty-state)

	// DescriptorOptions is non-empty only on an editable cell with at least two
	// selectable wordings for its current rating.
	DescriptorOptions []types.RadioOption
	SelectedEntryID   string

	Labels       outcome_matrix.NarrativeLabels
	CommonLabels any    // injected by ViewAdapter
	Nonce        string // injected by ViewAdapter (CSP nonce)
	WorkspaceID  string // injected by ViewAdapter (action-workspace signature)
}

// NewNarrativeAction creates the narrative drawer view (GET = form, POST = save).
// The GET gates on task_outcome:read (view the note); the POST gates on
// task_outcome:update (persist it) — the same family the underlying use cases
// enforce. Editability on GET and the write gate on POST both flow from
// resolveCellAuthority, so the drawer's view and its save agree by construction.
func NewNarrativeAction(deps *NarrativeDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		templateID := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			// View gate (safe method) — mirror the download drawer + grid view.
			if !perms.Can("task_outcome", "read") {
				return view.Forbidden("task_outcome:read")
			}
			outcomeID := strings.TrimSpace(viewCtx.Request.URL.Query().Get("outcome_id"))
			if outcomeID == "" {
				return view.HTMXError(deps.Labels.Errors.PermissionDenied)
			}

			// Editability is resolved by the shared authority core: the cell must be
			// in the acting principal's MINE matrix (owned) AND editable (phase
			// IN_PROGRESS, not frozen) AND the principal must hold task_outcome:update.
			// Any miss ⇒ the read-only variant; the client never decides.
			editable := false
			var cell srvCell
			if perms.Can("task_outcome", "update") {
				if auth, ok := resolveCellAuthority(ctx, authorityDeps{
					ResolveStaff:     deps.ResolveStaff,
					GetOutcomeMatrix: deps.GetOutcomeMatrix,
				}, templateID); ok {
					editable = auth.allowedUpdate[outcomeID]
					cell = auth.byOutcome[outcomeID]
				}
			}

			outcome := readOutcome(ctx, deps, outcomeID)
			note := outcome.GetDeterminationNote()
			if outcome == nil {
				// The single-outcome read is owner-scoped for a Teacher login, so a
				// reviewer (e.g. a subject coordinator) reading another teacher's cell
				// gets nothing. Fall back to the matrix the grid itself shows them:
				// the same scope rule, the same row scope, read-only.
				note = noteFromVisibleMatrix(ctx, deps, perms, templateID, outcomeID)
			}

			data := &NarrativeDrawerData{
				FormAction: route.ResolveURL(deps.Routes.NarrativeURL, "id", templateID),
				OutcomeID:  outcomeID,
				Editable:   editable,
				Note:       note,
				HasNote:    note != "",
				Labels:     deps.Labels.Narrative,
			}
			if editable {
				data.DescriptorOptions, data.SelectedEntryID = loadDescriptorOptions(ctx, deps, cell, outcome)
			}
			return view.OK("outcome-matrix-narrative-drawer", data)
		}

		// POST — save the note. Gate on the write family FIRST (fail-closed).
		if !perms.Can("task_outcome", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		outcomeID := strings.TrimSpace(viewCtx.Request.FormValue("outcome_id"))
		if outcomeID == "" {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		// Shared authority core — byte-identical to the value record gate. The cell
		// must be owned + editable in the freshly re-derived MINE matrix; a forged
		// or frozen / other-grader outcome_id is not in allowedUpdate ⇒ denied.
		auth, ok := resolveCellAuthority(ctx, authorityDeps{
			ResolveStaff:     deps.ResolveStaff,
			GetOutcomeMatrix: deps.GetOutcomeMatrix,
		}, templateID)
		if !ok || !auth.allowedUpdate[outcomeID] {
			log.Printf("[outcome-matrix] narrative save blocked: outcome %s not editable for the acting principal", outcomeID)
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		// A descriptor pick is its own path: the note and clear fields are not read.
		if strings.TrimSpace(viewCtx.Request.FormValue("use_descriptor")) != "" {
			return applyDescriptorSelection(ctx, deps, auth, outcomeID, strings.TrimSpace(viewCtx.Request.FormValue("rating_description_set_entry_id")))
		}

		// Empty submission (or the explicit Clear button) clears the note.
		note := strings.TrimSpace(viewCtx.Request.FormValue("note"))
		if strings.TrimSpace(viewCtx.Request.FormValue("clear")) != "" {
			note = ""
		}

		if !saveDeterminationNote(ctx, deps, auth.actingStaff, outcomeID, note) {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		// Success — signal the sheet to close. No table refresh (the grid is a
		// cell-grid, not a TableConfig); the per-cell icon-state refresh rides the
		// pyeza primitive + page wiring in a later stage.
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"HX-Trigger": `{"formSuccess":true}`},
		}
	})
}

// noteFromVisibleMatrix returns the narrative of outcomeID as the acting principal's
// grade sheet shows it: the ALL matrix when they hold workspace:list, otherwise
// their MINE matrix (own classes plus the sections they review). The matrix query
// applies the principal's row scope, so an outcome outside what the grid would
// show resolves to "". Display only: it never makes a cell editable, and saves
// still go through resolveCellAuthority and the recorded_by re-read.
func noteFromVisibleMatrix(ctx context.Context, deps *NarrativeDeps, perms *types.UserPermissions, templateID, outcomeID string) string {
	if deps.GetOutcomeMatrix == nil || templateID == "" || outcomeID == "" {
		return ""
	}
	scope := matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE
	if perms.Can(drawerScopeEntity, drawerScopeAction) {
		scope = matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL
	}
	matrix, err := deps.GetOutcomeMatrix(ctx, &matrixpb.GetOutcomeMatrixRequest{
		JobTemplateId: templateID,
		Scope:         scope,
	})
	if err != nil || matrix == nil {
		log.Printf("[outcome-matrix] narrative matrix fallback failed for template %s: %v", templateID, err)
		return ""
	}
	for _, row := range matrix.GetRows() {
		for _, cell := range row.GetCells() {
			if cell.GetOutcomeId() == outcomeID {
				return cell.GetDeterminationNote()
			}
		}
	}
	return ""
}

// readOutcome best-effort loads a cell's current outcome for the GET drawer.
// Returns nil on any miss (unwired closure, read error, or a staff-scoped read
// that resolves nothing), which the generated getters read as an empty note and
// no value — the drawer still renders (empty note / read-only). The drawer NEVER
// reads text_value (f9): determination_note is the one canonical narrative field
// post-cutover.
func readOutcome(ctx context.Context, deps *NarrativeDeps, outcomeID string) *taskoutcomepb.TaskOutcome {
	if deps.ReadTaskOutcome == nil {
		return nil
	}
	resp, err := deps.ReadTaskOutcome(ctx, &taskoutcomepb.ReadTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{Id: outcomeID},
	})
	if err != nil {
		log.Printf("[outcome-matrix] narrative read failed for outcome %s: %v", outcomeID, err)
		return nil
	}
	records := resp.GetData()
	if len(records) == 0 {
		return nil
	}
	return records[0]
}

// saveDeterminationNote applies the write-time IDOR re-guard then routes through
// task_outcome:update, setting ONLY determination_note (+ the use case's own
// date_modified). Mirrors record.go's updateCell guard structure exactly — a
// re-read of the persisted record confirms recorded_by == actingStaff before the
// write, defense-in-depth behind the MINE-derived allow-set. The typed grade
// value columns are left untouched: the update request carries only Id +
// DeterminationNote, so the partial-update adapter writes only that column (never
// the numeric/text/categorical/pass_fail value). Returns false on any
// guard/use-case failure.
func saveDeterminationNote(ctx context.Context, deps *NarrativeDeps, actingStaff, outcomeID, note string) bool {
	if deps.ReadTaskOutcome == nil || deps.UpdateTaskOutcome == nil {
		return false
	}
	readResp, err := deps.ReadTaskOutcome(ctx, &taskoutcomepb.ReadTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{Id: outcomeID},
	})
	if err != nil {
		log.Printf("[outcome-matrix] narrative read failed for outcome %s: %v", outcomeID, err)
		return false
	}
	records := readResp.GetData()
	if len(records) == 0 {
		return false
	}
	existing := records[0]

	// IDOR guard: the outcome MUST belong to the acting staff (parity with the
	// cell-edit updateCell guard).
	if existing.GetRecordedBy() != actingStaff {
		log.Printf("[outcome-matrix] narrative IDOR blocked: staff %s tried to edit the note on outcome %s owned by %s",
			actingStaff, outcomeID, existing.GetRecordedBy())
		return false
	}

	req := &taskoutcomepb.UpdateTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{
			Id:                outcomeID,
			DeterminationNote: &note,
		},
	}
	if _, err := deps.UpdateTaskOutcome(ctx, req); err != nil {
		log.Printf("[outcome-matrix] narrative update failed for outcome %s: %v", outcomeID, err)
		return false
	}
	return true
}
