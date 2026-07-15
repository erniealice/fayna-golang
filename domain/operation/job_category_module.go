package operation

import (
	"context"

	jobcategorypkg "github.com/erniealice/fayna-golang/domain/operation/job_category"
	jobcategorydetail "github.com/erniealice/fayna-golang/domain/operation/job_category/detail"
	jobcategorylist "github.com/erniealice/fayna-golang/domain/operation/job_category/list"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// JobCategoryModuleDeps holds all dependencies for the job category module.
type JobCategoryModuleDeps struct {
	Routes       jobcategorypkg.Routes
	Labels       jobcategorypkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateJobCategory func(ctx context.Context, req *jobcategorypb.CreateJobCategoryRequest) (*jobcategorypb.CreateJobCategoryResponse, error)
	ReadJobCategory   func(ctx context.Context, req *jobcategorypb.ReadJobCategoryRequest) (*jobcategorypb.ReadJobCategoryResponse, error)
	UpdateJobCategory func(ctx context.Context, req *jobcategorypb.UpdateJobCategoryRequest) (*jobcategorypb.UpdateJobCategoryResponse, error)
	DeleteJobCategory func(ctx context.Context, req *jobcategorypb.DeleteJobCategoryRequest) (*jobcategorypb.DeleteJobCategoryResponse, error)
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)
}

// JobCategoryModule holds all constructed job category views.
type JobCategoryModule struct {
	routes     jobcategorypkg.Routes
	List       view.View
	Detail     view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewJobCategoryModule creates a new job category module with all views wired.
func NewJobCategoryModule(deps *JobCategoryModuleDeps) *JobCategoryModule {
	entityDeps := &jobcategorypkg.ModuleDeps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CommonLabels:      deps.CommonLabels,
		TableLabels:       deps.TableLabels,
		CreateJobCategory: deps.CreateJobCategory,
		ReadJobCategory:   deps.ReadJobCategory,
		UpdateJobCategory: deps.UpdateJobCategory,
		DeleteJobCategory: deps.DeleteJobCategory,
		ListJobCategories: deps.ListJobCategories,
	}

	return &JobCategoryModule{
		routes: deps.Routes,
		List: jobcategorylist.NewView(&jobcategorylist.ListViewDeps{
			Routes:            deps.Routes,
			Labels:            deps.Labels,
			CommonLabels:      deps.CommonLabels,
			TableLabels:       deps.TableLabels,
			ListJobCategories: deps.ListJobCategories,
		}),
		Detail: jobcategorydetail.NewView(&jobcategorydetail.DetailViewDeps{
			Routes:          deps.Routes,
			Labels:          deps.Labels,
			CommonLabels:    deps.CommonLabels,
			ReadJobCategory: deps.ReadJobCategory,
		}),
		Add:        jobcategorypkg.NewAddAction(entityDeps),
		Edit:       jobcategorypkg.NewEditAction(entityDeps),
		Delete:     jobcategorypkg.NewDeleteAction(entityDeps),
		BulkDelete: jobcategorypkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all job category routes.
func (m *JobCategoryModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)

	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
}
