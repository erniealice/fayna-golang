package outcome_criteria

// outcome_criteria_labels.go — OutcomeCriteria label structs + DefaultOutcomeCriteriaLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// OutcomeCriteriaLabels holds all translatable strings for the outcome criteria module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Buttons ButtonLabels  `json:"buttons"`
	Columns ColumnLabels  `json:"columns"`
	Empty   EmptyLabels   `json:"empty"`
	Form    FormLabels    `json:"form"`
	Actions ActionLabels  `json:"actions"`
	Detail  DetailLabels  `json:"detail"`
	Tabs    TabLabels     `json:"tabs"`
	Confirm ConfirmLabels `json:"confirm"`
	Errors  ErrorLabels   `json:"errors"`
}

type PageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"heading_active"`
	HeadingInactive string `json:"heading_inactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"caption_active"`
	CaptionInactive string `json:"caption_inactive"`
}

type ButtonLabels struct {
	AddCriterion string `json:"add_criterion"`
}

type ColumnLabels struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Scope   string `json:"scope"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type EmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"active_title"`
	ActiveMessage   string `json:"active_message"`
	InactiveTitle   string `json:"inactive_title"`
	InactiveMessage string `json:"inactive_message"`
}

type FormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"name_placeholder"`
	Type            string `json:"type"`
	Scope           string `json:"scope"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"description_placeholder"`
	Required        string `json:"required"`
	Weight          string `json:"weight"`
	TypeInfo        string `json:"type_info"`
	ScopeInfo       string `json:"scope_info"`
	WeightInfo      string `json:"weight_info"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle    string `json:"page_title"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Scope        string `json:"scope"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Required     string `json:"required"`
	Weight       string `json:"weight"`
	CreatedDate  string `json:"created_date"`
	ModifiedDate string `json:"modified_date"`
}

type TabLabels struct {
	Info        string `json:"info"`
	Thresholds  string `json:"thresholds"`
	Options     string `json:"options"`
	Templates   string `json:"templates"`
	Versions    string `json:"versions"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type ConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"delete_message"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	InvalidFormData  string `json:"invalid_form_data"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
	NoPermission     string `json:"no_permission"`
}

// DefaultOutcomeCriteriaLabels returns OutcomeCriteriaLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading:         "Criteria Library",
			HeadingActive:   "Active Criteria",
			HeadingInactive: "Inactive Criteria",
			Caption:         "Manage reusable outcome evaluation criteria",
			CaptionActive:   "Manage your active outcome criteria",
			CaptionInactive: "View inactive or deprecated criteria",
		},
		Buttons: ButtonLabels{
			AddCriterion: "Add Criterion",
		},
		Columns: ColumnLabels{
			Name:    "Name",
			Type:    "Type",
			Scope:   "Scope",
			Version: "Version",
			Status:  "Status",
		},
		Empty: EmptyLabels{
			Title:           "No criteria found",
			Message:         "No outcome criteria to display.",
			ActiveTitle:     "No active criteria",
			ActiveMessage:   "Create your first outcome criterion to get started.",
			InactiveTitle:   "No inactive criteria",
			InactiveMessage: "Deactivated criteria will appear here.",
		},
		Form: FormLabels{
			Name:            "Name",
			NamePlaceholder: "Enter criterion name",
			Type:            "Criteria Type",
			Scope:           "Scope",
			Description:     "Description",
			DescPlaceholder: "Enter criterion description...",
			Required:        "Required",
			Weight:          "Weight",
			TypeInfo:        "The evaluation method used to measure this criterion (e.g. numeric score, pass/fail).",
			ScopeInfo:       "Whether this criterion applies at the task, phase, or job level.",
			WeightInfo:      "Relative importance of this criterion when computing an aggregate score.",
		},
		Actions: ActionLabels{
			View:   "View Criterion",
			Edit:   "Edit Criterion",
			Delete: "Delete Criterion",
		},
		Detail: DetailLabels{
			PageTitle:    "Criterion Details",
			Name:         "Name",
			Description:  "Description",
			Type:         "Criteria Type",
			Scope:        "Scope",
			Version:      "Version",
			Status:       "Status",
			Required:     "Required",
			Weight:       "Weight",
			CreatedDate:  "Created",
			ModifiedDate: "Last Modified",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Thresholds:  "Thresholds",
			Options:     "Options",
			Templates:   "Templates",
			Versions:    "Versions",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Criterion",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Outcome criterion not found",
			IDRequired:       "Criterion ID is required",
			NoPermission:     "No permission",
		},
	}
}
