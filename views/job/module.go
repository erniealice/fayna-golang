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
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatetaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	subscriptionpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription"

	jobaction "github.com/erniealice/fayna-golang/views/job/action"
	jobdashboard "github.com/erniealice/fayna-golang/views/job/dashboard"
	jobdetail "github.com/erniealice/fayna-golang/views/job/detail"
	joblist "github.com/erniealice/fayna-golang/views/job/list"
)

// ModuleDeps holds all dependencies for the job module.
type ModuleDeps struct {
	Routes       fayna.JobRoutes
	Labels       fayna.JobLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// GetInUseIDs checks which job IDs are referenced by child tables
	// (job_activity, job_phase, job_settlement). When non-nil, matched rows
	// have their delete action disabled and are excluded from bulk-delete
	// selections via data-deletable="false".
	GetInUseIDs func(ctx context.Context, ids []string) (map[string]bool, error)

	// 2026-04-29 milestone-billing plan §5/§6 — Activities tab on Job detail
	// renders an "+ Add Activity" CTA + per-row Edit CTA targeting the
	// JobActivity drawer. Both sets are optional; empty = legacy behaviour
	// (table-card only, no CTAs).
	JobActivityRoutes fayna.JobActivityRoutes
	JobActivityLabels fayna.JobActivityLabels

	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	JobTemplateRoutes       fayna.JobTemplateRoutes
	GetJobDashboardPageData func(ctx context.Context, req *jobdashboard.Request) (*jobdashboard.Response, error)

	// Job CRUD
	CreateJob func(ctx context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error)
	ReadJob   func(ctx context.Context, req *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error)
	UpdateJob func(ctx context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error)
	DeleteJob func(ctx context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error)
	ListJobs  func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)

	// Job phase operations (list only — CRUD is now owned by the job_phase module)
	ListJobPhases func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)

	// Job task operations
	ListJobTasks  func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	UpdateJobTask func(ctx context.Context, req *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error)

	// Job activity operations
	ListJobActivities func(ctx context.Context, req *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)

	// Job settlement operations
	ListJobSettlements func(ctx context.Context, req *jobsettlementpb.ListJobSettlementsRequest) (*jobsettlementpb.ListJobSettlementsResponse, error)

	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 — subscription
	// origin breadcrumb deps. Both nil/empty = breadcrumb hidden.
	ReadSubscription      func(ctx context.Context, req *subscriptionpb.ReadSubscriptionRequest) (*subscriptionpb.ReadSubscriptionResponse, error)
	SubscriptionDetailURL string

	// Budget tab — reads the JobTemplate-derived phase/task hour plan for
	// the budget tab. All three deps are optional; nil = budget tab renders
	// empty state. Wave 3 will replace these with a JobInputPlan reader.
	// TODO(composition): wire in packages/fayna-golang/block/wiring.go.
	ReadJobTemplate                 func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	ListJobTemplatePhasesByTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)
	ListJobTemplateTasksByPhase     func(ctx context.Context, req *jobtemplatetaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplatetaskpb.ListJobTemplateTasksByPhaseResponse, error)

	// Actuals tab — aggregated cost rollup from GetJobActivityRollup.
	// Optional; nil = actuals tab renders empty state.
	// TODO(composition): wire in packages/fayna-golang/block/wiring.go.
	GetJobActivityRollup func(ctx context.Context, req *jobactivitypb.GetJobActivityRollupRequest) (*jobactivitypb.GetJobActivityRollupResponse, error)

	// Search endpoints for the job drawer client + location pickers.
	// Set from fayna.JobRoutes by the fayna block (JobClientSearchURL / JobLocationSearchURL).
	ClientSearchURL   string
	LocationSearchURL string

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
	AssignTask       view.View

	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	Dashboard view.View
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
		Routes:       deps.Routes,
		Labels:       deps.Labels,
		CommonLabels: deps.CommonLabels,
		TableLabels:  deps.TableLabels,
		// 2026-04-29 milestone-billing plan §5/§6.
		JobActivityRoutes:  deps.JobActivityRoutes,
		JobActivityLabels:  deps.JobActivityLabels,
		ReadJob:            deps.ReadJob,
		ListJobPhases:      deps.ListJobPhases,
		ListJobTasks:       deps.ListJobTasks,
		UpdateJobTask:      deps.UpdateJobTask,
		ListJobActivities:  deps.ListJobActivities,
		ListJobSettlements: deps.ListJobSettlements,
		// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4.
		ReadSubscription:      deps.ReadSubscription,
		SubscriptionDetailURL: deps.SubscriptionDetailURL,
		// Budget tab deps (optional; nil = empty state).
		ReadJobTemplate:                 deps.ReadJobTemplate,
		ListJobTemplatePhasesByTemplate: deps.ListJobTemplatePhasesByTemplate,
		ListJobTemplateTasksByPhase:     deps.ListJobTemplateTasksByPhase,
		// Actuals tab dep (optional; nil = empty state).
		GetJobActivityRollup: deps.GetJobActivityRollup,
	}

	actionDeps := &jobaction.Deps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CreateJob:         deps.CreateJob,
		ReadJob:           deps.ReadJob,
		UpdateJob:         deps.UpdateJob,
		DeleteJob:         deps.DeleteJob,
		ListJobs:          deps.ListJobs,
		ClientSearchURL:   deps.ClientSearchURL,
		LocationSearchURL: deps.LocationSearchURL,
	}

	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	dashboardDeps := &jobdashboard.Deps{
		Routes:               deps.Routes,
		JobTemplateRoutes:    deps.JobTemplateRoutes,
		JobActivityRoutes:    deps.JobActivityRoutes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		GetDashboardPageData: deps.GetJobDashboardPageData,
	}

	return &Module{
		routes: deps.Routes,
		List: joblist.NewView(&joblist.ListViewDeps{
			Routes:       deps.Routes,
			ListJobs:     deps.ListJobs,
			GetInUseIDs:  deps.GetInUseIDs,
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
		AssignTask:       jobdetail.NewAssignTaskAction(detailDeps),
		// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
		Dashboard: jobdashboard.NewView(dashboardDeps),
	}
}

// RegisterRoutes registers all job routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	// Phase 3 — Pyeza dashboard block + per-app live dashboards plan.
	if m.Dashboard != nil && m.routes.DashboardURL != "" {
		r.GET(m.routes.DashboardURL, m.Dashboard)
	}
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
	// Task actions
	if m.routes.TaskAssignURL != "" {
		r.POST(m.routes.TaskAssignURL, m.AssignTask)
	}
	// Note: /action/job-phase/set-status is now owned by the job_phase module.
	// Register it via JobPhaseModule.RegisterRoutes instead of here.
}
