package activity_expense

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.activity_expense",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "activity_expense"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "activity_expense.json", Key: "activity_expense"},
		LabelName: "ActivityExpenseLabels",
		Templates: TemplatesFS,
	}
}
