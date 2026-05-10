package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"
	jobform "github.com/erniealice/fayna-golang/views/job/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// NewEditAction creates the job edit action (GET = form, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			readResp, err := deps.ReadJob(ctx, &jobpb.ReadJobRequest{
				Data: &jobpb.Job{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job %s: %v", id, err)
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			currentStatus := record.GetStatus().String()
			return view.OK("job-drawer-form", &jobform.Data{
				FormAction:         route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:             true,
				ID:                 id,
				Name:               record.GetName(),
				ClientID:           record.GetClientId(),
				LocationID:         record.GetLocationId(),
				Status:             currentStatus,
				StatusOptions:      jobform.BuildStatusOptions(currentStatus),
				BillingRuleType:    record.GetBillingRuleType().String(),
				BillingRuleOptions: jobform.BuildBillingRuleOptions(record.GetBillingRuleType().String()),
				// ClientName/LocationName: not yet joined on the proto read response.
				// The auto-complete will re-resolve the label from the stored value on open.
				ClientSearchURL:   deps.ClientSearchURL,
				LocationSearchURL: deps.LocationSearchURL,
				Labels:            deps.Labels,
				CommonLabels:      nil, // injected by ViewAdapter
			})
		}

		// POST — update job
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		r := viewCtx.Request

		_, err := deps.UpdateJob(ctx, &jobpb.UpdateJobRequest{
			Data: &jobpb.Job{
				Id:              id,
				Name:            r.FormValue("name"),
				ClientId:        strPtr(r.FormValue("client_id")),
				LocationId:      strPtr(r.FormValue("location_id")),
				Status:          jobStatusToEnum(r.FormValue("status")),
				BillingRuleType: billingRuleTypeToEnum(r.FormValue("billing_rule_type")),
			},
		})
		if err != nil {
			log.Printf("Failed to update job %s: %v", id, err)
			return fayna.HTMXError(err.Error())
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
