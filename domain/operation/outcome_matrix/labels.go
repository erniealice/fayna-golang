package outcome_matrix

// labels.go — OutcomeMatrix label structs + DefaultLabels constructor.
//
// The label tree mirrors the lyngua files at
// translations/en/{general,education}/outcome_matrix.json (root key
// "outcomeMatrix"). The "grid" section embeds pyeza's CellGridLabels verbatim
// so the shared strings (saveButton/savedBanner/clientColumn/...) round-trip
// straight onto the component config; the three grid keys that CellGridLabels
// does NOT model (criterionColumn / scoreColumn / readOnlyTooltip) are added
// alongside it via struct embedding, so they unmarshal from the SAME "grid"
// JSON object level.

import (
	pyezatypes "github.com/erniealice/pyeza-golang/types"
)

// Labels holds all translatable strings for the outcome matrix module.
type Labels struct {
	Page   PageLabels   `json:"page"`
	Picker PickerLabels `json:"picker"`
	Scope  ScopeLabels  `json:"scope"`
	Grid   GridLabels   `json:"grid"`
	Errors ErrorLabels  `json:"errors"`
}

// PageLabels — page heading.
type PageLabels struct {
	Title string `json:"title"`
}

// PickerLabels — the template/subject picker widget label.
type PickerLabels struct {
	Template string `json:"template"`
}

// ScopeLabels — the scope-toggle options ("mine" vs "all").
type ScopeLabels struct {
	Mine string `json:"mine"`
	All  string `json:"all"`
}

// GridLabels embeds pyeza's CellGridLabels (the fields that map 1:1 onto the
// cell-grid component config) and adds the three grid-object keys that
// CellGridLabels does not model. Because CellGridLabels is embedded anonymously,
// its json-tagged fields promote to the same object level, so a single "grid"
// JSON object populates both.
type GridLabels struct {
	pyezatypes.CellGridLabels
	CriterionColumn string `json:"criterionColumn"`
	ScoreColumn     string `json:"scoreColumn"`
	ReadOnlyTooltip string `json:"readOnlyTooltip"`
}

// ErrorLabels — generic permission-denied string.
type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
}

// DefaultLabels returns Labels with English (general tier) defaults, byte-for-byte
// matching translations/en/general/outcome_matrix.json.
func DefaultLabels() Labels {
	return Labels{
		Page:   PageLabels{Title: "Outcome Matrix"},
		Picker: PickerLabels{Template: "Template"},
		Scope: ScopeLabels{
			Mine: "My clients",
			All:  "All clients",
		},
		Grid: GridLabels{
			CellGridLabels: pyezatypes.CellGridLabels{
				ClientColumn:   "Client",
				SaveButton:     "Save scores",
				SavingButton:   "Saving…",
				SavedBanner:    "Scores saved.",
				ErrorBanner:    "Save failed — please try again.",
				ReadOnlyMarker: "(read only)",
				EmptyGrid:      "No rows to display.",
			},
			CriterionColumn: "Criterion",
			ScoreColumn:     "Score",
			ReadOnlyTooltip: "Recorded by another staff member — read only",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
		},
	}
}
