package outcome_matrix

// routes.go — OutcomeMatrix route constants and Routes config struct.
//
// The matrix is always scoped to one job_template ({id} == job_template_id),
// mirroring grade_sheet.go's subjectFilter. MatrixURL serves both the full page
// and the HTMX content partial; RecordURL is the batch-save POST target.

const (
	MatrixURL = "/outcome-matrix/{id}"
	RecordURL = "/action/outcome-matrix/{id}/record"
)

// Routes holds all route paths for the outcome matrix view.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	MatrixURL string `json:"matrix_url"`
	RecordURL string `json:"record_url"`
}

// DefaultRoutes returns a Routes populated from the package-level route
// constants. ActiveNav "job" matches task_outcome / grade_sheet.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "outcome-matrix",

		MatrixURL: MatrixURL,
		RecordURL: RecordURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_matrix.matrix": r.MatrixURL,
		"outcome_matrix.record": r.RecordURL,
	}
}
