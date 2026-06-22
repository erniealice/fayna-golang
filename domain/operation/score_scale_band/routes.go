package score_scale_band

// Score scale band route constants.
const (
	ListURL       = "/score-scale-bands/list/{status}"
	DetailURL     = "/score-scale-bands/detail/{id}"
	AddURL        = "/action/score-scale-band/add"
	EditURL       = "/action/score-scale-band/edit/{id}"
	DeleteURL     = "/action/score-scale-band/delete"
	BulkDeleteURL = "/action/score-scale-band/bulk-delete"
	TabActionURL  = "/action/score-scale-band/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the score scale band views.
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
		ActiveSubNav: "score_scale_bands",

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
		"score_scale_band.list":        r.ListURL,
		"score_scale_band.detail":      r.DetailURL,
		"score_scale_band.add":         r.AddURL,
		"score_scale_band.edit":        r.EditURL,
		"score_scale_band.delete":      r.DeleteURL,
		"score_scale_band.bulk_delete": r.BulkDeleteURL,
		"score_scale_band.tab_action":  r.TabActionURL,
	}
}
