package operation

// outcome_criteria_labels.go — OutcomeCriteria label structs + DefaultOutcomeCriteriaLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// OutcomeCriteriaLabels holds all translatable strings for the outcome criteria module.
type OutcomeCriteriaLabels struct {
	Page    OutcomeCriteriaPageLabels    `json:"page"`
	Buttons OutcomeCriteriaButtonLabels  `json:"buttons"`
	Columns OutcomeCriteriaColumnLabels  `json:"columns"`
	Empty   OutcomeCriteriaEmptyLabels   `json:"empty"`
	Form    OutcomeCriteriaFormLabels    `json:"form"`
	Actions OutcomeCriteriaActionLabels  `json:"actions"`
	Detail  OutcomeCriteriaDetailLabels  `json:"detail"`
	Tabs    OutcomeCriteriaTabLabels     `json:"tabs"`
	Confirm OutcomeCriteriaConfirmLabels `json:"confirm"`
	Errors  OutcomeCriteriaErrorLabels   `json:"errors"`
}

type OutcomeCriteriaPageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type OutcomeCriteriaButtonLabels struct {
	AddCriterion string `json:"addCriterion"`
}

type OutcomeCriteriaColumnLabels struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Scope   string `json:"scope"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type OutcomeCriteriaEmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type OutcomeCriteriaFormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Type            string `json:"type"`
	Scope           string `json:"scope"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"descriptionPlaceholder"`
	Required        string `json:"required"`
	Weight          string `json:"weight"`
	TypeInfo        string `json:"typeInfo"`
	ScopeInfo       string `json:"scopeInfo"`
	WeightInfo      string `json:"weightInfo"`
}

type OutcomeCriteriaActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type OutcomeCriteriaDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Scope        string `json:"scope"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Required     string `json:"required"`
	Weight       string `json:"weight"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
}

type OutcomeCriteriaTabLabels struct {
	Info        string `json:"info"`
	Thresholds  string `json:"thresholds"`
	Options     string `json:"options"`
	Templates   string `json:"templates"`
	Versions    string `json:"versions"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type OutcomeCriteriaConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type OutcomeCriteriaErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
}

// DefaultOutcomeCriteriaLabels returns OutcomeCriteriaLabels with sensible English defaults.
func DefaultOutcomeCriteriaLabels() OutcomeCriteriaLabels {
	return OutcomeCriteriaLabels{
		Page: OutcomeCriteriaPageLabels{
			Heading:         "Criteria Library",
			HeadingActive:   "Active Criteria",
			HeadingInactive: "Inactive Criteria",
			Caption:         "Manage reusable outcome evaluation criteria",
			CaptionActive:   "Manage your active outcome criteria",
			CaptionInactive: "View inactive or deprecated criteria",
		},
		Buttons: OutcomeCriteriaButtonLabels{
			AddCriterion: "Add Criterion",
		},
		Columns: OutcomeCriteriaColumnLabels{
			Name:    "Name",
			Type:    "Type",
			Scope:   "Scope",
			Version: "Version",
			Status:  "Status",
		},
		Empty: OutcomeCriteriaEmptyLabels{
			Title:           "No criteria found",
			Message:         "No outcome criteria to display.",
			ActiveTitle:     "No active criteria",
			ActiveMessage:   "Create your first outcome criterion to get started.",
			InactiveTitle:   "No inactive criteria",
			InactiveMessage: "Deactivated criteria will appear here.",
		},
		Form: OutcomeCriteriaFormLabels{
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
		Actions: OutcomeCriteriaActionLabels{
			View:   "View Criterion",
			Edit:   "Edit Criterion",
			Delete: "Delete Criterion",
		},
		Detail: OutcomeCriteriaDetailLabels{
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
		Tabs: OutcomeCriteriaTabLabels{
			Info:        "Information",
			Thresholds:  "Thresholds",
			Options:     "Options",
			Templates:   "Templates",
			Versions:    "Versions",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: OutcomeCriteriaConfirmLabels{
			Delete:        "Delete Criterion",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: OutcomeCriteriaErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Outcome criterion not found",
			IDRequired:       "Criterion ID is required",
			NoPermission:     "No permission",
		},
	}
}
