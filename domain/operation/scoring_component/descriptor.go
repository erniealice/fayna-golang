package scoring_component

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.scoring_component",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "scoring_component"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "scoring_component.json", Key: "scoring_component"},
		LabelName: "ScoringComponentLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "scoring_component:list",
			Items: []compose.NavItem{
				{Key: "scoring_components", Route: "scoring_component.list", Params: map[string]string{"status": "active"}, Label: "Scoring Components", Icon: "icon-bar-chart-2", Permission: "scoring_component:list"},
			},
		},
	}
}
