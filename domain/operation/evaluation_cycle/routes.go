package evaluation_cycle

// routes.go — EvaluationCycle route constants and Routes config struct.
//
// Route keys use dot.snake_case on the proto domain name (STR-4):
// evaluation_cycle.{list,table,dashboard,add,open,close,tab.members}.
// Action prefix: /action/evaluation_cycle/...
// Page prefix: /app/evaluation-cycles/... (workspace-keyed to /w/{ws}/... at the
// route_config layer by the workspace-path middleware).

// Default route constants for evaluation_cycle views.
const (
	ListURL          = "/app/evaluation-cycles/list/{status}"
	DetailURL        = "/app/evaluation-cycles/detail/{id}"
	TableURL         = "/action/evaluation_cycle/table/{status}"
	AddURL           = "/action/evaluation_cycle/add"
	// Verb-first (verb/{id}) to match the centymo/entydad mutation-route
	// convention (consistent with the rest of the evaluation domain; avoids
	// any Go 1.22+ ServeMux ambiguity with verb-first peers at the same depth).
	OpenURL          = "/action/evaluation_cycle/open/{id}"
	CloseURL         = "/action/evaluation_cycle/close/{id}"
	MembersTabURL    = "/action/evaluation_cycle/tab/{id}/members"
)

// Routes holds all route paths for evaluation_cycle views and actions.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	TableURL      string `json:"table_url"`
	AddURL        string `json:"add_url"`
	OpenURL       string `json:"open_url"`
	CloseURL      string `json:"close_url"`
	MembersTabURL string `json:"members_tab_url"`
}

// DefaultRoutes returns a Routes populated from the package-level
// route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "performance",
		ActiveSubNav: "evaluation-cycles",

		ListURL:       ListURL,
		DetailURL:     DetailURL,
		TableURL:      TableURL,
		AddURL:        AddURL,
		OpenURL:       OpenURL,
		CloseURL:      CloseURL,
		MembersTabURL: MembersTabURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// evaluation_cycle routes. The detail page key is `evaluation_cycle.dashboard`
// (pages.md §F), the members tab is `evaluation_cycle.tab.members`.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"evaluation_cycle.list":         r.ListURL,
		"evaluation_cycle.dashboard":    r.DetailURL,
		"evaluation_cycle.table":        r.TableURL,
		"evaluation_cycle.add":          r.AddURL,
		"evaluation_cycle.open":         r.OpenURL,
		"evaluation_cycle.close":        r.CloseURL,
		"evaluation_cycle.tab.members":  r.MembersTabURL,
	}
}
