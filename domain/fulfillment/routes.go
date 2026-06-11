package fulfillment

// Default route constants for fulfillment views.
// Consumer apps can use these or define their own.
const (
	// Fulfillment dashboard route (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	FulfillmentDashboardURL = "/fulfillment/dashboard"

	FulfillmentListURL             = "/fulfillment/list/{status}"
	FulfillmentDetailURL           = "/fulfillment/detail/{id}"
	FulfillmentAddURL              = "/action/fulfillment/add"
	FulfillmentEditURL             = "/action/fulfillment/edit/{id}"
	FulfillmentDeleteURL           = "/action/fulfillment/delete"
	FulfillmentTransitionURL       = "/action/fulfillment/transition/{id}"
	FulfillmentReturnURL           = "/action/fulfillment/return/{id}"
	FulfillmentTabActionURL        = "/action/fulfillment/detail/{id}/tab/{tab}"
	FulfillmentAttachmentUploadURL = "/action/fulfillment/detail/{id}/attachments/upload"
	FulfillmentAttachmentDeleteURL = "/action/fulfillment/detail/{id}/attachments/delete"
)

// FulfillmentRoutes holds URL patterns for fulfillment views.
type FulfillmentRoutes struct {
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

// DefaultFulfillmentRoutes returns the standard fulfillment route configuration.
func DefaultFulfillmentRoutes() FulfillmentRoutes {
	return FulfillmentRoutes{
		DashboardURL:  FulfillmentDashboardURL,
		ListURL:       FulfillmentListURL,
		DetailURL:     FulfillmentDetailURL,
		AddURL:        FulfillmentAddURL,
		EditURL:       FulfillmentEditURL,
		DeleteURL:     FulfillmentDeleteURL,
		TransitionURL: FulfillmentTransitionURL,
		ReturnURL:     FulfillmentReturnURL,

		TabActionURL: FulfillmentTabActionURL,

		AttachmentUploadURL: FulfillmentAttachmentUploadURL,
		AttachmentDeleteURL: FulfillmentAttachmentDeleteURL,
	}
}

// RouteMap returns all fulfillment routes as a map for template URL resolution.
func (r FulfillmentRoutes) RouteMap() map[string]string {
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
