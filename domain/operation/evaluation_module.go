package operation

import (
	"context"

	evalpkg "github.com/erniealice/fayna-golang/domain/operation/evaluation"
	evaldetail "github.com/erniealice/fayna-golang/domain/operation/evaluation/detail"
	evallist "github.com/erniealice/fayna-golang/domain/operation/evaluation/list"
	evalportal "github.com/erniealice/fayna-golang/domain/operation/evaluation/portal"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
	resppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_response"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
)

// EvaluationModuleDeps holds all dependencies for the evaluation module. The
// Integrator (block catalog) injects every closure; all IDOR / CR-5 gates are
// enforced INSIDE the closures (use-case/adapter QUERY PREDICATE).
type EvaluationModuleDeps struct {
	Routes       evalpkg.Routes
	Labels       evalpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateEvaluation func(ctx context.Context, req *evalpb.CreateEvaluationRequest) (*evalpb.CreateEvaluationResponse, error)
	ReadEvaluation   func(ctx context.Context, req *evalpb.ReadEvaluationRequest) (*evalpb.ReadEvaluationResponse, error)
	UpdateEvaluation func(ctx context.Context, req *evalpb.UpdateEvaluationRequest) (*evalpb.UpdateEvaluationResponse, error)
	DeleteEvaluation func(ctx context.Context, req *evalpb.DeleteEvaluationRequest) (*evalpb.DeleteEvaluationResponse, error)
	ListEvaluations  func(ctx context.Context, req *evalpb.ListEvaluationsRequest) (*evalpb.ListEvaluationsResponse, error)

	GetListPageData func(ctx context.Context, req *evalpb.GetEvaluationListPageDataRequest) (*evalpb.GetEvaluationListPageDataResponse, error)
	GetItemPageData func(ctx context.Context, req *evalpb.GetEvaluationItemPageDataRequest) (*evalpb.GetEvaluationItemPageDataResponse, error)

	// Client-portal "Rate My Team" — CLIENT-scoped page-data closure (IDOR
	// predicate is inside the closure; the view supplies no client_id).
	GetPortalPageData func(ctx context.Context, req *evalpb.GetEvaluationListPageDataRequest) (*evalpb.GetEvaluationListPageDataResponse, error)

	SignOffEvaluation func(ctx context.Context, req *evalpb.UpdateEvaluationRequest) (*evalpb.UpdateEvaluationResponse, error)
	ArchiveEvaluation func(ctx context.Context, req *evalpb.UpdateEvaluationRequest) (*evalpb.UpdateEvaluationResponse, error)

	ListEvaluationResponses  func(ctx context.Context, req *resppb.ListEvaluationResponsesRequest) (*resppb.ListEvaluationResponsesResponse, error)
	CreateEvaluationResponse func(ctx context.Context, req *resppb.CreateEvaluationResponseRequest) (*resppb.CreateEvaluationResponseResponse, error)

	ListEvaluationTemplateItems func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)

	NewID func() string
}

// EvaluationModule holds all constructed evaluation views.
type EvaluationModule struct {
	routes evalpkg.Routes

	List          view.View
	Table         view.View
	Detail        view.View
	TabAction     view.View
	Add           view.View
	Edit          view.View
	DimensionSlot view.View
	SignOff       view.View
	Archive       view.View
	Delete        view.View
	BulkArchive   view.View
	Portal        view.View
	PortalTable   view.View
}

// NewEvaluationModule constructs all evaluation views wired with the injected
// use-case closures.
func NewEvaluationModule(deps *EvaluationModuleDeps) *EvaluationModule {
	listDeps := &evallist.ListViewDeps{
		Routes:          deps.Routes,
		Labels:          deps.Labels,
		CommonLabels:    deps.CommonLabels,
		TableLabels:     deps.TableLabels,
		ListEvaluations: deps.ListEvaluations,
		GetListPageData: deps.GetListPageData,
	}

	detailDeps := &evaldetail.DetailViewDeps{
		Routes:                  deps.Routes,
		Labels:                  deps.Labels,
		CommonLabels:            deps.CommonLabels,
		TableLabels:             deps.TableLabels,
		ReadEvaluation:          deps.ReadEvaluation,
		ListEvaluationResponses: deps.ListEvaluationResponses,
	}

	actionDeps := &evalpkg.ModuleDeps{
		Routes:                      deps.Routes,
		Labels:                      deps.Labels,
		CommonLabels:                deps.CommonLabels,
		TableLabels:                 deps.TableLabels,
		CreateEvaluation:            deps.CreateEvaluation,
		ReadEvaluation:              deps.ReadEvaluation,
		UpdateEvaluation:            deps.UpdateEvaluation,
		DeleteEvaluation:            deps.DeleteEvaluation,
		ListEvaluations:             deps.ListEvaluations,
		GetListPageData:             deps.GetListPageData,
		GetItemPageData:             deps.GetItemPageData,
		SignOffEvaluation:           deps.SignOffEvaluation,
		ArchiveEvaluation:           deps.ArchiveEvaluation,
		ListEvaluationResponses:     deps.ListEvaluationResponses,
		CreateEvaluationResponse:    deps.CreateEvaluationResponse,
		ListEvaluationTemplateItems: deps.ListEvaluationTemplateItems,
		NewID:                       deps.NewID,
	}

	portalGetData := deps.GetPortalPageData
	if portalGetData == nil {
		// Fall back to the list page-data closure (the Integrator may wire a
		// dedicated client-scoped variant); both carry the IDOR predicate.
		portalGetData = deps.GetListPageData
	}
	portalDeps := &evalportal.PortalViewDeps{
		Routes:          deps.Routes,
		Labels:          deps.Labels,
		CommonLabels:    deps.CommonLabels,
		TableLabels:     deps.TableLabels,
		GetListPageData: portalGetData,
	}

	return &EvaluationModule{
		routes:        deps.Routes,
		List:          evallist.NewView(listDeps),
		Table:         evallist.NewTableView(listDeps),
		Detail:        evaldetail.NewView(detailDeps),
		TabAction:     evaldetail.NewTabAction(detailDeps),
		Add:           evalpkg.NewAddAction(actionDeps),
		Edit:          evalpkg.NewEditAction(actionDeps),
		DimensionSlot: evalpkg.NewDimensionSlotAction(actionDeps),
		SignOff:       evalpkg.NewSignOffAction(actionDeps),
		Archive:       evalpkg.NewArchiveAction(actionDeps),
		Delete:        evalpkg.NewDeleteAction(actionDeps),
		BulkArchive:   evalpkg.NewBulkArchiveAction(actionDeps),
		Portal:        evalportal.NewView(portalDeps),
		PortalTable:   evalportal.NewTableView(portalDeps),
	}
}

// RegisterRoutes registers all evaluation routes (page + HTMX + action + portal).
func (m *EvaluationModule) RegisterRoutes(r view.RouteRegistrar) {
	// Page routes (workspace-keyed)
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)

	// HTMX swaps
	r.GET(m.routes.TableURL, m.Table)
	r.GET(m.routes.TabActionURL, m.TabAction)

	// Drawer-form (DF-1) + polymorphic dimension slot
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.GET(m.routes.DimensionSlotURL, m.DimensionSlot)

	// Named lifecycle verbs
	r.POST(m.routes.SignOffURL, m.SignOff)
	r.POST(m.routes.ArchiveURL, m.Archive)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkArchiveURL, m.BulkArchive)

	// Client-portal "Rate My Team"
	r.GET(m.routes.PortalURL, m.Portal)
	r.GET(m.routes.PortalTableURL, m.PortalTable)
}
