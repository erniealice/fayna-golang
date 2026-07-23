package action

import (
	"context"

	"github.com/erniealice/fayna-golang/domain/operation/job_template"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	productpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product"
)

// Deps holds dependencies for job template action handlers.
type Deps struct {
	Routes            job_template.Routes
	Labels            job_template.Labels
	CreateJobTemplate func(ctx context.Context, req *jobtemplatepb.CreateJobTemplateRequest) (*jobtemplatepb.CreateJobTemplateResponse, error)
	ReadJobTemplate   func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	UpdateJobTemplate func(ctx context.Context, req *jobtemplatepb.UpdateJobTemplateRequest) (*jobtemplatepb.UpdateJobTemplateResponse, error)
	DeleteJobTemplate func(ctx context.Context, req *jobtemplatepb.DeleteJobTemplateRequest) (*jobtemplatepb.DeleteJobTemplateResponse, error)

	// ListJobCategories populates the Category picker (optional FK; the field
	// stays generic here — vertical vocabulary, e.g. "Subject", lives in lyngua
	// only). Nil-safe: the picker renders with no options.
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
	// ListProducts populates the Output Product picker. No product-search
	// endpoint is reachable from fayna's wiring today, so this backs a plain
	// select (not an action-mode auto-complete) — see W1 report note.
	// Nil-safe: the picker renders with no options.
	ListProducts func(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error)
}
