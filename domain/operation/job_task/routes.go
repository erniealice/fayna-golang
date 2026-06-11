package job_task

// routes.go — JobTask route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// JobTask standalone module routes.
// The list page is a power-user/debugging surface — no sidebar entry.
// The detail page is reached via JobPhase detail's Tasks tab deep links.
const (
	ListURL               = "/job-tasks/list/{status}"
	DetailURL             = "/job-task/{id}"
	AddURL                = "/action/job-task/add"
	EditURL               = "/action/job-task/edit/{id}"
	DeleteURL             = "/action/job-task/delete"
	BulkDeleteURL         = "/action/job-task/bulk-delete"
	SetStatusURL          = "/action/job-task/set-status"
	BulkSetStatusURL      = "/action/job-task/bulk-set-status"
	TabActionURL          = "/action/job-task/detail/{id}/tab/{tab}"
	StaffSearchURL        = "/action/job-task/search/staff"
	ResourceSearchURL     = "/action/job-task/search/resources"
	TemplateTaskSearchURL = "/action/job-task/search/template-tasks"
	AttachmentUploadURL   = "/action/job-task/detail/{id}/attachments/upload"
	AttachmentDeleteURL   = "/action/job-task/detail/{id}/attachments/delete"
)

// Routes holds all route paths for the JobTask standalone module.
// The list page is a power-user/debugging surface with no sidebar entry.
// The detail page is reached via JobPhase detail's Tasks tab deep links.
type Routes struct {
	// Sidebar navigation context — mirrors the Job module (tasks anchor to the job nav).
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL          string `json:"list_url"`
	DetailURL        string `json:"detail_url"`
	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	DeleteURL        string `json:"delete_url"`
	BulkDeleteURL    string `json:"bulk_delete_url"`
	SetStatusURL     string `json:"set_status_url"`
	BulkSetStatusURL string `json:"bulk_set_status_url"`

	TabActionURL string `json:"tab_action_url"`

	// Search endpoints for the task drawer form pickers.
	StaffSearchURL        string `json:"staff_search_url"`
	ResourceSearchURL     string `json:"resource_search_url"`
	TemplateTaskSearchURL string `json:"template_task_search_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`

	// JobPhaseDetailURL — used for breadcrumb back to the parent phase.
	JobPhaseDetailURL string `json:"job_phase_detail_url"`
}

// DefaultRoutes returns a Routes populated from the
// package-level route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "jobs",

		ListURL:          ListURL,
		DetailURL:        DetailURL,
		AddURL:           AddURL,
		EditURL:          EditURL,
		DeleteURL:        DeleteURL,
		BulkDeleteURL:    BulkDeleteURL,
		SetStatusURL:     SetStatusURL,
		BulkSetStatusURL: BulkSetStatusURL,

		TabActionURL: TabActionURL,

		StaffSearchURL:        StaffSearchURL,
		ResourceSearchURL:     ResourceSearchURL,
		TemplateTaskSearchURL: TemplateTaskSearchURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,

		// JobPhaseDetailURL must be supplied by the consumer (from job_phase.DetailURL)
		JobPhaseDetailURL: "/job-phase/{id}",
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job task routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_task.list":              r.ListURL,
		"job_task.detail":            r.DetailURL,
		"job_task.add":               r.AddURL,
		"job_task.edit":              r.EditURL,
		"job_task.delete":            r.DeleteURL,
		"job_task.bulk_delete":       r.BulkDeleteURL,
		"job_task.set_status":        r.SetStatusURL,
		"job_task.bulk_set_status":   r.BulkSetStatusURL,
		"job_task.tab_action":        r.TabActionURL,
		"job_task.search.staff":      r.StaffSearchURL,
		"job_task.search.resource":   r.ResourceSearchURL,
		"job_task.search.template":   r.TemplateTaskSearchURL,
		"job_task.attachment.upload": r.AttachmentUploadURL,
		"job_task.attachment.delete": r.AttachmentDeleteURL,
	}
}
