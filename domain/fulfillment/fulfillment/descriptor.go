package fulfillment

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "fulfillment.fulfillment",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "fulfillment"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "fulfillment.json", Key: "fulfillment"},
		LabelName: "FulfillmentLabels",
		Templates: TemplatesFS,
	}
}
