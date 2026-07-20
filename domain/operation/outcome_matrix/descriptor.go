package outcome_matrix

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe is the compose-v2 self-description for the outcome matrix Unit.
// block/catalog.go's OutcomeMatrixUnit(uc, infra) calls this Describe() and
// wires GetOutcomeMatrix from espyna's Service aggregate.
//
// The matrix.matrix route is a parameterized route ({id} == job_template_id)
// with no fixed landing target, so it exposes no sidebar item — it is reached
// from a job_template context. The Nav DOES carry a single fixed-target item:
// the grade-sheet template-management settings surface (Wave C / P4), a
// standalone GET page. Its wording ("Grade Sheet Templates" / "Sheet Templates")
// lives only in lyngua sidebar.json; the key + route are generic. Mirrors the
// JOSDT outcome_summary "outcome-summary-templates" NavItem.
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
		Nav: compose.NavContrib{
			Permission: "job_template_document_template:list",
			Items: []compose.NavItem{
				// Grade-sheet template settings (Wave C / P4): a dedicated
				// template-management settings entry. Generic key + route; the
				// "Grade Sheet Templates"/"Sheet Templates" wording lives in lyngua
				// (sheet_template_settings_label / education tier). Gated on the
				// binding entity's OWN list code — the same code the settings page
				// GET and its list use case enforce — so the sidebar entry, the page
				// gate, and the Gatekeeper agree for split-role users (JOSDT Q4).
				{Key: "outcome-matrix-templates", Route: "outcome_matrix.template_settings", Label: "Sheet Templates", Icon: "icon-file-text", Permission: "job_template_document_template:list", LabelKey: "sheet_template_settings_label", IconKey: "sheet_template_settings_icon"},
			},
		},
	}
}
