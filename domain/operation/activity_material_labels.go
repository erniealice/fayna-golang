package operation

// activity_material_labels.go — ActivityMaterial label structs + DefaultActivityMaterialLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// ActivityMaterialLabels holds all translatable strings for the activity material module.
// ActivityMaterial is the charge detail for ENTRY_TYPE_MATERIAL job activities.
// TODO(P7 lyngua sweep): add lyngua JSON files for these labels.
type ActivityMaterialLabels struct {
	Page    ActivityMaterialPageLabels   `json:"page"`
	Buttons ActivityMaterialButtonLabels `json:"buttons"`
	Columns ActivityMaterialColumnLabels `json:"columns"`
	Empty   ActivityMaterialEmptyLabels  `json:"empty"`
	Form    ActivityMaterialFormLabels   `json:"form"`
	Detail  ActivityMaterialDetailLabels `json:"detail"`
	Errors  ActivityMaterialErrorLabels  `json:"errors"`
}

type ActivityMaterialPageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ActivityMaterialButtonLabels struct {
	Add  string `json:"add"`
	Edit string `json:"edit"`
}

type ActivityMaterialColumnLabels struct {
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unitOfMeasure"`
	LotNumber     string `json:"lotNumber"`
	Location      string `json:"location"`
}

type ActivityMaterialEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ActivityMaterialFormLabels struct {
	SectionMaterial string `json:"sectionMaterial"`
	Product         string `json:"product"`
	UnitOfMeasure   string `json:"unitOfMeasure"`
	LotNumber       string `json:"lotNumber"`
	Location        string `json:"location"`
	SubmitCreate    string `json:"submitCreate"`
	SubmitUpdate    string `json:"submitUpdate"`
}

type ActivityMaterialDetailLabels struct {
	PageTitle     string `json:"pageTitle"`
	TitlePrefix   string `json:"titlePrefix"`
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unitOfMeasure"`
	LotNumber     string `json:"lotNumber"`
	Location      string `json:"location"`
}

type ActivityMaterialErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultActivityMaterialLabels returns ActivityMaterialLabels with sensible English defaults.
func DefaultActivityMaterialLabels() ActivityMaterialLabels {
	return ActivityMaterialLabels{
		Page: ActivityMaterialPageLabels{
			Heading: "Material Charges",
			Caption: "Material usage entries per activity",
		},
		Buttons: ActivityMaterialButtonLabels{
			Add:  "Add Material",
			Edit: "Edit Material",
		},
		Columns: ActivityMaterialColumnLabels{
			Product:       "Product",
			UnitOfMeasure: "Unit",
			LotNumber:     "Lot #",
			Location:      "Location",
		},
		Empty: ActivityMaterialEmptyLabels{
			Title:   "No material entries",
			Message: "No material charge recorded for this activity.",
		},
		Form: ActivityMaterialFormLabels{
			SectionMaterial: "Material",
			Product:         "Product",
			UnitOfMeasure:   "Unit of Measure",
			LotNumber:       "Lot Number",
			Location:        "Location",
			SubmitCreate:    "Save",
			SubmitUpdate:    "Update",
		},
		Detail: ActivityMaterialDetailLabels{
			PageTitle:     "Material Charge",
			TitlePrefix:   "Material: ",
			Product:       "Product",
			UnitOfMeasure: "Unit of Measure",
			LotNumber:     "Lot Number",
			Location:      "Location",
		},
		Errors: ActivityMaterialErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Material charge record not found",
			IDRequired:       "Activity ID is required",
		},
	}
}
