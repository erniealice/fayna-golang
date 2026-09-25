package rating_description_set_entry

// routes.go — RatingDescriptionSetEntry route constants.
//
// No standalone list/detail page — the entry surfaces only via the parent
// rating_description_set detail's Descriptors matrix + this drawer (same
// shape as evaluation_template_item). MVP scope: Add + Edit only; Delete is
// deferred (TODO, see sequence log).

const (
	AddURL  = "/action/rating-description-set-entry/add"
	EditURL = "/action/rating-description-set-entry/edit/{id}"
)

// Routes holds the entry drawer route paths.
type Routes struct {
	AddURL  string `json:"add_url"`
	EditURL string `json:"edit_url"`
}

// DefaultRoutes returns a Routes populated from the package-level constants.
func DefaultRoutes() Routes {
	return Routes{
		AddURL:  AddURL,
		EditURL: EditURL,
	}
}

// RouteMap returns dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"rating_description_set_entry.add":  r.AddURL,
		"rating_description_set_entry.edit": r.EditURL,
	}
}
