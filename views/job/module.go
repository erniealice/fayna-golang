package job

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobsettlementpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_settlement"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"

	jobaction "github.com/erniealice/fayna-golang/views/job/action"
	jobdetail "github.com/erniealice/fayna-golang/views/job/detail"
	joblist "github.com/erniealice/fayna-golang/views/job/list"
)

// ModuleDeps holds all dependencies for the job module.
type ModuleDeps struct {
	Routes       fayna.JobRoutes
	Labels       fayna.JobLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Job CRUD
	CreateJob func(ctx context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error)
	ReadJob   func(ctx context.Context, req *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error)
	UpdateJob func(ctx context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error)
	DeleteJob func(ctx context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error)
	ListJobs  func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)

	// Job phase operations
	ListJobPhases func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)

	// Job task operations
	ListJobTasks func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)

	// Job activity operations
	ListJobActivities func(ctx context.Context, req *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)

	// Job settlement operations
	ListJobSettlements func(ctx context.Context, req *jobsettlementpb.ListJobSettlementsRequest) (*jobsettlementpb.ListJobSettlementsResponse, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// Module holds all constructed job views.
type Module struct {
	routes           fayna.JobRoutes
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

// NewModule creates a new job module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	detailDeps := &jobdetail.DetailViewDeps{
		AttachmentOps: attachment.AttachmentOps{
			UploadFile:       deps.UploadFile,
			ListAttachments:  deps.ListAttachments,
			CreateAttachment: deps.CreateAttachment,
			DeleteAttachment: deps.DeleteAttachment,
			NewAttachmentID:  deps.NewID,
		},
		Routes:             deps.Routes,
		Labels:             deps.Labels,
		CommonLabels:       deps.CommonLabels,
		TableLabels:        deps.TableLabels,
		ReadJob:            deps.ReadJob,
		ListJobPhases:      deps.ListJobPhases,
		ListJobTasks:       deps.ListJobTasks,
		ListJobActivities:  deps.ListJobActivities,
		ListJobSettlements: deps.ListJobSettlements,
	}

	actionDeps := &jobaction.Deps{
		Routes:    deps.Routes,
		Labels:    deps.Labels,
		CreateJob: deps.CreateJob,
		ReadJob:   deps.ReadJob,
		UpdateJob: deps.UpdateJob,
		DeleteJob: deps.DeleteJob,
		ListJobs:  deps.ListJobs,
	}

	return &Module{
		routes: deps.Routes,
		List: joblist.NewView(&joblist.ListViewDeps{
			Routes:       deps.Routes,
			ListJobs:     deps.ListJobs,
			Labels:       deps.Labels,
			CommonLabels: deps.CommonLabels,
			TableLabels:  deps.TableLabels,
		}),
		Detail:           jobdetail.NewView(detailDeps),
		TabAction:        jobdetail.NewTabAction(detailDeps),
		Add:              jobaction.NewAddAction(actionDeps),
		Edit:             jobaction.NewEditAction(actionDeps),
		Delete:           jobaction.NewDeleteAction(actionDeps),
		BulkDelete:       jobaction.NewBulkDeleteAction(actionDeps),
		SetStatus:        jobaction.NewSetStatusAction(actionDeps),
		BulkSetStatus:    jobaction.NewBulkSetStatusAction(actionDeps),
		AttachmentUpload: jobdetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: jobdetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all job routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	r.POST(m.routes.SetStatusURL, m.SetStatus)
	r.POST(m.routes.BulkSetStatusURL, m.BulkSetStatus)
	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
