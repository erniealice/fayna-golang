// Package form holds the template-facing data shape for the activity expense drawer form.
// No Deps, no repo imports, no context.Context parameters (drawer-form-subpackage-convention.md).
package form

import fayna "github.com/erniealice/fayna-golang"

// Data is the template-facing data shape for the activity expense drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// ActivityID is the PK of ActivityExpense and the FK to JobActivity (1:1).
// It is context-locked on Edit — emitted as a hidden input above sheet-body.
// On Add it is sourced from the parent JobActivity's ?activity_id query param.
//
// Note: expense_category (string) is deprecated in the proto in favour of
// expense_category_id (FK). Both are supported here for round-trip safety.
type Data struct {
	FormAction string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit     bool

	// ActivityID is both the PK of ActivityExpense and the FK to JobActivity.
	// Context-locked on Edit; sourced from ?activity_id on Add.
	ActivityID string

	// ExpenseCategoryID is the FK to expenditure_category.
	// ExpenseCategoryName is the display label for the pre-selected auto-complete option.
	ExpenseCategoryID   string
	ExpenseCategoryName string

	// VendorRef — free-text vendor reference / invoice number.
	VendorRef string

	// ReceiptURL — URL of the uploaded receipt document.
	ReceiptURL string

	// PaymentMethod — one of "employee", "company_card", "vendor_bill".
	PaymentMethod string

	// PaymentMethodOptions — pre-built select options for the payment method picker.
	// Built by form.BuildPaymentMethodOptions in options.go.
	PaymentMethodOptions []Option

	// MarkupPctOverride — optional markup percentage override (e.g. 10.00 = 10%).
	MarkupPctOverride float64

	// ExpenseCategorySearchURL — action-mode auto-complete endpoint.
	// Returns [{value, label}] JSON. Empty = flat filter fallback.
	ExpenseCategorySearchURL string

	Labels       fayna.ActivityExpenseLabels
	CommonLabels any
}

// Option is a single select option used in PaymentMethodOptions.
type Option struct {
	Value    string
	Label    string
	Selected bool
}
