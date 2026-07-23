package job_template_relation

// Labels holds all translatable strings for the job template relation module.
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
	AddRelation string `json:"add_relation"`
}

type ColumnLabels struct {
	ParentTemplateID string `json:"parent_template_id"`
	ChildTemplateID  string `json:"child_template_id"`
	RelationType     string `json:"relation_type"`
	SequenceOrder    string `json:"sequence_order"`
	Status           string `json:"status"`
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
	ParentTemplateID          string `json:"parent_template_id"`
	ParentTemplateIDInfo      string `json:"parent_template_id_info"`
	ParentTemplatePlaceholder string `json:"parent_template_placeholder"`
	ChildTemplateID           string `json:"child_template_id"`
	ChildTemplateIDInfo       string `json:"child_template_id_info"`
	ChildTemplatePlaceholder  string `json:"child_template_placeholder"`
	RelationType              string `json:"relation_type"`
	RelationTypeInfo          string `json:"relation_type_info"`
	RelationTypeSubTemplate   string `json:"relation_type_sub_template"`
	RelationTypeOnceAtStart   string `json:"relation_type_once_at_start"`
	SequenceOrder             string `json:"sequence_order"`
	SequenceOrderInfo         string `json:"sequence_order_info"`
	Active                    string `json:"active"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
	// Add is the CTA label on the job_template detail Spawn Graph tab.
	Add string `json:"add"`
}

type DetailLabels struct {
	PageTitle        string `json:"page_title"`
	ParentTemplateID string `json:"parent_template_id"`
	ChildTemplateID  string `json:"child_template_id"`
	RelationType     string `json:"relation_type"`
	SequenceOrder    string `json:"sequence_order"`
	Status           string `json:"status"`
	CreatedDate      string `json:"created_date"`
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
			Heading:         "Template Relations",
			HeadingActive:   "Active Relations",
			HeadingInactive: "Inactive Relations",
			Caption:         "Manage parent/child job template spawn edges",
			CaptionActive:   "Active template relation edges",
			CaptionInactive: "Inactive template relation edges",
		},
		Buttons: ButtonLabels{
			AddRelation: "Add Relation",
		},
		Columns: ColumnLabels{
			ParentTemplateID: "Parent Template",
			ChildTemplateID:  "Child Template",
			RelationType:     "Relation Type",
			SequenceOrder:    "Order",
			Status:           "Status",
		},
		Empty: EmptyLabels{
			Title:           "No relations found",
			Message:         "No template relations to display.",
			ActiveTitle:     "No active relations",
			ActiveMessage:   "Create your first template relation to get started.",
			InactiveTitle:   "No inactive relations",
			InactiveMessage: "Deactivated relations will appear here.",
		},
		Form: FormLabels{
			ParentTemplateID:          "Parent Template",
			ParentTemplateIDInfo:      "The template that spawns the child as a sub-engagement or once-at-start job.",
			ParentTemplatePlaceholder: "Select parent template...",
			ChildTemplateID:           "Child Template",
			ChildTemplateIDInfo:       "The template spawned under the parent.",
			ChildTemplatePlaceholder:  "Select child template...",
			RelationType:              "Relation Type",
			RelationTypeInfo:          "How the child spawns relative to the parent.",
			RelationTypeSubTemplate:   "Sub-Template",
			RelationTypeOnceAtStart:   "Once at Engagement Start",
			SequenceOrder:             "Sequence Order",
			SequenceOrderInfo:         "Deterministic spawn order among sibling relations.",
			Active:                    "Active",
		},
		Actions: ActionLabels{
			View:   "View Relation",
			Edit:   "Edit Relation",
			Delete: "Delete Relation",
			Add:    "+ Add Relation",
		},
		Detail: DetailLabels{
			PageTitle:        "Template Relation Details",
			ParentTemplateID: "Parent Template",
			ChildTemplateID:  "Child Template",
			RelationType:     "Relation Type",
			SequenceOrder:    "Sequence Order",
			Status:           "Status",
			CreatedDate:      "Created",
		},
		Tabs: TabLabels{
			Info:    "Information",
			History: "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Relation",
			DeleteMessage: "Are you sure you want to delete this template relation? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Template relation not found",
			IDRequired:       "Relation ID is required",
			NoPermission:     "No permission",
		},
	}
}
