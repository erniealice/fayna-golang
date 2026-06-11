package outcome_criteria

// routes.go — OutcomeCriteria route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Outcome Criteria routes (criteria library)
const (
	ListURL             = "/criteria/list/{status}"
	DetailURL           = "/criteria/detail/{id}"
	AddURL              = "/action/criterion/add"
	EditURL             = "/action/criterion/edit/{id}"
	DeleteURL           = "/action/criterion/delete"
	BulkDeleteURL       = "/action/criterion/bulk-delete"
	TabActionURL        = "/action/criterion/detail/{id}/tab/{tab}"
	AttachmentUploadURL = "/action/criterion/detail/{id}/attachments/upload"
	AttachmentDeleteURL = "/action/criterion/detail/{id}/attachments/delete"
)

// Routes holds all route paths for the outcome criteria (criteria library) views.
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

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultRoutes returns a Routes populated from
// the package-level route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "criteria",

		ListURL:       ListURL,
		DetailURL:     DetailURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		BulkDeleteURL: BulkDeleteURL,

		TabActionURL: TabActionURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// outcome criteria routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_criteria.list":              r.ListURL,
		"outcome_criteria.detail":            r.DetailURL,
		"outcome_criteria.add":               r.AddURL,
		"outcome_criteria.edit":              r.EditURL,
		"outcome_criteria.delete":            r.DeleteURL,
		"outcome_criteria.bulk_delete":       r.BulkDeleteURL,
		"outcome_criteria.tab_action":        r.TabActionURL,
		"outcome_criteria.attachment.upload": r.AttachmentUploadURL,
		"outcome_criteria.attachment.delete": r.AttachmentDeleteURL,
	}
}
