package template_task_criteria

import (
	"context"
	"fmt"
	"log"
	"net/http"

	ttcform "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

// NewAddAction creates the template task criteria add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("template-task-criteria-drawer-form", &ttcform.Data{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
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

		_, err := deps.CreateTemplateTaskCriteria(ctx, &ttcpb.CreateTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{
				JobTemplateTaskId: r.FormValue("job_template_task_id"),
				OutcomeCriteriaId: r.FormValue("outcome_criteria_id"),
				SequenceOrder:     sequenceOrder,
				Active:            true,
			},
		})
		if err != nil {
			log.Printf("Failed to create template task criteria: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("template-task-criteria-table")
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

			return view.OK("template-task-criteria-drawer-form", &ttcform.Data{
				FormAction:        route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:            true,
				ID:                id,
				JobTemplateTaskID: record.GetJobTemplateTaskId(),
				OutcomeCriteriaID: record.GetOutcomeCriteriaId(),
				SequenceOrder:     record.GetSequenceOrder(),
				Labels:            deps.Labels,
				CommonLabels:      nil, // injected by ViewAdapter
			})
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

		_, err := deps.UpdateTemplateTaskCriteria(ctx, &ttcpb.UpdateTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{
				Id:                id,
				JobTemplateTaskId: r.FormValue("job_template_task_id"),
				OutcomeCriteriaId: r.FormValue("outcome_criteria_id"),
				SequenceOrder:     sequenceOrder,
			},
		})
		if err != nil {
			log.Printf("Failed to update template task criteria %s: %v", id, err)
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

// NewDeleteAction creates the template task criteria delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "delete") {
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

		_, err := deps.DeleteTemplateTaskCriteria(ctx, &ttcpb.DeleteTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete template task criteria %s: %v", id, err)
			return view.HTMXError(err.Error())
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
