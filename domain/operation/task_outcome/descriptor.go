package task_outcome

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.task_outcome",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "task_outcome"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "task_outcome.json", Key: "task_outcome"},
		LabelName: "TaskOutcomeLabels",
		Templates: TemplatesFS,
	}
}
