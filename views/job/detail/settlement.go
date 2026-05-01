package detail

import (
	"context"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/types"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobsettlementpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_settlement"
)

// loadSettlementTab populates the PageData with settlement table data.
func loadSettlementTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, jobID string) {
	if deps.ListJobSettlements == nil {
		return
	}

	resp, err := deps.ListJobSettlements(ctx, &jobsettlementpb.ListJobSettlementsRequest{})
	if err != nil {
		log.Printf("Failed to list job settlements for job %s: %v", jobID, err)
		return
	}

	// Collect activity IDs for this job to filter settlements.
	// Settlements are linked via job_activity_id, not directly to job_id.
	activityIDs := map[string]bool{}
	currency := ""
	if deps.ListJobActivities != nil {
		actResp, actErr := deps.ListJobActivities(ctx, &jobactivitypb.ListJobActivitiesRequest{})
		if actErr != nil {
			log.Printf("Failed to list job activities for settlement filter on job %s: %v", jobID, actErr)
		} else {
			for _, a := range actResp.GetData() {
				if a.GetJobId() == jobID {
					activityIDs[a.GetId()] = true
					if currency == "" {
						currency = a.GetCurrency()
					}
				}
			}
		}
	}

	// Filter settlements that belong to this job's activities.
	var settlements []*jobsettlementpb.JobSettlement
	for _, s := range resp.GetData() {
		if activityIDs[s.GetJobActivityId()] {
			settlements = append(settlements, s)
		}
	}

	l := deps.Labels
	pageData.SettlementTable = buildSettlementTable(settlements, l, deps.TableLabels, currency)
}

// buildSettlementTable builds the settlement table config.
func buildSettlementTable(
	settlements []*jobsettlementpb.JobSettlement,
	l fayna.JobLabels,
	tableLabels types.TableLabels,
	currency string,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "target_type", Label: l.Detail.TargetType, WidthClass: "col-5xl"},
		{Key: "target_id", Label: l.Detail.TargetID, NoSort: true},
		{Key: "allocated_amount", Label: l.Detail.AllocatedAmt, WidthClass: "col-3xl"},
		{Key: "settlement_date", Label: l.Detail.SettleDate, WidthClass: "col-3xl"},
		{Key: "status", Label: l.Detail.SettleStatus, WidthClass: "col-2xl"},
	}

	rows := []types.TableRow{}
	for _, s := range settlements {
		id := s.GetId()
		targetType := settlementTargetTypeString(s.GetTargetType())
		targetID := s.GetTargetId()
		allocatedAmtCell := types.MoneyCell(float64(s.GetAllocatedAmount()), currency, true)
		settlementDate := s.GetSettlementDateString()
		status := settlementStatusString(s.GetStatus())

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "badge", Value: targetType, Variant: "info"},
				{Type: "text", Value: targetID},
				allocatedAmtCell,
				types.DateTimeCell(settlementDate, types.DateReadable),
				{Type: "badge", Value: status, Variant: settlementStatusVariant(status)},
			},
			DataAttrs: map[string]string{
				"target_type":      targetType,
				"allocated_amount": allocatedAmtCell.Value,
				"status":           status,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "settlement-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: false,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No settlements",
			Message: "No cost allocations have been settled for this job yet.",
		},
	}
}

// settlementTargetTypeString converts a SettlementTargetType enum to a display string.
func settlementTargetTypeString(t jobsettlementpb.SettlementTargetType) string {
	switch t {
	case jobsettlementpb.SettlementTargetType_SETTLEMENT_TARGET_TYPE_INVOICE_LINE:
		return "invoice_line"
	case jobsettlementpb.SettlementTargetType_SETTLEMENT_TARGET_TYPE_INVENTORY_ASSET:
		return "inventory_asset"
	case jobsettlementpb.SettlementTargetType_SETTLEMENT_TARGET_TYPE_WIP_ACCOUNT:
		return "wip_account"
	case jobsettlementpb.SettlementTargetType_SETTLEMENT_TARGET_TYPE_OVERHEAD_POOL:
		return "overhead_pool"
	case jobsettlementpb.SettlementTargetType_SETTLEMENT_TARGET_TYPE_WRITE_OFF:
		return "write_off"
	default:
		return "unspecified"
	}
}

// settlementStatusString converts a SettlementStatus enum to a display string.
func settlementStatusString(s jobsettlementpb.SettlementStatus) string {
	switch s {
	case jobsettlementpb.SettlementStatus_SETTLEMENT_STATUS_PENDING:
		return "pending"
	case jobsettlementpb.SettlementStatus_SETTLEMENT_STATUS_SETTLED:
		return "settled"
	case jobsettlementpb.SettlementStatus_SETTLEMENT_STATUS_REVERSED:
		return "reversed"
	default:
		return "pending"
	}
}

// settlementStatusVariant returns the badge variant for a settlement status string.
func settlementStatusVariant(status string) string {
	switch status {
	case "pending":
		return "warning"
	case "settled":
		return "success"
	case "reversed":
		return "danger"
	default:
		return "default"
	}
}
