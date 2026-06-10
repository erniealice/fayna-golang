package action

import (
	"context"
	"math"
	"net/http"

	jobactivityform "github.com/erniealice/fayna-golang/views/job_activity/form"

	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"

	"log"
)

// NewAddAction creates the job activity create action (GET = form, POST = create).
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("job-activity-drawer-form", &jobactivityform.Data{
				FormAction:   deps.Routes.AddURL,
				JobID:        viewCtx.Request.URL.Query().Get("job_id"),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		id := deps.NewID()

		entryTypeStr := r.FormValue("entry_type")
		entryType := parseEntryType(entryTypeStr)

		quantity := parseFormFloat(r.FormValue("quantity"))
		unitCost := parseFormFloat(r.FormValue("unit_cost"))
		billRate := parseFormFloat(r.FormValue("bill_rate"))
		billAmount := parseFormFloat(r.FormValue("bill_amount"))

		description := r.FormValue("description")

		unitCostCentavos := int64(math.Round(unitCost * 100))
		totalCostCentavos := int64(math.Round(quantity * unitCost * 100))
		billRateCentavos := int64(math.Round(billRate * 100))
		billAmountCentavos := int64(math.Round(billAmount * 100))
		// When operator provided a bill_rate but not a bill_amount, derive
		// bill_amount = quantity × bill_rate so the BILLABLE T&M path
		// (flow.md §6) doesn't require a manual second entry.
		if billAmountCentavos == 0 && billRateCentavos > 0 && quantity > 0 {
			billAmountCentavos = int64(math.Round(quantity * billRate * 100))
		}

		activity := &jobactivitypb.JobActivity{
			Id:             id,
			JobId:          r.FormValue("job_id"),
			EntryType:      entryType,
			Quantity:       quantity,
			UnitCost:       unitCostCentavos,
			TotalCost:      totalCostCentavos,
			Currency:       r.FormValue("currency"),
			Description:    &description,
			BillableStatus: parseBillableStatus(r.FormValue("billable_status")),
			ApprovalStatus: jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_DRAFT,
			Active:         true,
		}
		if billRateCentavos > 0 {
			activity.BillRate = &billRateCentavos
		}
		if billAmountCentavos > 0 {
			activity.BillAmount = &billAmountCentavos
		}

		_, err := deps.CreateJobActivity(ctx, &jobactivitypb.CreateJobActivityRequest{
			Data: activity,
		})
		if err != nil {
			log.Printf("Failed to create job activity: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activities-table")
	})
}
