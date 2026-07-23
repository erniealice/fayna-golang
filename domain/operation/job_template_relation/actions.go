package job_template_relation

import (
	"context"
	"fmt"
	"log"
	"net/http"

	relationform "github.com/erniealice/fayna-golang/domain/operation/job_template_relation/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	relationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
)

// spawnGraphTableID is the table id on the job_template detail Spawn Graph
// tab (job_template/detail/spawn_graph.go). refreshTableFor picks this over
// the module's own list-page table id when the drawer was opened in
// ContextTemplate, so the HX-Trigger refreshes the right DOM table.
const spawnGraphTableID = "jt-spawn-graph-table"

// refreshTableFor returns the table id to refresh on success, based on
// whether the drawer was opened from a job_template detail Spawn Graph tab
// (parentTemplateID non-empty) or the module's own list page.
func refreshTableFor(parentTemplateID string) string {
	if parentTemplateID != "" {
		return spawnGraphTableID
	}
	return "job-template-relations-table"
}

// NewAddAction creates the job template relation add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_relation", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			parentTemplateID := viewCtx.Request.URL.Query().Get("parent_template_id")
			data := &relationform.Data{
				FormAction:           deps.Routes.AddURL,
				Labels:               deps.Labels,
				Active:               true,
				ChildTemplateOptions: relationform.BuildTemplateOptions(ctx, deps.ListJobTemplates, ""),
				RelationTypeOptions:  relationform.BuildRelationTypeOptions(""),
				CommonLabels:         nil, // injected by ViewAdapter
			}
			if parentTemplateID != "" {
				data.Context = relationform.ContextTemplate
				data.ParentTemplateID = parentTemplateID
				data.ParentTemplateName = resolveTemplateName(ctx, deps, parentTemplateID)
			} else {
				data.Context = relationform.ContextStandalone
				data.ParentTemplateOptions = relationform.BuildTemplateOptions(ctx, deps.ListJobTemplates, "")
			}
			return view.OK("job-template-relation-drawer-form", data)
		}

		// POST — create job template relation
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

		if deps.CreateJobTemplateRelation == nil {
			return view.HTMXError("Create not available")
		}

		_, err := deps.CreateJobTemplateRelation(ctx, &relationpb.CreateJobTemplateRelationRequest{
			Data: &relationpb.JobTemplateRelation{
				ParentTemplateId: r.FormValue("parent_template_id"),
				ChildTemplateId:  r.FormValue("child_template_id"),
				RelationType:     relationTypeFromString(r.FormValue("relation_type")),
				SequenceOrder:    sequenceOrder,
				Active:           true, // default status=active on create
			},
		})
		if err != nil {
			log.Printf("Failed to create job template relation: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess(refreshTableFor(r.FormValue("parent_template_id")))
	})
}

// NewEditAction creates the job template relation edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_relation", "update") {
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
			if deps.ReadJobTemplateRelation == nil {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}

			readResp, err := deps.ReadJobTemplateRelation(ctx, &relationpb.ReadJobTemplateRelationRequest{
				Data: &relationpb.JobTemplateRelation{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job template relation %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			return view.OK("job-template-relation-drawer-form", &relationform.Data{
				FormAction:            route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:                true,
				ID:                    id,
				Context:               relationform.ContextStandalone,
				ParentTemplateID:      record.GetParentTemplateId(),
				ChildTemplateID:       record.GetChildTemplateId(),
				RelationType:          record.GetRelationType().String(),
				SequenceOrder:         record.GetSequenceOrder(),
				Active:                record.GetActive(),
				Labels:                deps.Labels,
				ParentTemplateOptions: relationform.BuildTemplateOptions(ctx, deps.ListJobTemplates, record.GetParentTemplateId()),
				ChildTemplateOptions:  relationform.BuildTemplateOptions(ctx, deps.ListJobTemplates, record.GetChildTemplateId()),
				RelationTypeOptions:   relationform.BuildRelationTypeOptions(record.GetRelationType().String()),
				CommonLabels:          nil, // injected by ViewAdapter
			})
		}

		// POST — update job template relation
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

		if deps.UpdateJobTemplateRelation == nil {
			return view.HTMXError("Update not available")
		}

		_, err := deps.UpdateJobTemplateRelation(ctx, &relationpb.UpdateJobTemplateRelationRequest{
			Data: &relationpb.JobTemplateRelation{
				Id:               id,
				ParentTemplateId: r.FormValue("parent_template_id"),
				ChildTemplateId:  r.FormValue("child_template_id"),
				RelationType:     relationTypeFromString(r.FormValue("relation_type")),
				SequenceOrder:    sequenceOrder,
				Active:           active,
			},
		})
		if err != nil {
			log.Printf("Failed to update job template relation %s: %v", id, err)
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

// NewDeleteAction creates the job template relation delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_relation", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		// return_table lets a caller outside the module's own list page (e.g.
		// the job_template detail Spawn Graph tab's per-row remove action)
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
		if deps.DeleteJobTemplateRelation == nil {
			return view.HTMXError("Delete not available")
		}

		_, err := deps.DeleteJobTemplateRelation(ctx, &relationpb.DeleteJobTemplateRelationRequest{
			Data: &relationpb.JobTemplateRelation{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job template relation %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		if returnTable != "" {
			return view.HTMXSuccess(returnTable)
		}
		return view.HTMXSuccess("job-template-relations-table")
	})
}

// NewBulkDeleteAction creates the job template relation bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_relation", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}
		if deps.DeleteJobTemplateRelation == nil {
			return view.HTMXError("Delete not available")
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteJobTemplateRelation(ctx, &relationpb.DeleteJobTemplateRelationRequest{
				Data: &relationpb.JobTemplateRelation{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job template relation %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("job-template-relations-table")
	})
}

// relationTypeFromString maps a JobTemplateRelationType enum name string
// (posted by the relation_type select) to the proto enum, defaulting to
// UNSPECIFIED for an empty/unknown value.
func relationTypeFromString(s string) relationpb.JobTemplateRelationType {
	if v, ok := relationpb.JobTemplateRelationType_value[s]; ok {
		return relationpb.JobTemplateRelationType(v)
	}
	return relationpb.JobTemplateRelationType_JOB_TEMPLATE_RELATION_TYPE_UNSPECIFIED
}

// resolveTemplateName reads the parent template's name for the read-only
// display row when Context == ContextTemplate. Nil-safe: an unwired
// ReadJobTemplate or a failed read falls back to the raw id.
func resolveTemplateName(ctx context.Context, deps *ModuleDeps, id string) string {
	if deps.ReadJobTemplate == nil {
		return id
	}
	resp, err := deps.ReadJobTemplate(ctx, &jobtemplatepb.ReadJobTemplateRequest{
		Data: &jobtemplatepb.JobTemplate{Id: id},
	})
	if err != nil || resp == nil {
		return id
	}
	data := resp.GetData()
	if len(data) == 0 {
		return id
	}
	if name := data[0].GetName(); name != "" {
		return name
	}
	return id
}
