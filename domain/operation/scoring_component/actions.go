package scoring_component

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	scoringform "github.com/erniealice/fayna-golang/domain/operation/scoring_component/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	scoringpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component"
)

// NewAddAction creates the scoring component add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("scoring-component-drawer-form", &scoringform.Data{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create scoring component
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		weight, _ := strconv.ParseFloat(r.FormValue("weight"), 64)
		seqOrder, _ := strconv.Atoi(r.FormValue("sequence_order"))

		_, err := deps.CreateScoringComponent(ctx, &scoringpb.CreateScoringComponentRequest{
			Data: &scoringpb.ScoringComponent{
				ScoringSchemeId:   r.FormValue("scoring_scheme_id"),
				Code:              r.FormValue("code"),
				Label:             r.FormValue("label"),
				Weight:            weight,
				SequenceOrder:     int32(seqOrder),
				ParentComponentId: strPtrIfNotEmpty(r.FormValue("parent_component_id")),
				Active:            true,
			},
		})
		if err != nil {
			log.Printf("Failed to create scoring component: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("scoring-components-table")
	})
}

// NewEditAction creates the scoring component edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component", "update") {
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

			readResp, err := deps.ReadScoringComponent(ctx, &scoringpb.ReadScoringComponentRequest{
				Data: &scoringpb.ScoringComponent{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read scoring component %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			return view.OK("scoring-component-drawer-form", &scoringform.Data{
				FormAction:        route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:            true,
				ID:                id,
				ScoringSchemeId:   record.GetScoringSchemeId(),
				Code:              record.GetCode(),
				Label:             record.GetLabel(),
				Weight:            record.GetWeight(),
				SequenceOrder:     int(record.GetSequenceOrder()),
				ParentComponentId: record.GetParentComponentId(),
				Active:            record.GetActive(),
				Labels:            deps.Labels,
				CommonLabels:      nil, // injected by ViewAdapter
			})
		}

		// POST — update scoring component
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

		weight, _ := strconv.ParseFloat(r.FormValue("weight"), 64)
		seqOrder, _ := strconv.Atoi(r.FormValue("sequence_order"))

		_, err := deps.UpdateScoringComponent(ctx, &scoringpb.UpdateScoringComponentRequest{
			Data: &scoringpb.ScoringComponent{
				Id:                id,
				ScoringSchemeId:   r.FormValue("scoring_scheme_id"),
				Code:              r.FormValue("code"),
				Label:             r.FormValue("label"),
				Weight:            weight,
				SequenceOrder:     int32(seqOrder),
				ParentComponentId: strPtrIfNotEmpty(r.FormValue("parent_component_id")),
				Active:            r.FormValue("active") == "true" || r.FormValue("active") == "on",
			},
		})
		if err != nil {
			log.Printf("Failed to update scoring component %s: %v", id, err)
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

// NewDeleteAction creates the scoring component delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component", "delete") {
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

		_, err := deps.DeleteScoringComponent(ctx, &scoringpb.DeleteScoringComponentRequest{
			Data: &scoringpb.ScoringComponent{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete scoring component %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("scoring-components-table")
	})
}

// NewBulkDeleteAction creates the scoring component bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteScoringComponent(ctx, &scoringpb.DeleteScoringComponentRequest{
				Data: &scoringpb.ScoringComponent{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete scoring component %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("scoring-components-table")
	})
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil.
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// fmtInt32 formats an int32 for display.
func fmtInt32(n int32) string {
	return fmt.Sprintf("%d", n)
}
