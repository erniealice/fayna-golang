package detail

import (
	"context"
	"fmt"
	"log"

	workrequest "github.com/erniealice/fayna-golang/domain/operation/work_request"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	wrpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/work_request"
)

// DetailViewDeps holds dependencies for the work request detail view.
type DetailViewDeps struct {
	Routes       workrequest.Routes
	Labels       workrequest.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	ReadWorkRequest func(ctx context.Context, req *wrpb.ReadWorkRequestRequest) (*wrpb.ReadWorkRequestResponse, error)
	GetItemPageData func(ctx context.Context, req *wrpb.GetWorkRequestItemPageDataRequest) (*wrpb.GetWorkRequestItemPageDataResponse, error)
}

// PageData holds the data for the work request detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	WorkRequest     map[string]any
	Labels          workrequest.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem
}

// workRequestToMap converts a WorkRequest protobuf to a map[string]any for
// template use.
func workRequestToMap(wr *wrpb.WorkRequest) map[string]any {
	return map[string]any{
		"id":                            wr.GetId(),
		"workspace_id":                  wr.GetWorkspaceId(),
		"client_id":                     wr.GetClientId(),
		"origin":                        wr.GetOrigin().String(),
		"request_number":                wr.GetRequestNumber(),
		"work_request_type_id":          wr.GetWorkRequestTypeId(),
		"status":                        workRequestStatusString(wr.GetStatus()),
		"status_variant":                statusVariant(wr.GetStatus()),
		"title":                         wr.GetTitle(),
		"description":                   wr.GetDescription(),
		"payload_json":                  wr.GetPayloadJson(),
		"requested_by_user_id":          wr.GetRequestedByUserId(),
		"assigned_to_workspace_user_id": wr.GetAssignedToWorkspaceUserId(),
		"priority":                      wr.GetPriority(),
		"sla_target_hours":              wr.GetSlaTargetHours(),
		"active":                        wr.GetActive(),
		"date_created_string":           wr.GetDateCreatedString(),
		"date_modified_string":          wr.GetDateModifiedString(),
	}
}

func workRequestStatusString(s wrpb.WorkRequestStatus) string {
	switch s {
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_NEW:
		return "new"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_SUBMITTED:
		return "submitted"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_IN_REVIEW:
		return "in-review"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_APPROVED:
		return "approved"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_DECLINED:
		return "declined"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_COMPLETED:
		return "completed"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_CANCELLED:
		return "cancelled"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_RETURNED_FOR_INFO:
		return "returned-for-info"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ON_HOLD:
		return "on-hold"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ESCALATED:
		return "escalated"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_PENDING_OVERRIDE:
		return "pending-override"
	default:
		return "new"
	}
}

func statusVariant(s wrpb.WorkRequestStatus) string {
	switch s {
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_NEW,
		wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_SUBMITTED:
		return "default"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_IN_REVIEW:
		return "info"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_APPROVED:
		return "success"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_DECLINED:
		return "danger"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_COMPLETED:
		return "success"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_CANCELLED:
		return "default"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_RETURNED_FOR_INFO:
		return "warning"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ON_HOLD:
		return "secondary"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ESCALATED:
		return "danger"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_PENDING_OVERRIDE:
		return "warning"
	default:
		return "default"
	}
}

// NewView creates the work request detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "read") {
			return view.Forbidden("work_request:read")
		}

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadWorkRequest(ctx, &wrpb.ReadWorkRequestRequest{
			Data: &wrpb.WorkRequest{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read work request %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load work request: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Work request %s not found", id)
			return view.Error(fmt.Errorf("work request not found"))
		}
		wrMap := workRequestToMap(data[0])

		title, _ := wrMap["title"].(string)
		reqNum, _ := wrMap["request_number"].(string)
		headerTitle := title
		if reqNum != "" {
			headerTitle = reqNum + " - " + title
		}

		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    headerTitle,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-inbox",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "work-request-detail-content",
			WorkRequest:     wrMap,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		return view.OK("work-request-detail", pageData)
	})
}

// buildTabItems returns the visible tab strip for the work request detail page.
// 4 tabs: Info (default) / Activity-timeline / Attachments / Messages-link.
func buildTabItems(l workrequest.Labels, id string, routes workrequest.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "timeline", Label: l.Tabs.Timeline, Href: base + "?tab=timeline", HxGet: action + "timeline", Icon: "icon-clock"},
		{Key: "attachments", Label: l.Tabs.Attachments, Href: base + "?tab=attachments", HxGet: action + "attachments", Icon: "icon-paperclip"},
		{Key: "messages", Label: l.Tabs.Messages, Href: base + "?tab=messages", HxGet: action + "messages", Icon: "icon-message-circle"},
	}
}

// NewTabAction creates the tab action view (partial -- returns only the tab
// content). Each tab re-checks work_request:read.
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "read") {
			return view.Forbidden("work_request:read")
		}

		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		resp, err := deps.ReadWorkRequest(ctx, &wrpb.ReadWorkRequestRequest{
			Data: &wrpb.WorkRequest{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read work request %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load work request: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Work request %s not found", id)
			return view.Error(fmt.Errorf("work request not found"))
		}
		wrMap := workRequestToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			WorkRequest: wrMap,
			Labels:      l,
			ActiveTab:   tab,
			TabItems:    buildTabItems(l, id, deps.Routes),
		}

		templateName := "work-request-tab-" + tab
		return view.OK(templateName, pageData)
	})
}
