package action

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// resultEvent is the HX-Trigger custom-event name the pyeza cell-grid AutoSave
// client (cell-grid.js) binds to for per-cell acks. It MUST equal
// CellGridConfig.ResultEventName()'s default and cell-grid.js's RESULT_EVENT
// constant. The grid renders it as the form's data-result-event attribute; this
// action currently only emits the default (the grid leaves ResultEvent unset).
const resultEvent = "omcell-result"

// NewRecordAction creates the outcome matrix save POST handler. It serves BOTH
// the manual whole-grid batch (legacy, no save_mode — the a11y/retry fallback)
// AND the W2 per-cell micro-batch (save_mode=cell) over the SAME endpoint and
// form vocabulary; the security core (server-side MINE re-derivation + IDOR
// ownership) is identical for both.
//
// ── Form encoding (both modes; unchanged from the batch contract) ──────────
//   - cells.{outcome_id}={value}          → UPDATE existing (IDOR-guarded)
//   - new.{job_task_id}:{criteria_id}={v} → CREATE new (status=active/active=true)
//   - save_mode=cell                      → opt into the per-cell ack response
//     (absent → legacy aggregate formError/formSuccess response)
// The value's SHAPE never chooses the persisted column — the criterion's
// criteria_type (re-derived from the server matrix, never the POST) does.
//
// ── W2 micro-batch protocol (Q-GSE-3 locked: micro-batch + per-cell acks) ───
// The AutoSave client (packages/pyeza-golang/web/js/components/cell-grid.js)
// coalesces dirty cells for ~150ms and POSTs them single-flight with a hidden
// save_mode=cell, each cell named exactly as above. This action processes each
// cell INDEPENDENTLY (a partial failure never aborts the batch nor 500s) and
// returns one ack per submitted cell via an HX-Trigger JSON event:
//
//	HX-Trigger: {"omcell-result":{"cells":[
//	  {"key":"cells.<id>","ok":true,"outcomeId":"<id>","value":"86","ratingFresh":true},
//	  {"key":"new.<jt>:<cr>","ok":true,"outcomeId":"<new-id>","value":"P","ratingFresh":true},
//	  {"key":"cells.<id>","ok":false,"error":"value_rejected"}
//	]}}
//
// Ack item fields (must match cell-grid.js handleResult verbatim):
//   - key         the posted field name (cells.* / new.*) — the client keys its
//                 live input by this to paint state.
//   - ok          the cell persisted.
//   - outcomeId   the canonical task_outcome id. MANDATORY on a CREATE (and on a
//                 lost-ack new.* retry resolved to an existing outcome): the
//                 client renames the input new.{jt}:{cr} → cells.{outcomeId} IN
//                 PLACE so the next save UPDATES instead of re-creating.
//   - value       the normalized canonical stored value; becomes the client's
//                 new saved-baseline (data-saved-value). pass_fail is normalized
//                 to "true"/"false" so it round-trips the <select> option values.
//   - ratingFresh (Q-GSE-5 inline recompute) present only for ACADEMIC cells:
//                 true  → the affected phase (and job) roll-up recomputed (or was
//                         authoritative/frozen and correctly left pinned);
//                 false → the grade PERSISTED but its summary recompute failed or
//                         is unwired — the rating is stale + retryable, NEVER a
//                         reason to report the saved cell as failed. Omitted for
//                         non-academic (deportment/text) cells (no roll-up rating).
//   - ratingNotRecomputed  optional bounded reason string for observability when a
//                 rating was deliberately not recomputed (authoritative_frozen /
//                 non_academic). The client ignores it; it documents the split.
//
// ── New-id handshake + idempotent create-retry ─────────────────────────────
// Because the response can be lost, a client may re-POST new.{jt}:{cr} for a cell
// it already created. The per-POST MINE re-derivation is authoritative: that
// address now carries an outcome owned by the acting staff, so it is in
// allowedUpdate (not allowedCreate) and this action resolves the retry to an
// UPDATE of the existing outcome — returning its outcomeId so the client renames
// — never a duplicate insert. Combined with the client's single-flight queue this
// needs no idempotency table.
//
// ── Recompute freshness split (Q-GSE-5) ────────────────────────────────────
// After the per-cell dispatch, the affected phase ids then job ids are deduped
// FROM THE SERVER-DERIVED matrix (OutcomeCell.job_phase_id / job_id — never a
// browser value) and ComputePhaseOutcome then ComputeJobOutcome run once each.
// A frozen (is_authoritative) job_outcome_summary is left pinned (recompute
// skipped, rating still fresh). A compute failure marks only that phase/job stale
// (ratingFresh:false on its cells); the grade itself stays saved.
func NewRecordAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		// Per-verb gates: update cells need task_outcome:update, new cells need
		// task_outcome:create — an update-only save must not be denied for a
		// missing create grant (and vice versa).
		hasCreate := perms.Can("task_outcome", "create")
		hasUpdate := perms.Can("task_outcome", "update")
		if !hasCreate && !hasUpdate {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		var actingStaff string
		if deps.ResolveStaff != nil {
			actingStaff, _ = deps.ResolveStaff(ctx)
		}
		if actingStaff == "" {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		cellMode := viewCtx.Request.Form.Get("save_mode") == "cell"

		// Re-derive the acting principal's MINE-scoped matrix server-side and
		// only accept cell addresses it contains. The POST body's outcome_id /
		// job_task_id keys are attacker-controlled; the grid the server itself
		// scopes to this principal is the ONLY authority on which cells are
		// addressable (a forged job_task_id outside the principal's roster must
		// never reach the create/update use cases).
		templateID := viewCtx.Request.PathValue("id")
		if templateID == "" || deps.GetOutcomeMatrix == nil {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		matrix, err := deps.GetOutcomeMatrix(ctx, &matrixpb.GetOutcomeMatrixRequest{
			JobTemplateId: templateID,
			Scope:         matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE,
		})
		if err != nil || matrix == nil {
			log.Printf("[outcome-matrix] scope re-derivation failed for template %s: %v", templateID, err)
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		// The column tree carries each criterion's enforcement entity — the
		// criteria_type there (not the value's shape) decides which typed
		// column a create writes, so a crafted value can never land in the
		// wrong column.
		typeByColKey := make(map[string]enums.CriteriaType)
		for _, phase := range matrix.GetPhases() {
			for _, task := range phase.GetTasks() {
				for _, crit := range task.GetCriteria() {
					typeByColKey[crit.GetColumnKey()] = crit.GetCriteria().GetCriteriaType()
				}
			}
		}

		allowedCreate := make(map[string]enums.CriteriaType) // "{job_task_id}:{criteria_id}" → criteria type
		allowedUpdate := make(map[string]bool)               // outcome_id → updatable
		byOutcome := make(map[string]srvCell)                // outcome_id → server cell (recompute keys)
		byCreateAddr := make(map[string]srvCell)             // "{job_task_id}:{criteria_id}" → server cell
		for _, row := range matrix.GetRows() {
			// Cells are keyed by column_key "{job_template_task_id}:{criteria_id}";
			// the criteria id half addresses the create, paired with the cell's
			// own job_task instance id.
			for colKey, cell := range row.GetCells() {
				if !cell.GetEditable() {
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
					jobPhaseID: cell.GetJobPhaseId(),
					jobID:      cell.GetJobId(),
				}
				if cell.GetOutcomeId() != "" {
					allowedUpdate[cell.GetOutcomeId()] = true
					byOutcome[cell.GetOutcomeId()] = sc
					// A recorded cell's create-address maps back to its outcome so
					// a lost-ack new.* retry resolves to an UPDATE, never a dup.
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

		// Per-cell dispatch. Each addressed cell yields exactly one ack (cellMode)
		// / counts toward saved|failed (legacy). A partial failure never aborts.
		var acks []cellAck
		for key, vals := range viewCtx.Request.Form {
			if len(vals) == 0 {
				continue
			}
			raw := strings.TrimSpace(vals[0])

			switch {
			case strings.HasPrefix(key, "cells."):
				outcomeID := strings.TrimPrefix(key, "cells.")
				if outcomeID == "" {
					continue
				}
				if raw == "" {
					// Blank on an existing cell is "no change", never an
					// overwrite — a blank pass/fail must not coerce to false.
					// Persisted clear is deferred (Q-GSE-11); in cell mode surface
					// an explicit item failure so the client doesn't hang on an
					// unacked cell (rather than silently pretending to save).
					if cellMode {
						acks = append(acks, cellAck{key: key, errMsg: "clear_not_supported"})
					}
					continue
				}
				if !hasUpdate || !allowedUpdate[outcomeID] {
					log.Printf("[outcome-matrix] update blocked: outcome %s not addressable for staff %s", outcomeID, actingStaff)
					acks = append(acks, cellAck{key: key, errMsg: "not_editable"})
					continue
				}
				normVal, ok := updateCell(ctx, deps, actingStaff, outcomeID, raw)
				sc := byOutcome[outcomeID]
				acks = append(acks, cellAck{
					key: key, ok: ok, outcomeID: outcomeID, value: normVal,
					academic: isAcademic(sc.ct), jobPhaseID: sc.jobPhaseID, jobID: sc.jobID,
					errMsg: failMsg(ok, "value_rejected"),
				})

			case strings.HasPrefix(key, "new."):
				addr := strings.TrimPrefix(key, "new.")
				jobTaskID, criteriaID, ok := splitCellAddr(addr)
				if !ok || raw == "" {
					// Empty new-cell: nothing to create, not an error. (The client
					// never queues a blank fresh cell — it isn't dirty — so cell
					// mode need not ack it.)
					continue
				}
				createAddr := jobTaskID + ":" + criteriaID
				if ct, addressable := allowedCreate[createAddr]; hasCreate && addressable {
					newID, normVal, done := createCell(ctx, deps, actingStaff, jobTaskID, criteriaID, ct, raw)
					sc := byCreateAddr[createAddr]
					acks = append(acks, cellAck{
						key: key, ok: done, outcomeID: newID, value: normVal,
						academic: isAcademic(ct), jobPhaseID: sc.jobPhaseID, jobID: sc.jobID,
						errMsg: failMsg(done, "value_rejected"),
					})
					continue
				}
				// Lost-ack idempotent retry: the address now carries an existing
				// outcome owned by the acting staff (created by a prior batch whose
				// ack was lost) → resolve to an UPDATE, return its id so the client
				// renames new.* → cells.*. Never a duplicate insert.
				if sc, ok := byCreateAddr[createAddr]; ok && sc.outcomeID != "" && hasUpdate && allowedUpdate[sc.outcomeID] {
					normVal, done := updateCell(ctx, deps, actingStaff, sc.outcomeID, raw)
					acks = append(acks, cellAck{
						key: key, ok: done, outcomeID: sc.outcomeID, value: normVal,
						academic: isAcademic(sc.ct), jobPhaseID: sc.jobPhaseID, jobID: sc.jobID,
						errMsg: failMsg(done, "value_rejected"),
					})
					continue
				}
				log.Printf("[outcome-matrix] create blocked: job_task %s criteria %s not addressable for staff %s", jobTaskID, criteriaID, actingStaff)
				acks = append(acks, cellAck{key: key, errMsg: "not_editable"})
			}
		}

		// Inline recompute (Q-GSE-5): dedup the affected phase ids then job ids
		// from the SERVER-DERIVED cells of the successfully-saved ACADEMIC writes
		// (never a browser value), recompute phase→job once each, and classify
		// each id recomputed | frozen | failed. ratingFresh per cell folds in the
		// worst of its phase + job result.
		phaseRes := recomputeIDs(ctx, deps.ComputePhaseOutcome, academicPhaseIDs(acks))
		jobRes := recomputeIDs(ctx, deps.ComputeJobOutcome, academicJobIDs(acks))

		if !cellMode {
			// Legacy aggregate response (manual batch / a11y retry fallback).
			saved, failed := 0, 0
			for _, a := range acks {
				if a.ok {
					saved++
				} else {
					failed++
				}
			}
			if failed > 0 {
				msg := fmt.Sprintf("Saved %d, %d failed", saved, failed)
				return view.ViewResult{
					StatusCode: http.StatusOK,
					Headers:    map[string]string{"HX-Trigger": fmt.Sprintf(`{"formError":%q}`, msg)},
				}
			}
			return view.ViewResult{
				StatusCode: http.StatusOK,
				Headers:    map[string]string{"HX-Trigger": `{"formSuccess":true}`},
			}
		}

		return cellResponse(acks, phaseRes, jobRes)
	})
}

// srvCell is the server-derived truth for one editable matrix cell: its identity
// plus the trusted recompute keys (job_phase_id / job_id) read from the matrix.
type srvCell struct {
	outcomeID  string
	jobTaskID  string
	criteriaID string
	ct         enums.CriteriaType
	jobPhaseID string
	jobID      string
}

// cellAck is the internal per-cell outcome collected during dispatch, projected
// into the omcell-result JSON (cellMode) or the saved/failed counts (legacy).
type cellAck struct {
	key        string
	ok         bool
	outcomeID  string
	value      string
	academic   bool
	jobPhaseID string
	jobID      string
	errMsg     string // set only when !ok (bounded, never echoes the value)
}

// recResult classifies one phase/job recompute for the ratingFresh split.
type recResult int

const (
	recOK     recResult = iota // recompute ran → fresh
	recFrozen                  // authoritative/frozen summary left pinned → fresh, not stale
	recFailed                  // compute failed or unwired → stale + retryable
)

// recomputeIDs runs one recompute closure over a deduped id set and classifies
// each. A nil closure (unwired app / non-postgres) marks every id failed →
// ratingFresh:false (fail-safe; the grade already persisted, never a 500).
func recomputeIDs(ctx context.Context, fn func(context.Context, string) (bool, error), ids []string) map[string]recResult {
	out := make(map[string]recResult, len(ids))
	for _, id := range ids {
		if fn == nil {
			out[id] = recFailed
			continue
		}
		recomputed, err := fn(ctx, id)
		switch {
		case err != nil:
			log.Printf("[outcome-matrix] recompute failed for %s: %v", id, err)
			out[id] = recFailed
		case recomputed:
			out[id] = recOK
		default:
			out[id] = recFrozen // (false,nil) → authoritative/frozen skip
		}
	}
	return out
}

// academicPhaseIDs / academicJobIDs dedup the trusted phase/job ids of the
// successfully-saved academic cells (empty ids excluded).
func academicPhaseIDs(acks []cellAck) []string { return dedupIDs(acks, func(a cellAck) string { return a.jobPhaseID }) }
func academicJobIDs(acks []cellAck) []string   { return dedupIDs(acks, func(a cellAck) string { return a.jobID }) }

func dedupIDs(acks []cellAck, pick func(cellAck) string) []string {
	seen := map[string]bool{}
	var out []string
	for _, a := range acks {
		if !a.ok || !a.academic {
			continue
		}
		if id := pick(a); id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// cellResponse projects the acks + recompute results into the omcell-result
// HX-Trigger JSON the AutoSave client parses.
func cellResponse(acks []cellAck, phaseRes, jobRes map[string]recResult) view.ViewResult {
	type item struct {
		Key                 string `json:"key"`
		OK                  bool   `json:"ok"`
		OutcomeID           string `json:"outcomeId,omitempty"`
		Value               string `json:"value,omitempty"`
		RatingFresh         *bool  `json:"ratingFresh,omitempty"`
		RatingNotRecomputed string `json:"ratingNotRecomputed,omitempty"`
		Error               string `json:"error,omitempty"`
	}
	items := make([]item, 0, len(acks))
	for _, a := range acks {
		it := item{Key: a.key, OK: a.ok, OutcomeID: a.outcomeID, Value: a.value}
		if !a.ok {
			it.Error = a.errMsg
			items = append(items, it)
			continue
		}
		if a.academic {
			// Fold the worst of the phase + job recompute result. A failure on
			// either → stale; a frozen skip on the job → fresh but flagged.
			worst := recOK
			if r, ok := phaseRes[a.jobPhaseID]; ok && r > worst {
				worst = r
			}
			if r, ok := jobRes[a.jobID]; ok && r > worst {
				worst = r
			}
			fresh := worst != recFailed
			it.RatingFresh = &fresh
			if worst == recFrozen {
				it.RatingNotRecomputed = "authoritative_frozen"
			}
		} else {
			// Non-academic (deportment/text) cell: no roll-up rating to refresh.
			it.RatingNotRecomputed = "non_academic"
		}
		items = append(items, it)
	}

	payload := map[string]any{resultEvent: map[string]any{"cells": items}}
	body, err := json.Marshal(payload)
	if err != nil {
		// Marshal cannot realistically fail on these types; degrade to a generic
		// success so the client at least clears its saving state.
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"HX-Trigger": `{"formSuccess":true}`},
		}
	}
	return view.ViewResult{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"HX-Trigger": string(body)},
	}
}

// isAcademic reports whether a criterion drives an academic (score-scaled)
// roll-up — only numeric criteria do. Deportment/text/categorical criteria have
// no scaled phase/job summary, so their cells are never recomputed (and never
// reported stale).
func isAcademic(ct enums.CriteriaType) bool {
	switch ct {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return true
	default:
		return false
	}
}

// failMsg returns "" on success, or the given bounded reason code on failure.
func failMsg(ok bool, reason string) string {
	if ok {
		return ""
	}
	return reason
}

// updateCell applies the IDOR guard then routes through task_outcome:update.
// Returns the normalized stored value + whether the write succeeded (false on
// any guard/parse/use-case failure — counted as a failure).
func updateCell(ctx context.Context, deps *Deps, actingStaff, outcomeID, raw string) (string, bool) {
	if deps.ReadTaskOutcome == nil || deps.UpdateTaskOutcome == nil {
		return "", false
	}
	readResp, err := deps.ReadTaskOutcome(ctx, &taskoutcomepb.ReadTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{Id: outcomeID},
	})
	if err != nil {
		log.Printf("[outcome-matrix] read failed for outcome %s: %v", outcomeID, err)
		return "", false
	}
	records := readResp.GetData()
	if len(records) == 0 {
		return "", false
	}
	existing := records[0]

	// IDOR guard: the outcome MUST belong to the acting staff.
	if existing.GetRecordedBy() != actingStaff {
		log.Printf("[outcome-matrix] IDOR blocked: staff %s tried to edit outcome %s owned by %s",
			actingStaff, outcomeID, existing.GetRecordedBy())
		return "", false
	}

	ct := existing.GetCriteriaType()
	req := &taskoutcomepb.UpdateTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{
			Id:                outcomeID,
			CriteriaVersionId: existing.GetCriteriaVersionId(),
		},
	}
	// Fail-closed typed parse on update too (parity with create): a value that
	// does not parse as the criterion's type is an item failure, never a silent
	// no-op that reports success.
	if !applyValueStrict(req.Data, ct, raw) {
		log.Printf("[outcome-matrix] update rejected: value does not parse as %v for outcome %s", ct, outcomeID)
		return "", false
	}

	if _, err := deps.UpdateTaskOutcome(ctx, req); err != nil {
		log.Printf("[outcome-matrix] update failed for outcome %s: %v", outcomeID, err)
		return "", false
	}
	return normalizedValue(ct, req.Data), true
}

// createCell routes a new-cell value through task_outcome:create. The caller
// has already verified the (job_task, criteria) address against the acting
// principal's server-derived matrix — a POST-supplied address alone is never
// authority — and supplies the column's criteria_type from that same matrix,
// so the value lands in the column the criterion dictates (a value's SHAPE is
// attacker-controlled and must never choose the column). Stamps RecordedBy
// (the ownership axis for scope=mine and the read-only gate) and Active=true
// (the proto-entity-status-conventions trap — a new outcome must default
// active). A value that does not parse as the criterion's type fails the cell.
// Returns the new task_outcome id (for the client's new.*→cells.* rename
// handshake) + the normalized stored value.
func createCell(ctx context.Context, deps *Deps, actingStaff, jobTaskID, criteriaID string, ct enums.CriteriaType, raw string) (string, string, bool) {
	if deps.CreateTaskOutcome == nil {
		return "", "", false
	}
	req := &taskoutcomepb.CreateTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{
			JobTaskId:         jobTaskID,
			CriteriaVersionId: criteriaID,
			RecordedBy:        actingStaff,
			Active:            true,
		},
	}
	if !applyValueStrict(req.Data, ct, raw) {
		// Do NOT log the raw value (it is user content); the type + criteria id
		// are sufficient to diagnose a rejected create.
		log.Printf("[outcome-matrix] create rejected: value does not parse as %v for criteria %s", ct, criteriaID)
		return "", "", false
	}

	resp, err := deps.CreateTaskOutcome(ctx, req)
	if err != nil {
		log.Printf("[outcome-matrix] create failed for job_task %s criteria %s: %v", jobTaskID, criteriaID, err)
		return "", "", false
	}
	// The new id backs the mandatory new.*→cells.* rename handshake; without it
	// the client would re-CREATE on its next save (a duplicate).
	newID := ""
	if data := resp.GetData(); len(data) > 0 && data[0] != nil {
		newID = data[0].GetId()
	}
	return newID, normalizedValue(ct, req.Data), true
}

// normalizedValue renders the stored typed value back to its canonical string
// form for the client's saved-baseline. pass_fail is "true"/"false" (matching
// the <select> option values so a saved cell round-trips, not the blank option).
func normalizedValue(ct enums.CriteriaType, data *taskoutcomepb.TaskOutcome) string {
	switch ct {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		if data.NumericValue != nil {
			return strconv.FormatFloat(*data.NumericValue, 'f', -1, 64)
		}
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		if data.PassFailValue != nil {
			if *data.PassFailValue {
				return "true"
			}
			return "false"
		}
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		if data.CategoricalValue != nil {
			return *data.CategoricalValue
		}
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		if data.TextValue != nil {
			return *data.TextValue
		}
	}
	return ""
}

// applyValueStrict sets exactly the typed field the criterion dictates and
// reports whether the raw value parsed as that type. Unknown/unspecified
// criteria types are rejected (fail-closed) — never inferred from the value.
func applyValueStrict(data *taskoutcomepb.TaskOutcome, ct enums.CriteriaType, raw string) bool {
	switch ct {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return false
		}
		data.NumericValue = &f
		return true
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		switch strings.ToLower(raw) {
		case "true", "pass", "on", "1", "yes":
			b := true
			data.PassFailValue = &b
		case "false", "fail", "off", "0", "no":
			b := false
			data.PassFailValue = &b
		default:
			return false
		}
		return true
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		v := raw
		data.CategoricalValue = &v
		return true
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		v := raw
		data.TextValue = &v
		return true
	default:
		return false
	}
}

// splitCellAddr parses "{job_task_id}:{outcome_criteria_id}" into its two parts.
func splitCellAddr(addr string) (jobTaskID, criteriaID string, ok bool) {
	i := strings.LastIndex(addr, ":")
	if i <= 0 || i >= len(addr)-1 {
		return "", "", false
	}
	return addr[:i], addr[i+1:], true
}
