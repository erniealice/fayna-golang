package evaluation_template

// routes.go — EvaluationTemplate route constants + Routes config struct.
//
// Route grammar (pages.md §0.1 / STR-4): page routes are workspace-keyed
// (rewritten to /w/{ws}/... by the workspace-path middleware); action routes
// live under /action/*. Route-map KEYS are dot.snake_case on the proto domain
// name ("evaluation_template.*"), never a lyngua skin.

// EvaluationTemplate routes (rubric library — staff-only authoring).
const (
	ListURL          = "/evaluation-templates/list/{status}"
	DetailURL        = "/evaluation-templates/detail/{id}"
	TableURL         = "/action/evaluation_template/table/{status}"
	AddURL           = "/action/evaluation_template/add"
	EditURL          = "/action/evaluation_template/edit/{id}"
	// Verb-first (verb/{id}) to match the centymo/entydad mutation-route
	// convention and avoid Go 1.22+ ServeMux ambiguity with EditURL above
	// (id-first {id}/verb and verb-first edit/{id} at the same depth cannot
	// disambiguate, e.g. "/action/evaluation_template/edit/activate").
	ActivateURL      = "/action/evaluation_template/activate/{id}"
	DeprecateURL     = "/action/evaluation_template/deprecate/{id}"
	CloneURL         = "/action/evaluation_template/clone/{id}"
	BulkDeprecateURL = "/action/evaluation_template/bulk/deprecate"
	TabActionURL     = "/action/evaluation_template/tab/{id}/{tab}"
)

// Routes holds all route paths for the evaluation template views.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL          string `json:"list_url"`
	DetailURL        string `json:"detail_url"`
	TableURL         string `json:"table_url"`
	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	ActivateURL      string `json:"activate_url"`
	DeprecateURL     string `json:"deprecate_url"`
	CloneURL         string `json:"clone_url"`
	BulkDeprecateURL string `json:"bulk_deprecate_url"`
	TabActionURL     string `json:"tab_action_url"`
}

// DefaultRoutes returns a Routes populated from the package-level constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "performance",
		ActiveSubNav: "evaluation_template",

		ListURL:          ListURL,
		DetailURL:        DetailURL,
		TableURL:         TableURL,
		AddURL:           AddURL,
		EditURL:          EditURL,
		ActivateURL:      ActivateURL,
		DeprecateURL:     DeprecateURL,
		CloneURL:         CloneURL,
		BulkDeprecateURL: BulkDeprecateURL,
		TabActionURL:     TabActionURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// evaluation template routes (route_config / boot warning silencing).
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"evaluation_template.list":            r.ListURL,
		"evaluation_template.detail":          r.DetailURL,
		"evaluation_template.table":           r.TableURL,
		"evaluation_template.add":             r.AddURL,
		"evaluation_template.edit":            r.EditURL,
		"evaluation_template.activate":        r.ActivateURL,
		"evaluation_template.deprecate":       r.DeprecateURL,
		"evaluation_template.clone":           r.CloneURL,
		"evaluation_template.bulk_deprecate":  r.BulkDeprecateURL,
		"evaluation_template.tab_action":      r.TabActionURL,
	}
}
