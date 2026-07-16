package outcome_summary

// routes.go — OutcomeSummary route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Outcome Summary routes (report cards)
const (
	ListURL = "/outcomes/summaries"
	// ListScopeURL is the activeness-scoped report-cards landing: {scope} ∈
	// {current, past} filters the price_schedule tabs by the GENERIC
	// price_schedule.active flag (current = active schedules, past = inactive).
	// It renders the SAME tabbed section landing as ListURL, restricted to one
	// activeness band; an empty/unknown scope degrades to the unfiltered landing.
	// Education overrides it to /report-cards/list/{scope}.
	ListScopeURL = "/outcomes/summaries/list/{scope}"
	JobURL       = "/jobs/detail/{id}/summary"
	// SectionURL is the per-section report-card grid (view-2): {id} is a
	// subscription_group id. Generic default here; the education tier overrides
	// it to /report-cards/section/{id} via education/route.json.
	SectionURL = "/outcomes/summaries/section/{id}"
	// SectionExportURL serves the per-section report-card grid as a CSV
	// download ({id} = subscription_group id; optional ?id=<client id> narrows
	// to one row). Education overrides it to /report-cards/section/{id}/export.
	SectionExportURL = "/outcomes/summaries/section/{id}/export"
	// ClientCardURL is the per-client report card (view-3): {id} = subscription_group
	// id, {client_id} = the client id. The generic "client" path noun is
	// lyngua-fied to "student" on education
	// (/report-cards/section/{id}/student/{client_id}).
	ClientCardURL = "/outcomes/summaries/section/{id}/client/{client_id}"
	// ClientDocumentURL streams the per-client report card as a .docx download
	// ({id} = subscription_group id, {client_id} = the client id).
	// Education overrides it to /report-cards/section/{id}/student/{client_id}/document.
	ClientDocumentURL = "/outcomes/summaries/section/{id}/client/{client_id}/document"
	// GroupDetailURL is the listed entity's own detail page (the
	// subscription_group detail — centymo's mount), used by the section view's
	// header caption link. Kept as a Routes field so the per-tier route.json
	// override (education: /sections/detail/{id}) rides the same binding as
	// every other route here; the default mirrors centymo's generic constant.
	GroupDetailURL = "/subscription-groups/detail/{id}"
	PhaseURL       = "/jobs/detail/{id}/phase/{phase_id}/summary"

	// TemplateSettingsURL is the standalone report-card template management page
	// (TB3): list of AY→document_template bindings + upload/publish/delete. A
	// dedicated settings surface (D3), NOT a tab on the landing. GET only — no
	// mutation, so it stays outside /action/. Education overrides it to
	// /report-cards/templates via education/route.json.
	TemplateSettingsURL = "/outcomes/summaries/templates"
	// The three template MUTATIONS live under /action/* so they inherit the CSRF
	// validator + signed workspace-form guard (both default-scoped to /action/ in
	// espyna's middleware chain). Registering them elsewhere silently bypasses
	// both guards (B4 codex finding #1). Education keeps its /report-cards/templates
	// display vocabulary UNDER the /action/ prefix (see education/route.json).
	//
	// TemplateUploadURL is the upload drawer (GET = form, POST = create a DRAFT
	// binding + its document_template artifact). Education →
	// /action/report-cards/templates/upload.
	TemplateUploadURL = "/action/outcome-summary/templates/upload"
	// TemplatePublishURL publishes a DRAFT binding (id in ?id= query, appended by
	// the table row-action JS) via the controlled publish transaction. Flat (no
	// path param) so it composes with the generic "activate" row action.
	// Education → /action/report-cards/templates/publish.
	TemplatePublishURL = "/action/outcome-summary/templates/publish"
	// TemplateDeleteURL deletes a binding (POST, id in form). Education →
	// /action/report-cards/templates/delete.
	TemplateDeleteURL = "/action/outcome-summary/templates/delete"
)

// Routes holds all route paths for outcome summary (report card) views.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	// ListActiveSubNav overrides ActiveSubNav for the standalone list page.
	// Job/phase summary pages highlight "jobs" while the list page highlights "report-cards".
	ListActiveSubNav string `json:"list_active_sub_nav"`

	ListURL           string `json:"list_url"`
	ListScopeURL      string `json:"list_scope_url"`
	JobSummaryURL     string `json:"job_summary_url"`
	SectionURL        string `json:"section_url"`
	SectionExportURL  string `json:"section_export_url"`
	ClientCardURL     string `json:"client_url"`
	ClientDocumentURL string `json:"client_document_url"`
	GroupDetailURL    string `json:"group_detail_url"`
	PhaseSummaryURL   string `json:"phase_summary_url"`

	// Report-card template settings (TB3 management surface).
	TemplateSettingsURL string `json:"template_settings_url"`
	TemplateUploadURL   string `json:"template_upload_url"`
	TemplatePublishURL  string `json:"template_publish_url"`
	TemplateDeleteURL   string `json:"template_delete_url"`
}

// DefaultRoutes returns a Routes populated from
// the package-level route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:        "job",
		ActiveSubNav:     "jobs",
		ListActiveSubNav: "report-cards",

		ListURL:           ListURL,
		ListScopeURL:      ListScopeURL,
		JobSummaryURL:     JobURL,
		SectionURL:        SectionURL,
		SectionExportURL:  SectionExportURL,
		ClientCardURL:     ClientCardURL,
		ClientDocumentURL: ClientDocumentURL,
		GroupDetailURL:    GroupDetailURL,
		PhaseSummaryURL:   PhaseURL,

		TemplateSettingsURL: TemplateSettingsURL,
		TemplateUploadURL:   TemplateUploadURL,
		TemplatePublishURL:  TemplatePublishURL,
		TemplateDeleteURL:   TemplateDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// outcome summary routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_summary.list":            r.ListURL,
		"outcome_summary.list_scope":      r.ListScopeURL,
		"outcome_summary.job":             r.JobSummaryURL,
		"outcome_summary.section":         r.SectionURL,
		"outcome_summary.section_export":  r.SectionExportURL,
		"outcome_summary.client_card":     r.ClientCardURL,
		"outcome_summary.client_document": r.ClientDocumentURL,
		"outcome_summary.group_detail":    r.GroupDetailURL,
		"outcome_summary.phase":           r.PhaseSummaryURL,

		"outcome_summary.template_settings": r.TemplateSettingsURL,
		"outcome_summary.template_upload":   r.TemplateUploadURL,
		"outcome_summary.template_publish":  r.TemplatePublishURL,
		"outcome_summary.template_delete":   r.TemplateDeleteURL,
	}
}
