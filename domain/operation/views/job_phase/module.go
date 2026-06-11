// Package job_phase provides the view module for the JobPhase entity.
//
// Shape 3 (sibling, per-handler split). Reached via Job detail Phases tab
// deep links (/app/job-phase/{id}). The list page (/app/job-phases/list/{status})
// is a power-user/debugging surface with no sidebar entry.
//
// The set-status handler at /action/job-phase/set-status is the canonical owner
// of this route (LIFTED from views/job/action/job_phase_set_status.go). The URL
// pattern is preserved so existing E2E tests in phase5-revenue-recognition continue
// to work without modification.
package job_phase

import (
	"context"

	operation "github.com/erniealice/fayna-golang/domain/operation"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"

	jobphaseaction "github.com/erniealice/fayna-golang/domain/operation/views/job_phase/action"
	jobphasedetail "github.com/erniealice/fayna-golang/domain/operation/views/job_phase/detail"
	jobphaselist "github.com/erniealice/fayna-golang/domain/operation/views/job_phase/list"
)

// ModuleDeps holds all dependencies for the job_phase module.
type ModuleDeps struct {
	Routes       operation.JobPhaseRoutes
	Labels       operation.JobPhaseLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// GetInUseIDs blocks deletion of phases that are referenced by job_task rows.
	GetInUseIDs func(ctx context.Context, ids []string) (map[string]bool, error)

	// Job phase CRUD
	CreateJobPhase func(ctx context.Context, req *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error)
	ReadJobPhase   func(ctx context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error)
	UpdateJobPhase func(ctx context.Context, req *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error)
	DeleteJobPhase func(ctx context.Context, req *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error)
	ListJobPhases  func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)

	// Activities tab — filtered in-memory from all job activities.
	// TODO(WaveN): replace with dedicated ListJobActivitiesByPhase RPC.
	ListJobActivities func(ctx context.Context, req *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)

	// Tasks tab — list tasks for this phase via the job_task module.
	ListJobTasksByPhase func(ctx context.Context, req *jobtaskpb.ListJobTasksByPhaseRequest) (*jobtaskpb.ListJobTasksByPhaseResponse, error)

	// Attachment operations (optional — nil = attachments tab hidden/empty)
	attachment.AttachmentOps

	// Audit history (optional — nil = history tab hidden/empty)
	// AuditOps.ListAuditHistory is the single source of truth; do not add a
	// duplicate standalone ListAuditHistory field here.
	auditlog.AuditOps
}

// Module holds all constructed job_phase views.
type Module struct {
	routes           operation.JobPhaseRoutes
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

// NewModule creates a new job_phase module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	actionDeps := &jobphaseaction.Deps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CreateJobPhase:    deps.CreateJobPhase,
		ReadJobPhase:      deps.ReadJobPhase,
		UpdateJobPhase:    deps.UpdateJobPhase,
		DeleteJobPhase:    deps.DeleteJobPhase,
		ListJobPhases:     deps.ListJobPhases,
		ResourceSearchURL: deps.Routes.ResourceSearchURL,
	}

	detailDeps := &jobphasedetail.DetailViewDeps{
		AttachmentOps:       deps.AttachmentOps,
		AuditOps:            deps.AuditOps,
		Routes:              deps.Routes,
		Labels:              deps.Labels,
		CommonLabels:        deps.CommonLabels,
		TableLabels:         deps.TableLabels,
		ReadJobPhase:        deps.ReadJobPhase,
		ListJobActivities:   deps.ListJobActivities,
		ListJobTasksByPhase: deps.ListJobTasksByPhase,
	}

	return &Module{
		routes: deps.Routes,
		List: jobphaselist.NewView(&jobphaselist.ListViewDeps{
			Routes:        deps.Routes,
			ListJobPhases: deps.ListJobPhases,
			GetInUseIDs:   deps.GetInUseIDs,
			Labels:        deps.Labels,
			CommonLabels:  deps.CommonLabels,
			TableLabels:   deps.TableLabels,
		}),
		Detail:           jobphasedetail.NewView(detailDeps),
		TabAction:        jobphasedetail.NewTabAction(detailDeps),
		Add:              jobphaseaction.NewAddAction(actionDeps),
		Edit:             jobphaseaction.NewEditAction(actionDeps),
		Delete:           jobphaseaction.NewDeleteAction(actionDeps),
		BulkDelete:       jobphaseaction.NewBulkDeleteAction(actionDeps),
		SetStatus:        jobphaseaction.NewSetStatusAction(actionDeps),
		BulkSetStatus:    jobphaseaction.NewBulkSetStatusAction(actionDeps),
		AttachmentUpload: jobphasedetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: jobphasedetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all job_phase routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
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

	// The set-status URL is /action/job-phase/set-status — the canonical route
	// that existing E2E tests in phase5-revenue-recognition reference. This module
	// is now the owner; views/job/module.go no longer registers this route.
	r.POST(m.routes.SetStatusURL, m.SetStatus)
	r.POST(m.routes.BulkSetStatusURL, m.BulkSetStatus)

	// Attachments (optional)
	if m.AttachmentUpload != nil && m.routes.AttachmentUploadURL != "" {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
