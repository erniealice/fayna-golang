package job_outcome_line

// routes.go — JobOutcomeLine route constants and Routes config struct.

const (
	ListURL       = "/job-outcome-lines/list/{status}"
	DetailURL     = "/job-outcome-lines/detail/{id}"
	AddURL        = "/action/job-outcome-line/add"
	EditURL       = "/action/job-outcome-line/edit/{id}"
	DeleteURL     = "/action/job-outcome-line/delete"
	BulkDeleteURL = "/action/job-outcome-line/bulk-delete"
	TabActionURL  = "/action/job-outcome-line/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the job outcome line views.
type Routes struct {
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`

	TabActionURL string `json:"tab_action_url"`
}

// DefaultRoutes returns a Routes populated from the package-level route constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "job_outcome_lines",

		ListURL:       ListURL,
		DetailURL:     DetailURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		BulkDeleteURL: BulkDeleteURL,

		TabActionURL: TabActionURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_outcome_line.list":        r.ListURL,
		"job_outcome_line.detail":      r.DetailURL,
		"job_outcome_line.add":         r.AddURL,
		"job_outcome_line.edit":        r.EditURL,
		"job_outcome_line.delete":      r.DeleteURL,
		"job_outcome_line.bulk_delete": r.BulkDeleteURL,
		"job_outcome_line.tab_action":  r.TabActionURL,
	}
}
