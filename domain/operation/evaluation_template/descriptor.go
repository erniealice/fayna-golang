package evaluation_template

import "github.com/erniealice/pyeza-golang/compose"

// Describe returns the compose Unit for the evaluation_template module.
// Staff-only authoring surface — Nav permission is evaluation_template:list
// (clients have NO evaluation_template:* per acceptance #8).
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.evaluation_template",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "evaluation_template"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "evaluation_template.json", Key: "evaluationTemplate"},
		LabelName: "EvaluationTemplateLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "evaluation_template:list",
			Items: []compose.NavItem{
				{Key: "evaluation_template", Route: "evaluation_template.list", Params: map[string]string{"status": "active"}, Label: "Evaluation Templates", Icon: "icon-clipboard", Permission: "evaluation_template:list"},
			},
		},
	}
}
