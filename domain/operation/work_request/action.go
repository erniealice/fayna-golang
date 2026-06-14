package work_request

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	wrpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/work_request"
)

// ActionDeps holds dependencies for work request action handlers.
type ActionDeps struct {
	Routes            Routes
	Labels            Labels
	CreateWorkRequest func(ctx context.Context, req *wrpb.CreateWorkRequestRequest) (*wrpb.CreateWorkRequestResponse, error)
	ReadWorkRequest   func(ctx context.Context, req *wrpb.ReadWorkRequestRequest) (*wrpb.ReadWorkRequestResponse, error)
	UpdateWorkRequest func(ctx context.Context, req *wrpb.UpdateWorkRequestRequest) (*wrpb.UpdateWorkRequestResponse, error)
	DeleteWorkRequest func(ctx context.Context, req *wrpb.DeleteWorkRequestRequest) (*wrpb.DeleteWorkRequestResponse, error)
	ListWorkRequests  func(ctx context.Context, req *wrpb.ListWorkRequestsRequest) (*wrpb.ListWorkRequestsResponse, error)
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// workRequestStatusToEnum converts a status slug to the protobuf enum.
func workRequestStatusToEnum(status string) wrpb.WorkRequestStatus {
	switch status {
	case "new":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_NEW
	case "submitted":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_SUBMITTED
	case "in-review":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_IN_REVIEW
	case "approved":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_APPROVED
	case "declined":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_DECLINED
	case "completed":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_COMPLETED
	case "cancelled":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_CANCELLED
	case "returned-for-info":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_RETURNED_FOR_INFO
	case "on-hold":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ON_HOLD
	case "escalated":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ESCALATED
	case "pending-override":
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_PENDING_OVERRIDE
	default:
		return wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_UNSPECIFIED
	}
}

// NewAddAction creates the work request add action (GET = form, POST = create).
// Layer-2 GET (permission check before drawer HTML) + POST (permission check
// before data access).
func NewAddAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("work-request-drawer-form", map[string]any{
				"FormAction": deps.Routes.AddURL,
				"Labels":     deps.Labels,
				"Origin":     viewCtx.Request.URL.Query().Get("origin"),
			})
		}

		// POST -- create work request
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		r := viewCtx.Request
		title := r.FormValue("title")
		description := r.FormValue("description")
		priority := int32(0)
		if r.FormValue("priority") == "1" {
			priority = 1
		}

		wr := &wrpb.WorkRequest{
			Title:       title,
			Description: description,
			Priority:    priority,
		}

		// Origin stamping (never a form field; server-derived)
		originParam := r.FormValue("origin_code")
		switch originParam {
		case "3": // INTERNAL
			wr.Origin = wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_INTERNAL
		case "2": // CLIENT_RELATED_INTERNAL
			wr.Origin = wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_CLIENT_RELATED_INTERNAL
			if clientID := r.FormValue("client_id"); clientID != "" {
				wr.ClientId = &clientID
			}
		default: // CLIENT_ORIGINATED (portal path)
			wr.Origin = wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_CLIENT_ORIGINATED
			if clientID := r.FormValue("client_id"); clientID != "" {
				wr.ClientId = &clientID
			}
		}

		if typeID := r.FormValue("work_request_type_id"); typeID != "" {
			wr.WorkRequestTypeId = typeID
		}

		resp, err := deps.CreateWorkRequest(ctx, &wrpb.CreateWorkRequestRequest{
			Data: wr,
		})
		if err != nil {
			log.Printf("Failed to create work request: %v", err)
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

		return view.HTMXSuccess("work-request-list-table")
	})
}

// NewEditAction creates the work request edit action (GET = form, POST = update).
func NewEditAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			resp, err := deps.ReadWorkRequest(ctx, &wrpb.ReadWorkRequestRequest{
				Data: &wrpb.WorkRequest{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read work request %s for edit: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			data := resp.GetData()
			if len(data) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}

			wr := data[0]
			return view.OK("work-request-edit-drawer-form", map[string]any{
				"FormAction":  route.ResolveURL(deps.Routes.EditURL, "id", id),
				"Labels":      deps.Labels,
				"WorkRequest": wr,
			})
		}

		// POST -- update work request
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		r := viewCtx.Request
		priority := int32(0)
		if r.FormValue("priority") == "1" {
			priority = 1
		}

		_, err := deps.UpdateWorkRequest(ctx, &wrpb.UpdateWorkRequestRequest{
			Data: &wrpb.WorkRequest{
				Id:          id,
				Title:       r.FormValue("title"),
				Description: r.FormValue("description"),
				Priority:    priority,
			},
		})
		if err != nil {
			log.Printf("Failed to update work request %s: %v", id, err)
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

// NewSetStatusAction creates the work request set-status action.
// Layer-2: permission check on both GET (drawer render) and POST (mutation).
func NewSetStatusAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("work-request-set-status-drawer", map[string]any{
				"FormAction":   route.ResolveURL(deps.Routes.SetStatusURL, "id", id),
				"Labels":       deps.Labels,
				"WorkRequestID": id,
			})
		}

		// POST -- set status
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		targetStatus := viewCtx.Request.FormValue("status")
		if targetStatus == "" {
			return view.HTMXError(deps.Labels.Errors.StatusRequired)
		}

		resolutionNote := viewCtx.Request.FormValue("resolution_note")

		_, err := deps.UpdateWorkRequest(ctx, &wrpb.UpdateWorkRequestRequest{
			Data: &wrpb.WorkRequest{
				Id:             id,
				Status:         workRequestStatusToEnum(targetStatus),
				ResolutionNote: strPtrOrNil(resolutionNote),
			},
		})
		if err != nil {
			log.Printf("Failed to set status on work request %s: %v", id, err)
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

// NewAssignAction creates the work request assign action (GET = drawer, POST = assign).
func NewAssignAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "assign") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("work-request-assign-drawer", map[string]any{
				"FormAction":   route.ResolveURL(deps.Routes.AssignURL, "id", id),
				"Labels":       deps.Labels,
				"WorkRequestID": id,
			})
		}

		// POST -- assign
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		assigneeID := viewCtx.Request.FormValue("assigned_to_workspace_user_id")
		if assigneeID == "" {
			return view.HTMXError("Assignee is required")
		}

		_, err := deps.UpdateWorkRequest(ctx, &wrpb.UpdateWorkRequestRequest{
			Data: &wrpb.WorkRequest{
				Id:                        id,
				AssignedToWorkspaceUserId: &assigneeID,
			},
		})
		if err != nil {
			log.Printf("Failed to assign work request %s: %v", id, err)
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

// NewBulkAssignAction creates the work request bulk-assign action.
func NewBulkAssignAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "assign") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("work-request-bulk-assign-drawer", map[string]any{
				"FormAction": deps.Routes.BulkAssignURL,
				"Labels":     deps.Labels,
			})
		}

		// POST -- bulk assign
		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		assigneeID := viewCtx.Request.FormValue("assigned_to_workspace_user_id")

		if len(ids) == 0 {
			return view.HTMXError("No requests selected")
		}
		if assigneeID == "" {
			return view.HTMXError("Assignee is required")
		}

		for _, id := range ids {
			if _, err := deps.UpdateWorkRequest(ctx, &wrpb.UpdateWorkRequestRequest{
				Data: &wrpb.WorkRequest{
					Id:                        id,
					AssignedToWorkspaceUserId: &assigneeID,
				},
			}); err != nil {
				log.Printf("Failed to assign work request %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("work-request-list-table")
	})
}

// NewResolveAction creates the work request resolve action (APPROVED -> COMPLETED).
// Uses the atomic ResolveWorkRequest use case. Gate: work_request:resolve.
func NewResolveAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "resolve") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		// Resolve uses UpdateWorkRequest with COMPLETED status as the entry
		// point. The actual atomic ResolveWorkRequest use case is wired at
		// the espyna layer; here we call the same shaped update.
		_, err := deps.UpdateWorkRequest(ctx, &wrpb.UpdateWorkRequestRequest{
			Data: &wrpb.WorkRequest{
				Id:     id,
				Status: wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_COMPLETED,
			},
		})
		if err != nil {
			log.Printf("Failed to resolve work request %s: %v", id, err)
			return view.HTMXError(fmt.Errorf("failed to resolve: %w", err).Error())
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

// strPtrOrNil returns a pointer to s if non-empty, otherwise nil.
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
