package job_category

// labels.go — JobCategory label structs + defaults. Field json tags match the
// lyngua job_category.json keys (general + education tiers) so a per-tier
// override binds; a missing tag silently falls back to the compiled default.

// Labels holds all translatable strings for the job category module.
type Labels struct {
	Page    PageLabels   `json:"page"`
	Buttons ButtonLabels `json:"buttons"`
	Columns ColumnLabels `json:"columns"`
	Empty   EmptyLabels  `json:"empty"`
	Form    FormLabels   `json:"form"`
	Detail  DetailLabels `json:"detail"`
	Errors  ErrorLabels  `json:"errors"`
}

type PageLabels struct {
	Title         string `json:"title"`
	Subtitle      string `json:"subtitle"`
	ActiveTitle   string `json:"active_title"`
	InactiveTitle string `json:"inactive_title"`
}

type ButtonLabels struct {
	Activate   string `json:"activate"`
	Add        string `json:"add"`
	BulkDelete string `json:"bulk_delete"`
	Deactivate string `json:"deactivate"`
	Delete     string `json:"delete"`
	Edit       string `json:"edit"`
	View       string `json:"view"`
}

type ColumnLabels struct {
	Actions     string `json:"actions"`
	Code        string `json:"code"`
	DateCreated string `json:"date_created"`
	Name        string `json:"name"`
	SortOrder   string `json:"sort_order"`
	Status      string `json:"status"`
}

type EmptyLabels struct {
	Message string `json:"message"`
	Title   string `json:"title"`
}

type FormLabels struct {
	Active               string `json:"active"`
	ActiveInfo           string `json:"active_info"`
	Code                 string `json:"code"`
	CodeInfo             string `json:"code_info"`
	CodePlaceholder      string `json:"code_placeholder"`
	Name                 string `json:"name"`
	NameInfo             string `json:"name_info"`
	NamePlaceholder      string `json:"name_placeholder"`
	SortOrder            string `json:"sort_order"`
	SortOrderInfo        string `json:"sort_order_info"`
	SortOrderPlaceholder string `json:"sort_order_placeholder"`
	SectionIdentity      string `json:"section_identity"`
}

type DetailLabels struct {
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	NoCode       string `json:"no_code"`
	Title        string `json:"title"`
}

type ErrorLabels struct {
	CreateFailed string `json:"create_failed"`
	DeleteFailed string `json:"delete_failed"`
	InUse        string `json:"in_use"`
	LoadFailed   string `json:"load_failed"`
	NotFound     string `json:"not_found"`
	Unauthorized string `json:"unauthorized"`
	UpdateFailed string `json:"update_failed"`
}

// DefaultLabels returns Labels with sensible generic English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Title:         "Categories",
			Subtitle:      "Manage your job category taxonomy",
			ActiveTitle:   "Active Categories",
			InactiveTitle: "Inactive Categories",
		},
		Buttons: ButtonLabels{
			Activate:   "Activate",
			Add:        "Add Category",
			BulkDelete: "Delete Categories",
			Deactivate: "Deactivate",
			Delete:     "Delete Category",
			Edit:       "Edit Category",
			View:       "View",
		},
		Columns: ColumnLabels{
			Actions:     "Actions",
			Code:        "Code",
			DateCreated: "Date Created",
			Name:        "Name",
			SortOrder:   "Order",
			Status:      "Status",
		},
		Empty: EmptyLabels{
			Message: "No categories to display.",
			Title:   "No Categories",
		},
		Form: FormLabels{
			Active:               "Active",
			ActiveInfo:           "Inactive categories are hidden from new assignments.",
			Code:                 "Code",
			CodeInfo:             "Stable machine key for this category used in reporting.",
			CodePlaceholder:      "Enter a short code (e.g. academic)",
			Name:                 "Name",
			NameInfo:             "A short display name for this category.",
			NamePlaceholder:      "Enter category name",
			SortOrder:            "Order",
			SortOrderInfo:        "Controls the tab / list order (lower first).",
			SortOrderPlaceholder: "10",
			SectionIdentity:      "Category details",
		},
		Detail: DetailLabels{
			DateCreated:  "Date Created",
			DateModified: "Date Modified",
			NoCode:       "—",
			Title:        "Category",
		},
		Errors: ErrorLabels{
			CreateFailed: "Failed to create category",
			DeleteFailed: "Failed to delete category",
			InUse:        "This category is in use and cannot be deleted.",
			LoadFailed:   "Failed to load category",
			NotFound:     "Category not found",
			Unauthorized: "You are not authorized to perform this action",
			UpdateFailed: "Failed to update category",
		},
	}
}
