package job_template_task

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_template_task",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_template_task"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_template_task.json", Key: "job_template_task"},
		LabelName: "JobTemplateTaskLabels",
		Templates: TemplatesFS,
	}
}
