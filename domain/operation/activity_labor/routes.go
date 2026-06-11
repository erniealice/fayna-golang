package activity_labor

// routes.go — ActivityLabor route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Activity Labor routes (charge detail for ENTRY_TYPE_LABOR job activities).
// ListURL is registered but NOT in the sidebar — power-user / debug only.
// The primary surface is the JobActivity detail page's charge tab.
const (
	ListURL        = "/activity-labor/list"
	DetailURL      = "/activity-labor/{id}"
	AddURL         = "/action/activity-labor/add"
	EditURL        = "/action/activity-labor/edit/{id}"
	DeleteURL      = "/action/activity-labor/delete"
	StaffSearchURL = "/action/activity-labor/search/staff"
)

// Routes holds URL patterns for the activity labor sibling module.
// ActiveNav is intentionally empty — this module is NOT in the sidebar.
// Entry point is the JobActivity detail page's charge tab (entry_type=LABOR).
type Routes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.
	ActiveNav string `json:"active_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	// StaffSearchURL — JSON endpoint for the staff auto-complete picker.
	// Returns [{value, label}]. Empty when staff use case is unavailable.
	StaffSearchURL string `json:"staff_search_url"`
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

		StaffSearchURL: StaffSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// activity labor routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"activity_labor.list":         r.ListURL,
		"activity_labor.detail":       r.DetailURL,
		"activity_labor.add":          r.AddURL,
		"activity_labor.edit":         r.EditURL,
		"activity_labor.delete":       r.DeleteURL,
		"activity_labor.search.staff": r.StaffSearchURL,
	}
}
