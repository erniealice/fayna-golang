package evaluation_template

// labels.go — EvaluationTemplate label structs + DefaultLabels constructor.
//
// View-local label structs (Option-B): the translatable strings for the
// evaluation_template module live here, not in the lyngua Go package. The
// lyngua provider overlays JSON onto these via the compose Unit LabelJSON
// binding (root key "evaluationTemplate"). Staff-only authoring surface —
// clients have NO evaluation_template:* permission (acceptance #8).

// Labels holds all translatable strings for the evaluation template module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Buttons ButtonLabels  `json:"buttons"`
	Columns ColumnLabels  `json:"columns"`
	Empty   EmptyLabels   `json:"empty"`
	Form    FormLabels    `json:"form"`
	Actions ActionLabels  `json:"actions"`
	Detail  DetailLabels  `json:"detail"`
	Tabs    TabLabels     `json:"tabs"`
	Items   ItemLabels    `json:"items"`
	Confirm ConfirmLabels `json:"confirm"`
	Errors  ErrorLabels   `json:"errors"`
}

type PageLabels struct {
	Heading           string `json:"heading"`
	HeadingDraft      string `json:"headingDraft"`
	HeadingActive     string `json:"headingActive"`
	HeadingDeprecated string `json:"headingDeprecated"`
	Caption           string `json:"caption"`
	CaptionDraft      string `json:"captionDraft"`
	CaptionActive     string `json:"captionActive"`
	CaptionDeprecated string `json:"captionDeprecated"`
}

type ButtonLabels struct {
	AddTemplate string `json:"addTemplate"`
	AddQuestion string `json:"addQuestion"`
}

type ColumnLabels struct {
	Name             string `json:"name"`
	EvaluationType   string `json:"evaluationType"`
	RelationshipType string `json:"relationshipType"`
	Version          string `json:"version"`
	Status           string `json:"status"`
	Visibility       string `json:"visibility"`
	Items            string `json:"items"`
	Created          string `json:"created"`
}

type EmptyLabels struct {
	Title             string `json:"title"`
	Message           string `json:"message"`
	DraftTitle        string `json:"draftTitle"`
	DraftMessage      string `json:"draftMessage"`
	ActiveTitle       string `json:"activeTitle"`
	ActiveMessage     string `json:"activeMessage"`
	DeprecatedTitle   string `json:"deprecatedTitle"`
	DeprecatedMessage string `json:"deprecatedMessage"`
	ItemsTitle        string `json:"itemsTitle"`
	ItemsMessage      string `json:"itemsMessage"`
}

type FormLabels struct {
	Name                  string `json:"name"`
	NamePlaceholder       string `json:"namePlaceholder"`
	Description           string `json:"description"`
	DescPlaceholder       string `json:"descriptionPlaceholder"`
	EvaluationType        string `json:"evaluationType"`
	RelationshipType      string `json:"relationshipType"`
	Visibility            string `json:"visibility"`
	EvaluationTypeInfo    string `json:"evaluationTypeInfo"`
	RelationshipTypeInfo  string `json:"relationshipTypeInfo"`
	VisibilityInfo        string `json:"visibilityInfo"`
}

type ActionLabels struct {
	View      string `json:"view"`
	Activate  string `json:"activate"`
	Deprecate string `json:"deprecate"`
	Clone     string `json:"clone"`
}

type DetailLabels struct {
	PageTitle        string `json:"pageTitle"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	EvaluationType   string `json:"evaluationType"`
	RelationshipType string `json:"relationshipType"`
	Version          string `json:"version"`
	Status           string `json:"status"`
	Visibility       string `json:"visibility"`
	CopiedFrom       string `json:"copiedFrom"`
	CreatedDate      string `json:"createdDate"`
	ModifiedDate     string `json:"modifiedDate"`
}

type TabLabels struct {
	Info  string `json:"info"`
	Items string `json:"items"`
}

// ItemLabels — rubric-builder (evaluation_template_item) labels surfaced on the
// template Items tab.
type ItemLabels struct {
	Heading        string `json:"heading"`
	Caption        string `json:"caption"`
	Sequence       string `json:"sequence"`
	Criterion      string `json:"criterion"`
	CriteriaType   string `json:"criteriaType"`
	Weight         string `json:"weight"`
	QuestionLabel  string `json:"questionLabel"`
	QuestionPrompt string `json:"questionPrompt"`
	Required       string `json:"required"`
	Edit           string `json:"edit"`
	Remove         string `json:"remove"`
	NotScored      string `json:"notScored"`
}

type ConfirmLabels struct {
	Deprecate         string `json:"deprecate"`
	DeprecateMessage  string `json:"deprecateMessage"`
	RemoveItem        string `json:"removeItem"`
	RemoveItemMessage string `json:"removeItemMessage"`
}

type ErrorLabels struct {
	PermissionDenied  string `json:"permissionDenied"`
	InvalidFormData   string `json:"invalidFormData"`
	NotFound          string `json:"notFound"`
	IDRequired        string `json:"idRequired"`
	NoPermission      string `json:"noPermission"`
	WeightedNonNumeric string `json:"weightedNonNumeric"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading:           "Evaluation Templates",
			HeadingDraft:      "Draft Templates",
			HeadingActive:     "Active Templates",
			HeadingDeprecated: "Deprecated Templates",
			Caption:           "Manage reusable evaluation rubrics",
			CaptionDraft:      "Templates still being authored",
			CaptionActive:     "Templates available for new reviews",
			CaptionDeprecated: "Retired templates (existing drafts still submittable)",
		},
		Buttons: ButtonLabels{
			AddTemplate: "New Template",
			AddQuestion: "Add Question",
		},
		Columns: ColumnLabels{
			Name:             "Name",
			EvaluationType:   "Type",
			RelationshipType: "Relationship",
			Version:          "Version",
			Status:           "Status",
			Visibility:       "Visibility",
			Items:            "Items",
			Created:          "Created",
		},
		Empty: EmptyLabels{
			Title:             "No templates found",
			Message:           "No evaluation templates to display.",
			DraftTitle:        "No draft templates",
			DraftMessage:      "Create your first template to get started.",
			ActiveTitle:       "No active templates",
			ActiveMessage:     "Activate a draft template to make it pickable.",
			DeprecatedTitle:   "No deprecated templates",
			DeprecatedMessage: "Deprecated templates will appear here.",
			ItemsTitle:        "No rubric questions",
			ItemsMessage:      "Add a question to build the rubric.",
		},
		Form: FormLabels{
			Name:                 "Name",
			NamePlaceholder:      "Enter template name",
			Description:          "Description",
			DescPlaceholder:      "Enter template description...",
			EvaluationType:       "Evaluation Type",
			RelationshipType:     "Relationship Type",
			Visibility:           "Visibility",
			EvaluationTypeInfo:   "The purpose of this evaluation (e.g. performance review, CSAT).",
			RelationshipTypeInfo: "Who evaluates whom (e.g. client rates associate).",
			VisibilityInfo:       "Who can see the submitted scores.",
		},
		Actions: ActionLabels{
			View:      "View Template",
			Activate:  "Activate",
			Deprecate: "Deprecate",
			Clone:     "Clone",
		},
		Detail: DetailLabels{
			PageTitle:        "Template Details",
			Name:             "Name",
			Description:      "Description",
			EvaluationType:   "Evaluation Type",
			RelationshipType: "Relationship Type",
			Version:          "Version",
			Status:           "Status",
			Visibility:       "Visibility",
			CopiedFrom:       "Cloned From",
			CreatedDate:      "Created",
			ModifiedDate:     "Last Modified",
		},
		Tabs: TabLabels{
			Info:  "Information",
			Items: "Rubric",
		},
		Items: ItemLabels{
			Heading:        "Rubric Questions",
			Caption:        "Ordered criteria scored on each review",
			Sequence:       "Order",
			Criterion:      "Criterion",
			CriteriaType:   "Response Type",
			Weight:         "Weight",
			QuestionLabel:  "Question Label",
			QuestionPrompt: "Helper Text",
			Required:       "Required",
			Edit:           "Edit Question",
			Remove:         "Remove Question",
			NotScored:      "(not scored)",
		},
		Confirm: ConfirmLabels{
			Deprecate:         "Deprecate Template",
			DeprecateMessage:  "Are you sure you want to deprecate \"%s\"? It will no longer be pickable.",
			RemoveItem:        "Remove Question",
			RemoveItemMessage: "Remove this question from the rubric?",
		},
		Errors: ErrorLabels{
			PermissionDenied:   "You do not have permission to perform this action",
			InvalidFormData:    "Invalid form data. Please check your inputs and try again.",
			NotFound:           "Evaluation template not found",
			IDRequired:         "Template ID is required",
			NoPermission:       "No permission",
			WeightedNonNumeric: "A weighted question must use a numeric response type. Remove the weight or change the criterion.",
		},
	}
}
