package work_request

// labels.go -- WorkRequest label structs + DefaultLabels constructor.
//
// Covers: page headings, entity singular/plural, all 11 status labels,
// 3 origin labels, 4 SLA states, table column headers, action button labels,
// detail tabs, empty states, form labels, and error messages.

// Labels holds all translatable strings for the work_request module.
type Labels struct {
	Page    PageLabels   `json:"page"`
	Entity  EntityLabels `json:"entity"`
	Status  StatusLabels `json:"status"`
	Origin  OriginLabels `json:"origin"`
	SLA     SLALabels    `json:"sla"`
	Columns ColumnLabels `json:"columns"`
	Actions ActionLabels `json:"actions"`
	Detail  DetailLabels `json:"detail"`
	Tabs    TabLabels    `json:"tabs"`
	KPI     KPILabels    `json:"kpi"`
	Empty   EmptyLabels  `json:"empty"`
	Form    FormLabels   `json:"form"`
	Errors  ErrorLabels  `json:"errors"`
	Filters FilterLabels `json:"filters"`
}

// PageLabels holds translatable strings for the inbox page headings.
type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

// EntityLabels holds singular/plural entity name labels.
type EntityLabels struct {
	Singular string `json:"singular"`
	Plural   string `json:"plural"`
}

// StatusLabels holds translatable strings for all 11 work request statuses.
type StatusLabels struct {
	New             string `json:"new"`
	Submitted       string `json:"submitted"`
	InReview        string `json:"in_review"`
	Approved        string `json:"approved"`
	Declined        string `json:"declined"`
	Completed       string `json:"completed"`
	Cancelled       string `json:"cancelled"`
	ReturnedForInfo string `json:"returned_for_info"`
	OnHold          string `json:"on_hold"`
	Escalated       string `json:"escalated"`
	PendingOverride string `json:"pending_override"`
}

// OriginLabels holds translatable strings for the 3 request origins.
type OriginLabels struct {
	ClientOriginated      string `json:"client_originated"`
	ClientRelatedInternal string `json:"client_related_internal"`
	Internal              string `json:"internal"`
}

// SLALabels holds translatable strings for the 4 SLA states.
type SLALabels struct {
	OnTrack  string `json:"on_track"`
	AtRisk   string `json:"at_risk"`
	Breached string `json:"breached"`
	NoSLA    string `json:"no_sla"`
}

// ColumnLabels holds translatable strings for inbox table column headers.
type ColumnLabels struct {
	RequestNumber string `json:"request_number"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	Origin        string `json:"origin"`
	Client        string `json:"client"`
	Priority      string `json:"priority"`
	SLA           string `json:"sla"`
	Assigned      string `json:"assigned"`
	Raised        string `json:"raised"`
}

// ActionLabels holds translatable strings for action button labels.
type ActionLabels struct {
	LogRequest     string `json:"log_request"`
	NewInternal    string `json:"new_internal"`
	Assign         string `json:"assign"`
	BulkAssign     string `json:"bulk_assign"`
	SetStatus      string `json:"set_status"`
	Resolve        string `json:"resolve"`
	Edit           string `json:"edit"`
	View           string `json:"view"`
	Open           string `json:"open"`
	InReview       string `json:"in_review"`
	ReturnForInfo  string `json:"return_for_info"`
	PutOnHold      string `json:"put_on_hold"`
	Escalate       string `json:"escalate"`
	SendToOverride string `json:"send_to_override"`
	Approve        string `json:"approve"`
	Decline        string `json:"decline"`
	Resume         string `json:"resume"`
}

// DetailLabels holds translatable strings for the request detail page.
type DetailLabels struct {
	PageTitle      string `json:"page_title"`
	SectionInfo    string `json:"section_info"`
	RequestedBy    string `json:"requested_by"`
	AssignedTo     string `json:"assigned_to"`
	Type           string `json:"type"`
	Priority       string `json:"priority"`
	PriorityNormal string `json:"priority_normal"`
	PriorityHigh   string `json:"priority_high"`
}

// TabLabels holds translatable strings for detail page tab labels.
type TabLabels struct {
	Info        string `json:"info"`
	Timeline    string `json:"timeline"`
	Attachments string `json:"attachments"`
	Messages    string `json:"messages"`
}

// KPILabels holds translatable strings for the inbox KPI stat cards.
type KPILabels struct {
	OpenTotal    string `json:"open_total"`
	SLABreached  string `json:"sla_breached"`
	AvgResponse  string `json:"avg_response"`
	ResolvedWeek string `json:"resolved_week"`
}

// EmptyLabels holds translatable strings for empty state messaging.
type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

// FormLabels holds translatable strings for the request drawer form.
type FormLabels struct {
	TitlePlaceholder          string `json:"title_placeholder"`
	DescriptionPlaceholder    string `json:"description_placeholder"`
	JustificationPlaceholder  string `json:"justification_placeholder"`
	AssignToPlaceholder       string `json:"assign_to_placeholder"`
	ResolutionNotePlaceholder string `json:"resolution_note_placeholder"`
}

// FilterLabels holds translatable strings for inbox filter chips.
type FilterLabels struct {
	AllOpen      string `json:"all_open"`
	SLABreached  string `json:"sla_breached"`
	HighPriority string `json:"high_priority"`
	NewToday     string `json:"new_today"`
	MyQueue      string `json:"my_queue"`
	Unassigned   string `json:"unassigned"`
}

// ErrorLabels holds translatable strings for error messaging.
type ErrorLabels struct {
	NotFound         string `json:"not_found"`
	PermissionDenied string `json:"permission_denied"`
	IDRequired       string `json:"id_required"`
	InvalidForm      string `json:"invalid_form"`
	StatusRequired   string `json:"status_required"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Requests",
			Caption: "Manage work requests and approvals",
		},
		Entity: EntityLabels{
			Singular: "Request",
			Plural:   "Requests",
		},
		Status: StatusLabels{
			New:             "New",
			Submitted:       "Submitted",
			InReview:        "In Review",
			Approved:        "Approved",
			Declined:        "Declined",
			Completed:       "Completed",
			Cancelled:       "Cancelled",
			ReturnedForInfo: "Returned for Info",
			OnHold:          "On Hold",
			Escalated:       "Escalated",
			PendingOverride: "Pending Override",
		},
		Origin: OriginLabels{
			ClientOriginated:      "Client",
			ClientRelatedInternal: "Client-related",
			Internal:              "Internal",
		},
		SLA: SLALabels{
			OnTrack:  "On Track",
			AtRisk:   "At Risk",
			Breached: "Breached",
			NoSLA:    "No SLA",
		},
		Columns: ColumnLabels{
			RequestNumber: "Request",
			Title:         "Title",
			Status:        "Status",
			Origin:        "Origin",
			Client:        "Client",
			Priority:      "Priority",
			SLA:           "SLA",
			Assigned:      "Assigned",
			Raised:        "Raised",
		},
		Actions: ActionLabels{
			LogRequest:     "Log Request",
			NewInternal:    "New Internal Request",
			Assign:         "Assign",
			BulkAssign:     "Bulk Assign",
			SetStatus:      "Set Status",
			Resolve:        "Resolve",
			Edit:           "Edit",
			View:           "View",
			Open:           "Open",
			InReview:       "In Review",
			ReturnForInfo:  "Return for Info",
			PutOnHold:      "Put on Hold",
			Escalate:       "Escalate",
			SendToOverride: "Send to Override",
			Approve:        "Approve",
			Decline:        "Decline",
			Resume:         "Resume",
		},
		Detail: DetailLabels{
			PageTitle:      "Request Details",
			SectionInfo:    "Request Information",
			RequestedBy:    "Requested By",
			AssignedTo:     "Assigned To",
			Type:           "Type",
			Priority:       "Priority",
			PriorityNormal: "Normal",
			PriorityHigh:   "High",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Timeline:    "Activity",
			Attachments: "Attachments",
			Messages:    "Messages",
		},
		KPI: KPILabels{
			OpenTotal:    "Open",
			SLABreached:  "SLA Breached",
			AvgResponse:  "Avg. Response",
			ResolvedWeek: "Resolved This Week",
		},
		Empty: EmptyLabels{
			Title:   "No requests found",
			Message: "No work requests to display.",
		},
		Form: FormLabels{
			TitlePlaceholder:          "Enter request title",
			DescriptionPlaceholder:    "Describe the request",
			JustificationPlaceholder:  "Provide justification (min 10 characters)",
			AssignToPlaceholder:       "Select assignee",
			ResolutionNotePlaceholder: "Add a resolution note (optional)",
		},
		Filters: FilterLabels{
			AllOpen:      "All Open",
			SLABreached:  "SLA Breached",
			HighPriority: "High Priority",
			NewToday:     "New Today",
			MyQueue:      "My Queue",
			Unassigned:   "Unassigned",
		},
		Errors: ErrorLabels{
			NotFound:         "Request not found",
			PermissionDenied: "You do not have permission to perform this action",
			IDRequired:       "Request ID is required",
			InvalidForm:      "Invalid form data",
			StatusRequired:   "Target status is required",
		},
	}
}
