package outcome_matrix

import (
	"strconv"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// Synthetic roll-up ColumnKeys shared by the grid page (list) and the save ack
// (action), so a repaint after an autosave addresses exactly the cells the page
// rendered. Collision-proof against real cell keys, which are always
// "{job_template_task_id}:{outcome_criteria_id}" (two UUIDs).
const (
	RatingKeyPrefix = "rating:"
	TotalKeyPrefix  = "total:"
	FinalColumnKey  = "rating:final"
)

// FormatSummaryScore renders a stored raw composite for display. PRESENCE, not
// value, decides blankness: an UNSET summary_score returns "" (the component
// renders the standard "—"), while a stored 0 returns "0". 'f' with precision -1
// prints the shortest round-tripping decimal (28 → "28"); rounding is a
// write-path duty of the scoring scheme, never a display decision.
func FormatSummaryScore(pe *matrixpb.OutcomeSummaryPhaseEntry) string {
	if pe == nil || pe.SummaryScore == nil {
		return ""
	}
	return strconv.FormatFloat(pe.GetSummaryScore(), 'f', -1, 64)
}

// RollupCells projects one roster row into the derived cells the grid shows
// for it (column key → display value). Every phase in phaseIDs gets an entry —
// "" when the roster carries no stored composite — so a retired summary clears
// its cells instead of leaving a stale value painted.
func RollupCells(rr *matrixpb.OutcomeSummaryRosterRow, phaseIDs []string) map[string]string {
	out := make(map[string]string, 2*len(phaseIDs)+1)
	for _, id := range phaseIDs {
		out[TotalKeyPrefix+id] = ""
		out[RatingKeyPrefix+id] = ""
	}
	if rr == nil {
		out[FinalColumnKey] = ""
		return out
	}
	for _, pe := range rr.GetPhases() {
		id := pe.GetJobTemplatePhaseId()
		out[TotalKeyPrefix+id] = FormatSummaryScore(pe)
		out[RatingKeyPrefix+id] = pe.GetScaledLabel()
	}
	out[FinalColumnKey] = rr.GetYearFinalLabel()
	return out
}
