package score_scale

// Score Scale routes
const (
	ListURL       = "/score-scales/list/{status}"
	DetailURL     = "/score-scales/detail/{id}"
	AddURL        = "/action/score-scale/add"
	EditURL       = "/action/score-scale/edit/{id}"
	DeleteURL     = "/action/score-scale/delete"
	BulkDeleteURL = "/action/score-scale/bulk-delete"
	TabActionURL  = "/action/score-scale/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the score scale views.
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

// DefaultRoutes returns Routes populated from the package-level route constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "score_scales",

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
		"score_scale.list":        r.ListURL,
		"score_scale.detail":      r.DetailURL,
		"score_scale.add":         r.AddURL,
		"score_scale.edit":        r.EditURL,
		"score_scale.delete":      r.DeleteURL,
		"score_scale.bulk_delete": r.BulkDeleteURL,
		"score_scale.tab_action":  r.TabActionURL,
	}
}
