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
	HeadingActive   string `json:"heading_active"`
	HeadingInactive string `json:"heading_inactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"caption_active"`
	CaptionInactive string `json:"caption_inactive"`
}

type ButtonLabels struct {
	AddScoreScale string `json:"add_score_scale"`
}

type ColumnLabels struct {
	Name          string `json:"name"`
	ScaleKind     string `json:"scale_kind"`
	VersionStatus string `json:"version_status"`
	Version       string `json:"version"`
	InputUnit     string `json:"input_unit"`
	OutputUnit    string `json:"output_unit"`
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
	Name              string `json:"name"`
	NamePlaceholder   string `json:"name_placeholder"`
	ScaleKind         string `json:"scale_kind"`
	ScaleKindInfo     string `json:"scale_kind_info"`
	InputUnit         string `json:"input_unit"`
	InputUnitHint     string `json:"input_unit_hint"`
	InputMin          string `json:"input_min"`
	InputMax          string `json:"input_max"`
	OutputUnit        string `json:"output_unit"`
	OutputUnitHint    string `json:"output_unit_hint"`
	ScaleGroupId      string `json:"scale_group_id"`
	ScaleGroupIdHint  string `json:"scale_group_id_hint"`
	VersionStatusInfo string `json:"version_status_info"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle     string `json:"page_title"`
	Name          string `json:"name"`
	ScaleKind     string `json:"scale_kind"`
	VersionStatus string `json:"version_status"`
	Version       string `json:"version"`
	InputUnit     string `json:"input_unit"`
	InputMin      string `json:"input_min"`
	InputMax      string `json:"input_max"`
	OutputUnit    string `json:"output_unit"`
	ScaleGroupId  string `json:"scale_group_id"`
	Active        string `json:"active"`
	CreatedBy     string `json:"created_by"`
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
