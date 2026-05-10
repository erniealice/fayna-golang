// Package job_template_task provides the drawer-only view module for the JobTemplateTask entity.
//
// Shape: drawer-only subset (no list page, no standalone detail page, no sidebar entry).
// Operators reach this module only via the JobTemplate detail Tasks tab
// Add / Edit / Delete CTAs.
//
// Routes registered: Add (GET+POST), Edit (GET+POST), Delete (POST), BulkDelete (POST).
package job_template_task

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"
	jobtemplateTaskaction "github.com/erniealice/fayna-golang/views/job_template_task/action"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
)

// ModuleDeps holds all dependencies for the job_template_task drawer-only module.
type ModuleDeps struct {
	Routes       fayna.JobTemplateTaskRoutes
	Labels       fayna.JobTemplateTaskLabels
	CommonLabels pyeza.CommonLabels

	// Job template task CRUD
	CreateJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.CreateJobTemplateTaskRequest) (*jobtemplateTaskpb.CreateJobTemplateTaskResponse, error)
	ReadJobTemplateTask   func(ctx context.Context, req *jobtemplateTaskpb.ReadJobTemplateTaskRequest) (*jobtemplateTaskpb.ReadJobTemplateTaskResponse, error)
	UpdateJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.UpdateJobTemplateTaskRequest) (*jobtemplateTaskpb.UpdateJobTemplateTaskResponse, error)
	DeleteJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.DeleteJobTemplateTaskRequest) (*jobtemplateTaskpb.DeleteJobTemplateTaskResponse, error)
}

// Module holds all constructed job_template_task views.
type Module struct {
	routes     fayna.JobTemplateTaskRoutes
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewModule creates the job_template_task drawer-only module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	actionDeps := &jobtemplateTaskaction.Deps{
		Routes:                deps.Routes,
		Labels:                deps.Labels,
		CreateJobTemplateTask: deps.CreateJobTemplateTask,
		ReadJobTemplateTask:   deps.ReadJobTemplateTask,
		UpdateJobTemplateTask: deps.UpdateJobTemplateTask,
		DeleteJobTemplateTask: deps.DeleteJobTemplateTask,
		ResourceSearchURL:     deps.Routes.ResourceSearchURL,
	}

	return &Module{
		routes:     deps.Routes,
		Add:        jobtemplateTaskaction.NewAddAction(actionDeps),
		Edit:       jobtemplateTaskaction.NewEditAction(actionDeps),
		Delete:     jobtemplateTaskaction.NewDeleteAction(actionDeps),
		BulkDelete: jobtemplateTaskaction.NewBulkDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all job_template_task routes.
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
