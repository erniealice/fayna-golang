package operation

import (
	"context"

	jobtemplatepkg "github.com/erniealice/fayna-golang/domain/operation/job_template"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplaterelationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	productpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product"

	jobtemplateaction "github.com/erniealice/fayna-golang/domain/operation/job_template/action"
	jobtemplatedetail "github.com/erniealice/fayna-golang/domain/operation/job_template/detail"
	jobtemplatelist "github.com/erniealice/fayna-golang/domain/operation/job_template/list"
	jtrpkg "github.com/erniealice/fayna-golang/domain/operation/job_template_relation"
	ttcpkg "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"
)

// JobTemplateModuleDeps holds all dependencies for the job template module.
type JobTemplateModuleDeps struct {
	Routes       jobtemplatepkg.Routes
	Labels       jobtemplatepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// PhaseRoutes and TaskRoutes supply Edit/Delete/Add URLs for the Phases and Tasks
	// tabs on the JobTemplate detail page. Both are optional — zero-value structs
	// result in tabs with no CTA buttons (read-only view).
	PhaseRoutes JobTemplatePhaseRoutes
	TaskRoutes  JobTemplateTaskRoutes
	// CriteriaRoutes supplies Add/Delete URLs for the Standards tab's
	// "+ Add Standard" CTA and per-row remove actions. Optional — a zero-value
	// struct results in a read-only Standards tab.
	CriteriaRoutes ttcpkg.Routes
	// RelationRoutes supplies Add/Delete URLs for the Spawn Graph tab's
	// "+ Add Relation" CTA and per-row remove actions. Optional — a zero-value
	// struct results in a read-only Spawn Graph tab.
	RelationRoutes jtrpkg.Routes

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

	// ListJobCategories / ListProducts populate the Category / Output Product
	// pickers on the drawer form. Both optional — nil-safe (empty picker).
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
	ListProducts      func(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error)

	// Phase list (for detail phases tab)
	ListPhasesByJobTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)

	// Task + standards list stubs — wired in P6.template-children.
	// Nil is safe: the detail loaders render empty-state panels.
	ListTasksByPhase   func(ctx context.Context, req *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)
	ListCriteriaByTask func(ctx context.Context, req *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error)
	// ListRelationsByParent backs the Spawn Graph tab's roster table
	// (Q-TPL W4, NEW). Nil is safe: the tab loader renders an empty-state
	// panel.
	ListRelationsByParent func(ctx context.Context, req *jobtemplaterelationpb.ListJobTemplateRelationsByParentRequest) (*jobtemplaterelationpb.ListJobTemplateRelationsByParentResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// JobTemplateModule holds all constructed job template views.
type JobTemplateModule struct {
	routes           jobtemplatepkg.Routes
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

// NewJobTemplateModule creates the job template module with all views wired.
func NewJobTemplateModule(deps *JobTemplateModuleDeps) *JobTemplateModule {
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
		CriteriaRoutes:          deps.CriteriaRoutes,
		RelationRoutes:          deps.RelationRoutes,
		ReadJobTemplate:         deps.ReadJobTemplate,
		ListPhasesByJobTemplate: deps.ListPhasesByJobTemplate,
		ListTasksByPhase:        deps.ListTasksByPhase,
		ListCriteriaByTask:      deps.ListCriteriaByTask,
		ListRelationsByParent:   deps.ListRelationsByParent,
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
		ListJobCategories: deps.ListJobCategories,
		ListProducts:      deps.ListProducts,
	}

	return &JobTemplateModule{
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
func (m *JobTemplateModule) RegisterRoutes(r view.RouteRegistrar) {
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
