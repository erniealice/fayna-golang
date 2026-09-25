package rating_description_set_product_plan

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.rating_description_set_product_plan",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "rating_description_set_product_plan"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "rating_description_set_product_plan.json", Key: "rating_description_set_product_plan"},
		LabelName: "RatingDescriptionSetProductPlanLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "rating_description_set_product_plan:list",
			Items: []compose.NavItem{
				{Key: "rating-description-links", Route: "rating_description_set_product_plan.list", Params: map[string]string{"status": "linked"}, Label: "Descriptor Assignments", Icon: "icon-link", Permission: "rating_description_set_product_plan:list"},
			},
		},
	}
}
