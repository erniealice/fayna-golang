package operation

import (
	"context"

	scoringschemepkg "github.com/erniealice/fayna-golang/domain/operation/scoring_scheme"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	schemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"

	scoringschemedetail "github.com/erniealice/fayna-golang/domain/operation/scoring_scheme/detail"
	scoringschemelist "github.com/erniealice/fayna-golang/domain/operation/scoring_scheme/list"
)

// ScoringSchemeModuleDeps holds all dependencies for the scoring scheme module.
type ScoringSchemeModuleDeps struct {
	Routes       scoringschemepkg.Routes
	Labels       scoringschemepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Scoring scheme CRUD
	CreateScoringScheme func(ctx context.Context, req *schemepb.CreateScoringSchemeRequest) (*schemepb.CreateScoringSchemeResponse, error)
	ReadScoringScheme   func(ctx context.Context, req *schemepb.ReadScoringSchemeRequest) (*schemepb.ReadScoringSchemeResponse, error)
	UpdateScoringScheme func(ctx context.Context, req *schemepb.UpdateScoringSchemeRequest) (*schemepb.UpdateScoringSchemeResponse, error)
	DeleteScoringScheme func(ctx context.Context, req *schemepb.DeleteScoringSchemeRequest) (*schemepb.DeleteScoringSchemeResponse, error)
	ListScoringSchemes  func(ctx context.Context, req *schemepb.ListScoringSchemesRequest) (*schemepb.ListScoringSchemesResponse, error)

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// ScoringSchemeModule holds all constructed scoring scheme views.
type ScoringSchemeModule struct {
	routes     scoringschemepkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewScoringSchemeModule creates a new scoring scheme module with all views wired.
func NewScoringSchemeModule(deps *ScoringSchemeModuleDeps) *ScoringSchemeModule {
	detailDeps := &scoringschemedetail.DetailViewDeps{
		AuditOps:          deps.AuditOps,
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CommonLabels:      deps.CommonLabels,
		TableLabels:       deps.TableLabels,
		ReadScoringScheme: deps.ReadScoringScheme,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &scoringschemepkg.ModuleDeps{
		Routes:              deps.Routes,
		Labels:              deps.Labels,
		CommonLabels:        deps.CommonLabels,
		TableLabels:         deps.TableLabels,
		CreateScoringScheme: deps.CreateScoringScheme,
		ReadScoringScheme:   deps.ReadScoringScheme,
		UpdateScoringScheme: deps.UpdateScoringScheme,
		DeleteScoringScheme: deps.DeleteScoringScheme,
		ListScoringSchemes:  deps.ListScoringSchemes,
	}

	return &ScoringSchemeModule{
		routes: deps.Routes,
		List: scoringschemelist.NewView(&scoringschemelist.ListViewDeps{
			Routes:             deps.Routes,
			ListScoringSchemes: deps.ListScoringSchemes,
			Labels:             deps.Labels,
			CommonLabels:       deps.CommonLabels,
			TableLabels:        deps.TableLabels,
		}),
		Detail:     scoringschemedetail.NewView(detailDeps),
		TabAction:  scoringschemedetail.NewTabAction(detailDeps),
		Add:        scoringschemepkg.NewAddAction(entityDeps),
		Edit:       scoringschemepkg.NewEditAction(entityDeps),
		Delete:     scoringschemepkg.NewDeleteAction(entityDeps),
		BulkDelete: scoringschemepkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all scoring scheme routes.
func (m *ScoringSchemeModule) RegisterRoutes(r view.RouteRegistrar) {
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
