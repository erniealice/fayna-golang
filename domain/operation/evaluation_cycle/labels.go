package evaluation_cycle

// labels.go — EvaluationCycle label structs + DefaultLabels constructor.
//
// View-local label structs (Option-B): the lyngua JSON (camelCase root
// `evaluationCycle`) hydrates these via the ViewAdapter. Banner labels
// (sign-off due / closes / X of Y) live here per LBL-2 (v1, Phase E).

// Labels holds all translatable strings for the evaluation cycle module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Buttons ButtonLabels  `json:"buttons"`
	Columns ColumnLabels  `json:"columns"`
	Empty   EmptyLabels   `json:"empty"`
	Actions ActionLabels  `json:"actions"`
	Detail  DetailLabels  `json:"detail"`
	Tabs    TabLabels     `json:"tabs"`
	Banner  BannerLabels  `json:"banner"`
	Status  StatusLabels  `json:"status"`
	Form    FormLabels    `json:"form"`
	Confirm ConfirmLabels `json:"confirm"`
	Errors  ErrorLabels   `json:"errors"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ButtonLabels struct {
	NewCycle string `json:"newCycle"`
}

type ColumnLabels struct {
	Name       string `json:"name"`
	Engagement string `json:"engagement"`
	Period     string `json:"period"`
	Progress   string `json:"progress"`
	SignOffDue string `json:"signOffDue"`
	Closes     string `json:"closes"`
	Status     string `json:"status"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ActionLabels struct {
	View  string `json:"view"`
	Open  string `json:"open"`
	Close string `json:"close"`
}

type DetailLabels struct {
	PageTitle  string `json:"pageTitle"`
	Name       string `json:"name"`
	Engagement string `json:"engagement"`
	Period     string `json:"period"`
	SignOffDue string `json:"signOffDue"`
	Closes     string `json:"closes"`
	Status     string `json:"status"`
}

type TabLabels struct {
	Info    string `json:"info"`
	Members string `json:"members"`
}

// BannerLabels — the shared "X of Y" progress banner (rendered on cycle
// detail + perf panel + portal). Read-only projection; NOT a form.
type BannerLabels struct {
	ProgressFmt   string `json:"progressFmt"`   // "%d of %d complete"
	SignOffDueFmt string `json:"signOffDueFmt"` // "sign-offs due %s"
	ClosesFmt     string `json:"closesFmt"`     // "closes %s"
	Complete      string `json:"complete"`      // "complete"
}

type StatusLabels struct {
	Open    string `json:"open"`
	SignOff string `json:"signOff"`
	Closed  string `json:"closed"`
}

type FormLabels struct {
	Name           string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Subscription   string `json:"subscription"`
	PeriodStart    string `json:"periodStart"`
	PeriodEnd      string `json:"periodEnd"`
	SignOffDueDate string `json:"signOffDueDate"`
	CloseDate      string `json:"closeDate"`
}

type ConfirmLabels struct {
	Open         string `json:"open"`
	OpenMessage  string `json:"openMessage"`
	Close        string `json:"close"`
	CloseMessage string `json:"closeMessage"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Evaluation Cycles",
			Caption: "Manage performance review cycles and track completion",
		},
		Buttons: ButtonLabels{
			NewCycle: "New Cycle",
		},
		Columns: ColumnLabels{
			Name:       "Name",
			Engagement: "Engagement",
			Period:     "Period",
			Progress:   "Progress",
			SignOffDue: "Sign-off Due",
			Closes:     "Closes",
			Status:     "Status",
		},
		Empty: EmptyLabels{
			Title:   "No evaluation cycles",
			Message: "Create your first cycle to start a review period.",
		},
		Actions: ActionLabels{
			View:  "View Cycle",
			Open:  "Open Cycle",
			Close: "Close Cycle",
		},
		Detail: DetailLabels{
			PageTitle:  "Cycle Details",
			Name:       "Name",
			Engagement: "Engagement",
			Period:     "Period",
			SignOffDue: "Sign-off Due",
			Closes:     "Closes",
			Status:     "Status",
		},
		Tabs: TabLabels{
			Info:    "Information",
			Members: "Members",
		},
		Banner: BannerLabels{
			ProgressFmt:   "%d of %d complete",
			SignOffDueFmt: "sign-offs due %s",
			ClosesFmt:     "closes %s",
			Complete:      "complete",
		},
		Status: StatusLabels{
			Open:    "Open",
			SignOff: "Sign Off",
			Closed:  "Closed",
		},
		Form: FormLabels{
			Name:            "Cycle Name",
			NamePlaceholder: "e.g. H1 2026 Performance Review",
			Subscription:    "Engagement",
			PeriodStart:     "Period Start",
			PeriodEnd:       "Period End",
			SignOffDueDate:  "Sign-off Due Date",
			CloseDate:       "Close Date",
		},
		Confirm: ConfirmLabels{
			Open:         "Open Cycle",
			OpenMessage:  "Open this cycle? Members will be enrolled from active seats.",
			Close:        "Close Cycle",
			CloseMessage: "Close this cycle? No further evaluations can be submitted.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Evaluation cycle not found",
			IDRequired:       "Cycle ID is required",
		},
	}
}
