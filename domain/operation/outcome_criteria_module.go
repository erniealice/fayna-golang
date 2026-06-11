package operation

import (
	"context"

	outcomecritetriapkg "github.com/erniealice/fayna-golang/domain/operation/outcome_criteria"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"

	outcomecriteriadetail "github.com/erniealice/fayna-golang/domain/operation/outcome_criteria/detail"
	outcomecriterialist "github.com/erniealice/fayna-golang/domain/operation/outcome_criteria/list"
)

// OutcomeCriteriaModuleDeps holds all dependencies for the outcome criteria module.
type OutcomeCriteriaModuleDeps struct {
	Routes       outcomecritetriapkg.Routes
	Labels       outcomecritetriapkg.Labels
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

// OutcomeCriteriaModule holds all constructed outcome criteria views.
type OutcomeCriteriaModule struct {
	routes           outcomecritetriapkg.Routes
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

// NewOutcomeCriteriaModule creates a new outcome criteria module with all views wired.
func NewOutcomeCriteriaModule(deps *OutcomeCriteriaModuleDeps) *OutcomeCriteriaModule {
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

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &outcomecritetriapkg.ModuleDeps{
		Routes:                deps.Routes,
		Labels:                deps.Labels,
		CreateOutcomeCriteria: deps.CreateOutcomeCriteria,
		ReadOutcomeCriteria:   deps.ReadOutcomeCriteria,
		UpdateOutcomeCriteria: deps.UpdateOutcomeCriteria,
		DeleteOutcomeCriteria: deps.DeleteOutcomeCriteria,
		ListOutcomeCriterias:  deps.ListOutcomeCriterias,
		UploadFile:            deps.UploadFile,
		ListAttachments:       deps.ListAttachments,
		CreateAttachment:      deps.CreateAttachment,
		DeleteAttachment:      deps.DeleteAttachment,
		NewID:                 deps.NewID,
	}

	return &OutcomeCriteriaModule{
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
		Add:              outcomecritetriapkg.NewAddAction(entityDeps),
		Edit:             outcomecritetriapkg.NewEditAction(entityDeps),
		Delete:           outcomecritetriapkg.NewDeleteAction(entityDeps),
		BulkDelete:       outcomecritetriapkg.NewBulkDeleteAction(entityDeps),
		AttachmentUpload: outcomecriteriadetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: outcomecriteriadetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all outcome criteria routes.
func (m *OutcomeCriteriaModule) RegisterRoutes(r view.RouteRegistrar) {
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
