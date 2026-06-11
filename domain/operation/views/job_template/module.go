package job_template

import (
	"context"

	operation "github.com/erniealice/fayna-golang/domain/operation"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"

	jobtemplateaction "github.com/erniealice/fayna-golang/domain/operation/views/job_template/action"
	jobtemplatedetail "github.com/erniealice/fayna-golang/domain/operation/views/job_template/detail"
	jobtemplatelist "github.com/erniealice/fayna-golang/domain/operation/views/job_template/list"
)

// ModuleDeps holds all dependencies for the job template module.
type ModuleDeps struct {
	Routes       operation.JobTemplateRoutes
	Labels       operation.JobTemplateLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// PhaseRoutes and TaskRoutes supply Edit/Delete/Add URLs for the Phases and Tasks
	// tabs on the JobTemplate detail page. Both are optional — zero-value structs
	// result in tabs with no CTA buttons (read-only view).
	PhaseRoutes operation.JobTemplatePhaseRoutes
	TaskRoutes  operation.JobTemplateTaskRoutes

	// GetInUseIDs checks which job template IDs are referenced by jobs
	// (via job.job_template_id). When non-nil, matched rows have their
	// delete action disabled and are excluded from bulk-delete selections
	// via data-deletable="false".
	GetInUseIDs func(ctx context.Context, ids []string) (map[string]bool, error)

	// Typed job template use case functions
	CreateJobTemplate          func(ctx context.Context, req *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error)
	ReadJobTemplate            func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	UpdateJobTemplate          func(ctx context.Context, req *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error)
	DeleteJobTemplate          func(ctx context.Context, req *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error)
	GetJobTemplateListPageData func(ctx context.Context, req *jobtemplatepb.GetJobTemplateListPageDataRequest) (*jobtemplatepb.GetJobTemplateListPageDataResponse, error)

	// Phase list (for detail phases tab)
	ListPhasesByJobTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)

	// Task + standards list stubs — wired in P6.template-children.
	// Nil is safe: the detail loaders render empty-state panels.
	ListTasksByPhase   func(ctx context.Context, req *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)
	ListCriteriaByTask func(ctx context.Context, req *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// Module holds all constructed job template views.
type Module struct {
	routes           operation.JobTemplateRoutes
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
		PhaseRoutes:             deps.PhaseRoutes,
		TaskRoutes:              deps.TaskRoutes,
		ReadJobTemplate:         deps.ReadJobTemplate,
		ListPhasesByJobTemplate: deps.ListPhasesByJobTemplate,
		ListTasksByPhase:        deps.ListTasksByPhase,
		ListCriteriaByTask:      deps.ListCriteriaByTask,
		Labels:                  deps.Labels,
		CommonLabels:            deps.CommonLabels,
		TableLabels:             deps.TableLabels,
	}

	listView := jobtemplatelist.NewView(&jobtemplatelist.ListViewDeps{
		Routes:                     deps.Routes,
		GetJobTemplateListPageData: deps.GetJobTemplateListPageData,
		GetInUseIDs:                deps.GetInUseIDs,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
	})

	actionDeps := &jobtemplateaction.Deps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CreateJobTemplate: deps.CreateJobTemplate,
		ReadJobTemplate:   deps.ReadJobTemplate,
		UpdateJobTemplate: deps.UpdateJobTemplate,
		DeleteJobTemplate: deps.DeleteJobTemplate,
	}

	return &Module{
		routes:           deps.Routes,
		List:             listView,
		Detail:           jobtemplatedetail.NewView(detailDeps),
		TabAction:        jobtemplatedetail.NewTabAction(detailDeps),
		Add:              jobtemplateaction.NewAddAction(actionDeps),
		Edit:             jobtemplateaction.NewEditAction(actionDeps),
		Delete:           jobtemplateaction.NewDeleteAction(actionDeps),
		BulkDelete:       jobtemplateaction.NewBulkDeleteAction(actionDeps),
		AttachmentUpload: jobtemplatedetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: jobtemplatedetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all job template routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)
	// CRUD actions
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	if m.routes.BulkDeleteURL != "" {
		r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	}
	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
