package job_template

import "github.com/erniealice/pyeza-golang/compose"

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
	}
}
