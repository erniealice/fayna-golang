package job_template_task

// routes.go — JobTemplateTask route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// JobTemplateTask drawer-only module routes.
// No list page, no detail page, no sidebar entry.
// Reached via JobTemplate detail Tasks tab Add/Edit/Delete CTAs.
const (
	AddURL        = "/action/job-template-task/add"
	EditURL       = "/action/job-template-task/edit/{id}"
	DeleteURL     = "/action/job-template-task/delete"
	BulkDeleteURL = "/action/job-template-task/bulk-delete"
	// ResourceSearchURL reuses the job_phase resource search endpoint.
	ResourceSearchURL = "/action/job-phase/search/resources"
)

// Routes holds URL patterns for the job_template_task drawer-only module.
// No list page, no detail page, no sidebar entry.
// Operators reach this only via the JobTemplate detail Tasks tab Add/Edit/Delete CTAs.
type Routes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.

	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`

	// ResourceSearchURL — action-mode auto-complete for the resource FK picker.
	// Reuses the job_phase resource search endpoint (same resource entity).
	ResourceSearchURL string `json:"resource_search_url"`
}

// DefaultRoutes returns a Routes populated from the package-level
// route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		BulkDeleteURL: BulkDeleteURL,

		ResourceSearchURL: ResourceSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job template task routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_template_task.add":             r.AddURL,
		"job_template_task.edit":            r.EditURL,
		"job_template_task.delete":          r.DeleteURL,
		"job_template_task.bulk_delete":     r.BulkDeleteURL,
		"job_template_task.search.resource": r.ResourceSearchURL,
	}
}
