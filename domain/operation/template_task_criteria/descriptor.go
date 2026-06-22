package template_task_criteria

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.template_task_criteria",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "template_task_criteria"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "template_task_criteria.json", Key: "template_task_criteria"},
		LabelName: "TemplateTaskCriteriaLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "template_task_criteria:list",
			Items: []compose.NavItem{
				{Key: "template_task_criteria", Route: "template_task_criteria.list", Params: map[string]string{"status": "active"}, Label: "Template Task Criteria", Icon: "icon-link", Permission: "template_task_criteria:list"},
			},
		},
	}
}
