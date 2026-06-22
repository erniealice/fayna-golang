package template_task_criteria

// Labels holds all translatable strings for the template task criteria module.
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
	JobTemplateTaskID string `json:"jobTemplateTaskId"`
	OutcomeCriteriaID string `json:"outcomeCriteriaId"`
	SequenceOrder     string `json:"sequenceOrder"`
	Status            string `json:"status"`
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
	JobTemplateTaskID     string `json:"jobTemplateTaskId"`
	JobTemplateTaskIDInfo string `json:"jobTemplateTaskIdInfo"`
	OutcomeCriteriaID     string `json:"outcomeCriteriaId"`
	OutcomeCriteriaIDInfo string `json:"outcomeCriteriaIdInfo"`
	SequenceOrder         string `json:"sequenceOrder"`
	SequenceOrderInfo     string `json:"sequenceOrderInfo"`
	RequiredOverride      string `json:"requiredOverride"`
	RequiredOverrideInfo  string `json:"requiredOverrideInfo"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	JobTemplateTaskID string `json:"jobTemplateTaskId"`
	OutcomeCriteriaID string `json:"outcomeCriteriaId"`
	SequenceOrder     string `json:"sequenceOrder"`
	RequiredOverride  string `json:"requiredOverride"`
	Status            string `json:"status"`
	CreatedDate       string `json:"createdDate"`
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
			Heading:         "Template Task Criteria",
			HeadingActive:   "Active Links",
			HeadingInactive: "Inactive Links",
			Caption:         "Manage criteria linked to job template tasks",
			CaptionActive:   "Active template task criteria links",
			CaptionInactive: "Inactive template task criteria links",
		},
		Buttons: ButtonLabels{
			AddLink: "Add Link",
		},
		Columns: ColumnLabels{
			JobTemplateTaskID: "Job Template Task",
			OutcomeCriteriaID: "Outcome Criterion",
			SequenceOrder:     "Order",
			Status:            "Status",
		},
		Empty: EmptyLabels{
			Title:           "No links found",
			Message:         "No template task criteria links to display.",
			ActiveTitle:     "No active links",
			ActiveMessage:   "Create your first template task criteria link to get started.",
			InactiveTitle:   "No inactive links",
			InactiveMessage: "Deactivated links will appear here.",
		},
		Form: FormLabels{
			JobTemplateTaskID:     "Job Template Task ID",
			JobTemplateTaskIDInfo: "The job template task this criteria link belongs to.",
			OutcomeCriteriaID:     "Outcome Criteria ID",
			OutcomeCriteriaIDInfo: "The outcome criterion being linked to the template task.",
			SequenceOrder:         "Sequence Order",
			SequenceOrderInfo:     "Display order of this criterion within the task.",
			RequiredOverride:      "Required Override",
			RequiredOverrideInfo:  "Override whether this criterion is required for the task.",
		},
		Actions: ActionLabels{
			View:   "View Link",
			Edit:   "Edit Link",
			Delete: "Delete Link",
		},
		Detail: DetailLabels{
			PageTitle:         "Template Task Criteria Details",
			JobTemplateTaskID: "Job Template Task",
			OutcomeCriteriaID: "Outcome Criterion",
			SequenceOrder:     "Sequence Order",
			RequiredOverride:  "Required Override",
			Status:            "Status",
			CreatedDate:       "Created",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Link",
			DeleteMessage: "Are you sure you want to delete this template task criteria link? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Template task criteria link not found",
			IDRequired:       "Link ID is required",
			NoPermission:     "No permission",
		},
	}
}
