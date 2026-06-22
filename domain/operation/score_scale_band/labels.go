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
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type ButtonLabels struct {
	AddBand string `json:"addBand"`
}

type ColumnLabels struct {
	OutputLabel   string `json:"outputLabel"`
	SequenceOrder string `json:"sequenceOrder"`
	InputMin      string `json:"inputMin"`
	InputMax      string `json:"inputMax"`
	OutputValue   string `json:"outputValue"`
	Determination string `json:"determination"`
	Status        string `json:"status"`
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
	ScoreScaleId           string `json:"scoreScaleId"`
	SequenceOrder          string `json:"sequenceOrder"`
	InputMin               string `json:"inputMin"`
	InputMinPlaceholder    string `json:"inputMinPlaceholder"`
	InputMax               string `json:"inputMax"`
	InputMaxPlaceholder    string `json:"inputMaxPlaceholder"`
	InputMatch             string `json:"inputMatch"`
	InputMatchPlaceholder  string `json:"inputMatchPlaceholder"`
	OutputValue            string `json:"outputValue"`
	OutputLabel            string `json:"outputLabel"`
	OutputLabelPlaceholder string `json:"outputLabelPlaceholder"`
	BandRole               string `json:"bandRole"`
	BandRolePlaceholder    string `json:"bandRolePlaceholder"`
	Determination          string `json:"determination"`
	DeterminationInfo      string `json:"determinationInfo"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle     string `json:"pageTitle"`
	ScoreScaleId  string `json:"scoreScaleId"`
	SequenceOrder string `json:"sequenceOrder"`
	InputMin      string `json:"inputMin"`
	InputMax      string `json:"inputMax"`
	InputMatch    string `json:"inputMatch"`
	OutputValue   string `json:"outputValue"`
	OutputLabel   string `json:"outputLabel"`
	BandRole      string `json:"bandRole"`
	Determination string `json:"determination"`
	Status        string `json:"status"`
	CreatedDate   string `json:"createdDate"`
	ModifiedDate  string `json:"modifiedDate"`
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
