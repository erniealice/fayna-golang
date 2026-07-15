package job_category

// routes.go — JobCategory route constants + Routes config struct. JobCategory is
// the per-workspace job taxonomy reference entity (mirrors PlanGroup / Line):
// list / detail / drawer CRUD. Generic identifiers; vertical wording ("Class
// Categories") lives only in lyngua.

const (
	ListURL       = "/job-categories/list/{status}"
	DetailURL     = "/job-categories/detail/{id}"
	AddURL        = "/action/job-category/add"
	EditURL       = "/action/job-category/edit/{id}"
	DeleteURL     = "/action/job-category/delete"
	BulkDeleteURL = "/action/job-category/bulk-delete"
)

// Routes holds all route paths for the job category views.
type Routes struct {
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`
}

// DefaultRoutes returns a Routes populated from the package-level constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "job-categories",

		ListURL:       ListURL,
		DetailURL:     DetailURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		BulkDeleteURL: BulkDeleteURL,
	}
}

// RouteMap returns dot-notation keys → route paths for all job category routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_category.list":        r.ListURL,
		"job_category.detail":      r.DetailURL,
		"job_category.add":         r.AddURL,
		"job_category.edit":        r.EditURL,
		"job_category.delete":      r.DeleteURL,
		"job_category.bulk_delete": r.BulkDeleteURL,
	}
}
