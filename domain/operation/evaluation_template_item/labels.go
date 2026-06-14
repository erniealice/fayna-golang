package evaluation_template_item

// labels.go — EvaluationTemplateItem (rubric item) label structs.
// View-local (Option-B); lyngua overlays JSON via the compose Unit binding
// (root key "evaluationTemplateItem").

// Labels holds all translatable strings for the rubric-item drawer.
type Labels struct {
	Form    FormLabels    `json:"form"`
	Confirm ConfirmLabels `json:"confirm"`
	Errors  ErrorLabels   `json:"errors"`
}

type FormLabels struct {
	Criterion             string `json:"criterion"`
	CriterionPlaceholder  string `json:"criterionPlaceholder"`
	CriteriaType          string `json:"criteriaType"`
	SequenceOrder         string `json:"sequenceOrder"`
	WeightOverride        string `json:"weightOverride"`
	QuestionLabel         string `json:"questionLabel"`
	QuestionLabelPlaceholder string `json:"questionLabelPlaceholder"`
	QuestionPrompt        string `json:"questionPrompt"`
	RequiredOverride      string `json:"requiredOverride"`
	CriteriaTypeInfo      string `json:"criteriaTypeInfo"`
	WeightInfo            string `json:"weightInfo"`
}

type ConfirmLabels struct {
	Remove        string `json:"remove"`
	RemoveMessage string `json:"removeMessage"`
}

type ErrorLabels struct {
	PermissionDenied   string `json:"permissionDenied"`
	InvalidFormData    string `json:"invalidFormData"`
	NotFound           string `json:"notFound"`
	IDRequired         string `json:"idRequired"`
	TemplateIDRequired string `json:"templateIDRequired"`
	CriterionRequired  string `json:"criterionRequired"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Form: FormLabels{
			Criterion:                "Criterion",
			CriterionPlaceholder:     "Search evaluation criteria...",
			CriteriaType:             "Response Type",
			SequenceOrder:            "Order",
			WeightOverride:           "Weight",
			QuestionLabel:            "Question Label",
			QuestionLabelPlaceholder: "Override the criterion label (optional)",
			QuestionPrompt:           "Helper Text",
			RequiredOverride:         "Required",
			CriteriaTypeInfo:         "The response type is inherited from the selected criterion and cannot be changed here.",
			WeightInfo:               "A weighted question must use a numeric response type (BLOCKER-2 enforced on activation).",
		},
		Confirm: ConfirmLabels{
			Remove:        "Remove Question",
			RemoveMessage: "Remove this question from the rubric?",
		},
		Errors: ErrorLabels{
			PermissionDenied:   "You do not have permission to perform this action",
			InvalidFormData:    "Invalid form data. Please check your inputs and try again.",
			NotFound:           "Rubric question not found",
			IDRequired:         "Item ID is required",
			TemplateIDRequired: "Template ID is required",
			CriterionRequired:  "A criterion is required",
		},
	}
}
