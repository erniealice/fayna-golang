package form

import (
	fayna "github.com/erniealice/fayna-golang"
	"github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the job drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fayna.JobLabels verbatim — templates read
// .Labels.Columns.Name, .Labels.Form.NamePlaceholder, etc. directly.
type Data struct {
	FormAction string
	IsEdit     bool
	ID         string
	Name       string
	ClientID   string
	LocationID string

	// Lifecycle
	Status             string
	StatusOptions      []types.SelectOption
	BillingRuleType    string
	BillingRuleOptions []types.SelectOption

	// Auto-complete display labels pre-filled on Edit.
	ClientName   string
	LocationName string

	// Auto-complete search endpoints for the client and location pickers.
	// Served by the fayna block at JobClientSearchURL / JobLocationSearchURL.
	ClientSearchURL   string
	LocationSearchURL string

	Labels       fayna.JobLabels
	CommonLabels any
}
