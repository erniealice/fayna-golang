// Package action contains HTTP/HTMX handlers for the job_template_phase drawer-only module.
// Dependency-bearing helpers that need full Deps live here; pure-function builders live
// in the sibling form/ package.
package action

import (
	"context"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	fayna "github.com/erniealice/fayna-golang"
)

// Deps holds dependencies shared across all job_template_phase action handlers.
type Deps struct {
	Routes fayna.JobTemplatePhaseRoutes
	Labels fayna.JobTemplatePhaseLabels

	// Job template phase CRUD
	CreateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.CreateJobTemplatePhaseRequest) (*jobtemplatephasepb.CreateJobTemplatePhaseResponse, error)
	ReadJobTemplatePhase   func(ctx context.Context, req *jobtemplatephasepb.ReadJobTemplatePhaseRequest) (*jobtemplatephasepb.ReadJobTemplatePhaseResponse, error)
	UpdateJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.UpdateJobTemplatePhaseRequest) (*jobtemplatephasepb.UpdateJobTemplatePhaseResponse, error)
	DeleteJobTemplatePhase func(ctx context.Context, req *jobtemplatephasepb.DeleteJobTemplatePhaseRequest) (*jobtemplatephasepb.DeleteJobTemplatePhaseResponse, error)

	// ResourceSearchURL for the resource picker in the Add/Edit drawer.
	ResourceSearchURL string
}
