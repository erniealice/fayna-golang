package job_template

// routes.go — JobTemplate route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Job Template routes
const (
	ListURL             = "/job-templates/list/{status}"
	DetailURL           = "/job-templates/detail/{id}"
	AddURL              = "/action/job-template/add"
	EditURL             = "/action/job-template/edit/{id}"
	DeleteURL           = "/action/job-template/delete"
	BulkDeleteURL       = "/action/job-template/bulk-delete"
	SetStatusURL        = "/action/job-template/set-status"
	BulkSetStatusURL    = "/action/job-template/bulk-set-status"
	TabActionURL        = "/action/job-template/detail/{id}/tab/{tab}"
	AttachmentUploadURL = "/action/job-template/detail/{id}/attachments/upload"
	AttachmentDeleteURL = "/action/job-template/detail/{id}/attachments/delete"
)

// Routes holds all route paths for job template views and actions.
type Routes struct {
	// Sidebar navigation context
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

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultRoutes returns a Routes populated from the package-level
// route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "job-templates",

		ListURL:          ListURL,
		DetailURL:        DetailURL,
		AddURL:           AddURL,
		EditURL:          EditURL,
		DeleteURL:        DeleteURL,
		BulkDeleteURL:    BulkDeleteURL,
		SetStatusURL:     SetStatusURL,
		BulkSetStatusURL: BulkSetStatusURL,

		TabActionURL: TabActionURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job template routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_template.list":            r.ListURL,
		"job_template.detail":          r.DetailURL,
		"job_template.add":             r.AddURL,
		"job_template.edit":            r.EditURL,
		"job_template.delete":          r.DeleteURL,
		"job_template.bulk_delete":     r.BulkDeleteURL,
		"job_template.set_status":      r.SetStatusURL,
		"job_template.bulk_set_status": r.BulkSetStatusURL,

		"job_template.tab_action": r.TabActionURL,

		"job_template.attachment.upload": r.AttachmentUploadURL,
		"job_template.attachment.delete": r.AttachmentDeleteURL,
	}
}
