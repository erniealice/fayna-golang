package reporting_checkpoint

// Reporting Checkpoint routes
const (
	ListURL       = "/reporting-checkpoints/list/{status}"
	DetailURL     = "/reporting-checkpoints/detail/{id}"
	AddURL        = "/action/reporting-checkpoint/add"
	EditURL       = "/action/reporting-checkpoint/edit/{id}"
	DeleteURL     = "/action/reporting-checkpoint/delete"
	BulkDeleteURL = "/action/reporting-checkpoint/bulk-delete"
	TabActionURL  = "/action/reporting-checkpoint/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the reporting checkpoint views.
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
		ActiveSubNav: "reporting_checkpoints",

		ListURL:       ListURL,
		DetailURL:     DetailURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		BulkDeleteURL: BulkDeleteURL,

		TabActionURL: TabActionURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// reporting checkpoint routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"reporting_checkpoint.list":        r.ListURL,
		"reporting_checkpoint.detail":      r.DetailURL,
		"reporting_checkpoint.add":         r.AddURL,
		"reporting_checkpoint.edit":        r.EditURL,
		"reporting_checkpoint.delete":      r.DeleteURL,
		"reporting_checkpoint.bulk_delete": r.BulkDeleteURL,
		"reporting_checkpoint.tab_action":  r.TabActionURL,
	}
}
