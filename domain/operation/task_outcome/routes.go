package task_outcome

// routes.go — TaskOutcome route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Task Outcome routes (outcome recording on job tasks)
const (
	ListURL             = "/outcomes/list/{status}"
	DetailURL           = "/outcomes/detail/{id}"
	AddURL              = "/action/outcome/add"
	EditURL             = "/action/outcome/edit/{id}"
	DeleteURL           = "/action/outcome/delete"
	TabActionURL        = "/action/outcome/detail/{id}/tab/{tab}"
	AttachmentUploadURL = "/action/outcome/detail/{id}/attachments/upload"
	AttachmentDeleteURL = "/action/outcome/detail/{id}/attachments/delete"
)

// Routes holds all route paths for task outcome (outcome recording) views.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultRoutes returns a Routes populated from the
// package-level route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "outcomes",

		ListURL:   ListURL,
		DetailURL: DetailURL,
		AddURL:    AddURL,
		EditURL:   EditURL,
		DeleteURL: DeleteURL,

		TabActionURL: TabActionURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// task outcome routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"task_outcome.list":              r.ListURL,
		"task_outcome.detail":            r.DetailURL,
		"task_outcome.add":               r.AddURL,
		"task_outcome.edit":              r.EditURL,
		"task_outcome.delete":            r.DeleteURL,
		"task_outcome.tab_action":        r.TabActionURL,
		"task_outcome.attachment.upload": r.AttachmentUploadURL,
		"task_outcome.attachment.delete": r.AttachmentDeleteURL,
	}
}
