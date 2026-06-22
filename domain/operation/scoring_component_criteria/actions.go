package scoring_component_criteria

import (
	"context"
	"log"
	"net/http"

	sccform "github.com/erniealice/fayna-golang/domain/operation/scoring_component_criteria/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	sccpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component_criteria"
)

// NewAddAction creates the scoring component criteria add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component_criteria", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("scoring-component-criteria-drawer-form", &sccform.Data{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create scoring component criteria
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request

		_, err := deps.CreateScoringComponentCriteria(ctx, &sccpb.CreateScoringComponentCriteriaRequest{
			Data: &sccpb.ScoringComponentCriteria{
				ScoringSchemeId:    r.FormValue("scoring_scheme_id"),
				ScoringComponentId: r.FormValue("scoring_component_id"),
				OutcomeCriteriaId:  r.FormValue("outcome_criteria_id"),
				Active:             true,
			},
		})
		if err != nil {
			log.Printf("Failed to create scoring component criteria: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("scoring-component-criteria-table")
	})
}

// NewEditAction creates the scoring component criteria edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component_criteria", "update") {
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

			readResp, err := deps.ReadScoringComponentCriteria(ctx, &sccpb.ReadScoringComponentCriteriaRequest{
				Data: &sccpb.ScoringComponentCriteria{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read scoring component criteria %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			return view.OK("scoring-component-criteria-drawer-form", &sccform.Data{
				FormAction:         route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:             true,
				ID:                 id,
				ScoringSchemeID:    record.GetScoringSchemeId(),
				ScoringComponentID: record.GetScoringComponentId(),
				OutcomeCriteriaID:  record.GetOutcomeCriteriaId(),
				Labels:             deps.Labels,
				CommonLabels:       nil, // injected by ViewAdapter
			})
		}

		// POST — update scoring component criteria
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

		_, err := deps.UpdateScoringComponentCriteria(ctx, &sccpb.UpdateScoringComponentCriteriaRequest{
			Data: &sccpb.ScoringComponentCriteria{
				Id:                 id,
				ScoringSchemeId:    r.FormValue("scoring_scheme_id"),
				ScoringComponentId: r.FormValue("scoring_component_id"),
				OutcomeCriteriaId:  r.FormValue("outcome_criteria_id"),
			},
		})
		if err != nil {
			log.Printf("Failed to update scoring component criteria %s: %v", id, err)
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

// NewDeleteAction creates the scoring component criteria delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component_criteria", "delete") {
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

		_, err := deps.DeleteScoringComponentCriteria(ctx, &sccpb.DeleteScoringComponentCriteriaRequest{
			Data: &sccpb.ScoringComponentCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete scoring component criteria %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("scoring-component-criteria-table")
	})
}

// NewBulkDeleteAction creates the scoring component criteria bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component_criteria", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteScoringComponentCriteria(ctx, &sccpb.DeleteScoringComponentCriteriaRequest{
				Data: &sccpb.ScoringComponentCriteria{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete scoring component criteria %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("scoring-component-criteria-table")
	})
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil.
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
