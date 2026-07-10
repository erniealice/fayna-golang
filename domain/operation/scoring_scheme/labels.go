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
	HeadingActive   string `json:"heading_active"`
	HeadingInactive string `json:"heading_inactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"caption_active"`
	CaptionInactive string `json:"caption_inactive"`
}

type ButtonLabels struct {
	AddScheme string `json:"add_scheme"`
}

type ColumnLabels struct {
	Name            string `json:"name"`
	CompositeMethod string `json:"composite_method"`
	Version         string `json:"version"`
	Status          string `json:"status"`
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
	Name                    string `json:"name"`
	NamePlaceholder         string `json:"name_placeholder"`
	CompositeMethod         string `json:"composite_method"`
	CompositeMethodInfo     string `json:"composite_method_info"`
	RoundingMode            string `json:"rounding_mode"`
	RoundingModeInfo        string `json:"rounding_mode_info"`
	WeightsMustSumToOne     string `json:"weights_must_sum_to_one"`
	WeightsMustSumToOneInfo string `json:"weights_must_sum_to_one_info"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle           string `json:"page_title"`
	Name                string `json:"name"`
	CompositeMethod     string `json:"composite_method"`
	RoundingMode        string `json:"rounding_mode"`
	WeightsMustSumToOne string `json:"weights_must_sum_to_one"`
	Version             string `json:"version"`
	Status              string `json:"status"`
	SchemeGroupId       string `json:"scheme_group_id"`
	ScoreScaleId        string `json:"score_scale_id"`
	CreatedDate         string `json:"created_date"`
	ModifiedDate        string `json:"modified_date"`
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
	DeleteMessage string `json:"delete_message"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	InvalidFormData  string `json:"invalid_form_data"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
	NoPermission     string `json:"no_permission"`
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
