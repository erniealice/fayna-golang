package job

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job.json", Key: "job"},
		LabelName: "JobLabels",
		Templates: TemplatesFS,
	}
}
