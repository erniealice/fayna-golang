package operation

// job_template_labels.go — JobTemplate label structs + DefaultJobTemplateLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobTemplateLabels holds all translatable strings for the job template module.
type JobTemplateLabels struct {
	Page        JobTemplatePageLabels       `json:"page"`
	Buttons     JobTemplateButtonLabels     `json:"buttons"`
	Columns     JobTemplateColumnLabels     `json:"columns"`
	Empty       JobTemplateEmptyLabels      `json:"empty"`
	Form        JobTemplateFormLabels       `json:"form"`
	Actions     JobTemplateActionLabels     `json:"actions"`
	Detail      JobTemplateDetailLabels     `json:"detail"`
	Tabs        JobTemplateTabLabels        `json:"tabs"`
	Confirm     JobTemplateConfirmLabels    `json:"confirm"`
	Errors      JobTemplateErrorLabels      `json:"errors"`
	BulkActions JobTemplateBulkActionLabels `json:"bulkActions"`
}

type JobTemplatePageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type JobTemplateButtonLabels struct {
	AddJobTemplate string `json:"addJobTemplate"`
}

type JobTemplateColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type JobTemplateEmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type JobTemplateActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
	// Add is the CTA label on the Phases tab.
	Add string `json:"add"`
	// AddTask is the CTA label on the Tasks tab.
	AddTask string `json:"addTask"`
}

type JobTemplateErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
	InUse            string `json:"inUse"`
	InvalidForm      string `json:"invalidForm"`
}

type JobTemplateFormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"descriptionPlaceholder"`
	Active          string `json:"active"`
}

type JobTemplateDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
}

type JobTemplateTabLabels struct {
	Info        string `json:"info"`
	Phases      string `json:"phases"`
	Tasks       string `json:"tasks"`
	Standards   string `json:"standards"`
	Attachments string `json:"attachments"`
	AuditTrail  string `json:"auditTrail"`
	History     string `json:"history"`
}

type JobTemplateConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

// JobTemplateBulkActionLabels holds translatable strings for job template bulk-action controls.
type JobTemplateBulkActionLabels struct {
	Delete                 string `json:"delete"`
	BulkDelete             string `json:"bulkDelete"`
	BulkDeleteConfirmTitle string `json:"bulkDeleteConfirmTitle"`
	BulkDeleteConfirmMsg   string `json:"bulkDeleteConfirmMsg"`
	SelectAll              string `json:"selectAll"`
	SelectedCount          string `json:"selectedCount"`
	Cancel                 string `json:"cancel"`
}

// DefaultJobTemplateLabels returns JobTemplateLabels with sensible English defaults.
func DefaultJobTemplateLabels() JobTemplateLabels {
	return JobTemplateLabels{
		Page: JobTemplatePageLabels{
			Heading:         "Job Templates",
			HeadingActive:   "Active Job Templates",
			HeadingInactive: "Inactive Job Templates",
			Caption:         "Manage reusable execution plans",
			CaptionActive:   "Manage your active job templates",
			CaptionInactive: "View inactive or archived job templates",
		},
		Buttons: JobTemplateButtonLabels{
			AddJobTemplate: "Add Template",
		},
		Columns: JobTemplateColumnLabels{
			Name:        "Name",
			Description: "Description",
			Status:      "Status",
		},
		Empty: JobTemplateEmptyLabels{
			Title:           "No job templates found",
			Message:         "No job templates to display.",
			ActiveTitle:     "No active job templates",
			ActiveMessage:   "Create your first job template to get started.",
			InactiveTitle:   "No inactive job templates",
			InactiveMessage: "Deactivated templates will appear here.",
		},
		Form: JobTemplateFormLabels{
			Name:            "Template Name",
			NamePlaceholder: "Enter template name",
			Description:     "Description",
			DescPlaceholder: "Enter template description...",
			Active:          "Active",
		},
		Actions: JobTemplateActionLabels{
			View:    "View Template",
			Edit:    "Edit Template",
			Delete:  "Delete Template",
			Add:     "+ Add Phase",
			AddTask: "+ Add Task",
		},
		Detail: JobTemplateDetailLabels{
			PageTitle:    "Job Template Details",
			Name:         "Name",
			Description:  "Description",
			Status:       "Status",
			CreatedDate:  "Created",
			ModifiedDate: "Last Modified",
		},
		Tabs: JobTemplateTabLabels{
			Info:        "Information",
			Phases:      "Phases",
			Tasks:       "Tasks",
			Standards:   "Standards",
			Attachments: "Attachments",
			AuditTrail:  "Audit Trail",
			History:     "History",
		},
		Confirm: JobTemplateConfirmLabels{
			Delete:        "Delete Template",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobTemplateErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Job template not found",
			IDRequired:       "Job template ID is required",
			NoPermission:     "No permission",
		},
	}
}
