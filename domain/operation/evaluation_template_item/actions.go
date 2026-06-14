package evaluation_template_item

import (
	"context"
	"log"
	"net/http"
	"strconv"

	itemform "github.com/erniealice/fayna-golang/domain/operation/evaluation_template_item/form"

	"github.com/erniealice/pyeza-golang/route"
	pyezatypes "github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
)

// NewAddAction — GET renders the rubric-item drawer (criterion autocomplete),
// POST creates an item. Gated evaluation_template:update (rubric authoring is a
// template-edit operation). workspace_id is copied from the parent template by
// the espyna use case (H1 parity).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		templateID := viewCtx.Request.PathValue("template_id")
		if templateID == "" {
			templateID = viewCtx.Request.URL.Query().Get("template_id")
		}

		if viewCtx.Request.Method == http.MethodGet {
			if templateID == "" {
				return view.HTMXError(deps.Labels.Errors.TemplateIDRequired)
			}
			return view.OK("evaluation_template_item-drawer-form", &itemform.Data{
				FormAction:           route.ResolveURL(deps.Routes.AddURL),
				EvaluationTemplateID: templateID,
				SequenceOrder:        nextSequence(ctx, deps, templateID),
				CriteriaOptions:      loadCriteriaOptions(ctx, deps),
				Labels:               deps.Labels,
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		r := viewCtx.Request
		if templateID == "" {
			templateID = r.FormValue("evaluation_template_id")
		}
		if templateID == "" {
			return view.HTMXError(deps.Labels.Errors.TemplateIDRequired)
		}
		if r.FormValue("outcome_criteria_id") == "" {
			return view.HTMXError(deps.Labels.Errors.CriterionRequired)
		}

		_, err := deps.CreateEvaluationTemplateItem(ctx, &itempb.CreateEvaluationTemplateItemRequest{
			Data: &itempb.EvaluationTemplateItem{
				EvaluationTemplateId: templateID,
				OutcomeCriteriaId:    r.FormValue("outcome_criteria_id"),
				SequenceOrder:        parseInt32(r.FormValue("sequence_order")),
				WeightOverride:       parseFloatPtr(r.FormValue("weight_override")),
				QuestionLabel:        strPtrIfNotEmpty(r.FormValue("question_label")),
				QuestionPrompt:       strPtrIfNotEmpty(r.FormValue("question_prompt")),
				RequiredOverride:     boolPtr(r.FormValue("required_override") == "true" || r.FormValue("required_override") == "on"),
				Active:               true,
			},
		})
		if err != nil {
			log.Printf("Failed to create rubric item for template %s: %v", templateID, err)
			return view.HTMXError(err.Error())
		}

		// Refresh the template detail Items tab.
		return view.HTMXSuccess("tabContent")
	})
}

// NewEditAction — GET pre-filled drawer, POST updates a rubric item.
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		itemID := viewCtx.Request.PathValue("item_id")
		if itemID == "" {
			itemID = viewCtx.Request.URL.Query().Get("item_id")
		}

		if viewCtx.Request.Method == http.MethodGet {
			if itemID == "" {
				return view.HTMXError(deps.Labels.Errors.IDRequired)
			}
			readResp, err := deps.ReadEvaluationTemplateItem(ctx, &itempb.ReadEvaluationTemplateItemRequest{
				Data: &itempb.EvaluationTemplateItem{Id: itemID},
			})
			if err != nil {
				log.Printf("Failed to read rubric item %s: %v", itemID, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			rows := readResp.GetData()
			if len(rows) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			rec := rows[0]

			weight := ""
			if rec.WeightOverride != nil {
				weight = strconv.FormatFloat(rec.GetWeightOverride(), 'f', -1, 64)
			}

			return view.OK("evaluation_template_item-drawer-form", &itemform.Data{
				FormAction:           route.ResolveURL(deps.Routes.EditURL, "item_id", itemID),
				IsEdit:               true,
				ID:                   itemID,
				EvaluationTemplateID: rec.GetEvaluationTemplateId(),
				OutcomeCriteriaID:    rec.GetOutcomeCriteriaId(),
				SequenceOrder:        rec.GetSequenceOrder(),
				WeightOverride:       weight,
				QuestionLabel:        rec.GetQuestionLabel(),
				QuestionPrompt:       rec.GetQuestionPrompt(),
				RequiredOverride:     rec.GetRequiredOverride(),
				CriteriaOptions:      loadCriteriaOptions(ctx, deps),
				CriteriaTypeLabel:    criteriaTypeForID(ctx, deps, rec.GetOutcomeCriteriaId()),
				Labels:               deps.Labels,
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		r := viewCtx.Request
		if itemID == "" {
			itemID = r.FormValue("id")
		}
		if itemID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.UpdateEvaluationTemplateItem(ctx, &itempb.UpdateEvaluationTemplateItemRequest{
			Data: &itempb.EvaluationTemplateItem{
				Id:                itemID,
				OutcomeCriteriaId: r.FormValue("outcome_criteria_id"),
				SequenceOrder:     parseInt32(r.FormValue("sequence_order")),
				WeightOverride:    parseFloatPtr(r.FormValue("weight_override")),
				QuestionLabel:     strPtrIfNotEmpty(r.FormValue("question_label")),
				QuestionPrompt:    strPtrIfNotEmpty(r.FormValue("question_prompt")),
				RequiredOverride:  boolPtr(r.FormValue("required_override") == "true" || r.FormValue("required_override") == "on"),
			},
		})
		if err != nil {
			log.Printf("Failed to update rubric item %s: %v", itemID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("tabContent")
	})
}

// NewRemoveAction — POST soft-deletes (active=false) a rubric item.
func NewRemoveAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		itemID := viewCtx.Request.PathValue("item_id")
		if itemID == "" {
			_ = viewCtx.Request.ParseForm()
			itemID = viewCtx.Request.FormValue("item_id")
		}
		if itemID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		_, err := deps.DeleteEvaluationTemplateItem(ctx, &itempb.DeleteEvaluationTemplateItemRequest{
			Data: &itempb.EvaluationTemplateItem{Id: itemID},
		})
		if err != nil {
			log.Printf("Failed to remove rubric item %s: %v", itemID, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("tabContent")
	})
}

// --- helpers ---

func loadCriteriaOptions(ctx context.Context, deps *ModuleDeps) []pyezatypes.SelectOption {
	if deps.ListOutcomeCriterias == nil {
		return nil
	}
	resp, err := deps.ListOutcomeCriterias(ctx, &criteriapb.ListOutcomeCriteriasRequest{})
	if err != nil {
		log.Printf("Failed to load criteria for rubric-item drawer: %v", err)
		return nil
	}
	return itemform.BuildCriteriaOptions(resp.GetData())
}

func criteriaTypeForID(ctx context.Context, deps *ModuleDeps, criterionID string) string {
	if criterionID == "" || deps.ListOutcomeCriterias == nil {
		return ""
	}
	resp, err := deps.ListOutcomeCriterias(ctx, &criteriapb.ListOutcomeCriteriasRequest{})
	if err != nil {
		return ""
	}
	for _, c := range resp.GetData() {
		if c.GetId() == criterionID {
			return itemform.CriteriaTypeDisplay(c.GetCriteriaType())
		}
	}
	return ""
}

// nextSequence returns the tail sequence_order + 1 for a template's active items.
func nextSequence(ctx context.Context, deps *ModuleDeps, templateID string) int32 {
	if deps.ListEvaluationTemplateItems == nil {
		return 1
	}
	resp, err := deps.ListEvaluationTemplateItems(ctx, &itempb.ListEvaluationTemplateItemsRequest{})
	if err != nil {
		return 1
	}
	var max int32
	for _, it := range resp.GetData() {
		if it.GetEvaluationTemplateId() == templateID && it.GetActive() && it.GetSequenceOrder() > max {
			max = it.GetSequenceOrder()
		}
	}
	return max + 1
}

func parseInt32(s string) int32 {
	v, _ := strconv.ParseInt(s, 10, 32)
	return int32(v)
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func boolPtr(b bool) *bool { return &b }

func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
