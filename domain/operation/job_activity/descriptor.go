package job_activity

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_activity",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_activity"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_activity.json", Key: "job_activity"},
		LabelName: "JobActivityLabels",
		Templates: TemplatesFS,
	}
}
