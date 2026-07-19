package outcome_summary

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.outcome_summary",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "outcome_summary"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "outcome_summary.json", Key: "outcome_summary"},
		LabelName: "OutcomeSummaryLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "job_outcome_summary:list",
			Items: []compose.NavItem{
				{Key: "report-cards", Route: "outcome_summary.list", Label: "Outcome Reports", Icon: "icon-bar-chart", Permission: "job_outcome_summary:list", LabelKey: "job_outcome_summary_label", IconKey: "job_outcome_summary_icon"},
				// Activeness-scoped facets of the report-cards landing: {scope}
				// filters the price_schedule (period) tabs by the generic
				// price_schedule.active flag — current = active schedules, past =
				// inactive. Generic entity-named label keys (job_outcome_summary_*);
				// the "Current"/"Past" wording lives in lyngua. An app whose sidebar
				// does not Pick these leaves them offered-but-unrendered.
				{Key: "report-cards-current", Route: "outcome_summary.list_scope", Params: map[string]string{"scope": "current"}, Label: "Current", Icon: "icon-calendar", Permission: "job_outcome_summary:list", LabelKey: "job_outcome_summary_current_label", IconKey: "job_outcome_summary_current_icon"},
				{Key: "report-cards-past", Route: "outcome_summary.list_scope", Params: map[string]string{"scope": "past"}, Label: "Past", Icon: "icon-clock", Permission: "job_outcome_summary:list", LabelKey: "job_outcome_summary_past_label", IconKey: "job_outcome_summary_past_icon"},
				// TB3 (D3): a dedicated template-management settings entry
				// positioned immediately AFTER the reports item. Generic key +
				// route; the "Report Card Templates" wording lives in lyngua
				// (template_settings_label / education tier).
				// Q4: gated on the binding entity's OWN list code — the same code
				// the settings page GET and its list use case enforce — so the
				// sidebar entry, the page gate, and the Gatekeeper agree for
				// split-role users.
				{Key: "outcome-summary-templates", Route: "outcome_summary.template_settings", Label: "Report Templates", Icon: "icon-file-text", Permission: "job_outcome_summary_document_template:list", LabelKey: "template_settings_label", IconKey: "template_settings_icon"},
			},
		},
	}
}
