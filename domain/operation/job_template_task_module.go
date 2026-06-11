package operation

import (
	"context"

	jobtemplateTaskpkg "github.com/erniealice/fayna-golang/domain/operation/job_template_task"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"

	jobtemplateTaskaction "github.com/erniealice/fayna-golang/domain/operation/job_template_task/action"
)

// JobTemplateTaskModuleDeps holds all dependencies for the job_template_task drawer-only module.
type JobTemplateTaskModuleDeps struct {
	Routes       jobtemplateTaskpkg.Routes
	Labels       jobtemplateTaskpkg.Labels
	CommonLabels pyeza.CommonLabels

	// Job template task CRUD
	CreateJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.CreateJobTemplateTaskRequest) (*jobtemplateTaskpb.CreateJobTemplateTaskResponse, error)
	ReadJobTemplateTask   func(ctx context.Context, req *jobtemplateTaskpb.ReadJobTemplateTaskRequest) (*jobtemplateTaskpb.ReadJobTemplateTaskResponse, error)
	UpdateJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.UpdateJobTemplateTaskRequest) (*jobtemplateTaskpb.UpdateJobTemplateTaskResponse, error)
	DeleteJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.DeleteJobTemplateTaskRequest) (*jobtemplateTaskpb.DeleteJobTemplateTaskResponse, error)
}

// JobTemplateTaskModule holds all constructed job_template_task views.
type JobTemplateTaskModule struct {
	routes     jobtemplateTaskpkg.Routes
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewJobTemplateTaskModule creates the job_template_task drawer-only module with all views wired.
func NewJobTemplateTaskModule(deps *JobTemplateTaskModuleDeps) *JobTemplateTaskModule {
	actionDeps := &jobtemplateTaskaction.Deps{
		Routes:                deps.Routes,
		Labels:                deps.Labels,
		CreateJobTemplateTask: deps.CreateJobTemplateTask,
		ReadJobTemplateTask:   deps.ReadJobTemplateTask,
		UpdateJobTemplateTask: deps.UpdateJobTemplateTask,
		DeleteJobTemplateTask: deps.DeleteJobTemplateTask,
		ResourceSearchURL:     deps.Routes.ResourceSearchURL,
	}

	return &JobTemplateTaskModule{
		routes:     deps.Routes,
		Add:        jobtemplateTaskaction.NewAddAction(actionDeps),
		Edit:       jobtemplateTaskaction.NewEditAction(actionDeps),
		Delete:     jobtemplateTaskaction.NewDeleteAction(actionDeps),
		BulkDelete: jobtemplateTaskaction.NewBulkDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all job_template_task routes.
// No list page, no detail page — drawer CTAs only.
func (m *JobTemplateTaskModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	if m.routes.BulkDeleteURL != "" {
		r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	}
}
