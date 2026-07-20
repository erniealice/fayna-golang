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
)

// Routes holds all route paths for the outcome matrix view.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	MatrixURL string `json:"matrix_url"`
	RecordURL string `json:"record_url"`
	ExportURL string `json:"export_url"`

	// Per-phase approval transition routes.
	SubmitURL  string `json:"submit_url"`
	VerifyURL  string `json:"verify_url"`
	PublishURL string `json:"publish_url"`
	ReturnURL  string `json:"return_url"`
}

// DefaultRoutes returns a Routes populated from the package-level route
// constants. ActiveNav "job" matches task_outcome / grade_sheet.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "outcome-matrix",

		MatrixURL: MatrixURL,
		RecordURL: RecordURL,
		ExportURL: ExportURL,

		SubmitURL:  SubmitURL,
		VerifyURL:  VerifyURL,
		PublishURL: PublishURL,
		ReturnURL:  ReturnURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_matrix.matrix":  r.MatrixURL,
		"outcome_matrix.record":  r.RecordURL,
		"outcome_matrix.export":  r.ExportURL,
		"outcome_matrix.submit":  r.SubmitURL,
		"outcome_matrix.verify":  r.VerifyURL,
		"outcome_matrix.publish": r.PublishURL,
		"outcome_matrix.return":  r.ReturnURL,
	}
}
