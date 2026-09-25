package rating_description_set_product_plan

import (
	"context"
	"log"
	"net/http"

	linkform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan/form"

	pyeza "github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	linkpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_product_plan"
)

// NewRelinkAction creates the relink action (GET = drawer, POST = process).
// Handles BOTH the first link (no current link — expected_current_link_id
// empty) and a replacement (expected_current_link_id = the current active
// link id), matching the espyna RelinkRatingDescriptionSetProductPlan use
// case contract (W3-ESPYNA.done).
func NewRelinkAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		q := viewCtx.Request.URL.Query()

		if viewCtx.Request.Method == http.MethodGet {
			perms := view.GetUserPermissions(ctx)
			expectedCurrent := q.Get("expected_current_link_id")
			verb := "create"
			if expectedCurrent != "" {
				verb = "update"
			}
			if !perms.Can("rating_description_set_product_plan", verb) {
				return view.HTMXError(deps.Labels.Errors.PermissionDenied)
			}

			productPlanID := q.Get("product_plan_id")
			priceScheduleID := q.Get("price_schedule_id")
			if priceScheduleID == "" {
				return view.HTMXError(deps.Labels.Errors.InvalidFormData)
			}

			var productPlanOptions, setOptions []pyeza.SelectOption
			if deps.GetFormPageData != nil {
				pageData, err := deps.GetFormPageData(ctx)
				if err != nil {
					log.Printf("Failed to load rating description set product plan form page data: %v", err)
					return view.HTMXError(deps.Labels.Errors.LoadFailed)
				}
				productPlanOptions = linkform.BuildProductPlanOptionsFromPageData(pageData, productPlanID)
				setOptions = linkform.BuildPublishedSetOptionsFromPageData(pageData, "")
			} else {
				productPlanOptions = linkform.BuildProductPlanOptions(ctx, deps.ListProductPlans, productPlanID)
				setOptions = linkform.BuildPublishedSetOptions(ctx, deps.ListRatingDescriptionSets, "")
			}

			return view.OK("rating-description-set-product-plan-drawer-form", &linkform.Data{
				FormAction:             deps.Routes.RelinkURL,
				ProductPlanID:          productPlanID,
				ProductPlanLocked:      productPlanID != "",
				PriceScheduleID:        priceScheduleID,
				ExpectedCurrentLinkID:  expectedCurrent,
				ProductPlanOptions:     productPlanOptions,
				SetOptions:             setOptions,
				Labels:                 deps.Labels,
				CommonLabels:           nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		r := viewCtx.Request

		productPlanID := r.FormValue("product_plan_id")
		priceScheduleID := r.FormValue("price_schedule_id")
		setID := r.FormValue("rating_description_set_id")
		expectedCurrent := r.FormValue("expected_current_link_id")
		reason := r.FormValue("reason")

		if productPlanID == "" || priceScheduleID == "" || setID == "" {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		perms := view.GetUserPermissions(ctx)
		verb := "create"
		if expectedCurrent != "" {
			verb = "update"
		}
		if !perms.Can("rating_description_set_product_plan", verb) {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		req := &linkpb.RelinkRatingDescriptionSetProductPlanRequest{
			ProductPlanId:          productPlanID,
			PriceScheduleId:        priceScheduleID,
			RatingDescriptionSetId: setID,
			Reason:                 reason,
		}
		if expectedCurrent != "" {
			req.ExpectedCurrentLinkId = &expectedCurrent
		}

		_, err := deps.RelinkRatingDescriptionSetProductPlan(ctx, req)
		if err != nil {
			log.Printf("Failed to relink product plan %s to set %s: %v", productPlanID, setID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("rating-description-set-link-table")
	})
}

// NewUnlinkAction deactivates an active link (POST only). The offering
// becomes NO_LINK for the resolver.
func NewUnlinkAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set_product_plan", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		linkID := viewCtx.Request.URL.Query().Get("link_id")
		if linkID == "" {
			_ = viewCtx.Request.ParseForm()
			linkID = viewCtx.Request.FormValue("link_id")
		}
		if linkID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		reason := viewCtx.Request.FormValue("reason")

		_, err := deps.UnlinkRatingDescriptionSetProductPlan(ctx, &linkpb.UnlinkRatingDescriptionSetProductPlanRequest{
			LinkId: linkID,
			Reason: reason,
		})
		if err != nil {
			log.Printf("Failed to unlink %s: %v", linkID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("rating-description-set-link-table")
	})
}
