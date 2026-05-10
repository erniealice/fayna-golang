package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
)

// NewAddAction creates the activity material create action (GET = form, POST = create).
// The activity_id is sourced from the ?activity_id query param on GET and
// round-tripped as a hidden input on POST. ActivityMaterial.activity_id is the PK
// (1:1 with JobActivity) — there is no separate ID generation.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_material", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			activityID := viewCtx.Request.URL.Query().Get("activity_id")
			formData := buildEmptyFormData(activityID, deps.Routes, deps.Labels)
			formData.FormAction = addFormAction(deps.Routes)
			formData.CommonLabels = nil // injected by ViewAdapter
			return view.OK("activity-material-drawer-form", formData)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		activityID := r.FormValue("activity_id")
		if activityID == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if deps.CreateActivityMaterial == nil {
			// TODO: wire CreateActivityMaterial from espyna OperationUseCases.ActivityMaterial
			// when the use case is added. For now return a clear gap error.
			return fayna.HTMXError("CreateActivityMaterial use case not wired — add ActivityMaterial to espyna OperationUseCases")
		}

		lotNumber := r.FormValue("lot_number")
		locationID := r.FormValue("location_id")
		record := &activitymaterialpb.ActivityMaterial{
			ActivityId:    activityID,
			ProductId:     r.FormValue("product_id"),
			UnitOfMeasure: r.FormValue("unit_of_measure"),
			LotNumber:     &lotNumber,
			LocationId:    &locationID,
		}

		_, err := deps.CreateActivityMaterial(ctx, &activitymaterialpb.CreateActivityMaterialRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to create activity material for activity %s: %v", activityID, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activity-material-charge-section")
	})
}
