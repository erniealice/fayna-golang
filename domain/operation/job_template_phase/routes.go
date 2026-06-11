package job_template_phase

// routes.go — JobTemplatePhase route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// JobTemplatePhase drawer-only module routes.
// No list page, no detail page, no sidebar entry.
// Reached via JobTemplate detail Phases tab Add/Edit/Delete CTAs.
const (
	AddURL        = "/action/job-template-phase/add"
	EditURL       = "/action/job-template-phase/edit/{id}"
	DeleteURL     = "/action/job-template-phase/delete"
	BulkDeleteURL = "/action/job-template-phase/bulk-delete"
	// ResourceSearchURL reuses the job_phase resource search endpoint
	// (same underlying resource entity — no separate handler needed).
	ResourceSearchURL = "/action/job-phase/search/resources"
)

// Routes holds URL patterns for the job_template_phase drawer-only module.
// No list page, no detail page, no sidebar entry.
// Operators reach this only via the JobTemplate detail Phases tab Add/Edit/Delete CTAs.
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
// job template phase routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_template_phase.add":             r.AddURL,
		"job_template_phase.edit":            r.EditURL,
		"job_template_phase.delete":          r.DeleteURL,
		"job_template_phase.bulk_delete":     r.BulkDeleteURL,
		"job_template_phase.search.resource": r.ResourceSearchURL,
	}
}
