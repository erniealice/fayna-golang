package activity_material

// activity_material_labels.go — ActivityMaterial label structs + DefaultActivityMaterialLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// ActivityMaterialLabels holds all translatable strings for the activity material module.
// ActivityMaterial is the charge detail for ENTRY_TYPE_MATERIAL job activities.
// TODO(P7 lyngua sweep): add lyngua JSON files for these labels.
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
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ButtonLabels struct {
	Add  string `json:"add"`
	Edit string `json:"edit"`
}

type ColumnLabels struct {
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unit_of_measure"`
	LotNumber     string `json:"lot_number"`
	Location      string `json:"location"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
	SectionMaterial string `json:"section_material"`
	Product         string `json:"product"`
	UnitOfMeasure   string `json:"unit_of_measure"`
	LotNumber       string `json:"lot_number"`
	Location        string `json:"location"`
	SubmitCreate    string `json:"submit_create"`
	SubmitUpdate    string `json:"submit_update"`
}

type DetailLabels struct {
	PageTitle     string `json:"page_title"`
	TitlePrefix   string `json:"title_prefix"`
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unit_of_measure"`
	LotNumber     string `json:"lot_number"`
	Location      string `json:"location"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
}

// DefaultActivityMaterialLabels returns ActivityMaterialLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Material Charges",
			Caption: "Material usage entries per activity",
		},
		Buttons: ButtonLabels{
			Add:  "Add Material",
			Edit: "Edit Material",
		},
		Columns: ColumnLabels{
			Product:       "Product",
			UnitOfMeasure: "Unit",
			LotNumber:     "Lot #",
			Location:      "Location",
		},
		Empty: EmptyLabels{
			Title:   "No material entries",
			Message: "No material charge recorded for this activity.",
		},
		Form: FormLabels{
			SectionMaterial: "Material",
			Product:         "Product",
			UnitOfMeasure:   "Unit of Measure",
			LotNumber:       "Lot Number",
			Location:        "Location",
			SubmitCreate:    "Save",
			SubmitUpdate:    "Update",
		},
		Detail: DetailLabels{
			PageTitle:     "Material Charge",
			TitlePrefix:   "Material: ",
			Product:       "Product",
			UnitOfMeasure: "Unit of Measure",
			LotNumber:     "Lot Number",
			Location:      "Location",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Material charge record not found",
			IDRequired:       "Activity ID is required",
		},
	}
}
