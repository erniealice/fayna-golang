package score_scale_band

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.score_scale_band",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "score_scale_band"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "score_scale_band.json", Key: "score_scale_band"},
		LabelName: "ScoreScaleBandLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "score_scale_band:list",
			Items: []compose.NavItem{
				{Key: "score_scale_bands", Route: "score_scale_band.list", Params: map[string]string{"status": "active"}, Label: "Scale Bands", Icon: "icon-sliders", Permission: "score_scale_band:list"},
			},
		},
	}
}
