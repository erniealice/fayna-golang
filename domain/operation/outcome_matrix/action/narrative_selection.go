package action

// narrative_selection.go — the descriptor list in the narrative drawer.
//
// A grader may replace a cell's note with the wording another selectable set
// holds for the SAME rating value and criterion. The rating is never changed, and
// the next rating save (including a same-value re-save) returns the note to the
// default set's wording, because record.go is not involved here.
//
// The POST carries only an entry id. The text comes from the resolver, for the
// server-derived cell and the row's current value, so a forged id can select
// nothing the GET would not have offered. Saves never set
// IncludeSelectionOptions; only this file does. The resolver returns options only
// while the cell's own phase is in progress, so a phase sent for review shows no
// list and refuses a pick.

import (
	"context"
	"log"
	"net/http"
	"strings"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// selectableEntry is one wording a grader can pick for a cell's current value.
type selectableEntry struct {
	entryID   string
	setID     string
	text      string
	label     string
	isDefault bool
}

// selectableEntries asks the resolver for one cell's options and keeps the entry
// matching value in each set. Fewer than two means there is nothing to choose,
// which returns nil, as does every resolver failure or non-RESOLVED status.
func selectableEntries(ctx context.Context, deps *NarrativeDeps, cell srvCell, value float64) []selectableEntry {
	if deps.ResolveCellRatingDescriptions == nil || !cell.descriptionMode {
		return nil
	}
	resp, err := deps.ResolveCellRatingDescriptions(ctx, &matrixpb.ResolveCellRatingDescriptionsRequest{
		Cells: []*matrixpb.CellRatingRef{{
			JobId:             cell.jobID,
			JobTaskId:         cell.jobTaskID,
			OutcomeCriteriaId: cell.criteriaID,
		}},
		IncludeSelectionOptions: true,
	})
	if err != nil {
		log.Printf("[outcome-matrix] descriptor options unavailable for job=%s task=%s criteria=%s: %v", cell.jobID, cell.jobTaskID, cell.criteriaID, err)
		return nil
	}
	if resp == nil || !resp.GetSuccess() || len(resp.GetResults()) != 1 {
		log.Printf("[outcome-matrix] descriptor options unavailable for job=%s task=%s criteria=%s: resolver returned success=%v with %d results", cell.jobID, cell.jobTaskID, cell.criteriaID, resp.GetSuccess(), len(resp.GetResults()))
		return nil
	}
	res := resp.GetResults()[0]
	ref := res.GetCell()
	if ref.GetJobId() != cell.jobID || ref.GetJobTaskId() != cell.jobTaskID || ref.GetOutcomeCriteriaId() != cell.criteriaID {
		log.Printf("[outcome-matrix] descriptor options: resolver answered for another cell (job=%s task=%s criteria=%s)", cell.jobID, cell.jobTaskID, cell.criteriaID)
		return nil
	}
	if res.GetStatus() != enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED {
		return nil
	}
	if ratingDescriptionsFromList(res.GetDescriptions()).match(value) == nil {
		return nil
	}

	var out []selectableEntry
	for _, opt := range res.GetSelectionOptions() {
		m := ratingDescriptionsFromList(opt.GetDescriptions()).match(value)
		if m == nil || m.entryID == "" {
			continue
		}
		out = append(out, selectableEntry{
			entryID:   m.entryID,
			setID:     opt.GetRatingDescriptionSetId(),
			text:      m.description,
			label:     opt.GetOptionLabel(),
			isDefault: opt.GetIsDefault(),
		})
	}
	if len(out) < 2 {
		return nil
	}
	return out
}

// loadDescriptorOptions builds the GET list. Any failure hides the list and the
// drawer still renders. The pre-selected option is the one whose text equals the
// stored note, preferring the default when two sets hold identical text.
func loadDescriptorOptions(ctx context.Context, deps *NarrativeDeps, cell srvCell, outcome *taskoutcomepb.TaskOutcome) ([]types.RadioOption, string) {
	if outcome == nil || outcome.NumericValue == nil {
		return nil, ""
	}
	entries := selectableEntries(ctx, deps, cell, outcome.GetNumericValue())
	if len(entries) == 0 {
		return nil, ""
	}
	labels := deps.Labels.Narrative
	note := strings.TrimSpace(outcome.GetDeterminationNote())
	options := make([]types.RadioOption, 0, len(entries))
	selected := ""
	for _, e := range entries {
		template := labels.DescriptorOption
		if e.isDefault {
			template = labels.DescriptorOptionDefault
		}
		options = append(options, types.RadioOption{
			Value:       e.entryID,
			Label:       strings.ReplaceAll(template, "{label}", e.label),
			Description: e.text,
		})
		if e.text == note && (selected == "" || e.isDefault) {
			selected = e.entryID
		}
	}
	return options, selected
}

// applyDescriptorSelection writes the wording for a posted entry id. Every miss
// denies with nothing written. The write is conditional on the row just read, so
// a rating changed since then cannot take a note chosen for the old rating.
func applyDescriptorSelection(ctx context.Context, deps *NarrativeDeps, auth *cellAuthority, outcomeID, entryID string) view.ViewResult {
	denied := view.HTMXError(deps.Labels.Errors.PermissionDenied)
	notAvailable := view.HTMXError(deps.Labels.Errors.DescriptorNotAvailable)

	cell, ok := auth.byOutcome[outcomeID]
	if !ok || !auth.allowedUpdate[outcomeID] || !cell.descriptionMode {
		log.Printf("[outcome-matrix] descriptor pick blocked: outcome %s is not an editable description cell", outcomeID)
		return denied
	}
	if deps.UpdateTaskOutcomeIfUnchanged == nil || deps.ReadTaskOutcome == nil {
		log.Printf("[outcome-matrix] descriptor pick blocked for outcome %s: conditional write is unwired", outcomeID)
		return denied
	}

	readResp, err := deps.ReadTaskOutcome(ctx, &taskoutcomepb.ReadTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{Id: outcomeID},
	})
	if err != nil || len(readResp.GetData()) == 0 {
		log.Printf("[outcome-matrix] descriptor pick read failed for outcome %s: %v", outcomeID, err)
		return denied
	}
	existing := readResp.GetData()[0]
	if existing.GetRecordedBy() != auth.actingStaff {
		log.Printf("[outcome-matrix] descriptor pick IDOR blocked: staff %s on outcome %s owned by %s", auth.actingStaff, outcomeID, existing.GetRecordedBy())
		return denied
	}
	if existing.NumericValue == nil {
		return notAvailable
	}
	if entryID == "" {
		return view.HTMXError(deps.Labels.Errors.DescriptorChooseWording)
	}

	var picked *selectableEntry
	entries := selectableEntries(ctx, deps, cell, existing.GetNumericValue())
	for i := range entries {
		if entries[i].entryID == entryID {
			picked = &entries[i]
			break
		}
	}
	if picked == nil {
		log.Printf("[outcome-matrix] descriptor pick refused: entry %s is not selectable for outcome %s at its current value", entryID, outcomeID)
		return notAvailable
	}

	note := picked.text
	_, conflict, err := deps.UpdateTaskOutcomeIfUnchanged(ctx, &taskoutcomepb.UpdateTaskOutcomeRequest{
		Data: &taskoutcomepb.TaskOutcome{Id: outcomeID, DeterminationNote: &note},
	}, existing)
	if err != nil {
		log.Printf("[outcome-matrix] descriptor pick write failed for outcome %s: %v", outcomeID, err)
		return denied
	}
	if conflict {
		log.Printf("[outcome-matrix] descriptor pick CONFLICT for outcome %s: cell changed since read", outcomeID)
		return view.HTMXError(deps.Labels.Errors.CellChangedRetry)
	}

	log.Printf("[outcome-matrix] descriptor picked: outcome=%s entry=%s set=%s staff=%s", outcomeID, picked.entryID, picked.setID, auth.actingStaff)
	return view.ViewResult{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"HX-Trigger": `{"formSuccess":true}`},
	}
}
