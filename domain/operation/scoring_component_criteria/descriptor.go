package scoring_component_criteria

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.scoring_component_criteria",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "scoring_component_criteria"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "scoring_component_criteria.json", Key: "scoring_component_criteria"},
		LabelName: "ScoringComponentCriteriaLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "scoring_component_criteria:list",
			Items: []compose.NavItem{
				{Key: "scoring_component_criteria", Route: "scoring_component_criteria.list", Params: map[string]string{"status": "active"}, Label: "Scoring Component Criteria", Icon: "icon-link", Permission: "scoring_component_criteria:list"},
			},
		},
	}
}
