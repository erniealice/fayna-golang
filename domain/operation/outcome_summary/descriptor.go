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
				// TB3 (D3): a dedicated template-management settings entry
				// positioned immediately AFTER the reports item. Generic key +
				// route; the "Report Card Templates" wording lives in lyngua
				// (template_settings_label / education tier).
				{Key: "outcome-summary-templates", Route: "outcome_summary.template_settings", Label: "Report Templates", Icon: "icon-file-text", Permission: "job_outcome_summary:list", LabelKey: "template_settings_label", IconKey: "template_settings_icon"},
			},
		},
	}
}
