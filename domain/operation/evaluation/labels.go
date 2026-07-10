package evaluation

// labels.go — Evaluation label structs + DefaultLabels constructor.
//
// Covers: page headings, entity singular/plural, the 4 evaluation statuses
// (draft/submitted/signed_off/archived — Q-SIGNOFF-1 "Sign off" relabel), the 4
// evaluation types, table column headers, list/detail/drawer action labels,
// the polymorphic dimension-slot labels (one per criteria_type branch), the
// detail tabs (Info/Scores/SignOff/Audit), portal "Rate My Team" labels, empty
// states, drawer form labels, and error messages.
//
// All strings are view-local English defaults; lyngua JSON (root key
// "evaluation", camelCase) overlays them at compose time via descriptor.go.

// Labels holds all translatable strings for the evaluation module.
type Labels struct {
	Page      PageLabels      `json:"page"`
	Entity    EntityLabels    `json:"entity"`
	Status    StatusLabels    `json:"status"`
	Type      TypeLabels      `json:"type"`
	Columns   ColumnLabels    `json:"columns"`
	Actions   ActionLabels    `json:"actions"`
	Detail    DetailLabels    `json:"detail"`
	Tabs      TabLabels       `json:"tabs"`
	Scores    ScoresLabels    `json:"scores"`
	Dimension DimensionLabels `json:"dimension"`
	Drawer    DrawerLabels    `json:"drawer"`
	Portal    PortalLabels    `json:"portal"`
	Empty     EmptyLabels     `json:"empty"`
	Form      FormLabels      `json:"form"`
	Errors    ErrorLabels     `json:"errors"`
}

// PageLabels holds the list-page headings.
type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

// EntityLabels holds singular/plural entity names.
type EntityLabels struct {
	Singular string `json:"singular"`
	Plural   string `json:"plural"`
}

// StatusLabels holds the 4 evaluation statuses. SignedOff carries the
// Q-SIGNOFF-1 "Sign off" relabel (v1).
type StatusLabels struct {
	Draft     string `json:"draft"`
	Submitted string `json:"submitted"`
	SignedOff string `json:"signed_off"`
	Archived  string `json:"archived"`
}

// TypeLabels holds the 4 evaluation type chips.
type TypeLabels struct {
	PerformanceReview string `json:"performance_review"`
	CSAT              string `json:"csat"`
	CourseEval        string `json:"course_eval"`
	VendorScorecard   string `json:"vendor_scorecard"`
}

// ColumnLabels holds list table column headers.
type ColumnLabels struct {
	Associate string `json:"associate"`
	Client    string `json:"client"`
	Period    string `json:"period"`
	Type      string `json:"type"`
	Overall   string `json:"overall"`
	Status    string `json:"status"`
	Submitted string `json:"submitted"`
}

// ActionLabels holds list/detail/row action button labels.
type ActionLabels struct {
	View    string `json:"view"`
	SignOff string `json:"sign_off"`
	Archive string `json:"archive"`
	Delete  string `json:"delete"`
	Bulk    string `json:"bulk"`
}

// DetailLabels holds detail-header field labels.
type DetailLabels struct {
	PageTitle    string `json:"page_title"`
	Associate    string `json:"associate"`
	Client       string `json:"client"`
	Period       string `json:"period"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	OverallScore string `json:"overall_score"`
	Narrative    string `json:"narrative"`
}

// TabLabels holds detail-page tab labels.
type TabLabels struct {
	Info    string `json:"info"`
	Scores  string `json:"scores"`
	SignOff string `json:"sign_off"`
	Audit   string `json:"audit"`
}

// ScoresLabels holds the Scores tab (snapshot rows + weighted footer) labels.
type ScoresLabels struct {
	Criterion  string `json:"criterion"`
	Weight     string `json:"weight"`
	Type       string `json:"type"`
	Answer     string `json:"answer"`
	NotScored  string `json:"not_scored"`
	Weighted   string `json:"weighted"`
	OverallNil string `json:"overall_nil"`
}

// DimensionLabels holds the polymorphic dimension-slot labels — one helper per
// criteria_type branch (§A.2).
type DimensionLabels struct {
	Pass       string `json:"pass"`
	Fail       string `json:"fail"`
	Select     string `json:"select"`
	NotScored  string `json:"not_scored"`
	WeightChip string `json:"weight_chip"`
	ScoreOutOf string `json:"score_out_of"`
}

// DrawerLabels holds the score-submission drawer (DF-1) labels.
type DrawerLabels struct {
	Title        string `json:"title"`
	Template     string `json:"template"`
	TemplatePick string `json:"template_pick"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
	Narrative    string `json:"narrative"`
	SaveDraft    string `json:"save_draft"`
	Submit       string `json:"submit"`
	Cancel       string `json:"cancel"`
}

// PortalLabels holds the client-portal "Rate My Team" labels (§H).
type PortalLabels struct {
	Heading      string `json:"heading"`
	Caption      string `json:"caption"`
	StartReview  string `json:"start_review"`
	Rating       string `json:"rating"`
	RateBand     string `json:"rate_band"`
	EmptyTitle   string `json:"empty_title"`
	EmptyMessage string `json:"empty_message"`
}

// EmptyLabels holds empty-state messaging.
type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

// FormLabels holds drawer form placeholders.
type FormLabels struct {
	NarrativePlaceholder string `json:"narrative_placeholder"`
}

// ErrorLabels holds error messaging.
type ErrorLabels struct {
	NotFound         string `json:"not_found"`
	PermissionDenied string `json:"permission_denied"`
	IDRequired       string `json:"id_required"`
	InvalidForm      string `json:"invalid_form"`
	NoClient         string `json:"no_client"`
}

// DefaultLabels returns Labels with English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Reviews",
			Caption: "Performance evaluations across your engagements",
		},
		Entity: EntityLabels{
			Singular: "Review",
			Plural:   "Reviews",
		},
		Status: StatusLabels{
			Draft:     "Draft",
			Submitted: "Submitted",
			SignedOff: "Signed Off",
			Archived:  "Archived",
		},
		Type: TypeLabels{
			PerformanceReview: "Performance Review",
			CSAT:              "CSAT",
			CourseEval:        "Course Evaluation",
			VendorScorecard:   "Vendor Scorecard",
		},
		Columns: ColumnLabels{
			Associate: "Associate",
			Client:    "Client",
			Period:    "Period",
			Type:      "Type",
			Overall:   "Overall",
			Status:    "Status",
			Submitted: "Submitted",
		},
		Actions: ActionLabels{
			View:    "View",
			SignOff: "Sign off",
			Archive: "Archive",
			Delete:  "Delete",
			Bulk:    "Bulk Archive",
		},
		Detail: DetailLabels{
			PageTitle:    "Review Details",
			Associate:    "Associate",
			Client:       "Client",
			Period:       "Period",
			Type:         "Type",
			Status:       "Status",
			OverallScore: "Overall Score",
			Narrative:    "Narrative",
		},
		Tabs: TabLabels{
			Info:    "Information",
			Scores:  "Scores",
			SignOff: "Sign Off",
			Audit:   "Audit",
		},
		Scores: ScoresLabels{
			Criterion:  "Criterion",
			Weight:     "Weight",
			Type:       "Type",
			Answer:     "Answer",
			NotScored:  "(not scored)",
			Weighted:   "Weighted average",
			OverallNil: "Not yet scored",
		},
		Dimension: DimensionLabels{
			Pass:       "Pass",
			Fail:       "Fail",
			Select:     "Select…",
			NotScored:  "(not scored)",
			WeightChip: "Weight",
			ScoreOutOf: "out of",
		},
		Drawer: DrawerLabels{
			Title:        "Rate Performance",
			Template:     "Template",
			TemplatePick: "Select a template",
			PeriodStart:  "Period start",
			PeriodEnd:    "Period end",
			Narrative:    "Narrative",
			SaveDraft:    "Save draft",
			Submit:       "Sign off",
			Cancel:       "Cancel",
		},
		Portal: PortalLabels{
			Heading:      "Rate My Team",
			Caption:      "Rate the associates working on your engagement",
			StartReview:  "Start review",
			Rating:       "Latest rating",
			RateBand:     "Rate band",
			EmptyTitle:   "No associates to rate",
			EmptyMessage: "There are no active associates on your engagement yet.",
		},
		Empty: EmptyLabels{
			Title:   "No reviews found",
			Message: "No evaluations to display for this status.",
		},
		Form: FormLabels{
			NarrativePlaceholder: "Add an optional summary narrative",
		},
		Errors: ErrorLabels{
			NotFound:         "Review not found",
			PermissionDenied: "You do not have permission to perform this action",
			IDRequired:       "Review ID is required",
			InvalidForm:      "Invalid form data",
			NoClient:         "No client context — cannot load reviews",
		},
	}
}
