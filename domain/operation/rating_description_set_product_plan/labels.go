package rating_description_set_product_plan

// Labels holds all translatable strings for the rating description set
// product plan (AY setup / "Descriptor Assignments") module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Filters FilterLabels  `json:"filters"`
	Status  StatusLabels  `json:"status"`
	Columns ColumnLabels  `json:"columns"`
	Actions ActionLabels  `json:"actions"`
	Form    FormLabels    `json:"form"`
	Errors  ErrorLabels   `json:"errors"`
	Empty   EmptyLabels   `json:"empty"`
}

type PageLabels struct {
	Heading string `json:"heading"`
}

type FilterLabels struct {
	PriceSchedule string `json:"price_schedule"`
	Status        string `json:"status"`
}

type StatusLabels struct {
	Linked   string `json:"linked"`
	Unlinked string `json:"unlinked"`
	Invalid  string `json:"invalid"`
}

type ColumnLabels struct {
	Offering string `json:"offering"`
	Variant  string `json:"variant"`
	Set      string `json:"set"`
	Version  string `json:"version"`
	Status   string `json:"status"`
}

type ActionLabels struct {
	Link   string `json:"link"`
	Relink string `json:"relink"`
	Unlink string `json:"unlink"`
}

type FormLabels struct {
	Offering      string `json:"offering"`
	Set           string `json:"set"`
	Reason        string `json:"reason"`
	PriceSchedule string `json:"price_schedule"`
}

type ErrorLabels struct {
	Stale            string `json:"stale"`
	SetNotPublished  string `json:"set_not_published"`
	PermissionDenied string `json:"permission_denied"`
	InvalidFormData  string `json:"invalid_form_data"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
	// LoadFailed surfaces a picker-data load failure (codex-review-impl2.out.md
	// finding #6 — "surface load failures", not a silently empty picker).
	LoadFailed string `json:"load_failed"`
}

type EmptyLabels struct {
	Title            string `json:"title"`
	Message          string `json:"message"`
	NoScheduleTitle  string `json:"no_schedule_title"`
	NoScheduleMsg    string `json:"no_schedule_message"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Rating Description Links",
		},
		Filters: FilterLabels{
			PriceSchedule: "Price Schedule",
			Status:        "Status",
		},
		Status: StatusLabels{
			Linked:   "Linked",
			Unlinked: "Unlinked",
			Invalid:  "Invalid",
		},
		Columns: ColumnLabels{
			Offering: "Offering",
			Variant:  "Variant",
			Set:      "Rating Description Set",
			Version:  "Version",
			Status:   "Status",
		},
		Actions: ActionLabels{
			Link:   "Link Set",
			Relink: "Change Set",
			Unlink: "Unlink",
		},
		Form: FormLabels{
			Offering:      "Offering",
			Set:           "Rating Description Set",
			Reason:        "Reason for Change",
			PriceSchedule: "Academic Year",
		},
		Errors: ErrorLabels{
			Stale:            "This link has been changed by another user",
			SetNotPublished:  "Only published sets can be linked",
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Link not found",
			IDRequired:       "Link ID is required",
			LoadFailed:       "Failed to load form options. Please try again.",
		},
		Empty: EmptyLabels{
			Title:           "No linked offerings yet",
			Message:         "Link a published rating description set to an offering for this academic year.",
			NoScheduleTitle: "No academic year selected",
			NoScheduleMsg:   "Choose an academic year to see its descriptor assignments.",
		},
	}
}
