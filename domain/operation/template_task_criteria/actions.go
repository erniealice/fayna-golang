package template_task_criteria

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	ttcform "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	ttcrdpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria_rating_description"
)

// standardsTableID is the table id on the job_template detail Standards tab
// (job_template/detail/standards.go). refreshTableFor picks this over the
// module's own list-page table id when the drawer was opened in
// ContextTemplate, so the HX-Trigger refreshes the right DOM table.
const standardsTableID = "jt-standards-table"

// refreshTableFor returns the table id to refresh on success, based on which
// context the drawer was opened in (form.ContextTemplate vs the module's own
// list page).
func refreshTableFor(templateID string) string {
	if templateID != "" {
		return standardsTableID
	}
	return "template-task-criteria-table"
}

// NewAddAction creates the template task criteria add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			templateID := viewCtx.Request.URL.Query().Get("job_template_id")
			data := &ttcform.Data{
				FormAction:      deps.Routes.AddURL,
				Labels:          deps.Labels,
				Active:          true,
				CriteriaOptions: ttcform.BuildOutcomeCriteriaOptions(ctx, deps.ListOutcomeCriterias, ""),
				CommonLabels:    nil, // injected by ViewAdapter
			}
			populateRatingFormData(ctx, deps, data, nil, "")
			if templateID != "" {
				data.Context = ttcform.ContextTemplate
				data.TemplateID = templateID
				data.TaskOptions = ttcform.BuildTemplateTaskOptions(ctx, deps.ListPhasesByJobTemplate, deps.ListTasksByPhase, templateID, "")
			} else {
				data.Context = ttcform.ContextStandalone
			}
			return view.OK("template-task-criteria-drawer-form", data)
		}

		// POST — create template task criteria
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		sequenceOrder := int32(0)
		if v := r.FormValue("sequence_order"); v != "" {
			if n, err := fmt.Sscanf(v, "%d", &sequenceOrder); n == 0 || err != nil {
				sequenceOrder = 0
			}
		}
		// Default status="active" on create (proto-entity-status-conventions
		// silent-failure trap) — every list/table filter defaults to active=true.
		mode, scaleID := ratingConfigurationFromRequest(r)
		createResp, err := deps.CreateTemplateTaskCriteria(ctx, &ttcpb.CreateTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{
				JobTemplateTaskId: r.FormValue("job_template_task_id"),
				OutcomeCriteriaId: r.FormValue("outcome_criteria_id"),
				SequenceOrder:     sequenceOrder,
				Active:            true,
				RatingMode:        mode,
				RatingScaleId:     scaleID,
			},
		})
		if err != nil {
			log.Printf("Failed to create template task criteria: %v", err)
			return view.HTMXError(err.Error())
		}
		if createResp == nil || len(createResp.GetData()) == 0 {
			return view.HTMXError("Created template task criteria did not return an ID")
		}
		if err := saveRatingDescriptions(ctx, deps, createResp.GetData()[0].GetId(), mode, scaleID, r); err != nil {
			log.Printf("Failed to save rating descriptions: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess(refreshTableFor(r.FormValue("job_template_id")))
	})
}

// NewEditAction creates the template task criteria edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "update") {
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

			readResp, err := deps.ReadTemplateTaskCriteria(ctx, &ttcpb.ReadTemplateTaskCriteriaRequest{
				Data: &ttcpb.TemplateTaskCriteria{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read template task criteria %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			templateID := viewCtx.Request.URL.Query().Get("job_template_id")
			data := &ttcform.Data{
				FormAction:        route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:            true,
				ID:                id,
				Context:           ttcform.ContextStandalone,
				JobTemplateTaskID: record.GetJobTemplateTaskId(),
				OutcomeCriteriaID: record.GetOutcomeCriteriaId(),
				SequenceOrder:     record.GetSequenceOrder(),
				Active:            record.GetActive(),
				Labels:            deps.Labels,
				CriteriaOptions:   ttcform.BuildOutcomeCriteriaOptions(ctx, deps.ListOutcomeCriterias, record.GetOutcomeCriteriaId()),
				CommonLabels:      nil, // injected by ViewAdapter
			}
			if templateID != "" {
				data.Context = ttcform.ContextTemplate
				data.TemplateID = templateID
				data.TaskOptions = ttcform.BuildTemplateTaskOptions(
					ctx,
					deps.ListPhasesByJobTemplate,
					deps.ListTasksByPhase,
					templateID,
					record.GetJobTemplateTaskId(),
				)
			}
			populateRatingFormData(ctx, deps, data, record, id)
			return view.OK("template-task-criteria-drawer-form", data)
		}

		// POST — update template task criteria
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

		sequenceOrder := int32(0)
		if v := r.FormValue("sequence_order"); v != "" {
			if n, err := fmt.Sscanf(v, "%d", &sequenceOrder); n == 0 || err != nil {
				sequenceOrder = 0
			}
		}
		active := r.FormValue("active") == "true" || r.FormValue("active") == "1"
		mode, scaleID := ratingConfigurationFromRequest(r)

		_, err := deps.UpdateTemplateTaskCriteria(ctx, &ttcpb.UpdateTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{
				Id:                id,
				JobTemplateTaskId: r.FormValue("job_template_task_id"),
				OutcomeCriteriaId: r.FormValue("outcome_criteria_id"),
				SequenceOrder:     sequenceOrder,
				Active:            active,
				RatingMode:        mode,
				RatingScaleId:     scaleID,
			},
		})
		if err != nil {
			log.Printf("Failed to update template task criteria %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		if err := saveRatingDescriptions(ctx, deps, id, mode, scaleID, r); err != nil {
			log.Printf("Failed to save rating descriptions for %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		if templateID := r.FormValue("job_template_id"); templateID != "" {
			return view.HTMXSuccess(refreshTableFor(templateID))
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

func populateRatingFormData(ctx context.Context, deps *ModuleDeps, data *ttcform.Data, record *ttcpb.TemplateTaskCriteria, parentID string) {
	mode := ttcform.RatingModeStandard
	if record != nil && record.GetRatingMode() == enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION {
		mode = ttcform.RatingModeNumericWithDescription
	}
	scaleID := ""
	if record != nil {
		scaleID = record.GetRatingScaleId()
	}
	data.RatingMode = mode
	data.RatingModeOptions = ttcform.BuildRatingModeOptions(mode, ttcform.RatingModeLabels{
		Standard:               deps.Labels.Form.RatingModeStandard,
		NumericWithDescription: deps.Labels.Form.RatingModeNumericWithDescription,
	})
	data.RatingScaleID = scaleID
	data.RatingScaleOptions = ttcform.BuildScoreScaleOptions(ctx, deps.ListScoreScales, scaleID)
	data.RatingDescriptionRows = ttcform.BuildRatingDescriptionRows(
		ctx,
		deps.ListScoreScaleBands,
		deps.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteria,
		parentID,
		ttcform.BuildScoreScaleNameMap(ctx, deps.ListScoreScales),
	)
}

func ratingConfigurationFromRequest(r *http.Request) (*enums.RatingMode, *string) {
	mode := enums.RatingMode_RATING_MODE_STANDARD
	if r.FormValue("rating_mode") == ttcform.RatingModeNumericWithDescription {
		mode = enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION
	}
	scale := strings.TrimSpace(r.FormValue("rating_scale_id"))
	if mode != enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION || scale == "" {
		return &mode, nil
	}
	return &mode, &scale
}

func saveRatingDescriptions(ctx context.Context, deps *ModuleDeps, parentID string, mode *enums.RatingMode, scaleID *string, r *http.Request) error {
	if deps.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteria == nil ||
		deps.CreateTemplateTaskCriteriaRatingDescription == nil ||
		deps.UpdateTemplateTaskCriteriaRatingDescription == nil ||
		deps.DeleteTemplateTaskCriteriaRatingDescription == nil {
		if mode != nil && *mode == enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION {
			return fmt.Errorf("rating description authoring is unavailable")
		}
		return nil
	}
	existingResp, err := deps.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteria(ctx, &ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaRequest{TemplateTaskCriteriaId: parentID})
	if err != nil {
		return fmt.Errorf("failed to load existing rating descriptions: %w", err)
	}
	existing := map[string]*ttcrdpb.TemplateTaskCriteriaRatingDescription{}
	if existingResp != nil {
		for _, row := range existingResp.GetTemplateTaskCriteriaRatingDescriptions() {
			if row != nil {
				existing[row.GetScoreScaleBandId()] = row
			}
		}
	}

	selectedScale := ""
	if scaleID != nil {
		selectedScale = *scaleID
	}
	if mode == nil || *mode != enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION || selectedScale == "" {
		for _, row := range existing {
			if _, err := deps.DeleteTemplateTaskCriteriaRatingDescription(ctx, &ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionRequest{Data: &ttcrdpb.TemplateTaskCriteriaRatingDescription{Id: row.GetId()}}); err != nil {
				return fmt.Errorf("failed to clear rating description %s: %w", row.GetId(), err)
			}
		}
		return nil
	}

	bandIDs := r.Form["rating_description_band_id"]
	scaleIDs := r.Form["rating_description_scale_id"]
	descriptions := r.Form["rating_description_text"]
	seen := map[string]bool{}
	for i, bandID := range bandIDs {
		if i >= len(scaleIDs) || scaleIDs[i] != selectedScale {
			continue
		}
		seen[bandID] = true
		description := strings.TrimSpace(valueAt(descriptions, i))
		existingRow := existing[bandID]
		if description == "" {
			if existingRow != nil {
				if _, err := deps.DeleteTemplateTaskCriteriaRatingDescription(ctx, &ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionRequest{Data: &ttcrdpb.TemplateTaskCriteriaRatingDescription{Id: existingRow.GetId()}}); err != nil {
					return fmt.Errorf("failed to clear rating description %s: %w", bandID, err)
				}
			}
			continue
		}
		if existingRow != nil {
			_, err = deps.UpdateTemplateTaskCriteriaRatingDescription(ctx, &ttcrdpb.UpdateTemplateTaskCriteriaRatingDescriptionRequest{Data: &ttcrdpb.TemplateTaskCriteriaRatingDescription{
				Id: existingRow.GetId(), TemplateTaskCriteriaId: parentID, ScoreScaleBandId: bandID, Description: description, SequenceOrder: int32(i), Active: true,
			}})
		} else {
			_, err = deps.CreateTemplateTaskCriteriaRatingDescription(ctx, &ttcrdpb.CreateTemplateTaskCriteriaRatingDescriptionRequest{Data: &ttcrdpb.TemplateTaskCriteriaRatingDescription{
				TemplateTaskCriteriaId: parentID, ScoreScaleBandId: bandID, Description: description, SequenceOrder: int32(i), Active: true,
			}})
		}
		if err != nil {
			return fmt.Errorf("failed to save rating description %s: %w", bandID, err)
		}
	}
	for bandID, row := range existing {
		if !seen[bandID] {
			if _, err := deps.DeleteTemplateTaskCriteriaRatingDescription(ctx, &ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionRequest{Data: &ttcrdpb.TemplateTaskCriteriaRatingDescription{Id: row.GetId()}}); err != nil {
				return fmt.Errorf("failed to remove stale rating description %s: %w", bandID, err)
			}
		}
	}
	return nil
}

func valueAt(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return values[index]
}

// NewDeleteAction creates the template task criteria delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		// return_table lets a caller outside the module's own list page (e.g.
		// the job_template detail Standards tab's per-row remove action)
		// redirect the post-delete refresh at its own table instead.
		returnTable := viewCtx.Request.URL.Query().Get("return_table")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
			if returnTable == "" {
				returnTable = viewCtx.Request.FormValue("return_table")
			}
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteTemplateTaskCriteria(ctx, &ttcpb.DeleteTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete template task criteria %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		if returnTable != "" {
			return view.HTMXSuccess(returnTable)
		}
		return view.HTMXSuccess("template-task-criteria-table")
	})
}

// NewBulkDeleteAction creates the template task criteria bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteTemplateTaskCriteria(ctx, &ttcpb.DeleteTemplateTaskCriteriaRequest{
				Data: &ttcpb.TemplateTaskCriteria{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete template task criteria %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("template-task-criteria-table")
	})
}
