package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	outcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
)

// PageData holds the data for the task outcome detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Outcome         map[string]any
	Labels          fayna.TaskOutcomeLabels
}

// outcomeToMap converts a TaskOutcome protobuf to a map[string]any for template use.
func outcomeToMap(o *outcomepb.TaskOutcome) map[string]any {
	// Task name
	taskName := o.GetJobTaskId()
	if jt := o.GetJobTask(); jt != nil {
		if jt.GetName() != "" {
			taskName = jt.GetName()
		}
	}

	// Criteria name
	criteriaName := o.GetCriteriaVersionId()
	if cv := o.GetCriteriaVersion(); cv != nil {
		if cv.GetName() != "" {
			criteriaName = cv.GetName()
		}
	}

	// Value display
	value := buildValueDisplay(o)

	return map[string]any{
		"id":                    o.GetId(),
		"job_task_id":           o.GetJobTaskId(),
		"task_name":             taskName,
		"criteria_version_id":   o.GetCriteriaVersionId(),
		"criteria_name":         criteriaName,
		"criteria_type":         criteriaTypeString(o.GetCriteriaType()),
		"value":                 value,
		"determination":         determinationString(o.GetDetermination()),
		"determination_variant": determinationVariant(o.GetDetermination()),
		"determination_source":  o.GetDeterminationSource().String(),
		"determination_note":    o.GetDeterminationNote(),
		"recorded_by":           o.GetRecordedBy(),
		"recorded_by_name":      o.GetRecordedByName(),
		"revision_number":       o.GetRevisionNumber(),
		"active":                o.GetActive(),
		"date_created_string":   o.GetDateCreatedString(),
		"date_modified_string":  o.GetDateModifiedString(),
	}
}

func buildValueDisplay(o *outcomepb.TaskOutcome) string {
	switch o.GetCriteriaType() {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return fmt.Sprintf("%.2f", o.GetNumericValue())
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return o.GetTextValue()
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return o.GetCategoricalValue()
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		if o.GetPassFailValue() {
			return "Pass"
		}
		return "Fail"
	default:
		if o.GetTextValue() != "" {
			return o.GetTextValue()
		}
		return fmt.Sprintf("%.2f", o.GetNumericValue())
	}
}

func criteriaTypeString(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE:
		return "Numeric Range"
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "Numeric Score"
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "Pass/Fail"
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return "Categorical"
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return "Text"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "Multi-Check"
	default:
		return "Unspecified"
	}
}

func determinationString(d enums.Determination) string {
	switch d {
	case enums.Determination_DETERMINATION_PASS:
		return "pass"
	case enums.Determination_DETERMINATION_FAIL:
		return "fail"
	case enums.Determination_DETERMINATION_PASS_WITH_CONDITION:
		return "conditional"
	case enums.Determination_DETERMINATION_NOT_EVALUATED:
		return "not_evaluated"
	case enums.Determination_DETERMINATION_NOT_APPLICABLE:
		return "n_a"
	case enums.Determination_DETERMINATION_DEFERRED:
		return "deferred"
	default:
		return "unspecified"
	}
}

func determinationVariant(d enums.Determination) string {
	switch d {
	case enums.Determination_DETERMINATION_PASS:
		return "success"
	case enums.Determination_DETERMINATION_FAIL:
		return "danger"
	case enums.Determination_DETERMINATION_PASS_WITH_CONDITION:
		return "warning"
	case enums.Determination_DETERMINATION_NOT_EVALUATED:
		return "default"
	case enums.Determination_DETERMINATION_NOT_APPLICABLE:
		return "default"
	case enums.Determination_DETERMINATION_DEFERRED:
		return "info"
	default:
		return "default"
	}
}

// NewView creates the task outcome detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadTaskOutcome(ctx, &outcomepb.ReadTaskOutcomeRequest{
			Data: &outcomepb.TaskOutcome{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read task outcome %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load outcome: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Task outcome %s not found", id)
			return view.Error(fmt.Errorf("outcome not found"))
		}
		outcome := outcomeToMap(data[0])

		l := deps.Labels
		headerTitle := deps.Labels.Detail.PageTitle

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    headerTitle,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-clipboard-check",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "task-outcome-detail-content",
			Outcome:         outcome,
			Labels:          l,
		}

		return view.OK("task-outcome-detail", pageData)
	})
}
