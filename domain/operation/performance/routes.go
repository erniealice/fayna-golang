package performance

// routes.go — Performance admin-panel route constants and Routes config struct.
//
// Surface 6 (pages.md §G): a single composition page — the admin Performance
// panel (data-page="performance"). It is NOT a CRUD entity: no list/detail/add.
// The one page route key is `performance.dashboard` (STR-4: dot.snake_case on the
// proto-domain name "performance", never a lyngua skin).
//
// Page prefix: /app/performance (workspace-keyed to /w/{ws}/performance by the
// workspace-path middleware at the route_config layer).
// The per-row Start-review / View-last-review CTAs reuse the `evaluation.*` action
// + page routes (owned by the evaluation unit) — they are NOT redefined here.

// Default route constants for the performance panel.
const (
	DashboardURL = "/app/performance"
)

// Routes holds the route paths for the performance panel.
type Routes struct {
	// Sidebar navigation context.
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	DashboardURL string `json:"dashboard_url"`

	// Cross-unit routes the panel rows link to (owned by the evaluation unit).
	// Populated by the block so the view can resolve the per-row CTAs without
	// importing the evaluation package. Templates: {id}/{seat_id} substitution.
	EvaluationAddURL    string `json:"evaluation_add_url"`    // /action/evaluation/add (?seat_id=)
	EvaluationDetailURL string `json:"evaluation_detail_url"` // /app/evaluations/detail/{id}
}

// DefaultRoutes returns a Routes populated from the package-level route constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "performance",
		ActiveSubNav: "performance",

		DashboardURL: DashboardURL,

		EvaluationAddURL:    "/action/evaluation/add",
		EvaluationDetailURL: "/app/evaluations/detail/{id}",
	}
}

// RouteMap returns the dot-notation route keys for the performance panel.
// Only `performance.dashboard` is owned here (the CTA routes belong to the
// evaluation unit's RouteMap).
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"performance.dashboard": r.DashboardURL,
	}
}
