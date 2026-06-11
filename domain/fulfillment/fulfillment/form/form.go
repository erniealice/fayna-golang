package form

import fulfillment "github.com/erniealice/fayna-golang/domain/fulfillment/fulfillment"

// Data is the template-facing data shape for the fulfillment drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fulfillment.Labels verbatim — templates
// read .Labels.Errors.*, .Labels.Form.*, .Labels.Columns.* directly.
type Data struct {
	FormAction   string
	WorkspaceID  string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit       bool
	ID           string
	RevenueID    string
	SupplierID   string
	Method       string
	ScheduledAt  string // datetime-local input value ("2006-01-02T15:04"); empty when unset
	Notes        string
	Labels       fulfillment.Labels
	CommonLabels any
}
