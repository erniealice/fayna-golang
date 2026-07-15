package job_category

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// ModuleDeps holds the typed closures action builders + sub-packages need.
// Defined here (not in the module file) to avoid a self-import cycle.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateJobCategory func(ctx context.Context, req *jobcategorypb.CreateJobCategoryRequest) (*jobcategorypb.CreateJobCategoryResponse, error)
	ReadJobCategory   func(ctx context.Context, req *jobcategorypb.ReadJobCategoryRequest) (*jobcategorypb.ReadJobCategoryResponse, error)
	UpdateJobCategory func(ctx context.Context, req *jobcategorypb.UpdateJobCategoryRequest) (*jobcategorypb.UpdateJobCategoryResponse, error)
	DeleteJobCategory func(ctx context.Context, req *jobcategorypb.DeleteJobCategoryRequest) (*jobcategorypb.DeleteJobCategoryResponse, error)
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
}
