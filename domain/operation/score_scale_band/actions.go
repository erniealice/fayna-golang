package score_scale_band

import (
	"context"
	"log"
	"net/http"
	"strconv"

	ssbform "github.com/erniealice/fayna-golang/domain/operation/score_scale_band/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	ssbpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// NewAddAction creates the score scale band add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale_band", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("score-scale-band-drawer-form", &ssbform.Data{
				FormAction:           deps.Routes.AddURL,
				DeterminationOptions: ssbform.DefaultDeterminationOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil,
			})
		}

		// POST — create score scale band
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		seqOrder, _ := strconv.ParseInt(r.FormValue("sequence_order"), 10, 32)

		band := &ssbpb.ScoreScaleBand{
			ScoreScaleId:  r.FormValue("score_scale_id"),
			SequenceOrder: int32(seqOrder),
			OutputLabel:   r.FormValue("output_label"),
			Active:        true,
		}

		if v := r.FormValue("input_min"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				band.InputMin = &f
			}
		}
		if v := r.FormValue("input_max"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				band.InputMax = &f
			}
		}
		if v := r.FormValue("input_match"); v != "" {
			band.InputMatch = &v
		}
		if v := r.FormValue("output_value"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				band.OutputValue = &f
			}
		}
		if v := r.FormValue("band_role"); v != "" {
			band.BandRole = &v
		}
		if v := r.FormValue("determination"); v != "" {
			if ev, ok := enums.Determination_value[v]; ok {
				d := enums.Determination(ev)
				band.Determination = &d
			}
		}

		_, err := deps.CreateScoreScaleBand(ctx, &ssbpb.CreateScoreScaleBandRequest{
			Data: band,
		})
		if err != nil {
			log.Printf("Failed to create score scale band: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("score-scale-band-table")
	})
}

// NewEditAction creates the score scale band edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale_band", "update") {
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

			readResp, err := deps.ReadScoreScaleBand(ctx, &ssbpb.ReadScoreScaleBandRequest{
				Data: &ssbpb.ScoreScaleBand{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read score scale band %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			rec := readData[0]

			return view.OK("score-scale-band-drawer-form", &ssbform.Data{
				FormAction:           route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:               true,
				ID:                   id,
				ScoreScaleId:         rec.GetScoreScaleId(),
				SequenceOrder:        rec.GetSequenceOrder(),
				InputMin:             rec.InputMin,
				InputMax:             rec.InputMax,
				InputMatch:           rec.GetInputMatch(),
				OutputValue:          rec.OutputValue,
				OutputLabel:          rec.GetOutputLabel(),
				BandRole:             rec.GetBandRole(),
				Determination:        rec.GetDetermination().String(),
				DeterminationOptions: ssbform.DefaultDeterminationOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil,
			})
		}

		// POST — update score scale band
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

		seqOrder, _ := strconv.ParseInt(r.FormValue("sequence_order"), 10, 32)

		band := &ssbpb.ScoreScaleBand{
			Id:            id,
			ScoreScaleId:  r.FormValue("score_scale_id"),
			SequenceOrder: int32(seqOrder),
			OutputLabel:   r.FormValue("output_label"),
		}

		if v := r.FormValue("input_min"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				band.InputMin = &f
			}
		}
		if v := r.FormValue("input_max"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				band.InputMax = &f
			}
		}
		if v := r.FormValue("input_match"); v != "" {
			band.InputMatch = &v
		}
		if v := r.FormValue("output_value"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				band.OutputValue = &f
			}
		}
		if v := r.FormValue("band_role"); v != "" {
			band.BandRole = &v
		}
		if v := r.FormValue("determination"); v != "" {
			if ev, ok := enums.Determination_value[v]; ok {
				d := enums.Determination(ev)
				band.Determination = &d
			}
		}

		_, err := deps.UpdateScoreScaleBand(ctx, &ssbpb.UpdateScoreScaleBandRequest{
			Data: band,
		})
		if err != nil {
			log.Printf("Failed to update score scale band %s: %v", id, err)
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

// NewDeleteAction creates the score scale band delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale_band", "delete") {
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

		_, err := deps.DeleteScoreScaleBand(ctx, &ssbpb.DeleteScoreScaleBandRequest{
			Data: &ssbpb.ScoreScaleBand{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete score scale band %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("score-scale-band-table")
	})
}

// NewBulkDeleteAction creates the score scale band bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale_band", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteScoreScaleBand(ctx, &ssbpb.DeleteScoreScaleBandRequest{
				Data: &ssbpb.ScoreScaleBand{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete score scale band %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("score-scale-band-table")
	})
}
