package rating_description_set_entry

// Labels holds all translatable strings for the rating description set entry
// drawer.
type Labels struct {
	Form    FormLabels    `json:"form"`
	Columns ColumnLabels  `json:"columns"`
	Errors  ErrorLabels   `json:"errors"`
}

type FormLabels struct {
	Criterion   string `json:"criterion"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

type ColumnLabels struct {
	Criterion   string `json:"criterion"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

type ErrorLabels struct {
	BandNotInScale   string `json:"band_not_in_scale"`
	Duplicate        string `json:"duplicate"`
	PermissionDenied string `json:"permission_denied"`
	InvalidFormData  string `json:"invalid_form_data"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
	SetNotDraft      string `json:"set_not_draft"`
	// LoadFailed surfaces a picker-data load failure (codex-review-impl2.out.md
	// finding #6 — "surface load failures", not a silently empty picker).
	LoadFailed string `json:"load_failed"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Form: FormLabels{
			Criterion:   "Criterion",
			Level:       "Level",
			Description: "Description",
		},
		Columns: ColumnLabels{
			Criterion:   "Criterion",
			Level:       "Level",
			Description: "Description",
		},
		Errors: ErrorLabels{
			BandNotInScale:   "Level does not exist in the selected scale",
			Duplicate:        "This criterion and level combination already exists",
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Entry not found",
			IDRequired:       "Entry ID is required",
			SetNotDraft:      "Only a Draft set's entries can be changed",
			LoadFailed:       "Failed to load form options. Please try again.",
		},
	}
}
