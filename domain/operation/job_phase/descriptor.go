package job_phase

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_phase",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_phase"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_phase.json", Key: "job_phase"},
		LabelName: "JobPhaseLabels",
		Templates: TemplatesFS,
	}
}
