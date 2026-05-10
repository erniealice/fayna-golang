package action

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// Deps holds dependencies for job template action handlers.
type Deps struct {
	Routes             fayna.JobTemplateRoutes
	Labels             fayna.JobTemplateLabels
	CreateJobTemplate  func(ctx context.Context, req *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error)
	ReadJobTemplate    func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	UpdateJobTemplate  func(ctx context.Context, req *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error)
	DeleteJobTemplate  func(ctx context.Context, req *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error)
}
