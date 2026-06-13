package activity_labor

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.activity_labor",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "activity_labor"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "activity_labor.json", Key: "activity_labor"},
		LabelName: "ActivityLaborLabels",
		Templates: TemplatesFS,
	}
}
