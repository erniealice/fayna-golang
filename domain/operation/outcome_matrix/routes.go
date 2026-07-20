package outcome_matrix

// routes.go — OutcomeMatrix route constants and Routes config struct.
//
// The matrix is always scoped to one job_template ({id} == job_template_id),
// mirroring grade_sheet.go's subjectFilter. MatrixURL serves both the full page
// and the HTMX content partial; RecordURL is the batch-save POST target.

const (
	MatrixURL = "/outcome-matrix/{id}"
	RecordURL = "/action/outcome-matrix/{id}/record"

	// Sheet-level CSV download. A GET sibling of MatrixURL (NOT under /action/*
	// — safe method, no state change, so neither the CSRF hook nor the
	// action-workspace signature guard applies; the raw-handler registration
	// wraps it with the ViewAdapter's RBAC context, same as outcome_summary's
	// SectionExportURL). Honors the same ?scope= and ?hide= as the HTML view.
	ExportURL = "/outcome-matrix/{id}/export"

	// Export drawer (20260720 Q3): the GET view rendering the download form
	// (period + format selects + hidden scope/hide) into #sheetContent. It sits
	// under /action/* for slug tidiness, but it is a GET (safe method): the CSRF
	// hook and the action-workspace signature guard constrain NON-safe methods
	// only, so a GET drawer needs no signed form (same trust model as the raw
	// GET ExportURL). Registered with r.GET, not the action-POST path.
	DownloadDrawerURL = "/action/outcome-matrix/{id}/download"

	// Per-phase approval transition POST targets ({id} == job_template_id; the
	// job_template_phase_id rides in the form body). Each is a real, query-free
	// HTMX POST form signed by {{actionForm}} over its exact resolved path (the
	// landed Q7 rowActionTokens table path does NOT cover these bar buttons —
	// codex fresh finding). All live under /action/* so the CSRF + action-
	// workspace guards apply exactly as for RecordURL.
	SubmitURL  = "/action/outcome-matrix/{id}/submit"
	VerifyURL  = "/action/outcome-matrix/{id}/verify"
	PublishURL = "/action/outcome-matrix/{id}/publish"
	ReturnURL  = "/action/outcome-matrix/{id}/return"

	// TemplateSettingsURL is the standalone grade-sheet template management page
	// (Wave C / P4): a list of workspace→document_template bindings (scoped by
	// job_category + optional price_schedule) + upload/publish/delete. A dedicated
	// settings surface (Q5), NOT a tab on the grid. GET only — no mutation, so it
	// stays OUTSIDE /action/ (safe method). Education overrides it to
	// /grade-sheet/templates via education/route.json. Mirrors the JOSDT
	// outcome_summary.TemplateSettingsURL shape.
	TemplateSettingsURL = "/outcome-matrix/templates"
	// The three template MUTATIONS live under /action/* so they inherit the CSRF
	// validator + signed workspace-form guard (both default-scoped to /action/ in
	// espyna's middleware chain). Registering them elsewhere silently bypasses both
	// guards (JOSDT B4 precedent). Education keeps its /grade-sheet/templates
	// display vocabulary UNDER the /action/ prefix (see education/route.json).
	//
	// TemplateUploadURL is the upload drawer (GET = form, POST = create a DRAFT
	// binding + its document_template artifact). Education →
	// /action/grade-sheet/templates/upload.
	TemplateUploadURL = "/action/outcome-matrix/templates/upload"
	// TemplatePublishURL publishes a DRAFT binding (id in ?id= query, appended by
	// the table row-action JS) via the controlled publish transaction. Flat (no
	// path param) so it composes with the generic "activate" row action.
	// Education → /action/grade-sheet/templates/publish.
	TemplatePublishURL = "/action/outcome-matrix/templates/publish"
	// TemplateDeleteURL deletes a DRAFT binding (POST, id in form). Education →
	// /action/grade-sheet/templates/delete.
	TemplateDeleteURL = "/action/outcome-matrix/templates/delete"
)

// Routes holds all route paths for the outcome matrix view.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	MatrixURL         string `json:"matrix_url"`
	RecordURL         string `json:"record_url"`
	ExportURL         string `json:"export_url"`
	DownloadDrawerURL string `json:"download_drawer_url"`

	// Per-phase approval transition routes.
	SubmitURL  string `json:"submit_url"`
	VerifyURL  string `json:"verify_url"`
	PublishURL string `json:"publish_url"`
	ReturnURL  string `json:"return_url"`

	// Grade-sheet template settings (P4 management surface).
	TemplateSettingsURL string `json:"template_settings_url"`
	TemplateUploadURL   string `json:"template_upload_url"`
	TemplatePublishURL  string `json:"template_publish_url"`
	TemplateDeleteURL   string `json:"template_delete_url"`
}

// DefaultRoutes returns a Routes populated from the package-level route
// constants. ActiveNav "job" matches task_outcome / grade_sheet.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "outcome-matrix",

		MatrixURL:         MatrixURL,
		RecordURL:         RecordURL,
		ExportURL:         ExportURL,
		DownloadDrawerURL: DownloadDrawerURL,

		SubmitURL:  SubmitURL,
		VerifyURL:  VerifyURL,
		PublishURL: PublishURL,
		ReturnURL:  ReturnURL,

		TemplateSettingsURL: TemplateSettingsURL,
		TemplateUploadURL:   TemplateUploadURL,
		TemplatePublishURL:  TemplatePublishURL,
		TemplateDeleteURL:   TemplateDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_matrix.matrix":          r.MatrixURL,
		"outcome_matrix.record":          r.RecordURL,
		"outcome_matrix.export":          r.ExportURL,
		"outcome_matrix.download_drawer": r.DownloadDrawerURL,
		"outcome_matrix.submit":          r.SubmitURL,
		"outcome_matrix.verify":          r.VerifyURL,
		"outcome_matrix.publish":         r.PublishURL,
		"outcome_matrix.return":          r.ReturnURL,

		"outcome_matrix.template_settings": r.TemplateSettingsURL,
		"outcome_matrix.template_upload":   r.TemplateUploadURL,
		"outcome_matrix.template_publish":  r.TemplatePublishURL,
		"outcome_matrix.template_delete":   r.TemplateDeleteURL,
	}
}
