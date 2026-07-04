package task_outcome

import "github.com/erniealice/espyna-golang/consumer/compose"

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
		Nav: compose.NavContrib{
			Permission: "task_outcome:list",
			Items: []compose.NavItem{
				{Key: "outcome-log", Route: "task_outcome.list", Params: map[string]string{"status": "active"}, Label: "Outcome Log", Icon: "icon-clipboard", Permission: "task_outcome:list", LabelKey: "outcome_log_label", IconKey: "outcome_log_icon"},
			},
		},
	}
}
