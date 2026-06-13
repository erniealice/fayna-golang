package outcome_summary

import "github.com/erniealice/pyeza-golang/compose"

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
	}
}
