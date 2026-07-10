package outcome_matrix

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe is the compose-v2 self-description for the outcome matrix Unit.
// block/catalog.go's OutcomeMatrixUnit(uc, infra) calls this Describe() and
// wires GetOutcomeMatrix from espyna's Service aggregate.
//
// The Nav entry is intentionally omitted: outcome_matrix.matrix is a
// parameterized route ({id} == job_template_id) with no fixed landing target,
// so it is reached from a job_template context, not a standalone sidebar item.
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.outcome_matrix",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "outcome_matrix"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "outcome_matrix.json", Key: "outcome_matrix"},
		LabelName: "OutcomeMatrixLabels",
		Templates: TemplatesFS,
	}
}
