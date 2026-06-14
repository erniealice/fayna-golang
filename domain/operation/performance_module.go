package operation

import (
	"context"

	performancepkg "github.com/erniealice/fayna-golang/domain/operation/performance"
	performancelist "github.com/erniealice/fayna-golang/domain/operation/performance/list"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// PerformanceModuleDeps holds all dependencies for the performance admin panel
// module (Surface 6). The single page view is gated server-side (CR-5) inside
// the block-supplied GetPanelData closure — the view supplies no
// client_id/subscription_id. The view layer never calls espyna directly.
type PerformanceModuleDeps struct {
	Routes       performancepkg.Routes
	Labels       performancepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// GetPanelData is the block-supplied adapter over espyna's servicing-gated
	// GetPerformancePanelData (CR-5). Returns the view-local PanelData (espyna's
	// internal projection is not importable). nil → empty-state render.
	GetPanelData func(ctx context.Context) (*performancepkg.PanelData, error)
}

// PerformanceModule holds the constructed performance panel view.
type PerformanceModule struct {
	routes performancepkg.Routes
	Panel  view.View
}

// NewPerformanceModule constructs the performance panel module.
func NewPerformanceModule(deps *PerformanceModuleDeps) *PerformanceModule {
	listDeps := &performancelist.ListViewDeps{
		Routes:       deps.Routes,
		Labels:       deps.Labels,
		CommonLabels: deps.CommonLabels,
		TableLabels:  deps.TableLabels,
		GetPanelData: deps.GetPanelData,
	}
	return &PerformanceModule{
		routes: deps.Routes,
		Panel:  performancelist.NewView(listDeps),
	}
}

// RegisterRoutes registers the single performance dashboard page route.
func (m *PerformanceModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.DashboardURL, m.Panel)
}
