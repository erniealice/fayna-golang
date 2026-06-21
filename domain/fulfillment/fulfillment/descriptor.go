package fulfillment

import "github.com/erniealice/espyna-golang/consumer/compose"

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
		Nav: compose.NavContrib{
			Permission: "fulfillment:list",
			AppEntry: &compose.AppEntry{
				Key:        "fulfillment",
				Route:      "fulfillment.list",
				Label:      "Fulfillment",
				Icon:       "icon-truck",
				Permission: "fulfillment:list",
			},
			Items: []compose.NavItem{
				{Key: "fulfillment-pending", Route: "fulfillment.list", Params: map[string]string{"status": "pending"}, Label: "Pending", Icon: "icon-clock", Permission: "fulfillment:list"},
				{Key: "fulfillment-in-progress", Route: "fulfillment.list", Params: map[string]string{"status": "in_progress"}, Label: "Active", Icon: "icon-truck", Permission: "fulfillment:list"},
				{Key: "fulfillment-delivered", Route: "fulfillment.list", Params: map[string]string{"status": "delivered"}, Label: "Complete", Icon: "icon-check-circle", Permission: "fulfillment:list"},
			},
		},
	}
}
