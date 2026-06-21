package performance

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the compose Unit for the performance admin panel (Surface 6).
// The engine overlays lyngua JSON (route.json key "performance"; performance.json
// camelCase root "performance") onto the defaults, then the block sets Mount in
// phase 2. Single page route key: performance.dashboard.
//
// Nav: a single AppEntry "Performance" (data-page="performance" surface) gated on
// evaluation:dashboard (the panel-read permission, pages.md §G.1) — NOT a
// performance:* permission, which does not exist. Staff-only; CR-5 servicing
// gating is enforced server-side, the nav is NOT an auth boundary.
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.performance",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "performance"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "performance.json", Key: "performance"},
		LabelName: "PerformancePanelLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "evaluation:dashboard",
			AppEntry: &compose.AppEntry{
				Key:        "performance",
				Route:      "performance.dashboard",
				Label:      "Performance",
				Icon:       "icon-bar-chart",
				Permission: "evaluation:dashboard",
			},
		},
	}
}
