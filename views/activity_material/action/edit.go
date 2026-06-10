package action

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/erniealice/pyeza-golang/view"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
)

// NewEditAction creates the activity material update action (GET = form, POST = update).
// The {id} path value is the activity_id (PK of ActivityMaterial = FK to JobActivity).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_material", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		activityID := viewCtx.Request.PathValue("id")
		if activityID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			if deps.ReadActivityMaterial == nil {
				// Stub: render an empty form pre-filled with just the activity_id.
				formData := buildEmptyFormData(activityID, deps.Routes, deps.Labels)
				formData.IsEdit = true
				formData.FormAction = editFormAction(deps.Routes, activityID)
				formData.CommonLabels = nil
				return view.OK("activity-material-drawer-form", formData)
			}

			resp, err := deps.ReadActivityMaterial(ctx, &activitymaterialpb.ReadActivityMaterialRequest{
				Data: &activitymaterialpb.ActivityMaterial{ActivityId: activityID},
			})
			if err != nil {
				log.Printf("Failed to read activity material %s: %v", activityID, err)
				return view.HTMXError(fmt.Sprintf("Failed to load material record: %v", err))
			}
			data := resp.GetData()
			if len(data) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}

			formData := buildFormData(data[0], deps.Routes, deps.Labels)
			formData.FormAction = editFormAction(deps.Routes, activityID)
			formData.CommonLabels = nil
			return view.OK("activity-material-drawer-form", formData)
		}

		// POST — process the update.
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		if deps.UpdateActivityMaterial == nil {
			// TODO: wire UpdateActivityMaterial from espyna OperationUseCases.ActivityMaterial
			return view.HTMXError("UpdateActivityMaterial use case not wired — add ActivityMaterial to espyna OperationUseCases")
		}

		r := viewCtx.Request
		lotNumber := r.FormValue("lot_number")
		locationID := r.FormValue("location_id")
		record := &activitymaterialpb.ActivityMaterial{
			ActivityId:    activityID,
			ProductId:     r.FormValue("product_id"),
			UnitOfMeasure: r.FormValue("unit_of_measure"),
			LotNumber:     &lotNumber,
			LocationId:    &locationID,
		}

		_, err := deps.UpdateActivityMaterial(ctx, &activitymaterialpb.UpdateActivityMaterialRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to update activity material %s: %v", activityID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activity-material-charge-section")
	})
}
