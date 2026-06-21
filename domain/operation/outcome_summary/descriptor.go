package outcome_summary

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.outcome_summary",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "outcome_summary"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "outcome_summary.json", Key: "outcome_summary"},
		LabelName: "OutcomeSummaryLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "job_outcome_summary:list",
			Items: []compose.NavItem{
				{Key: "report-cards", Route: "outcome_summary.list", Label: "Outcome Reports", Icon: "icon-bar-chart", Permission: "job_outcome_summary:list"},
			},
		},
	}
}
