package detail

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erniealice/pyeza-golang/route"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// loadChargeTab populates charge-tab fields on pageData when the charge tab is active.
//
// Dispatch logic:
//   - LABOR       → read ActivityLabor, populate ChargeLabor.
//   - MATERIAL    → read ActivityMaterial, populate ChargeMaterial.
//   - EXPENSE     → read ActivityExpense, populate ChargeExpense.
//   - EQUIPMENT / SUBCONTRACT → Wave 3 placeholder (no dedicated charge table yet).
//   - Other       → no-op.
//
// The function is a no-op when tab is not "charge" — avoids the read when the
// operator is on a different tab.
func loadChargeTab(
	ctx context.Context,
	deps *DetailViewDeps,
	pageData *PageData,
	activityID string,
	entryType jobactivitypb.EntryType,
	activeTab string,
) {
	if activeTab != "charge" {
		return
	}

	switch entryType {
	case jobactivitypb.EntryType_ENTRY_TYPE_LABOR:
		loadLaborCharge(ctx, deps, pageData, activityID)

	case jobactivitypb.EntryType_ENTRY_TYPE_MATERIAL:
		loadMaterialCharge(ctx, deps, pageData, activityID)

	case jobactivitypb.EntryType_ENTRY_TYPE_EXPENSE:
		loadExpenseCharge(ctx, deps, pageData, activityID)

	case jobactivitypb.EntryType_ENTRY_TYPE_EQUIPMENT,
		jobactivitypb.EntryType_ENTRY_TYPE_SUBCONTRACT:
		// Wave 3 scope — no dedicated charge table yet.

	default:
		// UNSPECIFIED or unknown — no-op.
	}
}

// loadMaterialCharge reads the ActivityMaterial record and populates pageData.ChargeMaterial.
func loadMaterialCharge(ctx context.Context, deps *DetailViewDeps, pageData *PageData, activityID string) {
	if deps.ReadActivityMaterial == nil {
		return
	}

	resp, err := deps.ReadActivityMaterial(ctx, &activitymaterialpb.ReadActivityMaterialRequest{
		Data: &activitymaterialpb.ActivityMaterial{ActivityId: activityID},
	})
	if err != nil {
		log.Printf("loadMaterialCharge: failed to read activity material %s: %v", activityID, err)
		return
	}
	data := resp.GetData()
	if len(data) == 0 {
		return
	}

	record := data[0]

	productName := ""
	if p := record.GetProduct(); p != nil {
		productName = p.GetName()
	}

	editURL := route.ResolveURL(deps.ActivityMaterialRoutes.EditURL, "id", activityID)

	pageData.ChargeMaterial = &ActivityMaterialView{
		ProductID:     record.GetProductId(),
		ProductName:   productName,
		UnitOfMeasure: record.GetUnitOfMeasure(),
		LotNumber:     record.GetLotNumber(),
		LocationID:    record.GetLocationId(),
		EditURL:       editURL,
	}
}

// loadExpenseCharge reads the ActivityExpense record and populates pageData.ChargeExpense.
func loadExpenseCharge(ctx context.Context, deps *DetailViewDeps, pageData *PageData, activityID string) {
	if deps.ReadActivityExpense == nil {
		return
	}

	resp, err := deps.ReadActivityExpense(ctx, &activityexpensepb.ReadActivityExpenseRequest{
		Data: &activityexpensepb.ActivityExpense{ActivityId: activityID},
	})
	if err != nil {
		log.Printf("loadExpenseCharge: failed to read activity expense %s: %v", activityID, err)
		return
	}
	data := resp.GetData()
	if len(data) == 0 {
		return
	}

	record := data[0]

	editURL := route.ResolveURL(deps.ActivityExpenseRoutes.EditURL, "id", activityID)

	markupPct := ""
	if m := record.MarkupPctOverride; m != nil {
		markupPct = fmt.Sprintf("%.2f", *m)
	}

	pageData.ChargeExpense = &ActivityExpenseView{
		ExpenseCategoryID: record.GetExpenseCategoryId(),
		VendorRef:         record.GetVendorRef(),
		ReceiptURL:        record.GetReceiptUrl(),
		PaymentMethod:     record.GetPaymentMethod(),
		MarkupPctOverride: markupPct,
		EditURL:           editURL,
	}
}

// loadLaborCharge reads the ActivityLabor record and populates pageData.ChargeLabor.
func loadLaborCharge(ctx context.Context, deps *DetailViewDeps, pageData *PageData, activityID string) {
	if deps.ReadActivityLabor == nil {
		// Use case not wired — ChargeLabor stays nil; template renders a gap notice.
		return
	}

	resp, err := deps.ReadActivityLabor(ctx, &activitylaborpb.ReadActivityLaborRequest{
		Data: &activitylaborpb.ActivityLabor{ActivityId: activityID},
	})
	if err != nil {
		log.Printf("loadLaborCharge: failed to read activity labor %s: %v", activityID, err)
		return
	}
	data := resp.GetData()
	if len(data) == 0 {
		// No labor record yet — ChargeLabor stays nil; template renders empty state + Add CTA.
		return
	}

	record := data[0]

	// Resolve the Edit URL for the CTA button in the charge tab.
	editURL := route.ResolveURL(deps.ActivityLaborRoutes.EditURL, "id", activityID)

	pageData.ChargeLabor = &ActivityLaborView{
		StaffID:   record.GetStaffId(),
		Hours:     fmt.Sprintf("%.2f", record.GetHours()),
		RateType:  record.GetRateType().String(),
		TimeStart: formatChargeTimestamp(record.TimeStart),
		TimeEnd:   formatChargeTimestamp(record.TimeEnd),
		EditURL:   editURL,
	}
}

// formatChargeTimestamp converts a Unix int64 pointer to a human-readable string.
func formatChargeTimestamp(ts *int64) string {
	if ts == nil || *ts == 0 {
		return ""
	}
	return time.Unix(*ts, 0).Format("2006-01-02 15:04")
}
