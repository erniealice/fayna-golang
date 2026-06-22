package scoring_scheme

// Labels holds all translatable strings for the scoring scheme module.
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
	AddScheme string `json:"addScheme"`
}

type ColumnLabels struct {
	Name            string `json:"name"`
	CompositeMethod string `json:"compositeMethod"`
	Version         string `json:"version"`
	Status          string `json:"status"`
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
	Name                    string `json:"name"`
	NamePlaceholder         string `json:"namePlaceholder"`
	CompositeMethod         string `json:"compositeMethod"`
	CompositeMethodInfo     string `json:"compositeMethodInfo"`
	RoundingMode            string `json:"roundingMode"`
	RoundingModeInfo        string `json:"roundingModeInfo"`
	WeightsMustSumToOne     string `json:"weightsMustSumToOne"`
	WeightsMustSumToOneInfo string `json:"weightsMustSumToOneInfo"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle           string `json:"pageTitle"`
	Name                string `json:"name"`
	CompositeMethod     string `json:"compositeMethod"`
	RoundingMode        string `json:"roundingMode"`
	WeightsMustSumToOne string `json:"weightsMustSumToOne"`
	Version             string `json:"version"`
	Status              string `json:"status"`
	SchemeGroupId       string `json:"schemeGroupId"`
	ScoreScaleId        string `json:"scoreScaleId"`
	CreatedDate         string `json:"createdDate"`
	ModifiedDate        string `json:"modifiedDate"`
}

type TabLabels struct {
	Info        string `json:"info"`
	Criteria    string `json:"criteria"`
	Versions    string `json:"versions"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
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
			Heading:         "Scoring Schemes",
			HeadingActive:   "Active Scoring Schemes",
			HeadingInactive: "Inactive Scoring Schemes",
			Caption:         "Manage reusable scoring schemes for evaluation",
			CaptionActive:   "Manage your active scoring schemes",
			CaptionInactive: "View deprecated or inactive scoring schemes",
		},
		Buttons: ButtonLabels{
			AddScheme: "Add Scheme",
		},
		Columns: ColumnLabels{
			Name:            "Name",
			CompositeMethod: "Composite Method",
			Version:         "Version",
			Status:          "Status",
		},
		Empty: EmptyLabels{
			Title:           "No scoring schemes found",
			Message:         "No scoring schemes to display.",
			ActiveTitle:     "No active scoring schemes",
			ActiveMessage:   "Create your first scoring scheme to get started.",
			InactiveTitle:   "No inactive scoring schemes",
			InactiveMessage: "Deprecated schemes will appear here.",
		},
		Form: FormLabels{
			Name:                    "Name",
			NamePlaceholder:         "Enter scheme name",
			CompositeMethod:         "Composite Method",
			CompositeMethodInfo:     "The method used to compute the composite score from individual criterion scores.",
			RoundingMode:            "Rounding Mode",
			RoundingModeInfo:        "How fractional scores are rounded when computing final results.",
			WeightsMustSumToOne:     "Weights Must Sum to One",
			WeightsMustSumToOneInfo: "Enforce that criterion weights sum to exactly 1.0.",
		},
		Actions: ActionLabels{
			View:   "View Scheme",
			Edit:   "Edit Scheme",
			Delete: "Delete Scheme",
		},
		Detail: DetailLabels{
			PageTitle:           "Scheme Details",
			Name:                "Name",
			CompositeMethod:     "Composite Method",
			RoundingMode:        "Rounding Mode",
			WeightsMustSumToOne: "Weights Must Sum to One",
			Version:             "Version",
			Status:              "Status",
			SchemeGroupId:       "Scheme Group",
			ScoreScaleId:        "Score Scale",
			CreatedDate:         "Created",
			ModifiedDate:        "Last Modified",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Criteria:    "Criteria",
			Versions:    "Versions",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Scheme",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Scoring scheme not found",
			IDRequired:       "Scheme ID is required",
			NoPermission:     "No permission",
		},
	}
}
