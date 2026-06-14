package work_request

// labels.go -- WorkRequest label structs + DefaultLabels constructor.
//
// Covers: page headings, entity singular/plural, all 11 status labels,
// 3 origin labels, 4 SLA states, table column headers, action button labels,
// detail tabs, empty states, form labels, and error messages.

// Labels holds all translatable strings for the work_request module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Entity  EntityLabels  `json:"entity"`
	Status  StatusLabels  `json:"status"`
	Origin  OriginLabels  `json:"origin"`
	SLA     SLALabels     `json:"sla"`
	Columns ColumnLabels  `json:"columns"`
	Actions ActionLabels  `json:"actions"`
	Detail  DetailLabels  `json:"detail"`
	Tabs    TabLabels     `json:"tabs"`
	KPI     KPILabels     `json:"kpi"`
	Empty   EmptyLabels   `json:"empty"`
	Form    FormLabels    `json:"form"`
	Errors  ErrorLabels   `json:"errors"`
	Filters FilterLabels  `json:"filters"`
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
	InReview        string `json:"inReview"`
	Approved        string `json:"approved"`
	Declined        string `json:"declined"`
	Completed       string `json:"completed"`
	Cancelled       string `json:"cancelled"`
	ReturnedForInfo string `json:"returnedForInfo"`
	OnHold          string `json:"onHold"`
	Escalated       string `json:"escalated"`
	PendingOverride string `json:"pendingOverride"`
}

// OriginLabels holds translatable strings for the 3 request origins.
type OriginLabels struct {
	ClientOriginated      string `json:"clientOriginated"`
	ClientRelatedInternal string `json:"clientRelatedInternal"`
	Internal              string `json:"internal"`
}

// SLALabels holds translatable strings for the 4 SLA states.
type SLALabels struct {
	OnTrack  string `json:"onTrack"`
	AtRisk   string `json:"atRisk"`
	Breached string `json:"breached"`
	NoSLA    string `json:"noSla"`
}

// ColumnLabels holds translatable strings for inbox table column headers.
type ColumnLabels struct {
	RequestNumber string `json:"requestNumber"`
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
	LogRequest       string `json:"logRequest"`
	NewInternal      string `json:"newInternal"`
	Assign           string `json:"assign"`
	BulkAssign       string `json:"bulkAssign"`
	SetStatus        string `json:"setStatus"`
	Resolve          string `json:"resolve"`
	Edit             string `json:"edit"`
	View             string `json:"view"`
	Open             string `json:"open"`
	InReview         string `json:"inReview"`
	ReturnForInfo    string `json:"returnForInfo"`
	PutOnHold        string `json:"putOnHold"`
	Escalate         string `json:"escalate"`
	SendToOverride   string `json:"sendToOverride"`
	Approve          string `json:"approve"`
	Decline          string `json:"decline"`
	Resume           string `json:"resume"`
}

// DetailLabels holds translatable strings for the request detail page.
type DetailLabels struct {
	PageTitle      string `json:"pageTitle"`
	SectionInfo    string `json:"sectionInfo"`
	RequestedBy    string `json:"requestedBy"`
	AssignedTo     string `json:"assignedTo"`
	Type           string `json:"type"`
	Priority       string `json:"priority"`
	PriorityNormal string `json:"priorityNormal"`
	PriorityHigh   string `json:"priorityHigh"`
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
	OpenTotal    string `json:"openTotal"`
	SLABreached  string `json:"slaBreached"`
	AvgResponse  string `json:"avgResponse"`
	ResolvedWeek string `json:"resolvedWeek"`
}

// EmptyLabels holds translatable strings for empty state messaging.
type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

// FormLabels holds translatable strings for the request drawer form.
type FormLabels struct {
	TitlePlaceholder         string `json:"titlePlaceholder"`
	DescriptionPlaceholder   string `json:"descriptionPlaceholder"`
	JustificationPlaceholder string `json:"justificationPlaceholder"`
	AssignToPlaceholder      string `json:"assignToPlaceholder"`
	ResolutionNotePlaceholder string `json:"resolutionNotePlaceholder"`
}

// FilterLabels holds translatable strings for inbox filter chips.
type FilterLabels struct {
	AllOpen      string `json:"allOpen"`
	SLABreached  string `json:"slaBreached"`
	HighPriority string `json:"highPriority"`
	NewToday     string `json:"newToday"`
	MyQueue      string `json:"myQueue"`
	Unassigned   string `json:"unassigned"`
}

// ErrorLabels holds translatable strings for error messaging.
type ErrorLabels struct {
	NotFound         string `json:"notFound"`
	PermissionDenied string `json:"permissionDenied"`
	IDRequired       string `json:"idRequired"`
	InvalidForm      string `json:"invalidForm"`
	StatusRequired   string `json:"statusRequired"`
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
