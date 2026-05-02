package form

import fayna "github.com/erniealice/fayna-golang"

// Data is the template-facing data shape for the fulfillment drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fayna.FulfillmentLabels verbatim — templates
// read .Labels.Errors.*, .Labels.Form.*, .Labels.Columns.* directly.
type Data struct {
	FormAction   string
	IsEdit       bool
	ID           string
	RevenueID    string
	SupplierID   string
	Method       string
	Notes        string
	Labels       fayna.FulfillmentLabels
	CommonLabels any
}
