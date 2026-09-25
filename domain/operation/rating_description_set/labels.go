package rating_description_set

// Labels holds all translatable strings for the rating description set module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Buttons ButtonLabels  `json:"buttons"`
	Columns ColumnLabels  `json:"columns"`
	Status  StatusLabels  `json:"status"`
	Tabs    TabLabels     `json:"tabs"`
	Actions ActionLabels  `json:"actions"`
	Confirm ConfirmLabels `json:"confirm"`
	Errors  ErrorLabels   `json:"errors"`
	Form    FormLabels    `json:"form"`
	Matrix  MatrixLabels  `json:"matrix"`
	Empty   EmptyLabels   `json:"empty"`
}

type PageLabels struct {
	Heading           string `json:"heading"`
	Caption           string `json:"caption"`
	HeadingActive     string `json:"heading_active"`
	HeadingDeprecated string `json:"heading_deprecated"`
	HeadingAll        string `json:"heading_all"`
}

type ButtonLabels struct {
	Add string `json:"add"`
}

type ColumnLabels struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Scale   string `json:"scale"`
	Entries string `json:"entries"`
	Links   string `json:"links"`
}

type StatusLabels struct {
	Draft      string `json:"draft"`
	Published  string `json:"published"`
	Deprecated string `json:"deprecated"`
}

// TabLabels — keys per interfaces.md §8 (descriptors/assignments/versions/info).
// Only Descriptors is rendered this wave; the others are declared for label
// parity and future tabs (TODO).
type TabLabels struct {
	Descriptors string `json:"descriptors"`
	Assignments string `json:"assignments"`
	Versions    string `json:"versions"`
	Info        string `json:"info"`
}

type ActionLabels struct {
	Add       string `json:"add"`
	Publish   string `json:"publish"`
	Deprecate string `json:"deprecate"`
}

type ConfirmLabels struct {
	Publish          string `json:"publish"`
	PublishMessage   string `json:"publish_message"`
	Deprecate        string `json:"deprecate"`
	DeprecateMessage string `json:"deprecate_message"`
}

type ErrorLabels struct {
	PublishedImmutable string `json:"published_immutable"`
	HasLinks           string `json:"has_links"`
	NeedsEntries       string `json:"needs_entries"`
	InvalidTransition  string `json:"invalid_transition"`
	PermissionDenied   string `json:"permission_denied"`
	InvalidFormData    string `json:"invalid_form_data"`
	NotFound           string `json:"not_found"`
	IDRequired         string `json:"id_required"`
	// LoadFailed surfaces a picker-data load failure (codex-review-impl2.out.md
	// finding #6 — "surface load failures", not a silently empty picker).
	LoadFailed string `json:"load_failed"`
}

type FormLabels struct {
	Name      string `json:"name"`
	Code      string `json:"code"`
	Scale     string `json:"scale"`
	SourceRef string `json:"source_ref"`
}

type MatrixLabels struct {
	LevelZeroHint string `json:"level_zero_hint"`
	AddEntry      string `json:"add_entry"`
	EmptyCell     string `json:"empty_cell"`
	Filled        string `json:"filled"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading:           "Rating Description Sets",
			Caption:           "Manage rating description definitions",
			HeadingActive:     "Active Rating Description Sets",
			HeadingDeprecated: "Deprecated Rating Description Sets",
			HeadingAll:        "All Rating Description Sets",
		},
		Buttons: ButtonLabels{
			Add: "Add Set",
		},
		Columns: ColumnLabels{
			Name:    "Name",
			Version: "Version",
			Status:  "Status",
			Scale:   "Scale",
			Entries: "Entries",
			Links:   "Assigned to",
		},
		Status: StatusLabels{
			Draft:      "Draft",
			Published:  "Published",
			Deprecated: "Deprecated",
		},
		Tabs: TabLabels{
			Descriptors: "Descriptors",
			Assignments: "Assignments",
			Versions:    "Versions",
			Info:        "Information",
		},
		Actions: ActionLabels{
			Add:       "Add Set",
			Publish:   "Publish",
			Deprecate: "Deprecate",
		},
		Confirm: ConfirmLabels{
			Publish:          "Publish Set",
			PublishMessage:   "Published sets can't be changed. Continue?",
			Deprecate:        "Deprecate Set",
			DeprecateMessage: "Existing assignments keep working; it can't be newly assigned. Continue?",
		},
		Errors: ErrorLabels{
			PublishedImmutable: "Published sets cannot be edited",
			HasLinks:           "Cannot delete a set that has active links",
			NeedsEntries:       "Set must have at least one entry before publishing",
			InvalidTransition:  "Invalid status transition",
			PermissionDenied:   "You do not have permission to perform this action",
			InvalidFormData:    "Invalid form data. Please check your inputs and try again.",
			NotFound:           "Rating description set not found",
			IDRequired:         "Rating description set ID is required",
			LoadFailed:         "Failed to load form options. Please try again.",
		},
		Form: FormLabels{
			Name:      "Name",
			Code:      "Code",
			Scale:     "Scale",
			SourceRef: "Source Reference",
		},
		Matrix: MatrixLabels{
			LevelZeroHint: "Changing a rating to 0 clears the note; staff can then write one.",
			AddEntry:      "Add entry",
			EmptyCell:     "+ Add",
			Filled:        "Filled",
		},
		Empty: EmptyLabels{
			Title:   "No descriptor sets yet",
			Message: "Create your first descriptor set to get started.",
		},
	}
}
