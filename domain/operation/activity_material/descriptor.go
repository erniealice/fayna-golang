package activity_material

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.activity_material",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "activity_material"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "activity_material.json", Key: "activity_material"},
		LabelName: "ActivityMaterialLabels",
		Templates: TemplatesFS,
	}
}
