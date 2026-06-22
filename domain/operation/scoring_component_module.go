package operation

import (
	"context"

	scoringcomponentpkg "github.com/erniealice/fayna-golang/domain/operation/scoring_component"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	scoringpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component"

	scoringcomponentdetail "github.com/erniealice/fayna-golang/domain/operation/scoring_component/detail"
	scoringcomponentlist "github.com/erniealice/fayna-golang/domain/operation/scoring_component/list"
)

// ScoringComponentModuleDeps holds all dependencies for the scoring component module.
type ScoringComponentModuleDeps struct {
	Routes       scoringcomponentpkg.Routes
	Labels       scoringcomponentpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Scoring component CRUD
	CreateScoringComponent func(ctx context.Context, req *scoringpb.CreateScoringComponentRequest) (*scoringpb.CreateScoringComponentResponse, error)
	ReadScoringComponent   func(ctx context.Context, req *scoringpb.ReadScoringComponentRequest) (*scoringpb.ReadScoringComponentResponse, error)
	UpdateScoringComponent func(ctx context.Context, req *scoringpb.UpdateScoringComponentRequest) (*scoringpb.UpdateScoringComponentResponse, error)
	DeleteScoringComponent func(ctx context.Context, req *scoringpb.DeleteScoringComponentRequest) (*scoringpb.DeleteScoringComponentResponse, error)
	ListScoringComponents  func(ctx context.Context, req *scoringpb.ListScoringComponentsRequest) (*scoringpb.ListScoringComponentsResponse, error)

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// ScoringComponentModule holds all constructed scoring component views.
type ScoringComponentModule struct {
	routes     scoringcomponentpkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewScoringComponentModule creates a new scoring component module with all views wired.
func NewScoringComponentModule(deps *ScoringComponentModuleDeps) *ScoringComponentModule {
	detailDeps := &scoringcomponentdetail.DetailViewDeps{
		AuditOps:             deps.AuditOps,
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		TableLabels:          deps.TableLabels,
		ReadScoringComponent: deps.ReadScoringComponent,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &scoringcomponentpkg.ModuleDeps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
		TableLabels:            deps.TableLabels,
		CreateScoringComponent: deps.CreateScoringComponent,
		ReadScoringComponent:   deps.ReadScoringComponent,
		UpdateScoringComponent: deps.UpdateScoringComponent,
		DeleteScoringComponent: deps.DeleteScoringComponent,
		ListScoringComponents:  deps.ListScoringComponents,
	}

	return &ScoringComponentModule{
		routes: deps.Routes,
		List: scoringcomponentlist.NewView(&scoringcomponentlist.ListViewDeps{
			Routes:                deps.Routes,
			ListScoringComponents: deps.ListScoringComponents,
			Labels:                deps.Labels,
			CommonLabels:          deps.CommonLabels,
			TableLabels:           deps.TableLabels,
		}),
		Detail:     scoringcomponentdetail.NewView(detailDeps),
		TabAction:  scoringcomponentdetail.NewTabAction(detailDeps),
		Add:        scoringcomponentpkg.NewAddAction(entityDeps),
		Edit:       scoringcomponentpkg.NewEditAction(entityDeps),
		Delete:     scoringcomponentpkg.NewDeleteAction(entityDeps),
		BulkDelete: scoringcomponentpkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all scoring component routes.
func (m *ScoringComponentModule) RegisterRoutes(r view.RouteRegistrar) {
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
