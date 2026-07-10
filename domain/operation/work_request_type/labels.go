package work_request_type

// labels.go -- WorkRequestType label structs + DefaultLabels constructor.
//
// Covers: page headings, entity singular/plural, status labels (active/archived),
// category labels (person_scoped/account_scoped), table column headers,
// action button labels, form labels, empty states, and error messages.

// Labels holds all translatable strings for the work_request_type module.
type Labels struct {
	Page     PageLabels     `json:"page"`
	Entity   EntityLabels   `json:"entity"`
	Status   StatusLabels   `json:"status"`
	Category CategoryLabels `json:"category"`
	Columns  ColumnLabels   `json:"columns"`
	Actions  ActionLabels   `json:"actions"`
	Form     FormLabels     `json:"form"`
	Empty    EmptyLabels    `json:"empty"`
	Errors   ErrorLabels    `json:"errors"`
}

// PageLabels holds translatable strings for the catalog page headings.
type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

// EntityLabels holds singular/plural entity name labels.
type EntityLabels struct {
	Singular string `json:"singular"`
	Plural   string `json:"plural"`
}

// StatusLabels holds translatable strings for the 2 catalog statuses.
type StatusLabels struct {
	Active   string `json:"active"`
	Archived string `json:"archived"`
}

// CategoryLabels holds translatable strings for the 2 type categories.
type CategoryLabels struct {
	PersonScoped  string `json:"person_scoped"`
	AccountScoped string `json:"account_scoped"`
}

// ColumnLabels holds translatable strings for catalog table column headers.
type ColumnLabels struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	DefaultSLAHours string `json:"default_sla_hours"`
	Status          string `json:"status"`
	SortOrder       string `json:"sort_order"`
}

// ActionLabels holds translatable strings for action button labels.
type ActionLabels struct {
	Add       string `json:"add"`
	Edit      string `json:"edit"`
	Archive   string `json:"archive"`
	Unarchive string `json:"unarchive"`
}

// FormLabels holds translatable strings for the catalog drawer form.
type FormLabels struct {
	Code                       string `json:"code"`
	CodePlaceholder            string `json:"code_placeholder"`
	LabelKey                   string `json:"label_key"`
	LabelKeyPlaceholder        string `json:"label_key_placeholder"`
	DescriptionKey             string `json:"description_key"`
	DescriptionKeyPlaceholder  string `json:"description_key_placeholder"`
	Category                   string `json:"category"`
	RequiresResource           string `json:"requires_resource"`
	DefaultSLAHours            string `json:"default_sla_hours"`
	DefaultSLAHoursPlaceholder string `json:"default_sla_hours_placeholder"`
	SortOrder                  string `json:"sort_order"`
	SortOrderPlaceholder       string `json:"sort_order_placeholder"`
	IconKey                    string `json:"icon_key"`
	IconKeyPlaceholder         string `json:"icon_key_placeholder"`
	Status                     string `json:"status"`
}

// EmptyLabels holds translatable strings for empty state messaging.
type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

// ErrorLabels holds translatable strings for error messaging.
type ErrorLabels struct {
	NotFound         string `json:"not_found"`
	PermissionDenied string `json:"permission_denied"`
	IDRequired       string `json:"id_required"`
	InvalidForm      string `json:"invalid_form"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Request Types",
			Caption: "Manage work request type catalog",
		},
		Entity: EntityLabels{
			Singular: "Request Type",
			Plural:   "Request Types",
		},
		Status: StatusLabels{
			Active:   "Active",
			Archived: "Archived",
		},
		Category: CategoryLabels{
			PersonScoped:  "Person Scoped",
			AccountScoped: "Account Scoped",
		},
		Columns: ColumnLabels{
			Code:            "Code",
			Name:            "Name",
			Category:        "Category",
			DefaultSLAHours: "Default SLA (hours)",
			Status:          "Status",
			SortOrder:       "Sort Order",
		},
		Actions: ActionLabels{
			Add:       "Add Request Type",
			Edit:      "Edit",
			Archive:   "Archive",
			Unarchive: "Unarchive",
		},
		Form: FormLabels{
			Code:                       "Code",
			CodePlaceholder:            "e.g. salary_increase",
			LabelKey:                   "Label Key",
			LabelKeyPlaceholder:        "Lyngua label key",
			DescriptionKey:             "Description Key",
			DescriptionKeyPlaceholder:  "Lyngua description key",
			Category:                   "Category",
			RequiresResource:           "Requires Resource",
			DefaultSLAHours:            "Default SLA (hours)",
			DefaultSLAHoursPlaceholder: "e.g. 48",
			SortOrder:                  "Sort Order",
			SortOrderPlaceholder:       "e.g. 10",
			IconKey:                    "Icon Key",
			IconKeyPlaceholder:         "e.g. icon-file-text",
			Status:                     "Status",
		},
		Empty: EmptyLabels{
			Title:   "No request types found",
			Message: "No request types to display.",
		},
		Errors: ErrorLabels{
			NotFound:         "Request type not found",
			PermissionDenied: "You do not have permission to perform this action",
			IDRequired:       "Request type ID is required",
			InvalidForm:      "Invalid form data",
		},
	}
}
