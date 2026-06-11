// Package action contains HTTP/HTMX handlers for the job_template_task drawer-only module.
// Dependency-bearing helpers that need full Deps live here; pure-function builders live
// in the sibling form/ package.
package action

import (
	"context"

	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	operation "github.com/erniealice/fayna-golang/domain/operation"
)

// Deps holds dependencies shared across all job_template_task action handlers.
type Deps struct {
	Routes operation.JobTemplateTaskRoutes
	Labels operation.JobTemplateTaskLabels

	// Job template task CRUD
	CreateJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.CreateJobTemplateTaskRequest) (*jobtemplateTaskpb.CreateJobTemplateTaskResponse, error)
	ReadJobTemplateTask   func(ctx context.Context, req *jobtemplateTaskpb.ReadJobTemplateTaskRequest) (*jobtemplateTaskpb.ReadJobTemplateTaskResponse, error)
	UpdateJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.UpdateJobTemplateTaskRequest) (*jobtemplateTaskpb.UpdateJobTemplateTaskResponse, error)
	DeleteJobTemplateTask func(ctx context.Context, req *jobtemplateTaskpb.DeleteJobTemplateTaskRequest) (*jobtemplateTaskpb.DeleteJobTemplateTaskResponse, error)

	// ResourceSearchURL for the resource picker in the Add/Edit drawer.
	ResourceSearchURL string
}
