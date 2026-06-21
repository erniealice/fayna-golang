package work_request_type

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the compose Unit for the work_request_type catalog entity.
// The engine overlays lyngua JSON for routes and labels, then calls Mount
// (set by the block catalog) in phase 2.
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.work_request_type",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "workRequestType"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "work_request_type.json", Key: "workRequestType"},
		LabelName: "WorkRequestTypeLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "work_request_type:list",
			Items: []compose.NavItem{
				{Key: "request-types", Route: "work_request_type.list", Params: map[string]string{"status": "active"}, Label: "Request Types", Icon: "icon-list", Permission: "work_request_type:list"},
			},
		},
	}
}
