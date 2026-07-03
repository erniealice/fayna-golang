package action

import (
	"context"
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

// NewRecordAction creates the outcome matrix batch-save POST handler.
//
// Form encoding (per view-scope.md §2):
//   - cells.{outcome_id}={value}          → UPDATE existing (IDOR-guarded)
//   - new.{job_task_id}:{criteria_id}={v} → CREATE new (status=active/active=true)
//
// Partial failures are collected (OQ-7): a failed cell does not abort the batch
// nor 500 — the response reports saved/failed counts via HX-Trigger.
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
		for _, row := range matrix.GetRows() {
			// Cells are keyed by column_key "{job_template_task_id}:{criteria_id}";
			// the criteria id half addresses the create, paired with the cell's
			// own job_task instance id.
			for colKey, cell := range row.GetCells() {
				if !cell.GetEditable() {
					continue
				}
				if cell.GetOutcomeId() != "" {
					allowedUpdate[cell.GetOutcomeId()] = true
					continue
				}
				_, criteriaID, ok := splitCellAddr(colKey)
				if ok && cell.GetJobTaskId() != "" {
					allowedCreate[cell.GetJobTaskId()+":"+criteriaID] = typeByColKey[colKey]
				}
			}
		}

		var saved, failed int
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
					continue
				}
				if !hasUpdate || !allowedUpdate[outcomeID] {
					log.Printf("[outcome-matrix] update blocked: outcome %s not addressable for staff %s", outcomeID, actingStaff)
					failed++
					continue
				}
				if updateCell(ctx, deps, actingStaff, outcomeID, raw) {
					saved++
				} else {
					failed++
				}

			case strings.HasPrefix(key, "new."):
				addr := strings.TrimPrefix(key, "new.")
				jobTaskID, criteriaID, ok := splitCellAddr(addr)
				if !ok || raw == "" {
					// Empty new-cell: nothing to create, not an error.
					continue
				}
				ct, addressable := allowedCreate[jobTaskID+":"+criteriaID]
				if !hasCreate || !addressable {
					log.Printf("[outcome-matrix] create blocked: job_task %s criteria %s not addressable for staff %s", jobTaskID, criteriaID, actingStaff)
					failed++
					continue
				}
				if createCell(ctx, deps, actingStaff, jobTaskID, criteriaID, ct, raw) {
					saved++
				} else {
					failed++
				}
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
	})
}

// updateCell applies the IDOR guard then routes through task_outcome:update.
// Returns false on any guard/parse/use-case failure (counted as a failure).
func updateCell(ctx context.Context, deps *Deps, actingStaff, outcomeID, raw string) bool {
	if deps.ReadTaskOutcome == nil || deps.UpdateTaskOutcome == nil {
		return false
	}
	readResp, err := deps.ReadTaskOutcome(ctx, &taskoutcomepb.ReadTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{Id: outcomeID},
	})
	if err != nil {
		log.Printf("[outcome-matrix] read failed for outcome %s: %v", outcomeID, err)
		return false
	}
	records := readResp.GetData()
	if len(records) == 0 {
		return false
	}
	existing := records[0]

	// IDOR guard: the outcome MUST belong to the acting staff.
	if existing.GetRecordedBy() != actingStaff {
		log.Printf("[outcome-matrix] IDOR blocked: staff %s tried to edit outcome %s owned by %s",
			actingStaff, outcomeID, existing.GetRecordedBy())
		return false
	}

	req := &taskoutcomepb.UpdateTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{
			Id:                outcomeID,
			CriteriaVersionId: existing.GetCriteriaVersionId(),
		},
	}
	applyValueTyped(req.Data, existing.GetCriteriaType(), raw)

	if _, err := deps.UpdateTaskOutcome(ctx, req); err != nil {
		log.Printf("[outcome-matrix] update failed for outcome %s: %v", outcomeID, err)
		return false
	}
	return true
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
func createCell(ctx context.Context, deps *Deps, actingStaff, jobTaskID, criteriaID string, ct enums.CriteriaType, raw string) bool {
	if deps.CreateTaskOutcome == nil {
		return false
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
		log.Printf("[outcome-matrix] create rejected: value %q does not parse as %v for criteria %s", raw, ct, criteriaID)
		return false
	}

	if _, err := deps.CreateTaskOutcome(ctx, req); err != nil {
		log.Printf("[outcome-matrix] create failed for job_task %s criteria %s: %v", jobTaskID, criteriaID, err)
		return false
	}
	return true
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

// applyValueTyped sets the correct typed field on the outcome given the known
// criteria_type (update path — the criteria_type comes from the read record).
func applyValueTyped(data *taskoutcomepb.TaskOutcome, ct enums.CriteriaType, raw string) {
	switch ct {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			data.NumericValue = &f
		}
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		b := isTruthy(raw)
		data.PassFailValue = &b
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		v := raw
		data.CategoricalValue = &v
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		v := raw
		data.TextValue = &v
	default:
		applyValueHeuristic(data, raw)
	}
}

// applyValueHeuristic infers the value type from its shape (create path — the
// criteria_type is not available for a brand-new cell).
func applyValueHeuristic(data *taskoutcomepb.TaskOutcome, raw string) {
	if raw == "" {
		return
	}
	switch strings.ToLower(raw) {
	case "true", "on", "pass":
		b := true
		data.PassFailValue = &b
		return
	case "false", "off", "fail":
		b := false
		data.PassFailValue = &b
		return
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		data.NumericValue = &f
		return
	}
	v := raw
	data.TextValue = &v
}

func isTruthy(raw string) bool {
	switch strings.ToLower(raw) {
	case "true", "on", "pass", "1", "yes":
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
