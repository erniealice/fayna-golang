package evaluation

import (
	"context"
	"log"
	"net/http"
	"strconv"

	evaluationform "github.com/erniealice/fayna-golang/domain/operation/evaluation/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
	resppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_response"
)

// actions.go — package-level action builders for the evaluation drawer-form
// (DF-1, polymorphic) + the named lifecycle verbs.
//
// IDOR / CR-5 (Q-EVAL-IDOR-1): all row-scope, fail-closed, and visibility gates
// are enforced INSIDE the injected use-case closures (use-case/adapter QUERY
// PREDICATE). The view NEVER reads client_id from the form/URL — client_id,
// subject_staff_id, subscription_seat_id, subscription_id are all server-locked
// (the proto requests are stamped server-side). The drawer's hidden fields are
// display-only anchors; the espyna layer re-derives them from session +
// seat context (pages.md §A.3).

// NewAddAction handles the score-submission drawer: GET = render the drawer
// (+ dimension slot for the active template), POST = create a DRAFT evaluation
// + its responses.
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			fd := evaluationform.GetFormData(ctx, &evaluationform.Deps{
				ListEvaluationTemplateItems: deps.ListEvaluationTemplateItems,
				Routes:                      formRoutes(deps.Routes),
				Labels:                      deps.Labels,
			}, evaluationform.GetFormDataInput{
				FormAction: deps.Routes.AddURL,
				TemplateID: viewCtx.Request.URL.Query().Get("template_id"),
				SeatID:     viewCtx.Request.URL.Query().Get("seat_id"),
			})
			return view.OK("evaluation-drawer-form", fd)
		}

		// POST — create DRAFT evaluation + responses. client_id/seat/subject are
		// server-stamped by CreateEvaluation; never read from the form here.
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}
		r := viewCtx.Request

		ev := &evalpb.Evaluation{
			EvaluationTemplateId: strPtrIfNotEmpty(r.FormValue("evaluation_template_id")),
			PeriodStart:          r.FormValue("period_start"),
			PeriodEnd:            r.FormValue("period_end"),
			Status:               evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT,
			Narrative:            strPtrIfNotEmpty(r.FormValue("narrative")),
		}

		createResp, err := deps.CreateEvaluation(ctx, &evalpb.CreateEvaluationRequest{Data: ev})
		if err != nil {
			log.Printf("Failed to create evaluation: %v", err)
			return view.HTMXError(err.Error())
		}

		newID := ""
		if d := createResp.GetData(); len(d) > 0 {
			newID = d[0].GetId()
		}

		// Persist the per-dimension responses (one CreateEvaluationResponse per
		// rubric item). The answer column is keyed off the criteria_type.
		if newID != "" && deps.CreateEvaluationResponse != nil {
			persistResponses(ctx, deps, newID, r)
		}

		if newID != "" {
			return view.ViewResult{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"HX-Trigger":  `{"formSuccess":true}`,
					"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", newID),
				},
			}
		}
		return view.HTMXSuccess("evaluation-list-table")
	})
}

// NewEditAction handles editing a DRAFT evaluation: GET = pre-filled drawer,
// POST = submit (DRAFT→SUBMITTED, snapshot + ComputeEvaluationScore server-side).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			readResp, err := deps.ReadEvaluation(ctx, &evalpb.ReadEvaluationRequest{
				Data: &evalpb.Evaluation{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read evaluation %s for edit: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			data := readResp.GetData()
			if len(data) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			e := data[0]
			fd := evaluationform.GetFormData(ctx, &evaluationform.Deps{
				ListEvaluationTemplateItems: deps.ListEvaluationTemplateItems,
				Routes:                      formRoutes(deps.Routes),
				Labels:                      deps.Labels,
			}, evaluationform.GetFormDataInput{
				FormAction:  route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:      true,
				ID:          id,
				TemplateID:  e.GetEvaluationTemplateId(),
				PeriodStart: e.GetPeriodStart(),
				PeriodEnd:   e.GetPeriodEnd(),
				Narrative:   e.GetNarrative(),
			})
			return view.OK("evaluation-drawer-form", fd)
		}

		// POST — submit (DRAFT→SUBMITTED). The espyna layer snapshots the rubric,
		// computes overall_score, stamps submitted_at, and re-asserts IDOR gates.
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}
		r := viewCtx.Request

		_, err := deps.UpdateEvaluation(ctx, &evalpb.UpdateEvaluationRequest{
			Data: &evalpb.Evaluation{
				Id:        id,
				Narrative: strPtrIfNotEmpty(r.FormValue("narrative")),
				Status:    evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED,
			},
		})
		if err != nil {
			log.Printf("Failed to submit evaluation %s: %v", id, err)
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

// NewDimensionSlotAction renders the polymorphic dimension slot fragment for a
// given template (HTMX-swapped into #evaluation-dimension-slot on picker change).
func NewDimensionSlotAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		templateID := viewCtx.Request.URL.Query().Get("template_id")
		slot := evaluationform.GetDimensionSlot(ctx, &evaluationform.Deps{
			ListEvaluationTemplateItems: deps.ListEvaluationTemplateItems,
			Routes:                      formRoutes(deps.Routes),
			Labels:                      deps.Labels,
		}, templateID)
		return view.OK("evaluation-dimension-slot", slot)
	})
}

// NewSignOffAction handles SUBMITTED→SIGNED_OFF (is_owner-gated server-side).
func NewSignOffAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "sign_off") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		if deps.SignOffEvaluation == nil {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		_, err := deps.SignOffEvaluation(ctx, &evalpb.UpdateEvaluationRequest{
			Data: &evalpb.Evaluation{
				Id:     id,
				Status: evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF,
			},
		})
		if err != nil {
			log.Printf("Failed to sign off evaluation %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("evaluation-list-table")
	})
}

// NewArchiveAction handles SUBMITTED→ARCHIVED.
func NewArchiveAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		if err := archiveOne(ctx, deps, id); err != nil {
			log.Printf("Failed to archive evaluation %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("evaluation-list-table")
	})
}

// NewBulkArchiveAction archives all selected SUBMITTED evaluations.
func NewBulkArchiveAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		_ = viewCtx.Request.ParseMultipartForm(32 << 20)
		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No reviews selected")
		}
		for _, id := range ids {
			if err := archiveOne(ctx, deps, id); err != nil {
				log.Printf("Failed to archive evaluation %s: %v", id, err)
			}
		}
		return view.HTMXSuccess("evaluation-list-table")
	})
}

// NewDeleteAction handles a hard delete (staff {1,2}, evaluation:delete).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		_, err := deps.DeleteEvaluation(ctx, &evalpb.DeleteEvaluationRequest{
			Data: &evalpb.Evaluation{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete evaluation %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess("evaluation-list-table")
	})
}

// archiveOne routes through the dedicated ArchiveEvaluation closure when wired,
// otherwise falls back to the shaped UpdateEvaluation status transition.
func archiveOne(ctx context.Context, deps *ModuleDeps, id string) error {
	req := &evalpb.UpdateEvaluationRequest{
		Data: &evalpb.Evaluation{
			Id:     id,
			Status: evalpb.EvaluationStatus_EVALUATION_STATUS_ARCHIVED,
		},
	}
	if deps.ArchiveEvaluation != nil {
		_, err := deps.ArchiveEvaluation(ctx, req)
		return err
	}
	_, err := deps.UpdateEvaluation(ctx, req)
	return err
}

// persistResponses writes one evaluation_response per rubric dimension submitted
// in the form. Inputs are keyed dim_{criteria_id}; the answer column is chosen
// from the dim_type_{criteria_id} hidden field stamped by the slot template.
func persistResponses(ctx context.Context, deps *ModuleDeps, evalID string, r *http.Request) {
	for key, vals := range r.Form {
		if len(vals) == 0 || len(key) < 4 || key[:4] != "dim_" {
			continue
		}
		criteriaID := key[4:]
		resp := &resppb.EvaluationResponse{
			EvaluationId:      evalID,
			OutcomeCriteriaId: criteriaID,
		}
		dimType := r.FormValue("dim_type_" + criteriaID)
		switch dimType {
		case "NUMERIC_SCORE", "NUMERIC_RANGE":
			if f, err := strconv.ParseFloat(vals[0], 64); err == nil {
				resp.NumericValue = &f
			}
		case "PASS_FAIL":
			pf := vals[0] == "pass" || vals[0] == "true" || vals[0] == "on"
			resp.PassFailValue = &pf
		case "CATEGORICAL":
			resp.CategoricalValue = strPtrIfNotEmpty(vals[0])
		default:
			resp.TextValue = strPtrIfNotEmpty(vals[0])
		}
		if _, err := deps.CreateEvaluationResponse(ctx, &resppb.CreateEvaluationResponseRequest{Data: resp}); err != nil {
			log.Printf("Failed to create evaluation response for criteria %s: %v", criteriaID, err)
		}
	}
}

// formRoutes projects the entity Routes into the form package's narrow route
// contract (the dimension-slot endpoint the picker HTMX-triggers).
func formRoutes(r Routes) evaluationform.Routes {
	return evaluationform.Routes{
		DimensionSlotURL: r.DimensionSlotURL,
	}
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil.
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
