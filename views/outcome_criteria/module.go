package outcome_criteria

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"

	outcomecriteriadetail "github.com/erniealice/fayna-golang/views/outcome_criteria/detail"
	outcomecriterialist "github.com/erniealice/fayna-golang/views/outcome_criteria/list"
)

// ModuleDeps holds all dependencies for the outcome criteria module.
type ModuleDeps struct {
	Routes       fayna.OutcomeCriteriaRoutes
	Labels       fayna.OutcomeCriteriaLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Outcome criteria CRUD
	CreateOutcomeCriteria func(ctx context.Context, req *criteriapb.CreateOutcomeCriteriaRequest) (*criteriapb.CreateOutcomeCriteriaResponse, error)
	ReadOutcomeCriteria   func(ctx context.Context, req *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)
	UpdateOutcomeCriteria func(ctx context.Context, req *criteriapb.UpdateOutcomeCriteriaRequest) (*criteriapb.UpdateOutcomeCriteriaResponse, error)
	DeleteOutcomeCriteria func(ctx context.Context, req *criteriapb.DeleteOutcomeCriteriaRequest) (*criteriapb.DeleteOutcomeCriteriaResponse, error)
	ListOutcomeCriterias  func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// Module holds all constructed outcome criteria views.
type Module struct {
	routes           fayna.OutcomeCriteriaRoutes
	List             view.View
	Detail           view.View
	TabAction        view.View
	Add              view.View
	Edit             view.View
	Delete           view.View
	BulkDelete       view.View
	AttachmentUpload view.View
	AttachmentDelete view.View
}

// NewModule creates a new outcome criteria module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	detailDeps := &outcomecriteriadetail.DetailViewDeps{
		AttachmentOps: attachment.AttachmentOps{
			UploadFile:       deps.UploadFile,
			ListAttachments:  deps.ListAttachments,
			CreateAttachment: deps.CreateAttachment,
			DeleteAttachment: deps.DeleteAttachment,
			NewAttachmentID:  deps.NewID,
		},
		Routes:              deps.Routes,
		Labels:              deps.Labels,
		CommonLabels:        deps.CommonLabels,
		TableLabels:         deps.TableLabels,
		ReadOutcomeCriteria: deps.ReadOutcomeCriteria,
	}

	return &Module{
		routes: deps.Routes,
		List: outcomecriterialist.NewView(&outcomecriterialist.ListViewDeps{
			Routes:               deps.Routes,
			ListOutcomeCriterias: deps.ListOutcomeCriterias,
			Labels:               deps.Labels,
			CommonLabels:         deps.CommonLabels,
			TableLabels:          deps.TableLabels,
		}),
		Detail:           outcomecriteriadetail.NewView(detailDeps),
		TabAction:        outcomecriteriadetail.NewTabAction(detailDeps),
		Add:              newAddAction(deps),
		Edit:             newEditAction(deps),
		Delete:           newDeleteAction(deps),
		BulkDelete:       newBulkDeleteAction(deps),
		AttachmentUpload: outcomecriteriadetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: outcomecriteriadetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all outcome criteria routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)

	// CRUD actions (GET = drawer form, POST = process submission)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
