package action

import (
	"context"
	"log"
	"net/http"
	"time"

	fayna "github.com/erniealice/fayna-golang"
	fulfillmentform "github.com/erniealice/fayna-golang/views/fulfillment/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewEditAction creates the fulfillment edit action (GET = form, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("fulfillment", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			// Fetch current fulfillment data for the edit form via GetFulfillmentItemPageData.
			// We use GetFulfillmentItemPageData to avoid adding a separate GetFulfillment dep.
			readResp, err := deps.GetFulfillmentItemPageData(ctx, &fulfillmentpb.GetFulfillmentItemPageDataRequest{
				Id: id,
			})
			if err != nil {
				log.Printf("Failed to read fulfillment %s: %v", id, err)
				return fayna.HTMXError(deps.Labels.Errors.LoadFailed)
			}
			f := readResp.GetFulfillment()
			if f == nil {
				return fayna.HTMXError(deps.Labels.Errors.LoadFailed)
			}

			supplierID := ""
			if f.GetSupplierId() != "" {
				supplierID = f.GetSupplierId()
			}

			scheduledAt := ""
			if ts := f.GetScheduledAt(); ts != nil {
				scheduledAt = ts.AsTime().UTC().Format("2006-01-02T15:04")
			}

			return view.OK("fulfillment-edit-form", &fulfillmentform.Data{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				RevenueID:    f.GetRevenueId(),
				SupplierID:   supplierID,
				Method:       f.GetDeliveryMode(),
				ScheduledAt:  scheduledAt,
				Notes:        f.GetNotes(),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — update fulfillment fields (not status — that's transition)
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request

		supplierID := r.FormValue("supplier_id")

		var scheduledAtProto *timestamppb.Timestamp
		if raw := r.FormValue("scheduled_at"); raw != "" {
			parsed, err := time.Parse("2006-01-02T15:04", raw)
			if err != nil {
				return fayna.HTMXError("Invalid form data")
			}
			scheduledAtProto = timestamppb.New(parsed.UTC())
		}

		_, err := deps.UpdateFulfillment(ctx, &fulfillmentpb.UpdateFulfillmentRequest{
			Data: &fulfillmentpb.Fulfillment{
				Id:           id,
				SupplierId:   strPtr(supplierID),
				DeliveryMode: r.FormValue("delivery_mode"),
				ScheduledAt:  scheduledAtProto,
				Notes:        r.FormValue("notes"),
			},
		})
		if err != nil {
			log.Printf("Failed to update fulfillment %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true}`,
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", id),
			},
		}
	})
}
