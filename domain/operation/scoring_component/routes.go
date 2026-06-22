package scoring_component

// Scoring component route constants.
const (
	ListURL       = "/scoring-components/list/{status}"
	DetailURL     = "/scoring-components/detail/{id}"
	AddURL        = "/action/scoring-component/add"
	EditURL       = "/action/scoring-component/edit/{id}"
	DeleteURL     = "/action/scoring-component/delete"
	BulkDeleteURL = "/action/scoring-component/bulk-delete"
	TabActionURL  = "/action/scoring-component/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the scoring component views.
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
		ActiveSubNav: "scoring_components",

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
		"scoring_component.list":        r.ListURL,
		"scoring_component.detail":      r.DetailURL,
		"scoring_component.add":         r.AddURL,
		"scoring_component.edit":        r.EditURL,
		"scoring_component.delete":      r.DeleteURL,
		"scoring_component.bulk_delete": r.BulkDeleteURL,
		"scoring_component.tab_action":  r.TabActionURL,
	}
}
