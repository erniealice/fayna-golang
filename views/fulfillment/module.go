package fulfillment

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"

	fulfillmentaction "github.com/erniealice/fayna-golang/views/fulfillment/action"
	fulfillmentdashboard "github.com/erniealice/fayna-golang/views/fulfillment/dashboard"
	fulfillmentdetail "github.com/erniealice/fayna-golang/views/fulfillment/detail"
	fulfillmentlist "github.com/erniealice/fayna-golang/views/fulfillment/list"
)

// ModuleDeps holds all dependencies for the fulfillment module.
type ModuleDeps struct {
	Routes       fayna.FulfillmentRoutes
	Labels       fayna.FulfillmentLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Fulfillment CRUD
	CreateFulfillment func(ctx context.Context, req *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error)
	UpdateFulfillment func(ctx context.Context, req *fulfillmentpb.UpdateFulfillmentRequest) (*fulfillmentpb.UpdateFulfillmentResponse, error)
	DeleteFulfillment func(ctx context.Context, req *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error)

	// Fulfillment page data (enriched read operations)
	GetFulfillmentListPageData func(ctx context.Context, req *fulfillmentpb.GetFulfillmentListPageDataRequest) (*fulfillmentpb.GetFulfillmentListPageDataResponse, error)
	GetFulfillmentItemPageData func(ctx context.Context, req *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error)

	// Status transition
	TransitionStatus func(ctx context.Context, req *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error)

	// Return initiation
	CreateFulfillmentReturn func(ctx context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error)

	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	GetFulfillmentDashboardPageData func(ctx context.Context, req *fulfillmentdashboard.Request) (*fulfillmentdashboard.Response, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// Module holds all constructed fulfillment views.
type Module struct {
	routes           fayna.FulfillmentRoutes
	List             view.View
	Detail           view.View
	TabAction        view.View
	Add              view.View
	Edit             view.View
	Delete           view.View
	Transition       view.View
	Return           view.View
	AttachmentUpload view.View
	AttachmentDelete view.View

	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	Dashboard view.View
}

// NewModule creates a new fulfillment module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	detailDeps := &fulfillmentdetail.DetailViewDeps{
		AttachmentOps: attachment.AttachmentOps{
			UploadFile:       deps.UploadFile,
			ListAttachments:  deps.ListAttachments,
			CreateAttachment: deps.CreateAttachment,
			DeleteAttachment: deps.DeleteAttachment,
			NewAttachmentID:  deps.NewID,
		},
		Routes:                     deps.Routes,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
		GetFulfillmentItemPageData: deps.GetFulfillmentItemPageData,
	}

	actionDeps := &fulfillmentaction.Deps{
		Routes:                     deps.Routes,
		Labels:                     deps.Labels,
		CreateFulfillment:          deps.CreateFulfillment,
		UpdateFulfillment:          deps.UpdateFulfillment,
		DeleteFulfillment:          deps.DeleteFulfillment,
		GetFulfillmentItemPageData: deps.GetFulfillmentItemPageData,
		TransitionStatus:           deps.TransitionStatus,
		CreateFulfillmentReturn:    deps.CreateFulfillmentReturn,
	}

	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	dashboardDeps := &fulfillmentdashboard.Deps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		GetDashboardPageData: deps.GetFulfillmentDashboardPageData,
	}

	return &Module{
		routes: deps.Routes,
		List: fulfillmentlist.NewView(&fulfillmentlist.ListViewDeps{
			Routes:                     deps.Routes,
			Labels:                     deps.Labels,
			CommonLabels:               deps.CommonLabels,
			TableLabels:                deps.TableLabels,
			GetFulfillmentListPageData: deps.GetFulfillmentListPageData,
		}),
		Detail:           fulfillmentdetail.NewView(detailDeps),
		TabAction:        fulfillmentdetail.NewTabAction(detailDeps),
		Add:              fulfillmentaction.NewAddAction(actionDeps),
		Edit:             fulfillmentaction.NewEditAction(actionDeps),
		Delete:           fulfillmentaction.NewDeleteAction(actionDeps),
		Transition:       fulfillmentaction.NewTransitionAction(actionDeps),
		Return:           fulfillmentaction.NewReturnAction(actionDeps),
		AttachmentUpload: fulfillmentdetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: fulfillmentdetail.NewAttachmentDeleteAction(detailDeps),
		// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
		Dashboard: fulfillmentdashboard.NewView(dashboardDeps),
	}
}

// RegisterRoutes registers all fulfillment routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	if m.Dashboard != nil && m.routes.DashboardURL != "" {
		r.GET(m.routes.DashboardURL, m.Dashboard)
	}
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	if m.routes.TabActionURL != "" {
		r.GET(m.routes.TabActionURL, m.TabAction)
	}
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.TransitionURL, m.Transition)
	r.POST(m.routes.ReturnURL, m.Return)
	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
