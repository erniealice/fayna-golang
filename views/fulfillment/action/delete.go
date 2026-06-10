package action

import (
	"context"
	"log"

	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// NewDeleteAction creates the fulfillment delete action (POST only).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("fulfillment", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return view.HTMXError("ID is required")
		}

		_, err := deps.DeleteFulfillment(ctx, &fulfillmentpb.DeleteFulfillmentRequest{
			Id: id,
		})
		if err != nil {
			log.Printf("Failed to delete fulfillment %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("fulfillments-table")
	})
}
