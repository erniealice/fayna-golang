package outcome_criteria

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.outcome_criteria",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "outcome_criteria"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "outcome_criteria.json", Key: "outcome_criteria"},
		LabelName: "OutcomeCriteriaLabels",
		Templates: TemplatesFS,
	}
}
