package evaluation_cycle

import "github.com/erniealice/pyeza-golang/compose"

// Describe returns the compose Unit for the evaluation_cycle module.
// Key = operation.evaluation_cycle. Lyngua root key is camelCase
// `evaluationCycle` (LBL trap). Sidebar group: "Performance".
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.evaluation_cycle",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "evaluation_cycle"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "evaluation_cycle.json", Key: "evaluationCycle"},
		LabelName: "EvaluationCycleLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "evaluation_cycle:list",
			Items: []compose.NavItem{
				{Key: "evaluation-cycles", Route: "evaluation_cycle.list", Params: map[string]string{"status": "open"}, Label: "Evaluation Cycles", Icon: "icon-refresh-cw", Permission: "evaluation_cycle:list"},
			},
		},
	}
}
