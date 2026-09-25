package rating_description_set

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.rating_description_set",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "rating_description_set"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "rating_description_set.json", Key: "rating_description_set"},
		LabelName: "RatingDescriptionSetLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "rating_description_set:list",
			Items: []compose.NavItem{
				{Key: "rating-description-sets", Route: "rating_description_set.list", Params: map[string]string{"status": "active"}, Label: "Rubric Descriptors", Icon: "icon-file-text", Permission: "rating_description_set:list"},
			},
		},
	}
}
