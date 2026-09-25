package rating_description_set_product_plan

// routes.go — RatingDescriptionSetProductPlan (AY setup / "Descriptor
// Assignments") route constants and Routes config struct.
//
// MVP scope: list (active links for one academic year) + Relink (first link
// and replacement) + Unlink. Bulk assign / copy-previous-AY are deferred
// (TODO, see sequence log).

const (
	ListURL   = "/rating-description-set-links/list/{status}"
	RelinkURL = "/action/rating-description-set-product-plan/relink"
	UnlinkURL = "/action/rating-description-set-product-plan/unlink"
)

// Routes holds all route paths for the rating description set product plan
// (AY setup) views.
type Routes struct {
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL   string `json:"list_url"`
	RelinkURL string `json:"relink_url"`
	UnlinkURL string `json:"unlink_url"`
}

// DefaultRoutes returns Routes populated from the package-level route constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "rating_description_set_links",

		ListURL:   ListURL,
		RelinkURL: RelinkURL,
		UnlinkURL: UnlinkURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"rating_description_set_product_plan.list":   r.ListURL,
		"rating_description_set_product_plan.relink": r.RelinkURL,
		"rating_description_set_product_plan.unlink": r.UnlinkURL,
	}
}
