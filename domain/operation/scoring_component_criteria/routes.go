package scoring_component_criteria

// ScoringComponentCriteria routes
const (
	ListURL       = "/scoring-component-criteria/list/{status}"
	DetailURL     = "/scoring-component-criteria/detail/{id}"
	AddURL        = "/action/scoring-component-criteria/add"
	EditURL       = "/action/scoring-component-criteria/edit/{id}"
	DeleteURL     = "/action/scoring-component-criteria/delete"
	BulkDeleteURL = "/action/scoring-component-criteria/bulk-delete"
	TabActionURL  = "/action/scoring-component-criteria/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the scoring component criteria views.
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
		ActiveSubNav: "scoring_component_criteria",

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
		"scoring_component_criteria.list":        r.ListURL,
		"scoring_component_criteria.detail":      r.DetailURL,
		"scoring_component_criteria.add":         r.AddURL,
		"scoring_component_criteria.edit":        r.EditURL,
		"scoring_component_criteria.delete":      r.DeleteURL,
		"scoring_component_criteria.bulk_delete": r.BulkDeleteURL,
		"scoring_component_criteria.tab_action":  r.TabActionURL,
	}
}
