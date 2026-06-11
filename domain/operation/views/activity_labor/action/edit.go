package action

import (
	"context"
	"fmt"
	"log"
	"net/http"

	activitylaborform "github.com/erniealice/fayna-golang/domain/operation/views/activity_labor/form"

	"github.com/erniealice/pyeza-golang/view"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
)

// NewEditAction creates the activity labor update action (GET = form, POST = update).
// The {id} path value is the activity_id (PK of ActivityLabor = FK to JobActivity).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_labor", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		activityID := viewCtx.Request.PathValue("id")
		if activityID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			if deps.ReadActivityLabor == nil {
				// Stub: render an empty form pre-filled with just the activity_id.
				formData := buildEmptyFormData(activityID, deps.Routes, deps.Labels)
				formData.IsEdit = true
				formData.FormAction = editFormAction(deps.Routes, activityID)
				formData.CommonLabels = nil
				return view.OK("activity-labor-drawer-form", formData)
			}

			resp, err := deps.ReadActivityLabor(ctx, &activitylaborpb.ReadActivityLaborRequest{
				Data: &activitylaborpb.ActivityLabor{ActivityId: activityID},
			})
			if err != nil {
				log.Printf("Failed to read activity labor %s: %v", activityID, err)
				return view.HTMXError(fmt.Sprintf("Failed to load labor record: %v", err))
			}
			data := resp.GetData()
			if len(data) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}

			formData := buildFormData(data[0], deps.Routes, deps.Labels)
			formData.FormAction = editFormAction(deps.Routes, activityID)
			formData.CommonLabels = nil
			return view.OK("activity-labor-drawer-form", formData)
		}

		// POST — process the update.
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		if deps.UpdateActivityLabor == nil {
			// TODO: wire UpdateActivityLabor from espyna OperationUseCases.ActivityLabor
			return view.HTMXError("UpdateActivityLabor use case not wired — add ActivityLabor to espyna OperationUseCases")
		}

		r := viewCtx.Request
		hours := parseFormFloat(r.FormValue("hours"))
		rateType := activitylaborform.RateTypeFromString(r.FormValue("rate_type"))
		timeStart := parseDatetimeLocal(r.FormValue("time_start"))
		timeEnd := parseDatetimeLocal(r.FormValue("time_end"))

		record := &activitylaborpb.ActivityLabor{
			ActivityId: activityID,
			StaffId:    r.FormValue("staff_id"),
			Hours:      hours,
			RateType:   rateType,
			TimeStart:  timeStart,
			TimeEnd:    timeEnd,
		}

		_, err := deps.UpdateActivityLabor(ctx, &activitylaborpb.UpdateActivityLaborRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to update activity labor %s: %v", activityID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activity-labor-charge-section")
	})
}
