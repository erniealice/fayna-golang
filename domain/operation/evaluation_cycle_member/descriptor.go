package evaluation_cycle_member

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the compose Unit for the cycle-member surface.
//
// STR-1: evaluation_cycle_member gets the full 14-layer cascade but has NO
// standalone list page and NO routes — it surfaces ONLY via the evaluation_cycle
// detail Members tab + the X-of-Y banner projection. This is therefore a
// data-only / templates-only Unit (Routes nil, no Nav, no Mount): it contributes
// its members-tab.html templates FS + label JSON, nothing more.
func Describe() compose.Unit {
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.evaluation_cycle_member",
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "evaluation_cycle_member.json", Key: "evaluationCycleMember"},
		LabelName: "EvaluationCycleMemberLabels",
		Templates: TemplatesFS,
	}
}
