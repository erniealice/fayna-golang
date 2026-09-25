package rating_description_set_entry

import (
	"context"
	"log"
	"net/http"

	entryform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry/form"

	"github.com/erniealice/pyeza-golang/route"
	pyeza "github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	entrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
)

// NewAddAction creates the entry add action (GET = drawer, POST = create).
// Gated rating_description_set:update — entries are embedded in the parent
// set (interfaces.md §5: no separate entry permission codes); the espyna
// create use case (create_rating_description_set_entry.go) authorizes an
// entry write as an edit of the parent set's content, i.e.
// rating_description_set:update, NOT :create (codex-review-impl2.out.md
// finding #11 — the UI previously gated on :create, so a create-only editor
// could open the drawer but have every submit rejected, and an update-only
// editor was blocked from a page they were otherwise fully authorized to
// perform). Aligned here to match the use case exactly. The adapter enforces
// DRAFT-only + band-not-zero server-side (fails closed) regardless.
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		setID := viewCtx.Request.URL.Query().Get("rating_description_set_id")

		if viewCtx.Request.Method == http.MethodGet {
			if setID == "" {
				return view.HTMXError(deps.Labels.Errors.IDRequired)
			}
			criterionID := viewCtx.Request.URL.Query().Get("outcome_criteria_id")

			criteriaOptions, bandOptions, _, err := loadEntryDrawerFormData(ctx, deps, setID, "", criterionID, "")
			if err != nil {
				log.Printf("Failed to load rating description set entry form page data: %v", err)
				return view.HTMXError(deps.Labels.Errors.LoadFailed)
			}

			return view.OK("rating-description-set-entry-drawer-form", &entryform.Data{
				FormAction:             deps.Routes.AddURL,
				RatingDescriptionSetID: setID,
				OutcomeCriteriaID:      criterionID,
				CriteriaOptions:        criteriaOptions,
				BandOptions:            bandOptions,
				Labels:                 deps.Labels,
				CommonLabels:           nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		r := viewCtx.Request
		if setID == "" {
			setID = r.FormValue("rating_description_set_id")
		}
		criterionID := r.FormValue("outcome_criteria_id")
		bandID := r.FormValue("score_scale_band_id")
		description := r.FormValue("description")
		if setID == "" || criterionID == "" || bandID == "" || description == "" {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		_, err := deps.CreateRatingDescriptionSetEntry(ctx, &entrypb.CreateRatingDescriptionSetEntryRequest{
			Data: &entrypb.RatingDescriptionSetEntry{
				RatingDescriptionSetId: setID,
				OutcomeCriteriaId:      criterionID,
				ScoreScaleBandId:       bandID,
				Description:            description,
				Active:                 true,
			},
		})
		if err != nil {
			log.Printf("Failed to create rating description set entry for set %s: %v", setID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("tabContent")
	})
}

// NewEditAction — GET pre-filled drawer, POST updates an entry.
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set", "update") {
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

			// Loaded through a parent-update-authorized read
			// (rating_description_set:update) — NOT the entry's own :read
			// gate — so an update-only editor (no separate :read grant) can
			// still open this drawer (codex-review-impl4.out.md "Update-only
			// drawer" finding).
			criteriaOptions, bandOptions, entry, err := loadEntryDrawerFormData(ctx, deps, "", id, "", "")
			if err != nil {
				log.Printf("Failed to load rating description set entry form page data: %v", err)
				return view.HTMXError(deps.Labels.Errors.LoadFailed)
			}
			if entry == nil {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}

			return view.OK("rating-description-set-entry-drawer-form", &entryform.Data{
				FormAction:             route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:                 true,
				ID:                     id,
				RatingDescriptionSetID: entry.RatingDescriptionSetID,
				OutcomeCriteriaID:      entry.OutcomeCriteriaID,
				ScoreScaleBandID:       entry.ScoreScaleBandID,
				Description:            entry.Description,
				CriteriaOptions:        criteriaOptions,
				BandOptions:            bandOptions,
				Labels:                 deps.Labels,
				CommonLabels:           nil,
			})
		}

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
		description := r.FormValue("description")
		if description == "" {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		_, err := deps.UpdateRatingDescriptionSetEntry(ctx, &entrypb.UpdateRatingDescriptionSetEntryRequest{
			Data: &entrypb.RatingDescriptionSetEntry{
				Id:                 id,
				OutcomeCriteriaId:  r.FormValue("outcome_criteria_id"),
				ScoreScaleBandId:   r.FormValue("score_scale_band_id"),
				Description:        description,
			},
		})
		if err != nil {
			log.Printf("Failed to update rating description set entry %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("tabContent")
	})
}

// loadEntryDrawerFormData loads the entry drawer's criteria/band pickers
// and, for Edit, the entry itself — through ONE call authorized under
// rating_description_set:update (codex-review-impl4.out.md "Update-only
// drawer": an update-only editor without a separate :read grant must still
// be able to open both Add and Edit; a read-only-only viewer must not reach
// either). ratingDescriptionSetID is the Add drawer's known parent set;
// entryID is the Edit drawer's row id. Returns (criteriaOptions,
// bandOptions, entry, err) — entry is nil for Add, or for an Edit whose id
// does not resolve to a row (the caller treats a nil entry as Not Found,
// not an error).
//
// Prefers deps.GetFormPageData (the Drawer page-data use case); falls back
// to the narrower, separately :read-gated ReadRatingDescriptionSet /
// ReadRatingDescriptionSetEntry / ListOutcomeCriterias/ListScoreScaleBands
// closures ONLY when GetFormPageData is not wired, so the drawer stays
// functional (if permission-gated more narrowly) during a partial rollout —
// not reachable once the postgres provider registers this feature (see
// espyna usecases.go's all-or-nothing provider gate).
func loadEntryDrawerFormData(ctx context.Context, deps *ModuleDeps, ratingDescriptionSetID, entryID, selectedCriterion, selectedBand string) ([]pyeza.SelectOption, []pyeza.SelectOption, *entryform.EntrySnapshot, error) {
	if deps.GetFormPageData != nil {
		pageData, err := deps.GetFormPageData(ctx, ratingDescriptionSetID, entryID)
		if err != nil {
			return nil, nil, nil, err
		}
		selCriterion, selBand := selectedCriterion, selectedBand
		if pageData != nil && pageData.Entry != nil {
			selCriterion, selBand = pageData.Entry.OutcomeCriteriaID, pageData.Entry.ScoreScaleBandID
		}
		criteriaOptions := entryform.BuildCriteriaOptionsFromPageData(pageData, selCriterion)
		scaleID := ""
		if pageData != nil {
			scaleID = pageData.ScoreScaleID
		}
		bandOptions := entryform.BuildBandOptionsFromPageData(pageData, scaleID, selBand)
		var entry *entryform.EntrySnapshot
		if pageData != nil {
			entry = pageData.Entry
		}
		return criteriaOptions, bandOptions, entry, nil
	}

	// Legacy fallback (GetFormPageData not wired).
	var entry *entryform.EntrySnapshot
	if entryID != "" {
		if deps.ReadRatingDescriptionSetEntry == nil {
			return nil, nil, nil, nil
		}
		readResp, err := deps.ReadRatingDescriptionSetEntry(ctx, &entrypb.ReadRatingDescriptionSetEntryRequest{
			Data: &entrypb.RatingDescriptionSetEntry{Id: entryID},
		})
		if err != nil {
			return nil, nil, nil, err
		}
		rows := readResp.GetData()
		if len(rows) == 0 {
			return nil, nil, nil, nil
		}
		rec := rows[0]
		entry = &entryform.EntrySnapshot{
			ID:                     rec.GetId(),
			RatingDescriptionSetID: rec.GetRatingDescriptionSetId(),
			OutcomeCriteriaID:      rec.GetOutcomeCriteriaId(),
			ScoreScaleBandID:       rec.GetScoreScaleBandId(),
			Description:            rec.GetDescription(),
		}
		selectedCriterion, selectedBand = entry.OutcomeCriteriaID, entry.ScoreScaleBandID
		if ratingDescriptionSetID == "" {
			ratingDescriptionSetID = entry.RatingDescriptionSetID
		}
	}
	scaleID := scoreScaleIDForSet(ctx, deps, ratingDescriptionSetID)
	criteriaOptions := entryform.BuildCriteriaOptions(ctx, deps.ListOutcomeCriterias, selectedCriterion)
	bandOptions := entryform.BuildBandOptions(ctx, deps.ListScoreScaleBands, scaleID, selectedBand)
	return criteriaOptions, bandOptions, entry, nil
}

// scoreScaleIDForSet reads the parent set to find its score_scale_id, used
// to scope the Level (band) picker. Returns "" on any failure — the picker
// then falls back to the raw-id text input. LEGACY FALLBACK ONLY (see
// loadEntryDrawerFormData): the primary path resolves this server-side,
// under rating_description_set:update, inside the Drawer page-data call —
// this :read-gated lookup is only reached when GetFormPageData is nil.
func scoreScaleIDForSet(ctx context.Context, deps *ModuleDeps, setID string) string {
	if deps.ReadRatingDescriptionSet == nil || setID == "" {
		return ""
	}
	resp, err := deps.ReadRatingDescriptionSet(ctx, &setpb.ReadRatingDescriptionSetRequest{
		Data: &setpb.RatingDescriptionSet{Id: setID},
	})
	if err != nil || resp == nil || len(resp.GetData()) == 0 {
		return ""
	}
	return resp.GetData()[0].GetScoreScaleId()
}
