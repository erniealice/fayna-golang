package action

import (
	"context"
	"log"
	"net/http"

	activitylaborform "github.com/erniealice/fayna-golang/domain/operation/activity_labor/form"

	"github.com/erniealice/pyeza-golang/view"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
)

// NewAddAction creates the activity labor create action (GET = form, POST = create).
// The activity_id is sourced from the ?activity_id query param on GET and
// round-tripped as a hidden input on POST. ActivityLabor.activity_id is the PK
// (1:1 with JobActivity) — there is no separate ID generation.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_labor", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			activityID := viewCtx.Request.URL.Query().Get("activity_id")
			formData := buildEmptyFormData(activityID, deps.Routes, deps.Labels)
			formData.FormAction = addFormAction(deps.Routes)
			formData.CommonLabels = nil // injected by ViewAdapter
			return view.OK("activity-labor-drawer-form", formData)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		activityID := r.FormValue("activity_id")
		if activityID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if deps.CreateActivityLabor == nil {
			// TODO: wire CreateActivityLabor from espyna OperationUseCases.ActivityLabor
			// when the use case is added. For now return a clear gap error.
			return view.HTMXError("CreateActivityLabor use case not wired — add ActivityLabor to espyna OperationUseCases")
		}

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

		_, err := deps.CreateActivityLabor(ctx, &activitylaborpb.CreateActivityLaborRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to create activity labor for activity %s: %v", activityID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activity-labor-charge-section")
	})
}
