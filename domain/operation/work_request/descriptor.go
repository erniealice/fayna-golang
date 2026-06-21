package work_request

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the compose Unit for the work_request entity.
// The engine overlays lyngua JSON for routes and labels, then calls Mount
// (set by the block catalog) in phase 2.
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.work_request",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "workRequest"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "work_request.json", Key: "workRequest"},
		LabelName: "WorkRequestLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "work_request:list",
			AppEntry: &compose.AppEntry{
				Key:        "requests",
				Route:      "work_request.list",
				Params:     map[string]string{"status": "open"},
				Label:      "Requests",
				Icon:       "icon-inbox",
				Permission: "work_request:list",
			},
			Items: []compose.NavItem{
				{Key: "requests-open", Route: "work_request.list", Params: map[string]string{"status": "open"}, Label: "Open", Icon: "icon-inbox"},
				{Key: "requests-in-review", Route: "work_request.list", Params: map[string]string{"status": "in-review"}, Label: "In Review", Icon: "icon-eye"},
				{Key: "requests-approved", Route: "work_request.list", Params: map[string]string{"status": "approved"}, Label: "Approved", Icon: "icon-check-circle"},
				{Key: "requests-completed", Route: "work_request.list", Params: map[string]string{"status": "completed"}, Label: "Completed", Icon: "icon-check"},
			},
		},
	}
}
