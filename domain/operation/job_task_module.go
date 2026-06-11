package operation

import (
	"context"

	jobtaskpkg "github.com/erniealice/fayna-golang/domain/operation/job_task"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"

	jobtaskaction "github.com/erniealice/fayna-golang/domain/operation/job_task/action"
	jobtaskdetail "github.com/erniealice/fayna-golang/domain/operation/job_task/detail"
	jobtasklist "github.com/erniealice/fayna-golang/domain/operation/job_task/list"
)

// JobTaskModuleDeps holds all dependencies for the job_task module.
type JobTaskModuleDeps struct {
	Routes       jobtaskpkg.Routes
	Labels       jobtaskpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// GetInUseIDs blocks deletion of tasks that are referenced by job_activity rows.
	GetInUseIDs func(ctx context.Context, ids []string) (map[string]bool, error)

	// Job task CRUD
	CreateJobTask func(ctx context.Context, req *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error)
	ReadJobTask   func(ctx context.Context, req *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error)
	UpdateJobTask func(ctx context.Context, req *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error)
	DeleteJobTask func(ctx context.Context, req *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error)
	ListJobTasks  func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)

	// Activities tab — filtered by job_task_id.
	ListJobActivities func(ctx context.Context, req *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)

	// Attachment operations (optional — nil = attachments tab hidden/empty)
	attachment.AttachmentOps

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// JobTaskModule holds all constructed job_task views.
type JobTaskModule struct {
	routes           jobtaskpkg.Routes
	List             view.View
	Detail           view.View
	TabAction        view.View
	Add              view.View
	Edit             view.View
	Delete           view.View
	BulkDelete       view.View
	SetStatus        view.View
	BulkSetStatus    view.View
	AttachmentUpload view.View
	AttachmentDelete view.View
}

// NewJobTaskModule creates a new job_task module with all views wired.
func NewJobTaskModule(deps *JobTaskModuleDeps) *JobTaskModule {
	actionDeps := &jobtaskaction.Deps{
		Routes:                deps.Routes,
		Labels:                deps.Labels,
		CreateJobTask:         deps.CreateJobTask,
		ReadJobTask:           deps.ReadJobTask,
		UpdateJobTask:         deps.UpdateJobTask,
		DeleteJobTask:         deps.DeleteJobTask,
		ListJobTasks:          deps.ListJobTasks,
		StaffSearchURL:        deps.Routes.StaffSearchURL,
		ResourceSearchURL:     deps.Routes.ResourceSearchURL,
		TemplateTaskSearchURL: deps.Routes.TemplateTaskSearchURL,
	}

	detailDeps := &jobtaskdetail.DetailViewDeps{
		AttachmentOps:     deps.AttachmentOps,
		AuditOps:          deps.AuditOps,
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CommonLabels:      deps.CommonLabels,
		TableLabels:       deps.TableLabels,
		ReadJobTask:       deps.ReadJobTask,
		ListJobActivities: deps.ListJobActivities,
	}

	return &JobTaskModule{
		routes: deps.Routes,
		List: jobtasklist.NewView(&jobtasklist.ListViewDeps{
			Routes:       deps.Routes,
			ListJobTasks: deps.ListJobTasks,
			GetInUseIDs:  deps.GetInUseIDs,
			Labels:       deps.Labels,
			CommonLabels: deps.CommonLabels,
			TableLabels:  deps.TableLabels,
		}),
		Detail:           jobtaskdetail.NewView(detailDeps),
		TabAction:        jobtaskdetail.NewTabAction(detailDeps),
		Add:              jobtaskaction.NewAddAction(actionDeps),
		Edit:             jobtaskaction.NewEditAction(actionDeps),
		Delete:           jobtaskaction.NewDeleteAction(actionDeps),
		BulkDelete:       jobtaskaction.NewBulkDeleteAction(actionDeps),
		SetStatus:        jobtaskaction.NewSetStatusAction(actionDeps),
		BulkSetStatus:    jobtaskaction.NewBulkSetStatusAction(actionDeps),
		AttachmentUpload: jobtaskdetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: jobtaskdetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all job_task routes.
func (m *JobTaskModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)
	r.POST(m.routes.TabActionURL, m.TabAction)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	r.POST(m.routes.SetStatusURL, m.SetStatus)
	r.POST(m.routes.BulkSetStatusURL, m.BulkSetStatus)

	// Attachments (optional)
	if m.AttachmentUpload != nil && m.routes.AttachmentUploadURL != "" {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
