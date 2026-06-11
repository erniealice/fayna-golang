package activity_expense

// routes.go — ActivityExpense route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Activity Expense routes (charge detail for ENTRY_TYPE_EXPENSE job activities).
// ListURL is registered but NOT in the sidebar — power-user / debug only.
// The primary surface is the JobActivity detail page's charge tab.
const (
	ListURL                  = "/activity-expense/list"
	DetailURL                = "/activity-expense/{id}"
	AddURL                   = "/action/activity-expense/add"
	EditURL                  = "/action/activity-expense/edit/{id}"
	DeleteURL                = "/action/activity-expense/delete"
	ExpenseCategorySearchURL = "/action/activity-expense/search/expense-categories"
)

// Routes holds URL patterns for the activity expense sibling module.
// ActiveNav is intentionally empty — this module is NOT in the sidebar.
// Entry point is the JobActivity detail page's charge tab (entry_type=EXPENSE).
type Routes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.
	ActiveNav string `json:"active_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	// ExpenseCategorySearchURL — JSON endpoint for the expense category auto-complete picker.
	// Returns [{value, label}]. Empty when expenditure use case is unavailable.
	ExpenseCategorySearchURL string `json:"expense_category_search_url"`
}

// DefaultRoutes returns a Routes populated from the
// package-level route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav: "", // not in sidebar

		ListURL:   ListURL,
		DetailURL: DetailURL,
		AddURL:    AddURL,
		EditURL:   EditURL,
		DeleteURL: DeleteURL,

		ExpenseCategorySearchURL: ExpenseCategorySearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// activity expense routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"activity_expense.list":                    r.ListURL,
		"activity_expense.detail":                  r.DetailURL,
		"activity_expense.add":                     r.AddURL,
		"activity_expense.edit":                    r.EditURL,
		"activity_expense.delete":                  r.DeleteURL,
		"activity_expense.search.expense_category": r.ExpenseCategorySearchURL,
	}
}
