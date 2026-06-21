package evaluation

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the compose Unit for the evaluation entity.
//
// The engine overlays lyngua JSON (route.json key "evaluation"; evaluation.json
// root key "evaluation" — camelCase per LBL trap) onto Routes/Labels, then the
// block catalog (Integrator) calls Mount in phase 2 to wire the use-case
// closures (ModuleDeps) + register routes.
//
// Nav: the staff "Reviews" app entry is permission-gated on evaluation:list.
// The client-portal "Performance Reviews" entry (Q-PORTAL-2) is wired
// separately by the Integrator via portalSidebarBuilders — NOT here.
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.evaluation",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "evaluation"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "evaluation.json", Key: "evaluation"},
		LabelName: "EvaluationLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "evaluation:list",
			AppEntry: &compose.AppEntry{
				Key:        "reviews",
				Route:      "evaluation.list",
				Params:     map[string]string{"status": "submitted"},
				Label:      "Reviews",
				Icon:       "icon-clipboard-check",
				Permission: "evaluation:list",
			},
			Items: []compose.NavItem{
				{Key: "reviews-submitted", Route: "evaluation.list", Params: map[string]string{"status": "submitted"}, Label: "Submitted", Icon: "icon-clipboard-check", Permission: "evaluation:list"},
				{Key: "reviews-signed-off", Route: "evaluation.list", Params: map[string]string{"status": "signed_off"}, Label: "Signed Off", Icon: "icon-check-circle", Permission: "evaluation:list"},
				{Key: "reviews-draft", Route: "evaluation.list", Params: map[string]string{"status": "draft"}, Label: "Draft", Icon: "icon-edit", Permission: "evaluation:list"},
			},
		},
	}
}
