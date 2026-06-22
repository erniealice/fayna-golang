package score_scale

// Labels holds all translatable strings for the score scale module.
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
	AddScoreScale string `json:"addScoreScale"`
}

type ColumnLabels struct {
	Name          string `json:"name"`
	ScaleKind     string `json:"scaleKind"`
	VersionStatus string `json:"versionStatus"`
	Version       string `json:"version"`
	InputUnit     string `json:"inputUnit"`
	OutputUnit    string `json:"outputUnit"`
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
	Name              string `json:"name"`
	NamePlaceholder   string `json:"namePlaceholder"`
	ScaleKind         string `json:"scaleKind"`
	ScaleKindInfo     string `json:"scaleKindInfo"`
	InputUnit         string `json:"inputUnit"`
	InputUnitHint     string `json:"inputUnitHint"`
	InputMin          string `json:"inputMin"`
	InputMax          string `json:"inputMax"`
	OutputUnit        string `json:"outputUnit"`
	OutputUnitHint    string `json:"outputUnitHint"`
	ScaleGroupId      string `json:"scaleGroupId"`
	ScaleGroupIdHint  string `json:"scaleGroupIdHint"`
	VersionStatusInfo string `json:"versionStatusInfo"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle     string `json:"pageTitle"`
	Name          string `json:"name"`
	ScaleKind     string `json:"scaleKind"`
	VersionStatus string `json:"versionStatus"`
	Version       string `json:"version"`
	InputUnit     string `json:"inputUnit"`
	InputMin      string `json:"inputMin"`
	InputMax      string `json:"inputMax"`
	OutputUnit    string `json:"outputUnit"`
	ScaleGroupId  string `json:"scaleGroupId"`
	Active        string `json:"active"`
	CreatedBy     string `json:"createdBy"`
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
			Heading:         "Score Scales",
			HeadingActive:   "Active Score Scales",
			HeadingInactive: "Inactive Score Scales",
			Caption:         "Manage score scale definitions",
			CaptionActive:   "Manage your active score scales",
			CaptionInactive: "View deprecated or draft score scales",
		},
		Buttons: ButtonLabels{
			AddScoreScale: "Add Score Scale",
		},
		Columns: ColumnLabels{
			Name:          "Name",
			ScaleKind:     "Kind",
			VersionStatus: "Status",
			Version:       "Version",
			InputUnit:     "Input Unit",
			OutputUnit:    "Output Unit",
		},
		Empty: EmptyLabels{
			Title:           "No score scales found",
			Message:         "No score scales to display.",
			ActiveTitle:     "No active score scales",
			ActiveMessage:   "Create your first score scale to get started.",
			InactiveTitle:   "No inactive score scales",
			InactiveMessage: "Deprecated or draft score scales will appear here.",
		},
		Form: FormLabels{
			Name:              "Name",
			NamePlaceholder:   "Enter scale name",
			ScaleKind:         "Scale Kind",
			ScaleKindInfo:     "Whether the scale maps a numeric range or an exact value to an output.",
			InputUnit:         "Input Unit",
			InputUnitHint:     "Unit label for the raw input value (e.g. %, points).",
			InputMin:          "Input Minimum",
			InputMax:          "Input Maximum",
			OutputUnit:        "Output Unit",
			OutputUnitHint:    "Unit label for the mapped output value.",
			ScaleGroupId:      "Scale Group",
			ScaleGroupIdHint:  "ID of the scale group this scale belongs to.",
			VersionStatusInfo: "Publishing status of this scale version.",
		},
		Actions: ActionLabels{
			View:   "View Scale",
			Edit:   "Edit Scale",
			Delete: "Delete Scale",
		},
		Detail: DetailLabels{
			PageTitle:     "Score Scale Details",
			Name:          "Name",
			ScaleKind:     "Kind",
			VersionStatus: "Version Status",
			Version:       "Version",
			InputUnit:     "Input Unit",
			InputMin:      "Input Minimum",
			InputMax:      "Input Maximum",
			OutputUnit:    "Output Unit",
			ScaleGroupId:  "Scale Group",
			Active:        "Active",
			CreatedBy:     "Created By",
			CreatedDate:   "Created",
			ModifiedDate:  "Last Modified",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Score Scale",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Score scale not found",
			IDRequired:       "Score scale ID is required",
			NoPermission:     "No permission",
		},
	}
}
