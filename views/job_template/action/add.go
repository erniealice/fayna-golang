package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"
	jobtemplateform "github.com/erniealice/fayna-golang/views/job_template/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// NewAddAction creates the job template add action (GET = form, POST = create).
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("job-template-drawer-form", &jobtemplateform.Data{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create job template
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request

		resp, err := deps.CreateJobTemplate(ctx, &jobtemplatepb.CreateJobTemplateRequest{
			Data: &jobtemplatepb.JobTemplate{
				Name:        r.FormValue("name"),
				Description: strPtr(r.FormValue("description")),
			},
		})
		if err != nil {
			log.Printf("Failed to create job template: %v", err)
			return fayna.HTMXError(err.Error())
		}

		newID := ""
		if respData := resp.GetData(); len(respData) > 0 {
			newID = respData[0].GetId()
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

		return fayna.HTMXSuccess("job-templates-table")
	})
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}
