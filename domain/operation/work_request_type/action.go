package work_request_type

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	wrtpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/work_request_type"
)

// ActionDeps holds dependencies for work request type action handlers.
type ActionDeps struct {
	Routes                Routes
	Labels                Labels
	CreateWorkRequestType func(ctx context.Context, req *wrtpb.CreateWorkRequestTypeRequest) (*wrtpb.CreateWorkRequestTypeResponse, error)
	ReadWorkRequestType   func(ctx context.Context, req *wrtpb.ReadWorkRequestTypeRequest) (*wrtpb.ReadWorkRequestTypeResponse, error)
	UpdateWorkRequestType func(ctx context.Context, req *wrtpb.UpdateWorkRequestTypeRequest) (*wrtpb.UpdateWorkRequestTypeResponse, error)
	ListWorkRequestTypes  func(ctx context.Context, req *wrtpb.ListWorkRequestTypesRequest) (*wrtpb.ListWorkRequestTypesResponse, error)
}

// NewAddAction creates the work request type add action (GET = form, POST = create).
// Layer-2 GET (permission check before drawer HTML) + POST (permission check
// before data access).
func NewAddAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request_type", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("work-request-type-drawer-form", map[string]any{
				"FormAction": deps.Routes.AddURL,
				"Labels":     deps.Labels,
				"IsEdit":     false,
			})
		}

		// POST -- create work request type
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		r := viewCtx.Request
		wrt := &wrtpb.WorkRequestType{
			Code:           r.FormValue("code"),
			LabelKey:       r.FormValue("label_key"),
			DescriptionKey: r.FormValue("description_key"),
			IconKey:        r.FormValue("icon_key"),
		}

		// Category
		switch r.FormValue("category") {
		case "1":
			wrt.Category = wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_PERSON_SCOPED
		case "2":
			wrt.Category = wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_ACCOUNT_SCOPED
		}

		// Requires resource
		wrt.RequiresResource = r.FormValue("requires_resource") == "on" || r.FormValue("requires_resource") == "true"

		// Default SLA hours
		if sla := r.FormValue("default_sla_hours"); sla != "" {
			if v, err := strconv.ParseInt(sla, 10, 64); err == nil {
				wrt.DefaultSlaHours = v
			}
		}

		// Sort order
		if so := r.FormValue("sort_order"); so != "" {
			if v, err := strconv.ParseInt(so, 10, 32); err == nil {
				wrt.SortOrder = int32(v)
			}
		}

		resp, err := deps.CreateWorkRequestType(ctx, &wrtpb.CreateWorkRequestTypeRequest{
			Data: wrt,
		})
		if err != nil {
			log.Printf("Failed to create work request type: %v", err)
			return view.HTMXError(err.Error())
		}

		_ = resp
		return view.HTMXSuccess("work-request-type-list-table")
	})
}

// NewEditAction creates the work request type edit action (GET = form, POST = update).
// Layer-2 GET + POST gated by work_request_type:update.
func NewEditAction(deps *ActionDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request_type", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			resp, err := deps.ReadWorkRequestType(ctx, &wrtpb.ReadWorkRequestTypeRequest{
				Data: &wrtpb.WorkRequestType{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read work request type %s for edit: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			data := resp.GetData()
			if len(data) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}

			wrt := data[0]
			return view.OK("work-request-type-drawer-form", map[string]any{
				"FormAction":      route.ResolveURL(deps.Routes.EditURL, "id", id),
				"Labels":          deps.Labels,
				"IsEdit":          true,
				"WorkRequestType": wrt,
			})
		}

		// POST -- update work request type
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidForm)
		}

		r := viewCtx.Request
		wrt := &wrtpb.WorkRequestType{
			Id:             id,
			Code:           r.FormValue("code"),
			LabelKey:       r.FormValue("label_key"),
			DescriptionKey: r.FormValue("description_key"),
			IconKey:        r.FormValue("icon_key"),
		}

		// Category
		switch r.FormValue("category") {
		case "1":
			wrt.Category = wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_PERSON_SCOPED
		case "2":
			wrt.Category = wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_ACCOUNT_SCOPED
		}

		// Requires resource
		wrt.RequiresResource = r.FormValue("requires_resource") == "on" || r.FormValue("requires_resource") == "true"

		// Default SLA hours
		if sla := r.FormValue("default_sla_hours"); sla != "" {
			if v, err := strconv.ParseInt(sla, 10, 64); err == nil {
				wrt.DefaultSlaHours = v
			}
		}

		// Sort order
		if so := r.FormValue("sort_order"); so != "" {
			if v, err := strconv.ParseInt(so, 10, 32); err == nil {
				wrt.SortOrder = int32(v)
			}
		}

		// Status
		switch r.FormValue("status") {
		case "1":
			wrt.Status = wrtpb.WorkRequestTypeStatus_WORK_REQUEST_TYPE_STATUS_ACTIVE
			wrt.Active = true
		case "2":
			wrt.Status = wrtpb.WorkRequestTypeStatus_WORK_REQUEST_TYPE_STATUS_ARCHIVED
			wrt.Active = false
		}

		_, err := deps.UpdateWorkRequestType(ctx, &wrtpb.UpdateWorkRequestTypeRequest{
			Data: wrt,
		})
		if err != nil {
			log.Printf("Failed to update work request type %s: %v", id, err)
			return view.HTMXError(fmt.Errorf("failed to update: %w", err).Error())
		}

		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger": `{"formSuccess":true}`,
			},
		}
	})
}

