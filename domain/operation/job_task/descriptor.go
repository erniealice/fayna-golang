package job_task

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_task",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_task"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_task.json", Key: "job_task"},
		LabelName: "JobTaskLabels",
		Templates: TemplatesFS,
	}
}
