package work_request_type

// routes.go -- WorkRequestType route constants and Routes config struct.
//
// Route keys use dot.snake_case: work_request_type.list, work_request_type.table, etc.
// Action prefix: /action/work_request_type/...
// Page prefix: /app/request-types/... (workspace-keyed at route_config layer).

// Default route constants for work_request_type views.
const (
	ListURL  = "/app/request-types/list/{status}"
	TableURL = "/action/work_request_type/table/{status}"
	AddURL   = "/action/work_request_type/add"
	EditURL  = "/action/work_request_type/edit/{id}"
)

// Routes holds all route paths for work_request_type views and actions.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL  string `json:"list_url"`
	TableURL string `json:"table_url"`
	AddURL   string `json:"add_url"`
	EditURL  string `json:"edit_url"`
}

// DefaultRoutes returns a Routes populated from the package-level
// route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "request-types",
		ActiveSubNav: "request-types",

		ListURL:  ListURL,
		TableURL: TableURL,
		AddURL:   AddURL,
		EditURL:  EditURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// work_request_type routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"work_request_type.list":  r.ListURL,
		"work_request_type.table": r.TableURL,
		"work_request_type.add":   r.AddURL,
		"work_request_type.edit":  r.EditURL,
	}
}
