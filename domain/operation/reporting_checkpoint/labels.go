package reporting_checkpoint

// Labels holds all translatable strings for the reporting checkpoint module.
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
	AddCheckpoint string `json:"addCheckpoint"`
}

type ColumnLabels struct {
	Label         string `json:"label"`
	RoleCode      string `json:"roleCode"`
	SequenceOrder string `json:"sequenceOrder"`
	Version       string `json:"version"`
	VersionStatus string `json:"versionStatus"`
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
	Label               string `json:"label"`
	LabelPlaceholder    string `json:"labelPlaceholder"`
	CheckpointGroupID   string `json:"checkpointGroupId"`
	RoleCode            string `json:"roleCode"`
	RoleCodePlaceholder string `json:"roleCodePlaceholder"`
	SequenceOrder       string `json:"sequenceOrder"`
	WorkspaceID         string `json:"workspaceId"`
	PeriodID            string `json:"periodId"`
	IsTerminal          string `json:"isTerminal"`
	VersionStatus       string `json:"versionStatus"`
	VersionStatusInfo   string `json:"versionStatusInfo"`
	SequenceOrderInfo   string `json:"sequenceOrderInfo"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	Label             string `json:"label"`
	CheckpointGroupID string `json:"checkpointGroupId"`
	RoleCode          string `json:"roleCode"`
	SequenceOrder     string `json:"sequenceOrder"`
	Version           string `json:"version"`
	VersionStatus     string `json:"versionStatus"`
	WorkspaceID       string `json:"workspaceId"`
	PeriodID          string `json:"periodId"`
	IsTerminal        string `json:"isTerminal"`
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
			Heading:         "Reporting Checkpoints",
			HeadingActive:   "Active Checkpoints",
			HeadingInactive: "Inactive Checkpoints",
			Caption:         "Manage reporting checkpoints for workflow stages",
			CaptionActive:   "Manage your active reporting checkpoints",
			CaptionInactive: "View deprecated or draft checkpoints",
		},
		Buttons: ButtonLabels{
			AddCheckpoint: "Add Checkpoint",
		},
		Columns: ColumnLabels{
			Label:         "Label",
			RoleCode:      "Role",
			SequenceOrder: "Order",
			Version:       "Version",
			VersionStatus: "Status",
		},
		Empty: EmptyLabels{
			Title:           "No checkpoints found",
			Message:         "No reporting checkpoints to display.",
			ActiveTitle:     "No active checkpoints",
			ActiveMessage:   "Create your first reporting checkpoint to get started.",
			InactiveTitle:   "No inactive checkpoints",
			InactiveMessage: "Deprecated or draft checkpoints will appear here.",
		},
		Form: FormLabels{
			Label:               "Label",
			LabelPlaceholder:    "Enter checkpoint label",
			CheckpointGroupID:   "Checkpoint Group ID",
			RoleCode:            "Role Code",
			RoleCodePlaceholder: "Enter role code",
			SequenceOrder:       "Sequence Order",
			WorkspaceID:         "Workspace (optional)",
			PeriodID:            "Period (optional)",
			IsTerminal:          "Terminal Checkpoint",
			VersionStatus:       "Version Status",
			VersionStatusInfo:   "The publication state of this checkpoint version.",
			SequenceOrderInfo:   "Position of this checkpoint in the reporting sequence.",
		},
		Actions: ActionLabels{
			View:   "View Checkpoint",
			Edit:   "Edit Checkpoint",
			Delete: "Delete Checkpoint",
		},
		Detail: DetailLabels{
			PageTitle:         "Checkpoint Details",
			Label:             "Label",
			CheckpointGroupID: "Checkpoint Group ID",
			RoleCode:          "Role Code",
			SequenceOrder:     "Sequence Order",
			Version:           "Version",
			VersionStatus:     "Status",
			WorkspaceID:       "Workspace",
			PeriodID:          "Period",
			IsTerminal:        "Terminal",
			Active:            "Active",
			CreatedDate:       "Created",
			ModifiedDate:      "Last Modified",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Checkpoint",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Reporting checkpoint not found",
			IDRequired:       "Checkpoint ID is required",
			NoPermission:     "No permission",
		},
	}
}
