package scoring_component_criteria

// Labels holds all translatable strings for the scoring component criteria module.
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
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type ButtonLabels struct {
	AddLink string `json:"addLink"`
}

type ColumnLabels struct {
	ScoringSchemeID    string `json:"scoringSchemeId"`
	ScoringComponentID string `json:"scoringComponentId"`
	OutcomeCriteriaID  string `json:"outcomeCriteriaId"`
	Status             string `json:"status"`
}

type EmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type FormLabels struct {
	ScoringSchemeID        string `json:"scoringSchemeId"`
	ScoringSchemeIDInfo    string `json:"scoringSchemeIdInfo"`
	ScoringComponentID     string `json:"scoringComponentId"`
	ScoringComponentIDInfo string `json:"scoringComponentIdInfo"`
	OutcomeCriteriaID      string `json:"outcomeCriteriaId"`
	OutcomeCriteriaIDInfo  string `json:"outcomeCriteriaIdInfo"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle          string `json:"pageTitle"`
	ScoringSchemeID    string `json:"scoringSchemeId"`
	ScoringComponentID string `json:"scoringComponentId"`
	OutcomeCriteriaID  string `json:"outcomeCriteriaId"`
	Status             string `json:"status"`
	CreatedDate        string `json:"createdDate"`
	ModifiedDate       string `json:"modifiedDate"`
}

type TabLabels struct {
	Info    string `json:"info"`
	History string `json:"history"`
}

type ConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading:         "Scoring Component Criteria",
			HeadingActive:   "Active Links",
			HeadingInactive: "Inactive Links",
			Caption:         "Manage criteria linked to scoring components",
			CaptionActive:   "Active scoring component criteria links",
			CaptionInactive: "Inactive scoring component criteria links",
		},
		Buttons: ButtonLabels{
			AddLink: "Add Link",
		},
		Columns: ColumnLabels{
			ScoringSchemeID:    "Scoring Scheme",
			ScoringComponentID: "Scoring Component",
			OutcomeCriteriaID:  "Outcome Criterion",
			Status:             "Status",
		},
		Empty: EmptyLabels{
			Title:           "No links found",
			Message:         "No scoring component criteria links to display.",
			ActiveTitle:     "No active links",
			ActiveMessage:   "Create your first scoring component criteria link to get started.",
			InactiveTitle:   "No inactive links",
			InactiveMessage: "Deactivated links will appear here.",
		},
		Form: FormLabels{
			ScoringSchemeID:        "Scoring Scheme ID",
			ScoringSchemeIDInfo:    "The scoring scheme this link belongs to.",
			ScoringComponentID:     "Scoring Component ID",
			ScoringComponentIDInfo: "The scoring component being linked.",
			OutcomeCriteriaID:      "Outcome Criteria ID",
			OutcomeCriteriaIDInfo:  "The outcome criterion being linked to the scoring component.",
		},
		Actions: ActionLabels{
			View:   "View Link",
			Edit:   "Edit Link",
			Delete: "Delete Link",
		},
		Detail: DetailLabels{
			PageTitle:          "Scoring Component Criteria Details",
			ScoringSchemeID:    "Scoring Scheme",
			ScoringComponentID: "Scoring Component",
			OutcomeCriteriaID:  "Outcome Criterion",
			Status:             "Status",
			CreatedDate:        "Created",
			ModifiedDate:       "Last Modified",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Link",
			DeleteMessage: "Are you sure you want to delete this scoring component criteria link? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Scoring component criteria link not found",
			IDRequired:       "Link ID is required",
			NoPermission:     "No permission",
		},
	}
}
