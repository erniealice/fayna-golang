package score_scale

import (
	"context"
	"log"
	"net/http"
	"strconv"

	scaleform "github.com/erniealice/fayna-golang/domain/operation/score_scale/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
)

// NewAddAction creates the score scale add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("score-scale-drawer-form", &scaleform.Data{
				FormAction:           deps.Routes.AddURL,
				ScaleKindOptions:     scaleform.DefaultScaleKindOptions(),
				VersionStatusOptions: scaleform.DefaultVersionStatusOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil, // injected by ViewAdapter
			})
		}

		// POST — create score scale
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request

		scaleKindVal := enums.ScaleKind_value[r.FormValue("scale_kind")]
		versionStatusVal := enums.VersionStatus_value[r.FormValue("version_status")]

		data := &scalepb.ScoreScale{
			Name:          r.FormValue("name"),
			ScaleKind:     enums.ScaleKind(scaleKindVal),
			VersionStatus: enums.VersionStatus(versionStatusVal),
			InputUnit:     r.FormValue("input_unit"),
			OutputUnit:    r.FormValue("output_unit"),
			Active:        true,
		}

		if v := r.FormValue("scale_group_id"); v != "" {
			data.ScaleGroupId = v
		}

		if v := r.FormValue("input_min"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				data.InputMin = &f
			}
		}
		if v := r.FormValue("input_max"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				data.InputMax = &f
			}
		}

		_, err := deps.CreateScoreScale(ctx, &scalepb.CreateScoreScaleRequest{Data: data})
		if err != nil {
			log.Printf("Failed to create score scale: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("score-scale-table")
	})
}

// NewEditAction creates the score scale edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale", "update") {
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

			readResp, err := deps.ReadScoreScale(ctx, &scalepb.ReadScoreScaleRequest{
				Data: &scalepb.ScoreScale{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read score scale %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			var inputMinStr, inputMaxStr string
			if record.InputMin != nil {
				inputMinStr = strconv.FormatFloat(*record.InputMin, 'f', -1, 64)
			}
			if record.InputMax != nil {
				inputMaxStr = strconv.FormatFloat(*record.InputMax, 'f', -1, 64)
			}

			return view.OK("score-scale-drawer-form", &scaleform.Data{
				FormAction:           route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:               true,
				ID:                   id,
				Name:                 record.GetName(),
				ScaleKind:            record.GetScaleKind().String(),
				VersionStatus:        record.GetVersionStatus().String(),
				InputUnit:            record.GetInputUnit(),
				InputMin:             inputMinStr,
				InputMax:             inputMaxStr,
				OutputUnit:           record.GetOutputUnit(),
				ScaleGroupId:         record.GetScaleGroupId(),
				ScaleKindOptions:     scaleform.DefaultScaleKindOptions(),
				VersionStatusOptions: scaleform.DefaultVersionStatusOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil, // injected by ViewAdapter
			})
		}

		// POST — update score scale
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

		scaleKindVal := enums.ScaleKind_value[r.FormValue("scale_kind")]
		versionStatusVal := enums.VersionStatus_value[r.FormValue("version_status")]

		data := &scalepb.ScoreScale{
			Id:            id,
			Name:          r.FormValue("name"),
			ScaleKind:     enums.ScaleKind(scaleKindVal),
			VersionStatus: enums.VersionStatus(versionStatusVal),
			InputUnit:     r.FormValue("input_unit"),
			OutputUnit:    r.FormValue("output_unit"),
		}

		if v := r.FormValue("scale_group_id"); v != "" {
			data.ScaleGroupId = v
		}
		if v := r.FormValue("input_min"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				data.InputMin = &f
			}
		}
		if v := r.FormValue("input_max"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				data.InputMax = &f
			}
		}

		_, err := deps.UpdateScoreScale(ctx, &scalepb.UpdateScoreScaleRequest{Data: data})
		if err != nil {
			log.Printf("Failed to update score scale %s: %v", id, err)
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

// NewDeleteAction creates the score scale delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale", "delete") {
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

		_, err := deps.DeleteScoreScale(ctx, &scalepb.DeleteScoreScaleRequest{
			Data: &scalepb.ScoreScale{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete score scale %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("score-scale-table")
	})
}

// NewBulkDeleteAction creates the score scale bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteScoreScale(ctx, &scalepb.DeleteScoreScaleRequest{
				Data: &scalepb.ScoreScale{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete score scale %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("score-scale-table")
	})
}
