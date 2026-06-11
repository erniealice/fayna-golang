package job_phase

// routes.go — JobPhase route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// JobPhase standalone module routes.
// The list page is a power-user/debugging surface — no sidebar entry.
// The detail page is the canonical single-phase view with tab strip.
const (
	ListURL             = "/job-phases/list/{status}"
	DetailURL           = "/job-phase/{id}"
	AddURL              = "/action/job-phase/add"
	EditURL             = "/action/job-phase/edit/{id}"
	DeleteURL           = "/action/job-phase/delete"
	BulkDeleteURL       = "/action/job-phase/bulk-delete"
	BulkSetStatusURL    = "/action/job-phase/bulk-set-status"
	TabActionURL        = "/action/job-phase/detail/{id}/tab/{tab}"
	ResourceSearchURL   = "/action/job-phase/search/resources"
	AttachmentUploadURL = "/action/job-phase/detail/{id}/attachments/upload"
	AttachmentDeleteURL = "/action/job-phase/detail/{id}/attachments/delete"
	// SetStatusURL — operator-facing phase status flip (PENDING ↔ ACTIVE ↔ COMPLETED).
	// Reads `id` and `status` from query string. Drives the milestone-billing flow.
	SetStatusURL = "/action/job-phase/set-status"
)

// Routes holds all route paths for the JobPhase standalone module.
// The list page is a power-user/debugging surface with no sidebar entry.
// The detail page is reached via Job detail's Phases tab deep links.
type Routes struct {
	// Sidebar navigation context — mirrors the Job module (phases anchor to the job nav).
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

	// ResourceSearchURL — action-mode auto-complete for the resource FK picker.
	ResourceSearchURL string `json:"resource_search_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`

	// JobDetailURL — used for breadcrumb back to the parent job.
	JobDetailURL string `json:"job_detail_url"`
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

		ResourceSearchURL: ResourceSearchURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,

		// JobDetailURL must be supplied by the consumer (from job.DetailURL)
		JobDetailURL: "/jobs/detail/{id}",
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job phase routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_phase.list":              r.ListURL,
		"job_phase.detail":            r.DetailURL,
		"job_phase.add":               r.AddURL,
		"job_phase.edit":              r.EditURL,
		"job_phase.delete":            r.DeleteURL,
		"job_phase.bulk_delete":       r.BulkDeleteURL,
		"job_phase.set_status":        r.SetStatusURL,
		"job_phase.bulk_set_status":   r.BulkSetStatusURL,
		"job_phase.tab_action":        r.TabActionURL,
		"job_phase.search.resource":   r.ResourceSearchURL,
		"job_phase.attachment.upload": r.AttachmentUploadURL,
		"job_phase.attachment.delete": r.AttachmentDeleteURL,
	}
}
