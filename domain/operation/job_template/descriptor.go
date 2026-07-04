package job_template

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_template",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_template"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_template.json", Key: "job_template"},
		LabelName: "JobTemplateLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Items: []compose.NavItem{
				{Key: "job-templates", Route: "job_template.list", Params: map[string]string{"status": "active"}, Label: "Templates", Icon: "icon-copy", LabelKey: "job_templates_label", IconKey: "job_templates_icon"},
			},
		},
	}
}
