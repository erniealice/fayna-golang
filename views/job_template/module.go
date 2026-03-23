package job_template

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"

	jobtemplatedetail "github.com/erniealice/fayna-golang/views/job_template/detail"
	jobtemplatelist "github.com/erniealice/fayna-golang/views/job_template/list"
)

// ModuleDeps holds all dependencies for the job template module.
type ModuleDeps struct {
	Routes       fayna.JobTemplateRoutes
	Labels       fayna.JobTemplateLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Typed job template use case functions
	CreateJobTemplate              func(ctx context.Context, req *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error)
	ReadJobTemplate                func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	UpdateJobTemplate              func(ctx context.Context, req *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error)
	DeleteJobTemplate              func(ctx context.Context, req *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error)
	GetJobTemplateListPageData     func(ctx context.Context, req *jobtemplatepb.GetJobTemplateListPageDataRequest) (*jobtemplatepb.GetJobTemplateListPageDataResponse, error)

	// Phase list (for detail phases tab)
	ListPhasesByJobTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// Module holds all constructed job template views.
type Module struct {
	routes           fayna.JobTemplateRoutes
	List             view.View
	Detail           view.View
	TabAction        view.View
	AttachmentUpload view.View
	AttachmentDelete view.View
}

// NewModule creates the job template module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	detailDeps := &jobtemplatedetail.DetailViewDeps{
		AttachmentOps: attachment.AttachmentOps{
			UploadFile:       deps.UploadFile,
			ListAttachments:  deps.ListAttachments,
			CreateAttachment: deps.CreateAttachment,
			DeleteAttachment: deps.DeleteAttachment,
			NewAttachmentID:  deps.NewID,
		},
		Routes:                  deps.Routes,
		ReadJobTemplate:         deps.ReadJobTemplate,
		ListPhasesByJobTemplate: deps.ListPhasesByJobTemplate,
		Labels:                  deps.Labels,
		CommonLabels:            deps.CommonLabels,
		TableLabels:             deps.TableLabels,
	}

	listView := jobtemplatelist.NewView(&jobtemplatelist.ListViewDeps{
		Routes:                     deps.Routes,
		GetJobTemplateListPageData: deps.GetJobTemplateListPageData,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
	})

	return &Module{
		routes:           deps.Routes,
		List:             listView,
		Detail:           jobtemplatedetail.NewView(detailDeps),
		TabAction:        jobtemplatedetail.NewTabAction(detailDeps),
		AttachmentUpload: jobtemplatedetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: jobtemplatedetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all job template routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)
	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
