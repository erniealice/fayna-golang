package home

// Routes is the home surface's route set. The legacy dashboard, content, and
// ribbon paths remain aligned with each app-local home module. The routed-home
// contract adds four explicit section paths; its Lyngua route overlay merges
// idempotently with the app-shell nav unit because both declare the same keys
// and values.
type Routes struct {
	DashboardURL string `json:"dashboard_url"`
	ContentURL   string `json:"content_url"`
	RibbonURL    string `json:"ribbon_url"`

	// Routed-section URLs (20260809). DashboardURL now 302-redirects to
	// OverviewURL; ContentURL is retained for Overview tab switches and for the
	// app-local home modules that still register the overview-content endpoint.
	OverviewURL    string `json:"overview_url"`
	PulseURL       string `json:"pulse_url"`
	PerformanceURL string `json:"performance_url"`
	AttentionURL   string `json:"attention_url"`
}

// DefaultRoutes returns the post-P12 bare paths (workspace_path middleware
// dispatches /w/{slug}/home onto them).
func DefaultRoutes() Routes {
	return Routes{
		DashboardURL:   "/home",
		ContentURL:     "/home/content",
		RibbonURL:      "/action/home/ribbon",
		OverviewURL:    "/home/overview",
		PulseURL:       "/home/pulse",
		PerformanceURL: "/home/performance",
		AttentionURL:   "/home/attention",
	}
}

// RouteMap returns dot-notation keys for template route lookups. Keys match
// the app-level HomeRoutes.RouteMap exactly (idempotent merge contract).
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"home.dashboard_url":   r.DashboardURL,
		"home.content_url":     r.ContentURL,
		"home.ribbon_url":      r.RibbonURL,
		"home.overview_url":    r.OverviewURL,
		"home.pulse_url":       r.PulseURL,
		"home.performance_url": r.PerformanceURL,
		"home.attention_url":   r.AttentionURL,
	}
}
