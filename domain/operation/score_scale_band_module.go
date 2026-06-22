package operation

import (
	"context"

	scorescalebandpkg "github.com/erniealice/fayna-golang/domain/operation/score_scale_band"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	ssbpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"

	scorescalebanddetail "github.com/erniealice/fayna-golang/domain/operation/score_scale_band/detail"
	scorescalebandlist "github.com/erniealice/fayna-golang/domain/operation/score_scale_band/list"
)

// ScoreScaleBandModuleDeps holds all dependencies for the score scale band module.
type ScoreScaleBandModuleDeps struct {
	Routes       scorescalebandpkg.Routes
	Labels       scorescalebandpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Score scale band CRUD
	CreateScoreScaleBand func(ctx context.Context, req *ssbpb.CreateScoreScaleBandRequest) (*ssbpb.CreateScoreScaleBandResponse, error)
	ReadScoreScaleBand   func(ctx context.Context, req *ssbpb.ReadScoreScaleBandRequest) (*ssbpb.ReadScoreScaleBandResponse, error)
	UpdateScoreScaleBand func(ctx context.Context, req *ssbpb.UpdateScoreScaleBandRequest) (*ssbpb.UpdateScoreScaleBandResponse, error)
	DeleteScoreScaleBand func(ctx context.Context, req *ssbpb.DeleteScoreScaleBandRequest) (*ssbpb.DeleteScoreScaleBandResponse, error)
	ListScoreScaleBands  func(ctx context.Context, req *ssbpb.ListScoreScaleBandsRequest) (*ssbpb.ListScoreScaleBandsResponse, error)
}

// ScoreScaleBandModule holds all constructed score scale band views.
type ScoreScaleBandModule struct {
	routes     scorescalebandpkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewScoreScaleBandModule creates a new score scale band module with all views wired.
func NewScoreScaleBandModule(deps *ScoreScaleBandModuleDeps) *ScoreScaleBandModule {
	detailDeps := &scorescalebanddetail.DetailViewDeps{
		Routes:             deps.Routes,
		Labels:             deps.Labels,
		CommonLabels:       deps.CommonLabels,
		TableLabels:        deps.TableLabels,
		ReadScoreScaleBand: deps.ReadScoreScaleBand,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &scorescalebandpkg.ModuleDeps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		TableLabels:          deps.TableLabels,
		CreateScoreScaleBand: deps.CreateScoreScaleBand,
		ReadScoreScaleBand:   deps.ReadScoreScaleBand,
		UpdateScoreScaleBand: deps.UpdateScoreScaleBand,
		DeleteScoreScaleBand: deps.DeleteScoreScaleBand,
		ListScoreScaleBands:  deps.ListScoreScaleBands,
	}

	return &ScoreScaleBandModule{
		routes: deps.Routes,
		List: scorescalebandlist.NewView(&scorescalebandlist.ListViewDeps{
			Routes:              deps.Routes,
			ListScoreScaleBands: deps.ListScoreScaleBands,
			Labels:              deps.Labels,
			CommonLabels:        deps.CommonLabels,
			TableLabels:         deps.TableLabels,
		}),
		Detail:     scorescalebanddetail.NewView(detailDeps),
		TabAction:  scorescalebanddetail.NewTabAction(detailDeps),
		Add:        scorescalebandpkg.NewAddAction(entityDeps),
		Edit:       scorescalebandpkg.NewEditAction(entityDeps),
		Delete:     scorescalebandpkg.NewDeleteAction(entityDeps),
		BulkDelete: scorescalebandpkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all score scale band routes.
func (m *ScoreScaleBandModule) RegisterRoutes(r view.RouteRegistrar) {
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
