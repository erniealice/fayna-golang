package job_outcome_line

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_outcome_line",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_outcome_line"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_outcome_line.json", Key: "job_outcome_line"},
		LabelName: "JobOutcomeLineLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "job_outcome_line:list",
			Items: []compose.NavItem{
				{Key: "job_outcome_lines", Route: "job_outcome_line.list", Params: map[string]string{"status": "active"}, Label: "Outcome Lines", Icon: "icon-list", Permission: "job_outcome_line:list"},
			},
		},
	}
}
