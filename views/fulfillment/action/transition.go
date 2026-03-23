package action

import (
	"context"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// NewTransitionAction creates the fulfillment status transition action (POST only).
// It reads "event" from the form and calls TransitionStatus.
// AllowedEvents are determined by the use case — fayna just forwards the request.
func NewTransitionAction(deps *Deps) view.View {
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
		event := r.FormValue("event")
		if event == "" {
			return fayna.HTMXError("Event is required")
		}

		_, err := deps.TransitionStatus(ctx, &fulfillmentpb.TransitionStatusRequest{
			FulfillmentId:     id,
			Event:             event,
			Reason:            r.FormValue("reason"),
			ProviderStatus:    r.FormValue("provider_status"),
			ProviderReference: r.FormValue("provider_reference"),
		})
		if err != nil {
			log.Printf("Failed to transition fulfillment %s via event %s: %v", id, event, err)
			return fayna.HTMXError(deps.Labels.Errors.TransitionFailed)
		}

		// Redirect back to detail page to reflect the new status.
		return view.ViewResult{
			Headers: map[string]string{
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", id),
			},
		}
	})
}
