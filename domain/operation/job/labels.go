package job

// job_labels.go — Job label structs + DefaultLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// Labels holds all translatable strings for the job module.
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
	BulkActions BulkActionLabels `json:"bulk_actions"`
	// Dashboard labels for the Job live dashboard
	// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	Dashboard DashboardLabels `json:"dashboard"`
}

// DashboardLabels holds translatable strings for the Job live dashboard.
// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
type DashboardLabels struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// Stats
	StatActive    string `json:"stat_active"`
	StatDoneMonth string `json:"stat_done_month"`
	StatOverdue   string `json:"stat_overdue"`
	StatHoursWeek string `json:"stat_hours_week"`
	// Widgets
	HoursPerWeek      string `json:"hours_per_week"`
	UpcomingDeadlines string `json:"upcoming_deadlines"`
	RecentActivity    string `json:"recent_activity"`
	NoDeadlines       string `json:"no_deadlines"`
	NoActivity        string `json:"no_activity"`
	// Quick actions
	QuickNewJob      string `json:"quick_new_job"`
	QuickNewTemplate string `json:"quick_new_template"`
	QuickLogHours    string `json:"quick_log_hours"`
	QuickJobCalendar string `json:"quick_job_calendar"`
	// Common
	ViewAll   string `json:"view_all"`
	AxisHours string `json:"axis_hours"`
}

type PageLabels struct {
	Heading          string `json:"heading"`
	Caption          string `json:"caption"`
	HeadingDraft     string `json:"heading_draft"`
	CaptionDraft     string `json:"caption_draft"`
	HeadingActive    string `json:"heading_active"`
	CaptionActive    string `json:"caption_active"`
	HeadingCompleted string `json:"heading_completed"`
	CaptionCompleted string `json:"caption_completed"`
	HeadingClosed    string `json:"heading_closed"`
	CaptionClosed    string `json:"caption_closed"`
	HeadingPlanned   string `json:"heading_planned"`
	CaptionPlanned   string `json:"caption_planned"`
	HeadingReleased  string `json:"heading_released"`
	CaptionReleased  string `json:"caption_released"`
	HeadingOnHold    string `json:"heading_on_hold"`
	CaptionOnHold    string `json:"caption_on_hold"`
}

type ButtonLabels struct {
	AddJob string `json:"add_job"`
}

type ColumnLabels struct {
	Name     string `json:"name"`
	Client   string `json:"client"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Location string `json:"location"`
}

type EmptyLabels struct {
	Title             string `json:"title"`
	Message           string `json:"message"`
	PhasesTitle       string `json:"phases_title"`
	PhasesMessage     string `json:"phases_message"`
	ActivitiesTitle   string `json:"activities_title"`
	ActivitiesMessage string `json:"activities_message"`
	SettlementTitle   string `json:"settlement_title"`
	SettlementMessage string `json:"settlement_message"`
	OutcomesTitle     string `json:"outcomes_title"`
	OutcomesMessage   string `json:"outcomes_message"`

	// 2026-06-01 Wave 4.3 label sweep — budget & actuals tab empty states
	// (job/templates/detail.html job-tab-budget / job-tab-actuals).
	BudgetTitle           string `json:"budget_title"`
	BudgetMessage         string `json:"budget_message"`
	BudgetNoPhasesTitle   string `json:"budget_no_phases_title"`
	BudgetNoPhasesMessage string `json:"budget_no_phases_message"`
	BudgetNoTasks         string `json:"budget_no_tasks"`
	ActualsTitle          string `json:"actuals_title"`
	ActualsMessage        string `json:"actuals_message"`
}

type FormLabels struct {
	NamePlaceholder     string `json:"name_placeholder"`
	ClientPlaceholder   string `json:"client_placeholder"`
	LocationPlaceholder string `json:"location_placeholder"`
	NameInfo            string `json:"name_info"`
	ClientInfo          string `json:"client_info"`
	LocationInfo        string `json:"location_info"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle    string `json:"page_title"`
	TitlePrefix  string `json:"title_prefix"`
	SectionInfo  string `json:"section_info"`
	Approval     string `json:"approval"`
	Description  string `json:"description"`
	Quantity     string `json:"quantity"`
	UnitCost     string `json:"unit_cost"`
	TotalCost    string `json:"total_cost"`
	EntryDate    string `json:"entry_date"`
	EntryType    string `json:"entry_type"`
	PhaseName    string `json:"phase_name"`
	PhaseOrder   string `json:"phase_order"`
	PhaseStatus  string `json:"phase_status"`
	TaskName     string `json:"task_name"`
	TaskOrder    string `json:"task_order"`
	TaskStatus   string `json:"task_status"`
	AssignedTo   string `json:"assigned_to"`
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
	AllocatedAmt string `json:"allocated_amount"`
	SettleDate   string `json:"settlement_date"`
	SettleStatus string `json:"settlement_status"`

	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 / §9 — breadcrumb
	// link rendered when Job.origin_type = ORIGIN_TYPE_SUBSCRIPTION pointing
	// back to the centymo subscription detail page.
	OriginSubscriptionLink string `json:"origin_subscription_link"`

	// 2026-04-29 milestone-billing plan §4 — phase set-status surface on the
	// Job detail Phases tab. PhasesSectionTitle heads the per-phase mini-table;
	// PhaseMarkComplete is the row CTA; PhaseStatusPending / PhaseStatusActive /
	// PhaseStatusCompleted render the status badge.
	PhasesSectionTitle   string `json:"phases_section_title"`
	PhaseMarkComplete    string `json:"phase_mark_complete"`
	PhaseStatusPending   string `json:"phase_status_pending"`
	PhaseStatusActive    string `json:"phase_status_active"`
	PhaseStatusCompleted string `json:"phase_status_completed"`

	// 2026-06-01 Wave 4.3 label sweep — Budget tab (job-tab-budget).
	BudgetSectionTitle   string `json:"budget_section_title"`
	BudgetTask           string `json:"budget_task"`
	BudgetHours          string `json:"budget_hours"`
	BudgetSubtotalSuffix string `json:"budget_subtotal_suffix"`
	BudgetTotalHours     string `json:"budget_total_hours"`

	// 2026-06-01 Wave 4.3 label sweep — Actuals tab (job-tab-actuals).
	ActualsSectionTitle  string `json:"actuals_section_title"`
	ActualsCount         string `json:"actuals_count"`
	ActualsGrandTotal    string `json:"actuals_grand_total"`
	VarianceSectionTitle string `json:"variance_section_title"`
	VarianceBudgetHours  string `json:"variance_budget_hours"`
	VarianceActualsCost  string `json:"variance_actuals_cost"`
	VarianceNote         string `json:"variance_note"`
}

type TabLabels struct {
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

type ConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"delete_message"`
}

type ErrorLabels struct {
	NotFound             string `json:"not_found"`
	PermissionDenied     string `json:"permission_denied"`
	InUse                string `json:"in_use"`
	IDRequired           string `json:"id_required"`
	InvalidForm          string `json:"invalid_form"`
	NoIDs                string `json:"no_ids"`
	StatusRequired       string `json:"status_required"`
	TargetStatusRequired string `json:"target_status_required"`
}

// BulkActionLabels holds translatable strings for job bulk-action controls.
type BulkActionLabels struct {
	Delete                 string `json:"delete"`
	BulkDelete             string `json:"bulk_delete"`
	BulkDeleteConfirmTitle string `json:"bulk_delete_confirm_title"`
	BulkDeleteConfirmMsg   string `json:"bulk_delete_confirm_msg"`
	SetStatus              string `json:"set_status"`
	BulkSetStatus          string `json:"bulk_set_status"`
	SelectAll              string `json:"select_all"`
	SelectedCount          string `json:"selected_count"`
	Cancel                 string `json:"cancel"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
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
		Buttons: ButtonLabels{
			AddJob: "Add Job",
		},
		Columns: ColumnLabels{
			Name:     "Name",
			Client:   "Client",
			Status:   "Status",
			Created:  "Created",
			Location: "Location",
		},
		Empty: EmptyLabels{
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
		Form: FormLabels{
			NamePlaceholder:     "Enter job name",
			ClientPlaceholder:   "Select client",
			LocationPlaceholder: "Select location",
			NameInfo:            "The name of the job as it appears in lists and documents.",
			ClientInfo:          "The client this job is being performed for.",
			LocationInfo:        "The location or site where this job takes place.",
		},
		Actions: ActionLabels{
			View:   "View Job",
			Edit:   "Edit Job",
			Delete: "Delete Job",
		},
		Detail: DetailLabels{
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
		Tabs: TabLabels{
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
		Confirm: ConfirmLabels{
			Delete:        "Delete Job",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			NotFound:         "Job not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
		Dashboard: DashboardLabels{
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
