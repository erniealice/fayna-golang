package activity_material

// routes.go — ActivityMaterial route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Activity Material routes (charge detail for ENTRY_TYPE_MATERIAL job activities).
// ListURL is registered but NOT in the sidebar — power-user / debug only.
// The primary surface is the JobActivity detail page's charge tab.
const (
	ListURL           = "/activity-material/list"
	DetailURL         = "/activity-material/{id}"
	AddURL            = "/action/activity-material/add"
	EditURL           = "/action/activity-material/edit/{id}"
	DeleteURL         = "/action/activity-material/delete"
	ProductSearchURL  = "/action/activity-material/search/products"
	LocationSearchURL = "/action/activity-material/search/locations"
)

// Routes holds URL patterns for the activity material sibling module.
// ActiveNav is intentionally empty — this module is NOT in the sidebar.
// Entry point is the JobActivity detail page's charge tab (entry_type=MATERIAL).
type Routes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.
	ActiveNav string `json:"active_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	// ProductSearchURL — JSON endpoint for the product auto-complete picker.
	// Returns [{value, label}]. Empty when product use case is unavailable.
	ProductSearchURL string `json:"product_search_url"`

	// LocationSearchURL — JSON endpoint for the location auto-complete picker.
	// Returns [{value, label}]. Empty when location use case is unavailable.
	LocationSearchURL string `json:"location_search_url"`
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

		ProductSearchURL:  ProductSearchURL,
		LocationSearchURL: LocationSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// activity material routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"activity_material.list":            r.ListURL,
		"activity_material.detail":          r.DetailURL,
		"activity_material.add":             r.AddURL,
		"activity_material.edit":            r.EditURL,
		"activity_material.delete":          r.DeleteURL,
		"activity_material.search.product":  r.ProductSearchURL,
		"activity_material.search.location": r.LocationSearchURL,
	}
}
