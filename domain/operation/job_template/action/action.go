package action

import (
	"context"

	job_template "github.com/erniealice/fayna-golang/domain/operation/job_template"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// Deps holds dependencies for job template action handlers.
type Deps struct {
	Routes            job_template.Routes
	Labels            job_template.Labels
	CreateJobTemplate func(ctx context.Context, req *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error)
	ReadJobTemplate   func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	UpdateJobTemplate func(ctx context.Context, req *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error)
	DeleteJobTemplate func(ctx context.Context, req *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error)
}
