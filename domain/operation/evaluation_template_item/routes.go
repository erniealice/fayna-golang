package evaluation_template_item

// routes.go — EvaluationTemplateItem (rubric item) route constants.
//
// The item has NO standalone list/detail page — it surfaces only via the
// evaluation_template detail Items tab + this drawer. Route-map keys are
// dot.snake_case on the proto domain name ("evaluation_template_item.*").

const (
	AddURL    = "/action/evaluation_template_item/add"
	EditURL   = "/action/evaluation_template_item/edit/{item_id}"
	RemoveURL = "/action/evaluation_template_item/{item_id}/remove"
)

// Routes holds the rubric-item drawer route paths.
type Routes struct {
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	RemoveURL string `json:"remove_url"`
}

// DefaultRoutes returns a Routes populated from the package-level constants.
func DefaultRoutes() Routes {
	return Routes{
		AddURL:    AddURL,
		EditURL:   EditURL,
		RemoveURL: RemoveURL,
	}
}

// RouteMap returns dot-notation keys to route paths for the rubric-item routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"evaluation_template_item.add":    r.AddURL,
		"evaluation_template_item.edit":   r.EditURL,
		"evaluation_template_item.remove": r.RemoveURL,
	}
}
