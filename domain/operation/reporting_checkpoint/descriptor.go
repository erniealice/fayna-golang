package reporting_checkpoint

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.reporting_checkpoint",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "reporting_checkpoint"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "reporting_checkpoint.json", Key: "reporting_checkpoint"},
		LabelName: "ReportingCheckpointLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "reporting_checkpoint:list",
			Items: []compose.NavItem{
				{Key: "reporting_checkpoints", Route: "reporting_checkpoint.list", Params: map[string]string{"status": "active"}, Label: "Reporting Checkpoints", Icon: "icon-flag", Permission: "reporting_checkpoint:list"},
			},
		},
	}
}
