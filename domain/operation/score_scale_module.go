package operation

import (
	"context"

	scorescalepkg "github.com/erniealice/fayna-golang/domain/operation/score_scale"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"

	scorescaledetail "github.com/erniealice/fayna-golang/domain/operation/score_scale/detail"
	scorescalelist "github.com/erniealice/fayna-golang/domain/operation/score_scale/list"
)

// ScoreScaleModuleDeps holds all dependencies for the score scale module.
type ScoreScaleModuleDeps struct {
	Routes       scorescalepkg.Routes
	Labels       scorescalepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Score scale CRUD
	CreateScoreScale func(ctx context.Context, req *scalepb.CreateScoreScaleRequest) (*scalepb.CreateScoreScaleResponse, error)
	ReadScoreScale   func(ctx context.Context, req *scalepb.ReadScoreScaleRequest) (*scalepb.ReadScoreScaleResponse, error)
	UpdateScoreScale func(ctx context.Context, req *scalepb.UpdateScoreScaleRequest) (*scalepb.UpdateScoreScaleResponse, error)
	DeleteScoreScale func(ctx context.Context, req *scalepb.DeleteScoreScaleRequest) (*scalepb.DeleteScoreScaleResponse, error)
	ListScoreScales  func(ctx context.Context, req *scalepb.ListScoreScalesRequest) (*scalepb.ListScoreScalesResponse, error)
}

// ScoreScaleModule holds all constructed score scale views.
type ScoreScaleModule struct {
	routes     scorescalepkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewScoreScaleModule creates a new score scale module with all views wired.
func NewScoreScaleModule(deps *ScoreScaleModuleDeps) *ScoreScaleModule {
	detailDeps := &scorescaledetail.DetailViewDeps{
		Routes:         deps.Routes,
		Labels:         deps.Labels,
		CommonLabels:   deps.CommonLabels,
		TableLabels:    deps.TableLabels,
		ReadScoreScale: deps.ReadScoreScale,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &scorescalepkg.ModuleDeps{
		Routes:           deps.Routes,
		Labels:           deps.Labels,
		CommonLabels:     deps.CommonLabels,
		TableLabels:      deps.TableLabels,
		CreateScoreScale: deps.CreateScoreScale,
		ReadScoreScale:   deps.ReadScoreScale,
		UpdateScoreScale: deps.UpdateScoreScale,
		DeleteScoreScale: deps.DeleteScoreScale,
		ListScoreScales:  deps.ListScoreScales,
	}

	return &ScoreScaleModule{
		routes: deps.Routes,
		List: scorescalelist.NewView(&scorescalelist.ListViewDeps{
			Routes:          deps.Routes,
			ListScoreScales: deps.ListScoreScales,
			Labels:          deps.Labels,
			CommonLabels:    deps.CommonLabels,
			TableLabels:     deps.TableLabels,
		}),
		Detail:     scorescaledetail.NewView(detailDeps),
		TabAction:  scorescaledetail.NewTabAction(detailDeps),
		Add:        scorescalepkg.NewAddAction(entityDeps),
		Edit:       scorescalepkg.NewEditAction(entityDeps),
		Delete:     scorescalepkg.NewDeleteAction(entityDeps),
		BulkDelete: scorescalepkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all score scale routes.
func (m *ScoreScaleModule) RegisterRoutes(r view.RouteRegistrar) {
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
