package job_template_relation

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_template_relation",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_template_relation"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_template_relation.json", Key: "job_template_relation"},
		LabelName: "JobTemplateRelationLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "job_template_relation:list",
			Items: []compose.NavItem{
				{Key: "job_template_relation", Route: "job_template_relation.list", Params: map[string]string{"status": "active"}, Label: "Template Relations", Icon: "icon-git-branch", Permission: "job_template_relation:list"},
			},
		},
	}
}
