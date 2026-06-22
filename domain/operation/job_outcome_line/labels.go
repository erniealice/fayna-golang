package job_outcome_line

// labels.go — JobOutcomeLine label structs + DefaultLabels constructor.

// Labels holds all translatable strings for the job outcome line module.
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
	AddLine string `json:"addLine"`
}

type ColumnLabels struct {
	Label         string `json:"label"`
	ReportingRole string `json:"reportingRole"`
	OutputValue   string `json:"outputValue"`
	OutputLabel   string `json:"outputLabel"`
	Active        string `json:"active"`
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
	Label             string `json:"label"`
	LabelPlaceholder  string `json:"labelPlaceholder"`
	WeightOrCredits   string `json:"weightOrCredits"`
	OutputValue       string `json:"outputValue"`
	OutputLabel       string `json:"outputLabel"`
	ScoreScaleBandId  string `json:"scoreScaleBandId"`
	ReportingRole     string `json:"reportingRole"`
	ReportingRoleInfo string `json:"reportingRoleInfo"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle        string `json:"pageTitle"`
	Label            string `json:"label"`
	WeightOrCredits  string `json:"weightOrCredits"`
	OutputValue      string `json:"outputValue"`
	OutputLabel      string `json:"outputLabel"`
	ScoreScaleBandId string `json:"scoreScaleBandId"`
	ReportingRole    string `json:"reportingRole"`
	Active           string `json:"active"`
	CreatedDate      string `json:"createdDate"`
	ModifiedDate     string `json:"modifiedDate"`
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
			Heading:         "Outcome Lines",
			HeadingActive:   "Active Outcome Lines",
			HeadingInactive: "Inactive Outcome Lines",
			Caption:         "Manage job outcome lines",
			CaptionActive:   "Manage your active outcome lines",
			CaptionInactive: "View inactive outcome lines",
		},
		Buttons: ButtonLabels{
			AddLine: "Add Outcome Line",
		},
		Columns: ColumnLabels{
			Label:         "Label",
			ReportingRole: "Reporting Role",
			OutputValue:   "Output Value",
			OutputLabel:   "Output Label",
			Active:        "Active",
		},
		Empty: EmptyLabels{
			Title:           "No outcome lines found",
			Message:         "No job outcome lines to display.",
			ActiveTitle:     "No active outcome lines",
			ActiveMessage:   "Create your first outcome line to get started.",
			InactiveTitle:   "No inactive outcome lines",
			InactiveMessage: "Deactivated outcome lines will appear here.",
		},
		Form: FormLabels{
			Label:             "Label",
			LabelPlaceholder:  "Enter outcome line label",
			WeightOrCredits:   "Weight / Credits",
			OutputValue:       "Output Value",
			OutputLabel:       "Output Label",
			ScoreScaleBandId:  "Score Scale Band",
			ReportingRole:     "Reporting Role",
			ReportingRoleInfo: "The role this line plays in reporting (e.g. primary grade, transcript, percentile).",
		},
		Actions: ActionLabels{
			View:   "View Outcome Line",
			Edit:   "Edit Outcome Line",
			Delete: "Delete Outcome Line",
		},
		Detail: DetailLabels{
			PageTitle:        "Outcome Line Details",
			Label:            "Label",
			WeightOrCredits:  "Weight / Credits",
			OutputValue:      "Output Value",
			OutputLabel:      "Output Label",
			ScoreScaleBandId: "Score Scale Band",
			ReportingRole:    "Reporting Role",
			Active:           "Active",
			CreatedDate:      "Created",
			ModifiedDate:     "Last Modified",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Outcome Line",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Job outcome line not found",
			IDRequired:       "Outcome line ID is required",
			NoPermission:     "No permission",
		},
	}
}
