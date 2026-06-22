package score_scale

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.score_scale",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "score_scale"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "score_scale.json", Key: "score_scale"},
		LabelName: "ScoreScaleLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "score_scale:list",
			Items: []compose.NavItem{
				{Key: "score_scales", Route: "score_scale.list", Params: map[string]string{"status": "active"}, Label: "Score Scales", Icon: "icon-bar-chart-2", Permission: "score_scale:list"},
			},
		},
	}
}
