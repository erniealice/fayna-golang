package operation

import (
	"context"

	sccpkg "github.com/erniealice/fayna-golang/domain/operation/scoring_component_criteria"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	sccpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component_criteria"

	sccdetail "github.com/erniealice/fayna-golang/domain/operation/scoring_component_criteria/detail"
	scclist "github.com/erniealice/fayna-golang/domain/operation/scoring_component_criteria/list"
)

// ScoringComponentCriteriaModuleDeps holds all dependencies for the scoring component criteria module.
type ScoringComponentCriteriaModuleDeps struct {
	Routes       sccpkg.Routes
	Labels       sccpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Scoring component criteria CRUD
	CreateScoringComponentCriteria func(ctx context.Context, req *sccpb.CreateScoringComponentCriteriaRequest) (*sccpb.CreateScoringComponentCriteriaResponse, error)
	ReadScoringComponentCriteria   func(ctx context.Context, req *sccpb.ReadScoringComponentCriteriaRequest) (*sccpb.ReadScoringComponentCriteriaResponse, error)
	UpdateScoringComponentCriteria func(ctx context.Context, req *sccpb.UpdateScoringComponentCriteriaRequest) (*sccpb.UpdateScoringComponentCriteriaResponse, error)
	DeleteScoringComponentCriteria func(ctx context.Context, req *sccpb.DeleteScoringComponentCriteriaRequest) (*sccpb.DeleteScoringComponentCriteriaResponse, error)
	ListScoringComponentCriterias  func(ctx context.Context, req *sccpb.ListScoringComponentCriteriasRequest) (*sccpb.ListScoringComponentCriteriasResponse, error)

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// ScoringComponentCriteriaModule holds all constructed scoring component criteria views.
type ScoringComponentCriteriaModule struct {
	routes     sccpkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewScoringComponentCriteriaModule creates a new scoring component criteria module with all views wired.
func NewScoringComponentCriteriaModule(deps *ScoringComponentCriteriaModuleDeps) *ScoringComponentCriteriaModule {
	detailDeps := &sccdetail.DetailViewDeps{
		AuditOps:                     deps.AuditOps,
		Routes:                       deps.Routes,
		Labels:                       deps.Labels,
		CommonLabels:                 deps.CommonLabels,
		TableLabels:                  deps.TableLabels,
		ReadScoringComponentCriteria: deps.ReadScoringComponentCriteria,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &sccpkg.ModuleDeps{
		Routes:                         deps.Routes,
		Labels:                         deps.Labels,
		CommonLabels:                   deps.CommonLabels,
		TableLabels:                    deps.TableLabels,
		CreateScoringComponentCriteria: deps.CreateScoringComponentCriteria,
		ReadScoringComponentCriteria:   deps.ReadScoringComponentCriteria,
		UpdateScoringComponentCriteria: deps.UpdateScoringComponentCriteria,
		DeleteScoringComponentCriteria: deps.DeleteScoringComponentCriteria,
		ListScoringComponentCriterias:  deps.ListScoringComponentCriterias,
	}

	return &ScoringComponentCriteriaModule{
		routes: deps.Routes,
		List: scclist.NewView(&scclist.ListViewDeps{
			Routes:                        deps.Routes,
			ListScoringComponentCriterias: deps.ListScoringComponentCriterias,
			Labels:                        deps.Labels,
			CommonLabels:                  deps.CommonLabels,
			TableLabels:                   deps.TableLabels,
		}),
		Detail:     sccdetail.NewView(detailDeps),
		TabAction:  sccdetail.NewTabAction(detailDeps),
		Add:        sccpkg.NewAddAction(entityDeps),
		Edit:       sccpkg.NewEditAction(entityDeps),
		Delete:     sccpkg.NewDeleteAction(entityDeps),
		BulkDelete: sccpkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all scoring component criteria routes.
func (m *ScoringComponentCriteriaModule) RegisterRoutes(r view.RouteRegistrar) {
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
