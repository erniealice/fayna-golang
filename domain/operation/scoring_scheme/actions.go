package scoring_scheme

import (
	"context"
	"log"
	"net/http"

	scoringschemeform "github.com/erniealice/fayna-golang/domain/operation/scoring_scheme/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	schemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"
)

// NewAddAction creates the scoring scheme add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_scheme", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("scoring-scheme-drawer-form", &scoringschemeform.Data{
				FormAction:             deps.Routes.AddURL,
				CompositeMethodOptions: scoringschemeform.DefaultCompositeMethodOptions(),
				RoundingModeOptions:    scoringschemeform.DefaultRoundingModeOptions(),
				Labels:                 deps.Labels,
				CommonLabels:           nil, // injected by ViewAdapter
			})
		}

		// POST — create scoring scheme
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		compositeMethod := enums.ScoringMethod(enums.ScoringMethod_value[r.FormValue("composite_method")])
		weightsMustSumToOne := r.FormValue("weights_must_sum_to_one") == "true" || r.FormValue("weights_must_sum_to_one") == "on"

		scheme := &schemepb.ScoringScheme{
			Name:                r.FormValue("name"),
			CompositeMethod:     compositeMethod,
			WeightsMustSumToOne: weightsMustSumToOne,
			Active:              true,
		}

		// Optional: score_scale_id
		if v := r.FormValue("score_scale_id"); v != "" {
			scheme.ScoreScaleId = &v
		}
		// Optional: rounding_mode
		if v := r.FormValue("rounding_mode"); v != "" {
			rm := enums.RoundingMode(enums.RoundingMode_value[v])
			scheme.RoundingMode = &rm
		}

		_, err := deps.CreateScoringScheme(ctx, &schemepb.CreateScoringSchemeRequest{
			Data: scheme,
		})
		if err != nil {
			log.Printf("Failed to create scoring scheme: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("scoring-scheme-table")
	})
}

// NewEditAction creates the scoring scheme edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_scheme", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}

		if viewCtx.Request.Method == http.MethodGet {
			if id == "" {
				return view.HTMXError(deps.Labels.Errors.IDRequired)
			}

			readResp, err := deps.ReadScoringScheme(ctx, &schemepb.ReadScoringSchemeRequest{
				Data: &schemepb.ScoringScheme{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read scoring scheme %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			return view.OK("scoring-scheme-drawer-form", &scoringschemeform.Data{
				FormAction:             route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:                 true,
				ID:                     id,
				Name:                   record.GetName(),
				CompositeMethod:        record.GetCompositeMethod().String(),
				RoundingMode:           record.GetRoundingMode().String(),
				WeightsMustSumToOne:    record.GetWeightsMustSumToOne(),
				CompositeMethodOptions: scoringschemeform.DefaultCompositeMethodOptions(),
				RoundingModeOptions:    scoringschemeform.DefaultRoundingModeOptions(),
				Labels:                 deps.Labels,
				CommonLabels:           nil, // injected by ViewAdapter
			})
		}

		// POST — update scoring scheme
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		if id == "" {
			id = r.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		compositeMethod := enums.ScoringMethod(enums.ScoringMethod_value[r.FormValue("composite_method")])
		weightsMustSumToOne := r.FormValue("weights_must_sum_to_one") == "true" || r.FormValue("weights_must_sum_to_one") == "on"

		scheme := &schemepb.ScoringScheme{
			Id:                  id,
			Name:                r.FormValue("name"),
			CompositeMethod:     compositeMethod,
			WeightsMustSumToOne: weightsMustSumToOne,
		}

		// Optional: score_scale_id
		if v := r.FormValue("score_scale_id"); v != "" {
			scheme.ScoreScaleId = &v
		}
		// Optional: rounding_mode
		if v := r.FormValue("rounding_mode"); v != "" {
			rm := enums.RoundingMode(enums.RoundingMode_value[v])
			scheme.RoundingMode = &rm
		}

		_, err := deps.UpdateScoringScheme(ctx, &schemepb.UpdateScoringSchemeRequest{
			Data: scheme,
		})
		if err != nil {
			log.Printf("Failed to update scoring scheme %s: %v", id, err)
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

// NewDeleteAction creates the scoring scheme delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_scheme", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteScoringScheme(ctx, &schemepb.DeleteScoringSchemeRequest{
			Data: &schemepb.ScoringScheme{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete scoring scheme %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("scoring-scheme-table")
	})
}

// NewBulkDeleteAction creates the scoring scheme bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_scheme", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteScoringScheme(ctx, &schemepb.DeleteScoringSchemeRequest{
				Data: &schemepb.ScoringScheme{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete scoring scheme %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("scoring-scheme-table")
	})
}
