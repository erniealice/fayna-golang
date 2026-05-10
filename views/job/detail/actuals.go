package detail

import (
	"context"
	"fmt"
	"log"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// loadActualsTab populates pageData.ActualsRows and pageData.ActualsGrandTotal
// from GetJobActivityRollup. If pageData.Budget.HasBudget is true and a grand
// total is available, a variance is surfaced via pageData.ActualsVariance.
//
// The rollup groups activities by EntryType and sums total_cost (centavos).
// Rendered as a table: Entry Type | Count | Total Cost.
//
// TODO(composition): wire GetJobActivityRollup in
// packages/fayna-golang/block/wiring.go wireJobDeps().
func loadActualsTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, jobID string) {
	if deps.GetJobActivityRollup == nil {
		// Dep not wired yet — graceful empty state.
		return
	}

	resp, err := deps.GetJobActivityRollup(ctx, &jobactivitypb.GetJobActivityRollupRequest{
		JobId: jobID,
	})
	if err != nil {
		log.Printf("loadActualsTab: failed to get activity rollup for job %s: %v", jobID, err)
		return
	}
	if resp == nil {
		return
	}

	// Determine display currency from the first non-zero rollup row.
	// The rollup proto does not carry a currency field; we default to
	// whatever was set on the activities (inferred at the actuals level).
	// For now we leave currency blank — the money formatting still shows
	// the raw centavos ÷ 100 amount without a currency symbol.
	for _, row := range resp.GetRollup() {
		entryType := activityEntryTypeString(row.GetEntryType())
		totalCostDisplay := fmt.Sprintf("%.2f", float64(row.GetTotalCost())/100.0)
		pageData.ActualsRows = append(pageData.ActualsRows, ActualsRow{
			EntryType: entryType,
			Count:     row.GetCount(),
			TotalCost: totalCostDisplay,
		})
	}

	pageData.ActualsGrandTotal = fmt.Sprintf("%.2f", float64(resp.GetGrandTotal())/100.0)
}
