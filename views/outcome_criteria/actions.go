package outcome_criteria

import (
	"context"
	"log"
	"net/http"
	"strconv"

	fayna "github.com/erniealice/fayna-golang"
	outcomecriteriaform "github.com/erniealice/fayna-golang/views/outcome_criteria/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
)

// newAddAction creates the outcome criteria add action (GET = form, POST = create).
func newAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("outcome_criteria", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("outcome-criteria-drawer-form", &outcomecriteriaform.Data{
				FormAction:   deps.Routes.AddURL,
				TypeOptions:  outcomecriteriaform.DefaultTypeOptions(deps.Labels),
				ScopeOptions: outcomecriteriaform.DefaultScopeOptions(deps.Labels),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create outcome criteria
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		weight, _ := strconv.ParseFloat(r.FormValue("weight"), 64)

		_, err := deps.CreateOutcomeCriteria(ctx, &criteriapb.CreateOutcomeCriteriaRequest{
			Data: &criteriapb.OutcomeCriteria{
				Name:        r.FormValue("name"),
				Description: strPtrIfNotEmpty(r.FormValue("description")),
				Weight:      weight,
				Required:    r.FormValue("required") == "true" || r.FormValue("required") == "on",
				Active:      true,
			},
		})
		if err != nil {
			log.Printf("Failed to create outcome criteria: %v", err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("criteria-table")
	})
}

// newEditAction creates the outcome criteria edit action (GET = pre-filled form, POST = update).
func newEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("outcome_criteria", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}

		if viewCtx.Request.Method == http.MethodGet {
			if id == "" {
				return fayna.HTMXError(deps.Labels.Errors.IDRequired)
			}

			readResp, err := deps.ReadOutcomeCriteria(ctx, &criteriapb.ReadOutcomeCriteriaRequest{
				Data: &criteriapb.OutcomeCriteria{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read outcome criteria %s: %v", id, err)
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			return view.OK("outcome-criteria-drawer-form", &outcomecriteriaform.Data{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				Name:         record.GetName(),
				Type:         record.GetCriteriaType().String(),
				Scope:        record.GetScope().String(),
				Description:  record.GetDescription(),
				Required:     record.GetRequired(),
				Weight:       record.GetWeight(),
				TypeOptions:  outcomecriteriaform.DefaultTypeOptions(deps.Labels),
				ScopeOptions: outcomecriteriaform.DefaultScopeOptions(deps.Labels),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — update outcome criteria
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		if id == "" {
			id = r.FormValue("id")
		}
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		weight, _ := strconv.ParseFloat(r.FormValue("weight"), 64)

		_, err := deps.UpdateOutcomeCriteria(ctx, &criteriapb.UpdateOutcomeCriteriaRequest{
			Data: &criteriapb.OutcomeCriteria{
				Id:          id,
				Name:        r.FormValue("name"),
				Description: strPtrIfNotEmpty(r.FormValue("description")),
				Weight:      weight,
				Required:    r.FormValue("required") == "true" || r.FormValue("required") == "on",
			},
		})
		if err != nil {
			log.Printf("Failed to update outcome criteria %s: %v", id, err)
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

// newDeleteAction creates the outcome criteria delete action (POST only).
func newDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("outcome_criteria", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteOutcomeCriteria(ctx, &criteriapb.DeleteOutcomeCriteriaRequest{
			Data: &criteriapb.OutcomeCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete outcome criteria %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("criteria-table")
	})
}

// newBulkDeleteAction creates the outcome criteria bulk delete action (POST only).
func newBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("outcome_criteria", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return fayna.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteOutcomeCriteria(ctx, &criteriapb.DeleteOutcomeCriteriaRequest{
				Data: &criteriapb.OutcomeCriteria{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete outcome criteria %s: %v", id, err)
			}
		}

		return fayna.HTMXSuccess("criteria-table")
	})
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil.
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
