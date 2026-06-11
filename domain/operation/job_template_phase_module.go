package operation

import (
	"context"

	jobtemplatephasepkg "github.com/erniealice/fayna-golang/domain/operation/job_template_phase"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"

	jobtemplatephaseaction "github.com/erniealice/fayna-golang/domain/operation/job_template_phase/action"
)

// JobTemplatePhaseModuleDeps holds all dependencies for the job_template_phase drawer-only module.
type JobTemplatePhaseModuleDeps struct {
	Routes       jobtemplatephasepkg.Routes
	Labels       jobtemplatephasepkg.Labels
	CommonLabels pyeza.CommonLabels

	// Job template phase CRUD
	CreateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error)
	ReadJobTemplatePhase   func(ctx context.Context, req *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error)
	UpdateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error)
	DeleteJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error)
}

// JobTemplatePhaseModule holds all constructed job_template_phase views.
type JobTemplatePhaseModule struct {
	routes     jobtemplatephasepkg.Routes
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewJobTemplatePhaseModule creates the job_template_phase drawer-only module with all views wired.
func NewJobTemplatePhaseModule(deps *JobTemplatePhaseModuleDeps) *JobTemplatePhaseModule {
	actionDeps := &jobtemplatephaseaction.Deps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CreateJobTemplatePhase: deps.CreateJobTemplatePhase,
		ReadJobTemplatePhase:   deps.ReadJobTemplatePhase,
		UpdateJobTemplatePhase: deps.UpdateJobTemplatePhase,
		DeleteJobTemplatePhase: deps.DeleteJobTemplatePhase,
		ResourceSearchURL:      deps.Routes.ResourceSearchURL,
	}

	return &JobTemplatePhaseModule{
		routes:     deps.Routes,
		Add:        jobtemplatephaseaction.NewAddAction(actionDeps),
		Edit:       jobtemplatephaseaction.NewEditAction(actionDeps),
		Delete:     jobtemplatephaseaction.NewDeleteAction(actionDeps),
		BulkDelete: jobtemplatephaseaction.NewBulkDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all job_template_phase routes.
// No list page, no detail page — drawer CTAs only.
func (m *JobTemplatePhaseModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	if m.routes.BulkDeleteURL != "" {
		r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	}
}
