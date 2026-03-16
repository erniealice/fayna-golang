package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// FormData is the template data for the job drawer form.
type FormData struct {
	FormAction   string
	IsEdit       bool
	ID           string
	Name         string
	ClientID     string
	LocationID   string
	Labels       fayna.JobLabels
	CommonLabels any
}

// Deps holds dependencies for job action handlers.
type Deps struct {
	Routes    fayna.JobRoutes
	Labels    fayna.JobLabels
	CreateJob func(ctx context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error)
	ReadJob   func(ctx context.Context, req *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error)
	UpdateJob func(ctx context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error)
	DeleteJob func(ctx context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error)
	ListJobs  func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
}

// NewAddAction creates the job add action (GET = form, POST = create).
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("job-drawer-form", &FormData{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create job
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request

		resp, err := deps.CreateJob(ctx, &jobpb.CreateJobRequest{
			Data: &jobpb.Job{
				Name:       r.FormValue("name"),
				ClientId:   strPtr(r.FormValue("client_id")),
				LocationId: strPtr(r.FormValue("location_id")),
				Status:     enums.JobStatus_JOB_STATUS_DRAFT,
			},
		})
		if err != nil {
			log.Printf("Failed to create job: %v", err)
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

		return fayna.HTMXSuccess("jobs-table")
	})
}

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

			return view.OK("job-drawer-form", &FormData{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				Name:         record.GetName(),
				ClientID:     record.GetClientId(),
				LocationID:   record.GetLocationId(),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — update job
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request

		_, err := deps.UpdateJob(ctx, &jobpb.UpdateJobRequest{
			Data: &jobpb.Job{
				Id:         id,
				Name:       r.FormValue("name"),
				ClientId:   strPtr(r.FormValue("client_id")),
				LocationId: strPtr(r.FormValue("location_id")),
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

// NewDeleteAction creates the job delete action (POST only).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return fayna.HTMXError("ID is required")
		}

		_, err := deps.DeleteJob(ctx, &jobpb.DeleteJobRequest{
			Data: &jobpb.Job{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("jobs-table")
	})
}

// NewBulkDeleteAction creates the job bulk delete action (POST only).
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return fayna.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteJob(ctx, &jobpb.DeleteJobRequest{
				Data: &jobpb.Job{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job %s: %v", id, err)
			}
		}

		return fayna.HTMXSuccess("jobs-table")
	})
}

// NewSetStatusAction creates the job status update action (POST only).
func NewSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		targetStatus := viewCtx.Request.URL.Query().Get("status")

		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
			targetStatus = viewCtx.Request.FormValue("status")
		}
		if id == "" {
			return fayna.HTMXError("ID is required")
		}
		if targetStatus == "" {
			return fayna.HTMXError("Status is required")
		}

		statusEnum := jobStatusToEnum(targetStatus)

		_, err := deps.UpdateJob(ctx, &jobpb.UpdateJobRequest{
			Data: &jobpb.Job{Id: id, Status: statusEnum},
		})
		if err != nil {
			log.Printf("Failed to update job status %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("jobs-table")
	})
}

// NewBulkSetStatusAction creates the job bulk status update action (POST only).
func NewBulkSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		targetStatus := viewCtx.Request.FormValue("target_status")

		if len(ids) == 0 {
			return fayna.HTMXError("No IDs provided")
		}
		if targetStatus == "" {
			return fayna.HTMXError("Target status is required")
		}

		statusEnum := jobStatusToEnum(targetStatus)

		for _, id := range ids {
			if _, err := deps.UpdateJob(ctx, &jobpb.UpdateJobRequest{
				Data: &jobpb.Job{Id: id, Status: statusEnum},
			}); err != nil {
				log.Printf("Failed to update job status %s: %v", id, err)
			}
		}

		return fayna.HTMXSuccess("jobs-table")
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// jobStatusToEnum converts a status string to the protobuf JobStatus enum.
func jobStatusToEnum(status string) enums.JobStatus {
	switch status {
	case "draft":
		return enums.JobStatus_JOB_STATUS_DRAFT
	case "pending":
		return enums.JobStatus_JOB_STATUS_PENDING
	case "active":
		return enums.JobStatus_JOB_STATUS_ACTIVE
	case "paused":
		return enums.JobStatus_JOB_STATUS_PAUSED
	case "completed":
		return enums.JobStatus_JOB_STATUS_COMPLETED
	case "closed":
		return enums.JobStatus_JOB_STATUS_CLOSED
	default:
		return enums.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}
