package operation

// job_labels.go — Job label structs + DefaultJobLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobLabels holds all translatable strings for the job module.
type JobLabels struct {
	Page        JobPageLabels       `json:"page"`
	Buttons     JobButtonLabels     `json:"buttons"`
	Columns     JobColumnLabels     `json:"columns"`
	Empty       JobEmptyLabels      `json:"empty"`
	Form        JobFormLabels       `json:"form"`
	Actions     JobActionLabels     `json:"actions"`
	Detail      JobDetailLabels     `json:"detail"`
	Tabs        JobTabLabels        `json:"tabs"`
	Confirm     JobConfirmLabels    `json:"confirm"`
	Errors      JobErrorLabels      `json:"errors"`
	BulkActions JobBulkActionLabels `json:"bulkActions"`
	// Dashboard labels for the Job live dashboard
	// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	Dashboard JobDashboardLabels `json:"dashboard"`
}

// JobDashboardLabels holds translatable strings for the Job live dashboard.
// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
type JobDashboardLabels struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// Stats
	StatActive    string `json:"statActive"`
	StatDoneMonth string `json:"statDoneMonth"`
	StatOverdue   string `json:"statOverdue"`
	StatHoursWeek string `json:"statHoursWeek"`
	// Widgets
	HoursPerWeek      string `json:"hoursPerWeek"`
	UpcomingDeadlines string `json:"upcomingDeadlines"`
	RecentActivity    string `json:"recentActivity"`
	NoDeadlines       string `json:"noDeadlines"`
	NoActivity        string `json:"noActivity"`
	// Quick actions
	QuickNewJob      string `json:"quickNewJob"`
	QuickNewTemplate string `json:"quickNewTemplate"`
	QuickLogHours    string `json:"quickLogHours"`
	QuickJobCalendar string `json:"quickJobCalendar"`
	// Common
	ViewAll   string `json:"viewAll"`
	AxisHours string `json:"axisHours"`
}

type JobPageLabels struct {
	Heading          string `json:"heading"`
	Caption          string `json:"caption"`
	HeadingDraft     string `json:"headingDraft"`
	CaptionDraft     string `json:"captionDraft"`
	HeadingActive    string `json:"headingActive"`
	CaptionActive    string `json:"captionActive"`
	HeadingCompleted string `json:"headingCompleted"`
	CaptionCompleted string `json:"captionCompleted"`
	HeadingClosed    string `json:"headingClosed"`
	CaptionClosed    string `json:"captionClosed"`
	HeadingPlanned   string `json:"headingPlanned"`
	CaptionPlanned   string `json:"captionPlanned"`
	HeadingReleased  string `json:"headingReleased"`
	CaptionReleased  string `json:"captionReleased"`
	HeadingOnHold    string `json:"headingOnHold"`
	CaptionOnHold    string `json:"captionOnHold"`
}

type JobButtonLabels struct {
	AddJob string `json:"addJob"`
}

type JobColumnLabels struct {
	Name     string `json:"name"`
	Client   string `json:"client"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Location string `json:"location"`
}

type JobEmptyLabels struct {
	Title             string `json:"title"`
	Message           string `json:"message"`
	PhasesTitle       string `json:"phasesTitle"`
	PhasesMessage     string `json:"phasesMessage"`
	ActivitiesTitle   string `json:"activitiesTitle"`
	ActivitiesMessage string `json:"activitiesMessage"`
	SettlementTitle   string `json:"settlementTitle"`
	SettlementMessage string `json:"settlementMessage"`
	OutcomesTitle     string `json:"outcomesTitle"`
	OutcomesMessage   string `json:"outcomesMessage"`

	// 2026-06-01 Wave 4.3 label sweep — budget & actuals tab empty states
	// (job/templates/detail.html job-tab-budget / job-tab-actuals).
	BudgetTitle           string `json:"budgetTitle"`
	BudgetMessage         string `json:"budgetMessage"`
	BudgetNoPhasesTitle   string `json:"budgetNoPhasesTitle"`
	BudgetNoPhasesMessage string `json:"budgetNoPhasesMessage"`
	BudgetNoTasks         string `json:"budgetNoTasks"`
	ActualsTitle          string `json:"actualsTitle"`
	ActualsMessage        string `json:"actualsMessage"`
}

type JobFormLabels struct {
	NamePlaceholder     string `json:"namePlaceholder"`
	ClientPlaceholder   string `json:"clientPlaceholder"`
	LocationPlaceholder string `json:"locationPlaceholder"`
	NameInfo            string `json:"nameInfo"`
	ClientInfo          string `json:"clientInfo"`
	LocationInfo        string `json:"locationInfo"`
}

type JobActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type JobDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	TitlePrefix  string `json:"titlePrefix"`
	SectionInfo  string `json:"sectionInfo"`
	Approval     string `json:"approval"`
	Description  string `json:"description"`
	Quantity     string `json:"quantity"`
	UnitCost     string `json:"unitCost"`
	TotalCost    string `json:"totalCost"`
	EntryDate    string `json:"entryDate"`
	EntryType    string `json:"entryType"`
	PhaseName    string `json:"phaseName"`
	PhaseOrder   string `json:"phaseOrder"`
	PhaseStatus  string `json:"phaseStatus"`
	TaskName     string `json:"taskName"`
	TaskOrder    string `json:"taskOrder"`
	TaskStatus   string `json:"taskStatus"`
	AssignedTo   string `json:"assignedTo"`
	TargetType   string `json:"targetType"`
	TargetID     string `json:"targetId"`
	AllocatedAmt string `json:"allocatedAmount"`
	SettleDate   string `json:"settlementDate"`
	SettleStatus string `json:"settlementStatus"`

	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 / §9 — breadcrumb
	// link rendered when Job.origin_type = ORIGIN_TYPE_SUBSCRIPTION pointing
	// back to the centymo subscription detail page.
	OriginSubscriptionLink string `json:"originSubscriptionLink"`

	// 2026-04-29 milestone-billing plan §4 — phase set-status surface on the
	// Job detail Phases tab. PhasesSectionTitle heads the per-phase mini-table;
	// PhaseMarkComplete is the row CTA; PhaseStatusPending / PhaseStatusActive /
	// PhaseStatusCompleted render the status badge.
	PhasesSectionTitle   string `json:"phasesSectionTitle"`
	PhaseMarkComplete    string `json:"phaseMarkComplete"`
	PhaseStatusPending   string `json:"phaseStatusPending"`
	PhaseStatusActive    string `json:"phaseStatusActive"`
	PhaseStatusCompleted string `json:"phaseStatusCompleted"`

	// 2026-06-01 Wave 4.3 label sweep — Budget tab (job-tab-budget).
	BudgetSectionTitle   string `json:"budgetSectionTitle"`
	BudgetTask           string `json:"budgetTask"`
	BudgetHours          string `json:"budgetHours"`
	BudgetSubtotalSuffix string `json:"budgetSubtotalSuffix"`
	BudgetTotalHours     string `json:"budgetTotalHours"`

	// 2026-06-01 Wave 4.3 label sweep — Actuals tab (job-tab-actuals).
	ActualsSectionTitle  string `json:"actualsSectionTitle"`
	ActualsCount         string `json:"actualsCount"`
	ActualsGrandTotal    string `json:"actualsGrandTotal"`
	VarianceSectionTitle string `json:"varianceSectionTitle"`
	VarianceBudgetHours  string `json:"varianceBudgetHours"`
	VarianceActualsCost  string `json:"varianceActualsCost"`
	VarianceNote         string `json:"varianceNote"`
}

type JobTabLabels struct {
	Info        string `json:"info"`
	Phases      string `json:"phases"`
	Activities  string `json:"activities"`
	Settlement  string `json:"settlement"`
	Outcomes    string `json:"outcomes"`
	Attachments string `json:"attachments"`
	Budget      string `json:"budget"`
	Actuals     string `json:"actuals"`
	History     string `json:"history"`
}

type JobConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type JobErrorLabels struct {
	NotFound             string `json:"notFound"`
	PermissionDenied     string `json:"permissionDenied"`
	InUse                string `json:"inUse"`
	IDRequired           string `json:"idRequired"`
	InvalidForm          string `json:"invalidForm"`
	NoIDs                string `json:"noIds"`
	StatusRequired       string `json:"statusRequired"`
	TargetStatusRequired string `json:"targetStatusRequired"`
}

// JobBulkActionLabels holds translatable strings for job bulk-action controls.
type JobBulkActionLabels struct {
	Delete                 string `json:"delete"`
	BulkDelete             string `json:"bulkDelete"`
	BulkDeleteConfirmTitle string `json:"bulkDeleteConfirmTitle"`
	BulkDeleteConfirmMsg   string `json:"bulkDeleteConfirmMsg"`
	SetStatus              string `json:"setStatus"`
	BulkSetStatus          string `json:"bulkSetStatus"`
	SelectAll              string `json:"selectAll"`
	SelectedCount          string `json:"selectedCount"`
	Cancel                 string `json:"cancel"`
}

// DefaultJobLabels returns JobLabels with sensible English defaults.
func DefaultJobLabels() JobLabels {
	return JobLabels{
		Page: JobPageLabels{
			Heading:          "Jobs",
			Caption:          "Manage operational jobs",
			HeadingDraft:     "Draft Jobs",
			CaptionDraft:     "Review jobs that are still being prepared",
			HeadingActive:    "Active Jobs",
			CaptionActive:    "Track work currently in progress",
			HeadingCompleted: "Completed Jobs",
			CaptionCompleted: "Review jobs that have been completed",
			HeadingClosed:    "Closed Jobs",
			CaptionClosed:    "View archived or closed jobs",
		},
		Buttons: JobButtonLabels{
			AddJob: "Add Job",
		},
		Columns: JobColumnLabels{
			Name:     "Name",
			Client:   "Client",
			Status:   "Status",
			Created:  "Created",
			Location: "Location",
		},
		Empty: JobEmptyLabels{
			Title:             "No jobs found",
			Message:           "No jobs to display.",
			PhasesTitle:       "No phases",
			PhasesMessage:     "This job has no phases defined yet.",
			ActivitiesTitle:   "No activities",
			ActivitiesMessage: "No activity entries have been recorded for this job yet.",
			SettlementTitle:   "No settlements",
			SettlementMessage: "No cost allocations have been settled for this job yet.",
			OutcomesTitle:     "No outcomes",
			OutcomesMessage:   "No outcome evaluations have been recorded for this job yet.",
			// 2026-06-01 Wave 4.3 label sweep — budget & actuals tab empty states.
			BudgetTitle:           "No budget available",
			BudgetMessage:         "No template attached. Budget unavailable until a JobTemplate is linked to this matter.",
			BudgetNoPhasesTitle:   "No phases defined",
			BudgetNoPhasesMessage: "The linked template has no phases or tasks. Add phases to the template to see the budget breakdown.",
			BudgetNoTasks:         "No tasks",
			ActualsTitle:          "No actuals recorded",
			ActualsMessage:        "No activity entries have been posted for this job yet.",
		},
		Form: JobFormLabels{
			NamePlaceholder:     "Enter job name",
			ClientPlaceholder:   "Select client",
			LocationPlaceholder: "Select location",
			NameInfo:            "The name of the job as it appears in lists and documents.",
			ClientInfo:          "The client this job is being performed for.",
			LocationInfo:        "The location or site where this job takes place.",
		},
		Actions: JobActionLabels{
			View:   "View Job",
			Edit:   "Edit Job",
			Delete: "Delete Job",
		},
		Detail: JobDetailLabels{
			PageTitle:    "Job Details",
			TitlePrefix:  "Job ",
			SectionInfo:  "Job Information",
			Approval:     "Approval",
			Description:  "Description",
			Quantity:     "Quantity",
			UnitCost:     "Unit Cost",
			TotalCost:    "Total Cost",
			EntryDate:    "Date",
			EntryType:    "Entry Type",
			PhaseName:    "Phase",
			PhaseOrder:   "Order",
			PhaseStatus:  "Status",
			TaskName:     "Task",
			TaskOrder:    "Task Order",
			TaskStatus:   "Task Status",
			AssignedTo:   "Assigned To",
			TargetType:   "Target Type",
			TargetID:     "Target ID",
			AllocatedAmt: "Allocated Amount",
			SettleDate:   "Settlement Date",
			SettleStatus: "Settlement Status",
			// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 / §9.
			OriginSubscriptionLink: "Spawned from Subscription",
			// 2026-04-29 milestone-billing plan §4 — phase set-status surface.
			PhasesSectionTitle:   "Phases",
			PhaseMarkComplete:    "Mark Complete",
			PhaseStatusPending:   "Pending",
			PhaseStatusActive:    "Active",
			PhaseStatusCompleted: "Completed",
			// 2026-06-01 Wave 4.3 label sweep — Budget tab.
			BudgetSectionTitle:   "Estimated Hours by Phase",
			BudgetTask:           "Task",
			BudgetHours:          "Hours",
			BudgetSubtotalSuffix: "subtotal",
			BudgetTotalHours:     "Total estimated hours",
			// 2026-06-01 Wave 4.3 label sweep — Actuals tab.
			ActualsSectionTitle:  "Cost by Entry Type",
			ActualsCount:         "Count",
			ActualsGrandTotal:    "Grand Total",
			VarianceSectionTitle: "Budget vs Actuals",
			VarianceBudgetHours:  "Budget (estimated hours)",
			VarianceActualsCost:  "Actuals (total cost)",
			VarianceNote:         "Full money-vs-money variance available after Wave 3 (resource bill rates).",
		},
		Tabs: JobTabLabels{
			Info:        "Information",
			Phases:      "Phases",
			Activities:  "Activities",
			Settlement:  "Settlement",
			Outcomes:    "Outcomes",
			Attachments: "Attachments",
			Budget:      "Budget",
			Actuals:     "Actuals",
			History:     "History",
		},
		Confirm: JobConfirmLabels{
			Delete:        "Delete Job",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobErrorLabels{
			NotFound:         "Job not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
		Dashboard: JobDashboardLabels{
			Title:             "Jobs Dashboard",
			Subtitle:          "Active workload, upcoming deadlines, hours logged, and risk",
			StatActive:        "Active Jobs",
			StatDoneMonth:     "Done This Month",
			StatOverdue:       "Overdue",
			StatHoursWeek:     "Hours This Week",
			HoursPerWeek:      "Hours per Week",
			UpcomingDeadlines: "Upcoming Deadlines",
			RecentActivity:    "Recent Activity",
			NoDeadlines:       "No upcoming deadlines",
			NoActivity:        "No recent activity",
			QuickNewJob:       "New Job",
			QuickNewTemplate:  "New Job Template",
			QuickLogHours:     "Log Hours",
			QuickJobCalendar:  "Job Calendar",
			ViewAll:           "View All",
			AxisHours:         "Hours",
		},
	}
}
