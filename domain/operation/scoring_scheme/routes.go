package scoring_scheme

// Routes holds all route paths for the scoring scheme views.
const (
	ListURL       = "/scoring-schemes/list/{status}"
	DetailURL     = "/scoring-schemes/detail/{id}"
	AddURL        = "/action/scoring-scheme/add"
	EditURL       = "/action/scoring-scheme/edit/{id}"
	DeleteURL     = "/action/scoring-scheme/delete"
	BulkDeleteURL = "/action/scoring-scheme/bulk-delete"
	TabActionURL  = "/action/scoring-scheme/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the scoring scheme views.
type Routes struct {
	// Sidebar navigation context
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
		ActiveSubNav: "scoring_schemes",

		ListURL:       ListURL,
		DetailURL:     DetailURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		BulkDeleteURL: BulkDeleteURL,

		TabActionURL: TabActionURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all scoring scheme routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"scoring_scheme.list":        r.ListURL,
		"scoring_scheme.detail":      r.DetailURL,
		"scoring_scheme.add":         r.AddURL,
		"scoring_scheme.edit":        r.EditURL,
		"scoring_scheme.delete":      r.DeleteURL,
		"scoring_scheme.bulk_delete": r.BulkDeleteURL,
		"scoring_scheme.tab_action":  r.TabActionURL,
	}
}
