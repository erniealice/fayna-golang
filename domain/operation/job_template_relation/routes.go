package job_template_relation

// JobTemplateRelation routes
const (
	ListURL       = "/job-template-relations/list/{status}"
	DetailURL     = "/job-template-relations/detail/{id}"
	AddURL        = "/action/job-template-relation/add"
	EditURL       = "/action/job-template-relation/edit/{id}"
	DeleteURL     = "/action/job-template-relation/delete"
	BulkDeleteURL = "/action/job-template-relation/bulk-delete"
	TabActionURL  = "/action/job-template-relation/detail/{id}/tab/{tab}"
)

// Routes holds all route paths for the job template relation views.
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
		ActiveSubNav: "job_template_relation",

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
		"job_template_relation.list":        r.ListURL,
		"job_template_relation.detail":      r.DetailURL,
		"job_template_relation.add":         r.AddURL,
		"job_template_relation.edit":        r.EditURL,
		"job_template_relation.delete":      r.DeleteURL,
		"job_template_relation.bulk_delete": r.BulkDeleteURL,
		"job_template_relation.tab_action":  r.TabActionURL,
	}
}
