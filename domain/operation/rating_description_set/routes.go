package rating_description_set

// routes.go — RatingDescriptionSet route constants and Routes config struct.
//
// MVP scope (docs/plan/20260925-criterion-descriptors-by-program-year, sequence
// w3-fayna-views): list + detail (read-only matrix) + Add (create a DRAFT) +
// Publish + Deprecate. Edit/Delete/CreateVersion routes are deferred (TODO —
// see the plan sequence log) and are NOT declared here so RouteMap() only
// advertises what this package actually mounts.

const (
	ListURL      = "/rating-description-sets/list/{status}"
	DetailURL    = "/rating-description-sets/detail/{id}"
	AddURL       = "/action/rating-description-set/add"
	PublishURL   = "/action/rating-description-set/publish/{id}"
	DeprecateURL = "/action/rating-description-set/deprecate/{id}"
)

// Routes holds all route paths for the rating description set views.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL      string `json:"list_url"`
	DetailURL    string `json:"detail_url"`
	AddURL       string `json:"add_url"`
	PublishURL   string `json:"publish_url"`
	DeprecateURL string `json:"deprecate_url"`
}

// DefaultRoutes returns Routes populated from the package-level route constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "rating_description_sets",

		ListURL:      ListURL,
		DetailURL:    DetailURL,
		AddURL:       AddURL,
		PublishURL:   PublishURL,
		DeprecateURL: DeprecateURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"rating_description_set.list":      r.ListURL,
		"rating_description_set.detail":    r.DetailURL,
		"rating_description_set.add":       r.AddURL,
		"rating_description_set.publish":   r.PublishURL,
		"rating_description_set.deprecate": r.DeprecateURL,
	}
}
