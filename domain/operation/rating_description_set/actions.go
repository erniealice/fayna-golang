package rating_description_set

import (
	"context"
	"log"
	"net/http"

	setform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set/form"

	"github.com/erniealice/pyeza-golang/route"
	pyeza "github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
)

// NewAddAction creates the rating description set add action (GET = form,
// POST = create). CreateRatingDescriptionSet always creates a DRAFT at
// version 1 (espyna-golang.md lifecycle rules) — there is no version_status
// field on the form.
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			var scaleOptions []pyeza.SelectOption
			if deps.GetFormPageData != nil {
				pageData, err := deps.GetFormPageData(ctx)
				if err != nil {
					log.Printf("Failed to load rating description set form page data: %v", err)
					return view.HTMXError(deps.Labels.Errors.LoadFailed)
				}
				scaleOptions = setform.BuildScaleOptionsFromPageData(pageData, "")
			} else {
				// Fallback for as-yet-unwired environments (finding #6's fix
				// is the preferred path above; this keeps the drawer
				// functional, if score_scale:list-gated, until wiring lands).
				scaleOptions = setform.BuildScaleOptions(ctx, deps.ListScoreScales, "")
			}
			return view.OK("rating-description-set-drawer-form", &setform.Data{
				FormAction:   deps.Routes.AddURL,
				ScaleOptions: scaleOptions,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		r := viewCtx.Request

		name := r.FormValue("name")
		if name == "" {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		scaleID := r.FormValue("score_scale_id")
		if scaleID == "" {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		data := &setpb.RatingDescriptionSet{
			Name:         name,
			ScoreScaleId: scaleID,
			Active:       true,
		}
		if v := r.FormValue("code"); v != "" {
			data.Code = &v
		}
		if v := r.FormValue("source_ref"); v != "" {
			data.SourceRef = &v
		}

		resp, err := deps.CreateRatingDescriptionSet(ctx, &setpb.CreateRatingDescriptionSetRequest{Data: data})
		if err != nil {
			log.Printf("Failed to create rating description set: %v", err)
			return view.HTMXError(err.Error())
		}

		newID := ""
		if rows := resp.GetData(); len(rows) > 0 {
			newID = rows[0].GetId()
		}
		if newID == "" {
			return view.HTMXSuccess("rating-description-set-table")
		}

		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true}`,
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", newID),
			},
		}
	})
}

// NewPublishAction transitions a DRAFT set to PUBLISHED (POST only).
func NewPublishAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set", "publish") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.PublishRatingDescriptionSet(ctx, &setpb.PublishRatingDescriptionSetRequest{Id: id})
		if err != nil {
			log.Printf("Failed to publish rating description set %s: %v", id, err)
			return view.HTMXError(err.Error())
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

// NewDeprecateAction transitions a PUBLISHED set to DEPRECATED (POST only).
// Existing links keep resolving (Q19) — this action only flips status.
func NewDeprecateAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set", "deprecate") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeprecateRatingDescriptionSet(ctx, &setpb.DeprecateRatingDescriptionSetRequest{Id: id})
		if err != nil {
			log.Printf("Failed to deprecate rating description set %s: %v", id, err)
			return view.HTMXError(err.Error())
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
