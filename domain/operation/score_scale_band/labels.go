package score_scale_band

// Labels holds all translatable strings for the score scale band module.
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
	AddBand string `json:"add_band"`
}

type ColumnLabels struct {
	OutputLabel   string `json:"output_label"`
	SequenceOrder string `json:"sequence_order"`
	InputMin      string `json:"input_min"`
	InputMax      string `json:"input_max"`
	OutputValue   string `json:"output_value"`
	Determination string `json:"determination"`
	Status        string `json:"status"`
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
	ScoreScaleId           string `json:"score_scale_id"`
	SequenceOrder          string `json:"sequence_order"`
	InputMin               string `json:"input_min"`
	InputMinPlaceholder    string `json:"input_min_placeholder"`
	InputMax               string `json:"input_max"`
	InputMaxPlaceholder    string `json:"input_max_placeholder"`
	InputMatch             string `json:"input_match"`
	InputMatchPlaceholder  string `json:"input_match_placeholder"`
	OutputValue            string `json:"output_value"`
	OutputLabel            string `json:"output_label"`
	OutputLabelPlaceholder string `json:"output_label_placeholder"`
	BandRole               string `json:"band_role"`
	BandRolePlaceholder    string `json:"band_role_placeholder"`
	Determination          string `json:"determination"`
	DeterminationInfo      string `json:"determination_info"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle     string `json:"page_title"`
	ScoreScaleId  string `json:"score_scale_id"`
	SequenceOrder string `json:"sequence_order"`
	InputMin      string `json:"input_min"`
	InputMax      string `json:"input_max"`
	InputMatch    string `json:"input_match"`
	OutputValue   string `json:"output_value"`
	OutputLabel   string `json:"output_label"`
	BandRole      string `json:"band_role"`
	Determination string `json:"determination"`
	Status        string `json:"status"`
	CreatedDate   string `json:"created_date"`
	ModifiedDate  string `json:"modified_date"`
}

type TabLabels struct {
	Info    string `json:"info"`
	History string `json:"history"`
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
			Heading:         "Score Scale Bands",
			HeadingActive:   "Active Bands",
			HeadingInactive: "Inactive Bands",
			Caption:         "Manage score scale band definitions",
			CaptionActive:   "Active score scale bands",
			CaptionInactive: "Inactive score scale bands",
		},
		Buttons: ButtonLabels{
			AddBand: "Add Band",
		},
		Columns: ColumnLabels{
			OutputLabel:   "Output Label",
			SequenceOrder: "Order",
			InputMin:      "Input Min",
			InputMax:      "Input Max",
			OutputValue:   "Output Value",
			Determination: "Determination",
			Status:        "Status",
		},
		Empty: EmptyLabels{
			Title:           "No bands found",
			Message:         "No score scale bands to display.",
			ActiveTitle:     "No active bands",
			ActiveMessage:   "Add the first band to this scale.",
			InactiveTitle:   "No inactive bands",
			InactiveMessage: "Deactivated bands will appear here.",
		},
		Form: FormLabels{
			ScoreScaleId:           "Score Scale",
			SequenceOrder:          "Sequence Order",
			InputMin:               "Input Min",
			InputMinPlaceholder:    "e.g. 0",
			InputMax:               "Input Max",
			InputMaxPlaceholder:    "e.g. 100",
			InputMatch:             "Input Match",
			InputMatchPlaceholder:  "Exact match string",
			OutputValue:            "Output Value",
			OutputLabel:            "Output Label",
			OutputLabelPlaceholder: "e.g. Excellent",
			BandRole:               "Band Role",
			BandRolePlaceholder:    "e.g. passing",
			Determination:          "Determination",
			DeterminationInfo:      "The pass/fail determination applied when a score falls within this band.",
		},
		Actions: ActionLabels{
			View:   "View Band",
			Edit:   "Edit Band",
			Delete: "Delete Band",
		},
		Detail: DetailLabels{
			PageTitle:     "Band Details",
			ScoreScaleId:  "Score Scale",
			SequenceOrder: "Sequence Order",
			InputMin:      "Input Min",
			InputMax:      "Input Max",
			InputMatch:    "Input Match",
			OutputValue:   "Output Value",
			OutputLabel:   "Output Label",
			BandRole:      "Band Role",
			Determination: "Determination",
			Status:        "Status",
			CreatedDate:   "Created",
			ModifiedDate:  "Last Modified",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Band",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Score scale band not found",
			IDRequired:       "Band ID is required",
			NoPermission:     "No permission",
		},
	}
}
