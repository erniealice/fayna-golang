package action

import (
	"context"
	"log"
	"net/http"

	jobtemplateform "github.com/erniealice/fayna-golang/views/job_template/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// NewEditAction creates the job template edit action (GET = form, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			readResp, err := deps.ReadJobTemplate(ctx, &jobtemplatepb.ReadJobTemplateRequest{
				Data: &jobtemplatepb.JobTemplate{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job template %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			desc := ""
			if record.Description != nil {
				desc = *record.Description
			}

			return view.OK("job-template-drawer-form", &jobtemplateform.Data{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				Name:         record.GetName(),
				Description:  desc,
				Active:       record.GetActive(),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — update job template
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		active := r.FormValue("active") == "true" || r.FormValue("active") == "1"

		_, err := deps.UpdateJobTemplate(ctx, &jobtemplatepb.UpdateJobTemplateRequest{
			Data: &jobtemplatepb.JobTemplate{
				Id:          id,
				Name:        r.FormValue("name"),
				Description: strPtr(r.FormValue("description")),
				Active:      active,
			},
		})
		if err != nil {
			log.Printf("Failed to update job template %s: %v", id, err)
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
