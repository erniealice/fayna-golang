package evaluation_template_item

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the compose Unit for the rubric-item drawer module.
// No Nav — the item has no standalone page; it surfaces only via the
// evaluation_template detail Items tab + this drawer.
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.evaluation_template_item",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "evaluation_template_item"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "evaluation_template_item.json", Key: "evaluation_template_item"},
		LabelName: "EvaluationTemplateItemLabels",
		Templates: TemplatesFS,
	}
}
