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
		Nav: compose.NavContrib{
			Permission: "outcome_criteria:list",
			Items: []compose.NavItem{
				{Key: "criteria", Route: "outcome_criteria.list", Params: map[string]string{"status": "active"}, Label: "Criteria", Icon: "icon-check-square", Permission: "outcome_criteria:list"},
			},
		},
	}
}
