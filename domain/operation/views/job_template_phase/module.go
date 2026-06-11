// Package job_template_phase provides the drawer-only view module for the JobTemplatePhase entity.
//
// Shape: drawer-only subset (no list page, no standalone detail page, no sidebar entry).
// Operators reach this module only via the JobTemplate detail Phases tab
// Add / Edit / Delete CTAs.
//
// Routes registered: Add (GET+POST), Edit (GET+POST), Delete (POST), BulkDelete (POST).
package job_template_phase

import (
	"context"

	operation "github.com/erniealice/fayna-golang/domain/operation"
	jobtemplatephaseaction "github.com/erniealice/fayna-golang/domain/operation/views/job_template_phase/action"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
)

// ModuleDeps holds all dependencies for the job_template_phase drawer-only module.
type ModuleDeps struct {
	Routes       operation.JobTemplatePhaseRoutes
	Labels       operation.JobTemplatePhaseLabels
	CommonLabels pyeza.CommonLabels

	// Job template phase CRUD
	CreateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error)
	ReadJobTemplatePhase   func(ctx context.Context, req *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error)
	UpdateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error)
	DeleteJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error)
}

// Module holds all constructed job_template_phase views.
type Module struct {
	routes     operation.JobTemplatePhaseRoutes
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewModule creates the job_template_phase drawer-only module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	actionDeps := &jobtemplatephaseaction.Deps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CreateJobTemplatePhase: deps.CreateJobTemplatePhase,
		ReadJobTemplatePhase:   deps.ReadJobTemplatePhase,
		UpdateJobTemplatePhase: deps.UpdateJobTemplatePhase,
		DeleteJobTemplatePhase: deps.DeleteJobTemplatePhase,
		ResourceSearchURL:      deps.Routes.ResourceSearchURL,
	}

	return &Module{
		routes:     deps.Routes,
		Add:        jobtemplatephaseaction.NewAddAction(actionDeps),
		Edit:       jobtemplatephaseaction.NewEditAction(actionDeps),
		Delete:     jobtemplatephaseaction.NewDeleteAction(actionDeps),
		BulkDelete: jobtemplatephaseaction.NewBulkDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all job_template_phase routes.
// No list page, no detail page — drawer CTAs only.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	if m.routes.BulkDeleteURL != "" {
		r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	}
}
