package action

import (
	"context"
	"log"
	"strconv"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// NewReturnAction creates the fulfillment return initiation action (POST only).
// It creates a FulfillmentReturn record; return items are managed separately.
func NewReturnAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("fulfillment", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return fayna.HTMXError("Fulfillment ID is required")
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		reason := r.FormValue("reason")
		if reason == "" {
			return fayna.HTMXError("Reason is required")
		}

		returnData := &fulfillmentpb.FulfillmentReturn{
			FulfillmentId: id,
			Reason:        reason,
			Notes:         r.FormValue("notes"),
			Status:        "PENDING",
		}

		// Parse optional refund amount.
		if amtStr := r.FormValue("refund_amount"); amtStr != "" {
			if amt, err := strconv.ParseFloat(amtStr, 64); err == nil {
				returnData.RefundAmount = &amt
			}
		}

		// Parse optional currency.
		if currency := r.FormValue("currency"); currency != "" {
			returnData.Currency = currency
		}

		_, err := deps.CreateFulfillmentReturn(ctx, returnData)
		if err != nil {
			log.Printf("Failed to create return for fulfillment %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		// Redirect back to detail page returns tab.
		return view.ViewResult{
			Headers: map[string]string{
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", id) + "?tab=returns",
			},
		}
	})
}
