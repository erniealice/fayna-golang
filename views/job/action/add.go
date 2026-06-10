package action

import (
	"context"
	"log"
	"net/http"

	jobform "github.com/erniealice/fayna-golang/views/job/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// NewAddAction creates the job add action (GET = form, POST = create).
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			const defaultStatus = "JOB_STATUS_DRAFT"
			return view.OK("job-drawer-form", &jobform.Data{
				FormAction:         deps.Routes.AddURL,
				Status:             defaultStatus,
				StatusOptions:      jobform.BuildStatusOptions(defaultStatus),
				BillingRuleOptions: jobform.BuildBillingRuleOptions(""),
				ClientSearchURL:    deps.ClientSearchURL,
				LocationSearchURL:  deps.LocationSearchURL,
				Labels:             deps.Labels,
				CommonLabels:       nil, // injected by ViewAdapter
			})
		}

		// POST — create job
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		r := viewCtx.Request

		resp, err := deps.CreateJob(ctx, &jobpb.CreateJobRequest{
			Data: &jobpb.Job{
				Name:            r.FormValue("name"),
				ClientId:        strPtr(r.FormValue("client_id")),
				LocationId:      strPtr(r.FormValue("location_id")),
				Status:          jobStatusToEnum(r.FormValue("status")),
				BillingRuleType: billingRuleTypeToEnum(r.FormValue("billing_rule_type")),
			},
		})
		if err != nil {
			log.Printf("Failed to create job: %v", err)
			return view.HTMXError(err.Error())
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

		return view.HTMXSuccess("jobs-table")
	})
}
