package job_template

// job_template_labels.go — JobTemplate label structs + DefaultLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobTemplateLabels holds all translatable strings for the job template module.
type Labels struct {
	Page        PageLabels       `json:"page"`
	Buttons     ButtonLabels     `json:"buttons"`
	Columns     ColumnLabels     `json:"columns"`
	Empty       EmptyLabels      `json:"empty"`
	Form        FormLabels       `json:"form"`
	Actions     ActionLabels     `json:"actions"`
	Detail      DetailLabels     `json:"detail"`
	Tabs        TabLabels        `json:"tabs"`
	Confirm     ConfirmLabels    `json:"confirm"`
	Errors      ErrorLabels      `json:"errors"`
	BulkActions BulkActionLabels `json:"bulkActions"`
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
	AddJobTemplate string `json:"addJobTemplate"`
}

type ColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type EmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
	// Add is the CTA label on the Phases tab.
	Add string `json:"add"`
	// AddTask is the CTA label on the Tasks tab.
	AddTask string `json:"addTask"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
	InUse            string `json:"inUse"`
	InvalidForm      string `json:"invalidForm"`
}

type FormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"descriptionPlaceholder"`
	Active          string `json:"active"`
}

type DetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
}

type TabLabels struct {
	Info        string `json:"info"`
	Phases      string `json:"phases"`
	Tasks       string `json:"tasks"`
	Standards   string `json:"standards"`
	Attachments string `json:"attachments"`
	AuditTrail  string `json:"auditTrail"`
	History     string `json:"history"`
}

type ConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

// BulkActionLabels holds translatable strings for job template bulk-action controls.
type BulkActionLabels struct {
	Delete                 string `json:"delete"`
	BulkDelete             string `json:"bulkDelete"`
	BulkDeleteConfirmTitle string `json:"bulkDeleteConfirmTitle"`
	BulkDeleteConfirmMsg   string `json:"bulkDeleteConfirmMsg"`
	SelectAll              string `json:"selectAll"`
	SelectedCount          string `json:"selectedCount"`
	Cancel                 string `json:"cancel"`
}

// DefaultLabels returns JobTemplateLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading:         "Job Templates",
			HeadingActive:   "Active Job Templates",
			HeadingInactive: "Inactive Job Templates",
			Caption:         "Manage reusable execution plans",
			CaptionActive:   "Manage your active job templates",
			CaptionInactive: "View inactive or archived job templates",
		},
		Buttons: ButtonLabels{
			AddJobTemplate: "Add Template",
		},
		Columns: ColumnLabels{
			Name:        "Name",
			Description: "Description",
			Status:      "Status",
		},
		Empty: EmptyLabels{
			Title:           "No job templates found",
			Message:         "No job templates to display.",
			ActiveTitle:     "No active job templates",
			ActiveMessage:   "Create your first job template to get started.",
			InactiveTitle:   "No inactive job templates",
			InactiveMessage: "Deactivated templates will appear here.",
		},
		Form: FormLabels{
			Name:            "Template Name",
			NamePlaceholder: "Enter template name",
			Description:     "Description",
			DescPlaceholder: "Enter template description...",
			Active:          "Active",
		},
		Actions: ActionLabels{
			View:    "View Template",
			Edit:    "Edit Template",
			Delete:  "Delete Template",
			Add:     "+ Add Phase",
			AddTask: "+ Add Task",
		},
		Detail: DetailLabels{
			PageTitle:    "Job Template Details",
			Name:         "Name",
			Description:  "Description",
			Status:       "Status",
			CreatedDate:  "Created",
			ModifiedDate: "Last Modified",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Phases:      "Phases",
			Tasks:       "Tasks",
			Standards:   "Standards",
			Attachments: "Attachments",
			AuditTrail:  "Audit Trail",
			History:     "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Template",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Job template not found",
			IDRequired:       "Job template ID is required",
			NoPermission:     "No permission",
		},
	}
}
