package fulfillment

// Default route constants for fulfillment views.
// Consumer apps can use these or define their own.
const (
	// DashboardURL — read-only Fulfillment dashboard (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	DashboardURL = "/fulfillment/dashboard"

	ListURL             = "/fulfillment/list/{status}"
	DetailURL           = "/fulfillment/detail/{id}"
	AddURL              = "/action/fulfillment/add"
	EditURL             = "/action/fulfillment/edit/{id}"
	DeleteURL           = "/action/fulfillment/delete"
	TransitionURL       = "/action/fulfillment/transition/{id}"
	ReturnURL           = "/action/fulfillment/return/{id}"
	TabActionURL        = "/action/fulfillment/detail/{id}/tab/{tab}"
	AttachmentUploadURL = "/action/fulfillment/detail/{id}/attachments/upload"
	AttachmentDeleteURL = "/action/fulfillment/detail/{id}/attachments/delete"
)

// Routes holds URL patterns for fulfillment views.
type Routes struct {
	// DashboardURL — read-only Fulfillment dashboard (Phase 3 — Pyeza dashboard block plan).
	DashboardURL  string `json:"dashboard_url"`
	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	TransitionURL string `json:"transition_url"`
	ReturnURL     string `json:"return_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultRoutes returns the standard fulfillment route configuration.
func DefaultRoutes() Routes {
	return Routes{
		DashboardURL:  DashboardURL,
		ListURL:       ListURL,
		DetailURL:     DetailURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		DeleteURL:     DeleteURL,
		TransitionURL: TransitionURL,
		ReturnURL:     ReturnURL,

		TabActionURL: TabActionURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,
	}
}

// RouteMap returns all fulfillment routes as a map for template URL resolution.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"fulfillment.dashboard":         r.DashboardURL,
		"fulfillment.list":              r.ListURL,
		"fulfillment.detail":            r.DetailURL,
		"fulfillment.add":               r.AddURL,
		"fulfillment.edit":              r.EditURL,
		"fulfillment.delete":            r.DeleteURL,
		"fulfillment.transition":        r.TransitionURL,
		"fulfillment.return":            r.ReturnURL,
		"fulfillment.tab_action":        r.TabActionURL,
		"fulfillment.attachment.upload": r.AttachmentUploadURL,
		"fulfillment.attachment.delete": r.AttachmentDeleteURL,
	}
}
