package evaluation_template

import (
	"context"
	"log"
	"net/http"

	evaluationtemplateform "github.com/erniealice/fayna-golang/domain/operation/evaluation_template/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
	templatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template"
)

// NewAddAction — GET renders the template-header drawer, POST creates a DRAFT
// template. Staff-only (evaluation_template:create). Clients have no
// evaluation_template:* permission (acceptance #8) so this rejects for client
// principals at L2.
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("evaluation_template-drawer-form", &evaluationtemplateform.Data{
				FormAction:            deps.Routes.AddURL,
				EvaluationTypeOptions: evaluationtemplateform.DefaultEvaluationTypeOptions(),
				RelationshipOptions:   evaluationtemplateform.DefaultRelationshipOptions(),
				VisibilityOptions:     evaluationtemplateform.DefaultVisibilityOptions(),
				Labels:                deps.Labels,
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}
		r := viewCtx.Request

		_, err := deps.CreateEvaluationTemplate(ctx, &templatepb.CreateEvaluationTemplateRequest{
			Data: &templatepb.EvaluationTemplate{
				Name:             r.FormValue("name"),
				Description:      strPtrIfNotEmpty(r.FormValue("description")),
				EvaluationType:   parseEvaluationType(r.FormValue("evaluation_type")),
				RelationshipType: parseRelationshipType(r.FormValue("relationship_type")),
				VisibilityType:   parseVisibility(r.FormValue("visibility_type")),
				Status:           templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DRAFT,
				Active:           true,
			},
		})
		if err != nil {
			log.Printf("Failed to create evaluation template: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("evaluation_template-list-table")
	})
}

// NewEditAction — GET pre-filled drawer, POST updates a DRAFT template header.
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "update") {
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
			readResp, err := deps.ReadEvaluationTemplate(ctx, &templatepb.ReadEvaluationTemplateRequest{
				Data: &templatepb.EvaluationTemplate{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read evaluation template %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			rows := readResp.GetData()
			if len(rows) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			rec := rows[0]

			return view.OK("evaluation_template-drawer-form", &evaluationtemplateform.Data{
				FormAction:            route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:                true,
				ID:                    id,
				Name:                  rec.GetName(),
				Description:           rec.GetDescription(),
				EvaluationType:        rec.GetEvaluationType().String(),
				RelationshipType:      rec.GetRelationshipType().String(),
				VisibilityType:        rec.GetVisibilityType().String(),
				EvaluationTypeOptions: evaluationtemplateform.DefaultEvaluationTypeOptions(),
				RelationshipOptions:   evaluationtemplateform.DefaultRelationshipOptions(),
				VisibilityOptions:     evaluationtemplateform.DefaultVisibilityOptions(),
				Labels:                deps.Labels,
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

		_, err := deps.UpdateEvaluationTemplate(ctx, &templatepb.UpdateEvaluationTemplateRequest{
			Data: &templatepb.EvaluationTemplate{
				Id:               id,
				Name:             r.FormValue("name"),
				Description:      strPtrIfNotEmpty(r.FormValue("description")),
				EvaluationType:   parseEvaluationType(r.FormValue("evaluation_type")),
				RelationshipType: parseRelationshipType(r.FormValue("relationship_type")),
				VisibilityType:   parseVisibility(r.FormValue("visibility_type")),
			},
		})
		if err != nil {
			log.Printf("Failed to update evaluation template %s: %v", id, err)
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

// NewActivateAction — POST DRAFT->ACTIVE. BLOCKER-2: reject any item whose
// criteria_type is non-numeric AND carries a non-zero weight (a weighted
// non-numeric question cannot contribute to overall_score). The guard runs
// here AND is re-enforced server-side in the espyna ActivateEvaluationTemplate
// use case; the view check gives an inline message.
func NewActivateAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		id := pathOrFormID(viewCtx)
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if msg := weightedNonNumericGuard(ctx, deps, id); msg != "" {
			return view.HTMXError(msg)
		}

		_, err := deps.UpdateEvaluationTemplate(ctx, &templatepb.UpdateEvaluationTemplateRequest{
			Data: &templatepb.EvaluationTemplate{
				Id:     id,
				Status: templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_ACTIVE,
			},
		})
		if err != nil {
			log.Printf("Failed to activate evaluation template %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("evaluation_template-list-table")
	})
}

// NewDeprecateAction — POST ACTIVE->DEPRECATED.
func NewDeprecateAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		id := pathOrFormID(viewCtx)
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		_, err := deps.UpdateEvaluationTemplate(ctx, &templatepb.UpdateEvaluationTemplateRequest{
			Data: &templatepb.EvaluationTemplate{
				Id:     id,
				Status: templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DEPRECATED,
			},
		})
		if err != nil {
			log.Printf("Failed to deprecate evaluation template %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("evaluation_template-list-table")
	})
}

// NewCloneAction — POST creates a new DRAFT from an existing template
// (copied_from_id). Clone-flips-to-create: gates evaluation_template:create.
func NewCloneAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		id := pathOrFormID(viewCtx)
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		src := id
		_, err := deps.CreateEvaluationTemplate(ctx, &templatepb.CreateEvaluationTemplateRequest{
			Data: &templatepb.EvaluationTemplate{
				CopiedFromId: &src,
				Status:       templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DRAFT,
				Active:       true,
			},
		})
		if err != nil {
			log.Printf("Failed to clone evaluation template %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("evaluation_template-list-table")
	})
}

// NewBulkDeprecateAction — POST bulk ACTIVE->DEPRECATED.
func NewBulkDeprecateAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		_ = viewCtx.Request.ParseMultipartForm(32 << 20)
		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}
		for _, id := range ids {
			_, err := deps.UpdateEvaluationTemplate(ctx, &templatepb.UpdateEvaluationTemplateRequest{
				Data: &templatepb.EvaluationTemplate{
					Id:     id,
					Status: templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DEPRECATED,
				},
			})
			if err != nil {
				log.Printf("Failed to bulk-deprecate evaluation template %s: %v", id, err)
			}
		}
		return view.HTMXSuccess("evaluation_template-list-table")
	})
}

// weightedNonNumericGuard returns a non-empty inline error message when the
// template has any active item whose linked criterion is non-numeric AND
// carries a non-zero weight_override. Empty string = OK to activate.
func weightedNonNumericGuard(ctx context.Context, deps *ModuleDeps, templateID string) string {
	if deps.ListEvaluationTemplateItems == nil || deps.ListOutcomeCriterias == nil {
		return ""
	}
	itemsResp, err := deps.ListEvaluationTemplateItems(ctx, &itempb.ListEvaluationTemplateItemsRequest{})
	if err != nil {
		log.Printf("Failed to load items for activation guard %s: %v", templateID, err)
		return ""
	}
	critResp, err := deps.ListOutcomeCriterias(ctx, nil)
	if err != nil {
		log.Printf("Failed to load criteria for activation guard: %v", err)
		return ""
	}
	critType := map[string]enums.CriteriaType{}
	for _, c := range critResp.GetData() {
		critType[c.GetId()] = c.GetCriteriaType()
	}
	for _, it := range itemsResp.GetData() {
		if it.GetEvaluationTemplateId() != templateID || !it.GetActive() {
			continue
		}
		if it.WeightOverride == nil || it.GetWeightOverride() == 0 {
			continue
		}
		ct := critType[it.GetOutcomeCriteriaId()]
		if ct != enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE &&
			ct != enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE {
			return deps.Labels.Errors.WeightedNonNumeric
		}
	}
	return ""
}

func pathOrFormID(viewCtx *view.ViewContext) string {
	id := viewCtx.Request.PathValue("id")
	if id == "" {
		id = viewCtx.Request.URL.Query().Get("id")
	}
	if id == "" {
		_ = viewCtx.Request.ParseForm()
		id = viewCtx.Request.FormValue("id")
	}
	return id
}

func parseEvaluationType(s string) evalpb.EvaluationType {
	if v, ok := evalpb.EvaluationType_value["EVALUATION_TYPE_"+s]; ok {
		return evalpb.EvaluationType(v)
	}
	if v, ok := evalpb.EvaluationType_value[s]; ok {
		return evalpb.EvaluationType(v)
	}
	return evalpb.EvaluationType_EVALUATION_TYPE_UNSPECIFIED
}

func parseRelationshipType(s string) evalpb.RelationshipType {
	if v, ok := evalpb.RelationshipType_value["RELATIONSHIP_TYPE_"+s]; ok {
		return evalpb.RelationshipType(v)
	}
	if v, ok := evalpb.RelationshipType_value[s]; ok {
		return evalpb.RelationshipType(v)
	}
	return evalpb.RelationshipType_RELATIONSHIP_TYPE_UNSPECIFIED
}

func parseVisibility(s string) evalpb.VisibilityType {
	if v, ok := evalpb.VisibilityType_value["VISIBILITY_TYPE_"+s]; ok {
		return evalpb.VisibilityType(v)
	}
	if v, ok := evalpb.VisibilityType_value[s]; ok {
		return evalpb.VisibilityType(v)
	}
	return evalpb.VisibilityType_VISIBILITY_TYPE_UNSPECIFIED
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil.
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
