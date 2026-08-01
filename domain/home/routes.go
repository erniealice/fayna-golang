package home

// Routes — the home surface's route set. The defaults are the SAME three
// paths the app-local home module registers today
// (apps/*/internal/composition/route_config.go DefaultHomeRoutes), and the
// unit overlays lyngua route.json key "home" exactly like the app-level
// "app.home" nav unit — so the block's route-map contribution merges
// idempotently (same keys, same values) with the app-shell's nav unit and no
// new route surface appears (plan.md §4.4: no new routes).
type Routes struct {
	DashboardURL string `json:"dashboard_url"`
	ContentURL   string `json:"content_url"`
	RibbonURL    string `json:"ribbon_url"`
}

// DefaultRoutes returns the post-P12 bare paths (workspace_path middleware
// dispatches /w/{slug}/home onto them).
func DefaultRoutes() Routes {
	return Routes{
		DashboardURL: "/home",
		ContentURL:   "/home/content",
		RibbonURL:    "/action/home/ribbon",
	}
}

// RouteMap returns dot-notation keys for template route lookups. Keys match
// the app-level HomeRoutes.RouteMap exactly (idempotent merge contract).
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"home.dashboard_url": r.DashboardURL,
		"home.content_url":   r.ContentURL,
		"home.ribbon_url":    r.RibbonURL,
	}
}
