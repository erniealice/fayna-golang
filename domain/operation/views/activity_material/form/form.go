// Package form holds the template-facing data shape for the activity material drawer form.
// No Deps, no repo imports, no context.Context parameters (drawer-form-subpackage-convention.md).
package form

import operation "github.com/erniealice/fayna-golang/domain/operation"

// Data is the template-facing data shape for the activity material drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// ActivityID is the PK of ActivityMaterial and the FK to JobActivity (1:1).
// It is context-locked on Edit — emitted as a hidden input above sheet-body.
// On Add it is sourced from the parent JobActivity's ?activity_id query param.
type Data struct {
	FormAction string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit     bool

	// ActivityID is both the PK of ActivityMaterial and the FK to JobActivity.
	// Context-locked on Edit; sourced from ?activity_id on Add.
	ActivityID string

	// ProductID is the FK to product. ProductName is the display label for the
	// pre-selected product auto-complete option.
	ProductID   string
	ProductName string

	// UnitOfMeasure — free-text unit (e.g. "pcs", "kg", "L").
	UnitOfMeasure string

	// LotNumber — optional serialized inventory lot identifier.
	LotNumber string

	// LocationID is the FK to location. LocationName is the display label for
	// the pre-selected location auto-complete option.
	LocationID   string
	LocationName string

	// ProductSearchURL — action-mode auto-complete endpoint (returns [{value,label}] JSON).
	// Empty = product search not available; falls back to flat filter mode.
	ProductSearchURL string

	// LocationSearchURL — action-mode auto-complete endpoint (returns [{value,label}] JSON).
	// Empty = location search not available; falls back to flat filter mode.
	LocationSearchURL string

	Labels       operation.ActivityMaterialLabels
	CommonLabels any
}
