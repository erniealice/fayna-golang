package scoring_component

// Labels holds all translatable strings for the scoring component module.
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
	AddComponent string `json:"addComponent"`
}

type ColumnLabels struct {
	Code          string `json:"code"`
	Label         string `json:"label"`
	Weight        string `json:"weight"`
	SequenceOrder string `json:"sequenceOrder"`
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
	ScoringSchemeId       string `json:"scoringSchemeId"`
	Code                  string `json:"code"`
	CodePlaceholder       string `json:"codePlaceholder"`
	Label                 string `json:"label"`
	LabelPlaceholder      string `json:"labelPlaceholder"`
	Weight                string `json:"weight"`
	WeightInfo            string `json:"weightInfo"`
	SequenceOrder         string `json:"sequenceOrder"`
	SequenceOrderInfo     string `json:"sequenceOrderInfo"`
	ParentComponentId     string `json:"parentComponentId"`
	ParentComponentIdInfo string `json:"parentComponentIdInfo"`
	Active                string `json:"active"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	ScoringSchemeId   string `json:"scoringSchemeId"`
	Code              string `json:"code"`
	Label             string `json:"label"`
	Weight            string `json:"weight"`
	SequenceOrder     string `json:"sequenceOrder"`
	ParentComponentId string `json:"parentComponentId"`
	Active            string `json:"active"`
	CreatedDate       string `json:"createdDate"`
	ModifiedDate      string `json:"modifiedDate"`
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
			Heading:         "Scoring Components",
			HeadingActive:   "Active Components",
			HeadingInactive: "Inactive Components",
			Caption:         "Manage scoring components within scoring schemes",
			CaptionActive:   "Manage your active scoring components",
			CaptionInactive: "View inactive scoring components",
		},
		Buttons: ButtonLabels{
			AddComponent: "Add Component",
		},
		Columns: ColumnLabels{
			Code:          "Code",
			Label:         "Label",
			Weight:        "Weight",
			SequenceOrder: "Order",
			Active:        "Active",
		},
		Empty: EmptyLabels{
			Title:           "No components found",
			Message:         "No scoring components to display.",
			ActiveTitle:     "No active components",
			ActiveMessage:   "Create your first scoring component to get started.",
			InactiveTitle:   "No inactive components",
			InactiveMessage: "Deactivated scoring components will appear here.",
		},
		Form: FormLabels{
			ScoringSchemeId:       "Scoring Scheme",
			Code:                  "Code",
			CodePlaceholder:       "Enter component code",
			Label:                 "Label",
			LabelPlaceholder:      "Enter component label",
			Weight:                "Weight",
			WeightInfo:            "Relative weight of this component in the overall scoring scheme.",
			SequenceOrder:         "Sequence Order",
			SequenceOrderInfo:     "Display and evaluation order of this component.",
			ParentComponentId:     "Parent Component",
			ParentComponentIdInfo: "Optional parent component for nested component hierarchies.",
			Active:                "Active",
		},
		Actions: ActionLabels{
			View:   "View Component",
			Edit:   "Edit Component",
			Delete: "Delete Component",
		},
		Detail: DetailLabels{
			PageTitle:         "Component Details",
			ScoringSchemeId:   "Scoring Scheme",
			Code:              "Code",
			Label:             "Label",
			Weight:            "Weight",
			SequenceOrder:     "Sequence Order",
			ParentComponentId: "Parent Component",
			Active:            "Active",
			CreatedDate:       "Created",
			ModifiedDate:      "Last Modified",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Component",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Scoring component not found",
			IDRequired:       "Component ID is required",
			NoPermission:     "No permission",
		},
	}
}
