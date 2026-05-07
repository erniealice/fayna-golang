package task_outcome

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	outcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"

	taskoutcomedetail "github.com/erniealice/fayna-golang/views/task_outcome/detail"
	taskoutcomelist "github.com/erniealice/fayna-golang/views/task_outcome/list"
)

// ModuleDeps holds all dependencies for the task outcome module.
type ModuleDeps struct {
	Routes       fayna.TaskOutcomeRoutes
	Labels       fayna.TaskOutcomeLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Task outcome CRUD
	CreateTaskOutcome func(ctx context.Context, req *outcomepb.CreateTaskOutcomeRequest) (*outcomepb.CreateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(ctx context.Context, req *outcomepb.ReadTaskOutcomeRequest) (*outcomepb.ReadTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *outcomepb.UpdateTaskOutcomeRequest) (*outcomepb.UpdateTaskOutcomeResponse, error)
	DeleteTaskOutcome func(ctx context.Context, req *outcomepb.DeleteTaskOutcomeRequest) (*outcomepb.DeleteTaskOutcomeResponse, error)
	ListTaskOutcomes  func(ctx context.Context, req *outcomepb.ListTaskOutcomesRequest) (*outcomepb.ListTaskOutcomesResponse, error)

	// Outcome criteria read (for linking criteria details)
	ReadOutcomeCriteria func(ctx context.Context, req *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// Module holds all constructed task outcome views.
type Module struct {
	routes           fayna.TaskOutcomeRoutes
	List             view.View
	Detail           view.View
	TabAction        view.View
	Add              view.View
	Edit             view.View
	Delete           view.View
	AttachmentUpload view.View
	AttachmentDelete view.View
}

// NewModule creates a new task outcome module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	detailDeps := &taskoutcomedetail.DetailViewDeps{
		AttachmentOps: attachment.AttachmentOps{
			UploadFile:       deps.UploadFile,
			ListAttachments:  deps.ListAttachments,
			CreateAttachment: deps.CreateAttachment,
			DeleteAttachment: deps.DeleteAttachment,
			NewAttachmentID:  deps.NewID,
		},
		Routes:          deps.Routes,
		Labels:          deps.Labels,
		CommonLabels:    deps.CommonLabels,
		ReadTaskOutcome: deps.ReadTaskOutcome,
	}

	return &Module{
		routes: deps.Routes,
		List: taskoutcomelist.NewView(&taskoutcomelist.ListViewDeps{
			Routes:           deps.Routes,
			ListTaskOutcomes: deps.ListTaskOutcomes,
			Labels:           deps.Labels,
			CommonLabels:     deps.CommonLabels,
			TableLabels:      deps.TableLabels,
		}),
		Detail:           taskoutcomedetail.NewView(detailDeps),
		TabAction:        taskoutcomedetail.NewTabAction(detailDeps),
		Add:              newAddAction(deps),
		Edit:             newEditAction(deps),
		Delete:           newDeleteAction(deps),
		AttachmentUpload: taskoutcomedetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: taskoutcomedetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all task outcome routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	if m.routes.TabActionURL != "" {
		r.GET(m.routes.TabActionURL, m.TabAction)
	}

	// CRUD actions (GET = recording form, POST = process submission)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
