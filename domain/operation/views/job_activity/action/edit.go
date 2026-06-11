package action

import (
	"context"
	"log"
	"math"
	"net/http"

	jobactivityform "github.com/erniealice/fayna-golang/domain/operation/views/job_activity/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// NewEditAction creates the job activity update action (GET = pre-filled form, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			id := viewCtx.Request.URL.Query().Get("id")
			if id == "" {
				return view.HTMXError(deps.Labels.Errors.IDRequired)
			}

			readResp, err := deps.ReadJobActivity(ctx, &jobactivitypb.ReadJobActivityRequest{
				Data: &jobactivitypb.JobActivity{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job activity %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			desc := ""
			if record.Description != nil {
				desc = *record.Description
			}

			billRate := float64(0)
			if record.BillRate != nil {
				billRate = float64(*record.BillRate) / 100
			}
			billAmount := float64(0)
			if record.BillAmount != nil {
				billAmount = float64(*record.BillAmount) / 100
			}
			return view.OK("job-activity-drawer-form", &jobactivityform.Data{
				FormAction:     route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:         true,
				ID:             id,
				JobID:          record.GetJobId(),
				EntryType:      record.GetEntryType().String(),
				Description:    desc,
				Quantity:       record.GetQuantity(),
				UnitCost:       float64(record.GetUnitCost()) / 100,
				Currency:       record.GetCurrency(),
				BillableStatus: record.GetBillableStatus().String(),
				BillRate:       billRate,
				BillAmount:     billAmount,
				Labels:         deps.Labels,
				CommonLabels:   nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		id := r.FormValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

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
		}
		if billRateCentavos > 0 {
			activity.BillRate = &billRateCentavos
		}
		if billAmountCentavos > 0 {
			activity.BillAmount = &billAmountCentavos
		}

		_, err := deps.UpdateJobActivity(ctx, &jobactivitypb.UpdateJobActivityRequest{
			Data: activity,
		})
		if err != nil {
			log.Printf("Failed to update job activity %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activities-table")
	})
}
