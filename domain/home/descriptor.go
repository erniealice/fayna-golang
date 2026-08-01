package home

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the home surface's compose unit. It contributes the SAME
// route-map keys/values the app-level "app.home" nav unit contributes
// (idempotent MergeFrom), carries the package templates, and deliberately
// contributes NO Nav — the app-shell's "app.home" unit keeps sidebar
// ownership (plan.md §4.4). The Mount closure is bound by block/homeblock.go
// (the self-contained catalog for this block).
func Describe() compose.Unit {
	r := DefaultRoutes()
	return compose.Unit{
		Key:       "home.dashboard",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "home"},
		Templates: TemplatesFS,
	}
}
