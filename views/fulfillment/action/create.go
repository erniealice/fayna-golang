package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"
	fulfillmentform "github.com/erniealice/fayna-golang/views/fulfillment/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// NewAddAction creates the fulfillment add action (GET = form, POST = create).
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("fulfillment", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("fulfillment-drawer-form", &fulfillmentform.Data{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create fulfillment
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request

		supplierID := r.FormValue("supplier_id")
		resp, err := deps.CreateFulfillment(ctx, &fulfillmentpb.CreateFulfillmentRequest{
			Data: &fulfillmentpb.Fulfillment{
				RevenueId:         r.FormValue("revenue_id"),
				SupplierId:        strPtr(supplierID),
				DeliveryMode: r.FormValue("delivery_mode"),
				Notes:             r.FormValue("notes"),
				Status:            "PENDING",
			},
		})
		if err != nil {
			log.Printf("Failed to create fulfillment: %v", err)
			return fayna.HTMXError(err.Error())
		}

		newID := ""
		if data := resp.GetData(); data != nil {
			newID = data.GetId()
		}
		if newID != "" {
			return view.ViewResult{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"HX-Trigger":  `{"formSuccess":true}`,
					"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", newID),
				},
			}
		}

		return fayna.HTMXSuccess("fulfillments-table")
	})
}
