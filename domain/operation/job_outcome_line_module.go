package operation

import (
	"context"

	joboutcomelinepkg "github.com/erniealice/fayna-golang/domain/operation/job_outcome_line"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	joboutcomelinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"

	joboutcomelinedetail "github.com/erniealice/fayna-golang/domain/operation/job_outcome_line/detail"
	joboutcomelinelist "github.com/erniealice/fayna-golang/domain/operation/job_outcome_line/list"
)

// JobOutcomeLineModuleDeps holds all dependencies for the job outcome line module.
type JobOutcomeLineModuleDeps struct {
	Routes       joboutcomelinepkg.Routes
	Labels       joboutcomelinepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Job outcome line CRUD
	CreateJobOutcomeLine func(ctx context.Context, req *joboutcomelinepb.CreateJobOutcomeLineRequest) (*joboutcomelinepb.CreateJobOutcomeLineResponse, error)
	ReadJobOutcomeLine   func(ctx context.Context, req *joboutcomelinepb.ReadJobOutcomeLineRequest) (*joboutcomelinepb.ReadJobOutcomeLineResponse, error)
	UpdateJobOutcomeLine func(ctx context.Context, req *joboutcomelinepb.UpdateJobOutcomeLineRequest) (*joboutcomelinepb.UpdateJobOutcomeLineResponse, error)
	DeleteJobOutcomeLine func(ctx context.Context, req *joboutcomelinepb.DeleteJobOutcomeLineRequest) (*joboutcomelinepb.DeleteJobOutcomeLineResponse, error)
	ListJobOutcomeLines  func(ctx context.Context, req *joboutcomelinepb.ListJobOutcomeLinesRequest) (*joboutcomelinepb.ListJobOutcomeLinesResponse, error)

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// JobOutcomeLineModule holds all constructed job outcome line views.
type JobOutcomeLineModule struct {
	routes     joboutcomelinepkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewJobOutcomeLineModule creates a new job outcome line module with all views wired.
func NewJobOutcomeLineModule(deps *JobOutcomeLineModuleDeps) *JobOutcomeLineModule {
	detailDeps := &joboutcomelinedetail.DetailViewDeps{
		AuditOps:           deps.AuditOps,
		Routes:             deps.Routes,
		Labels:             deps.Labels,
		CommonLabels:       deps.CommonLabels,
		TableLabels:        deps.TableLabels,
		ReadJobOutcomeLine: deps.ReadJobOutcomeLine,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &joboutcomelinepkg.ModuleDeps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		TableLabels:          deps.TableLabels,
		CreateJobOutcomeLine: deps.CreateJobOutcomeLine,
		ReadJobOutcomeLine:   deps.ReadJobOutcomeLine,
		UpdateJobOutcomeLine: deps.UpdateJobOutcomeLine,
		DeleteJobOutcomeLine: deps.DeleteJobOutcomeLine,
		ListJobOutcomeLines:  deps.ListJobOutcomeLines,
	}

	return &JobOutcomeLineModule{
		routes: deps.Routes,
		List: joboutcomelinelist.NewView(&joboutcomelinelist.ListViewDeps{
			Routes:              deps.Routes,
			ListJobOutcomeLines: deps.ListJobOutcomeLines,
			Labels:              deps.Labels,
			CommonLabels:        deps.CommonLabels,
			TableLabels:         deps.TableLabels,
		}),
		Detail:     joboutcomelinedetail.NewView(detailDeps),
		TabAction:  joboutcomelinedetail.NewTabAction(detailDeps),
		Add:        joboutcomelinepkg.NewAddAction(entityDeps),
		Edit:       joboutcomelinepkg.NewEditAction(entityDeps),
		Delete:     joboutcomelinepkg.NewDeleteAction(entityDeps),
		BulkDelete: joboutcomelinepkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all job outcome line routes.
func (m *JobOutcomeLineModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)

	// CRUD actions (GET = drawer form, POST = process submission)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
}
