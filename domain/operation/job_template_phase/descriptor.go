package job_template_phase

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.job_template_phase",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "job_template_phase"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "job_template_phase.json", Key: "job_template_phase"},
		LabelName: "JobTemplatePhaseLabels",
		Templates: TemplatesFS,
	}
}
