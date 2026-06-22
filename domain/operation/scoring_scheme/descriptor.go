package scoring_scheme

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.scoring_scheme",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "scoring_scheme"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "scoring_scheme.json", Key: "scoring_scheme"},
		LabelName: "ScoringSchemeLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "scoring_scheme:list",
			Items: []compose.NavItem{
				{Key: "scoring_schemes", Route: "scoring_scheme.list", Params: map[string]string{"status": "active"}, Label: "Scoring Schemes", Icon: "icon-sliders", Permission: "scoring_scheme:list"},
			},
		},
	}
}
