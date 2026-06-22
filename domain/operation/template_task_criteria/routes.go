package template_task_criteria

// TemplateTaskCriteria routes
const (
	ListURL       = "/template-task-criteria/list/{status}"
	DetailURL     = "/template-task-criteria/detail/{id}"
	AddURL        = "/action/template-task-criteria/add"
	EditURL       = "/action/template-task-criteria/edit/{id}"
	DeleteURL     = "/action/template-task-criteria/delete"
	BulkDeleteURL = "/action/template-task-criteria/bulk-delete"
	TabActionURL  = "/action/template-task-criteria/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the template task criteria views.
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
		ActiveSubNav: "template_task_criteria",

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
		"template_task_criteria.list":        r.ListURL,
		"template_task_criteria.detail":      r.DetailURL,
		"template_task_criteria.add":         r.AddURL,
		"template_task_criteria.edit":        r.EditURL,
		"template_task_criteria.delete":      r.DeleteURL,
		"template_task_criteria.bulk_delete": r.BulkDeleteURL,
		"template_task_criteria.tab_action":  r.TabActionURL,
	}
}
