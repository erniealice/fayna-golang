package job_activity

// routes.go — JobActivity route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Job Activity (timesheet / cross-job activity log) routes
const (
	ListURL                = "/activities"
	DetailURL              = "/activities/detail/{id}"
	AddURL                 = "/action/activity/add"
	EditURL                = "/action/activity/edit/{id}"
	DeleteURL              = "/action/activity/delete"
	SubmitURL              = "/action/activity/submit"
	ApproveURL             = "/action/activity/approve"
	RejectURL              = "/action/activity/reject"
	PostURL                = "/action/activity/post"
	ReverseURL             = "/action/activity/reverse"
	BulkDeleteURL          = "/action/activity/bulk-delete"
	BulkGenerateInvoiceURL = "/action/activity/bulk-generate-invoice"
	TabActionURL           = "/action/activity/detail/{id}/tab/{tab}"
	AttachmentUploadURL    = "/action/activity/detail/{id}/attachments/upload"
	AttachmentDeleteURL    = "/action/activity/detail/{id}/attachments/delete"
)

// Routes holds all route paths for the job activity (timesheet)
// views and actions.
type Routes struct {
	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`
	SubmitURL     string `json:"submit_url"`
	ApproveURL    string `json:"approve_url"`
	RejectURL     string `json:"reject_url"`
	PostURL       string `json:"post_url"`
	ReverseURL    string `json:"reverse_url"`

	BulkGenerateInvoiceURL string `json:"bulk_generate_invoice_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultRoutes returns a Routes populated from the package-level
// route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ListURL:       ListURL,
		DetailURL:     DetailURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		BulkDeleteURL: BulkDeleteURL,
		SubmitURL:     SubmitURL,
		ApproveURL:    ApproveURL,
		RejectURL:     RejectURL,
		PostURL:       PostURL,
		ReverseURL:    ReverseURL,

		BulkGenerateInvoiceURL: BulkGenerateInvoiceURL,

		TabActionURL: TabActionURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job activity routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"job_activity.list":                  r.ListURL,
		"job_activity.detail":                r.DetailURL,
		"job_activity.add":                   r.AddURL,
		"job_activity.edit":                  r.EditURL,
		"job_activity.delete":                r.DeleteURL,
		"job_activity.bulk_delete":           r.BulkDeleteURL,
		"job_activity.submit":                r.SubmitURL,
		"job_activity.approve":               r.ApproveURL,
		"job_activity.reject":                r.RejectURL,
		"job_activity.post":                  r.PostURL,
		"job_activity.reverse":               r.ReverseURL,
		"job_activity.bulk_generate_invoice": r.BulkGenerateInvoiceURL,
		"job_activity.tab_action":            r.TabActionURL,
		"job_activity.attachment.upload":     r.AttachmentUploadURL,
		"job_activity.attachment.delete":     r.AttachmentDeleteURL,
	}
}
