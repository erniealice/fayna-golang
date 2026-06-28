package detail

import (
	"context"
	"fmt"
	"log"

	evaluation "github.com/erniealice/fayna-golang/domain/operation/evaluation"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	resppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_response"
)

// DetailViewDeps holds dependencies for the evaluation detail view (Surface 2).
// The Scores tab reads evaluation_response SNAPSHOT rows via ListEvaluationResponses.
type DetailViewDeps struct {
	Routes       evaluation.Routes
	Labels       evaluation.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	ReadEvaluation          func(ctx context.Context, req *evalpb.ReadEvaluationRequest) (*evalpb.ReadEvaluationResponse, error)
	ListEvaluationResponses func(ctx context.Context, req *resppb.ListEvaluationResponsesRequest) (*resppb.ListEvaluationResponsesResponse, error)
}

// PageData holds the data for the evaluation detail page + its tab partials.
type PageData struct {
	types.PageData
	ContentTemplate string
	Evaluation      map[string]any
	Responses       []ResponseRow
	WeightedAverage string
	Labels          evaluation.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem
}

// ResponseRow is a template-facing evaluation_response snapshot row (Scores tab).
type ResponseRow struct {
	OutcomeCriteriaID string
	Label             string
	Weight            float64
	TypeLabel         string
	Answer            string
	Scored            bool
	TestID            string
}

// evaluationToMap converts an Evaluation proto to a template map.
func evaluationToMap(e *evalpb.Evaluation) map[string]any {
	return map[string]any{
		"id":                  e.GetId(),
		"workspace_id":        e.GetWorkspaceId(),
		"client_id":           e.GetClientId(),
		"subscription_id":     e.GetSubscriptionId(),
		"subject_staff_id":    e.GetStaffId(),
		"evaluation_type":     e.GetEvaluationType().String(),
		"relationship_type":   e.GetRelationshipType().String(),
		"period_start":        e.GetPeriodStart(),
		"period_end":          e.GetPeriodEnd(),
		"status":              statusString(e.GetStatus()),
		"status_variant":      statusVariant(e.GetStatus()),
		"overall_score":       overallScore(e),
		"overall_score_id":    "overall-score-" + e.GetId(),
		"narrative":           e.GetNarrative(),
		"submitted_at":        e.GetSubmittedAt(),
		"signed_off_at":       e.GetSignedOffAt(),
		"date_created_string": e.GetDateCreatedString(),
	}
}

// NewView creates the evaluation detail view (Surface 2).
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "read") {
			return view.Forbidden("evaluation:read")
		}

		id := viewCtx.Request.PathValue("id")
		e, ok, res := loadEvaluation(ctx, deps, id)
		if !ok {
			return res
		}
		evMap := evaluationToMap(e)

		l := deps.Labels
		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}

		headerTitle := l.Detail.PageTitle
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    headerTitle,
				HeaderSubtitle: stringOrDash(e.GetStaffId()),
				HeaderIcon:     "icon-clipboard-check",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "evaluation-detail-content",
			Evaluation:      evMap,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        buildTabItems(l, id, deps.Routes),
		}

		return view.OK("evaluation-detail", pageData)
	})
}

// NewTabAction creates the per-tab partial view. Each tab re-runs the
// evaluation:read gate (permission-reflection-pattern). The Scores tab loads
// the evaluation_response snapshot rows + the weighted-average footer.
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "read") {
			return view.Forbidden("evaluation:read")
		}

		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		e, ok, res := loadEvaluation(ctx, deps, id)
		if !ok {
			return res
		}

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Evaluation: evaluationToMap(e),
			Labels:     l,
			ActiveTab:  tab,
			TabItems:   buildTabItems(l, id, deps.Routes),
		}

		// Scores tab — load the snapshot rows + compute the weighted footer.
		if tab == "scores" && deps.ListEvaluationResponses != nil {
			respList, err := deps.ListEvaluationResponses(ctx, &resppb.ListEvaluationResponsesRequest{})
			if err != nil {
				log.Printf("Failed to list evaluation responses for %s: %v", id, err)
				return view.Error(fmt.Errorf("failed to load scores: %w", err))
			}
			rows, weighted := buildResponseRows(respList.GetData(), id, l)
			pageData.Responses = rows
			pageData.WeightedAverage = weighted
		}

		return view.OK("evaluation-tab-"+tab, pageData)
	})
}

func loadEvaluation(ctx context.Context, deps *DetailViewDeps, id string) (*evalpb.Evaluation, bool, view.ViewResult) {
	resp, err := deps.ReadEvaluation(ctx, &evalpb.ReadEvaluationRequest{
		Data: &evalpb.Evaluation{Id: id},
	})
	if err != nil {
		log.Printf("Failed to read evaluation %s: %v", id, err)
		return nil, false, view.Error(fmt.Errorf("failed to load review: %w", err))
	}
	data := resp.GetData()
	if len(data) == 0 {
		log.Printf("Evaluation %s not found", id)
		return nil, false, view.Error(fmt.Errorf("review not found"))
	}
	return data[0], true, view.ViewResult{}
}

// buildResponseRows renders the snapshot rows and the weighted-average footer.
// Non-numeric dimensions are shown as "(not scored)" and excluded from the
// weighted average (mirrors ComputeEvaluationScore / BLOCKER-2, §D.1).
func buildResponseRows(items []*resppb.EvaluationResponse, evalID string, l evaluation.Labels) ([]ResponseRow, string) {
	rows := make([]ResponseRow, 0, len(items))
	var weightedSum, weightTotal float64
	for _, r := range items {
		if r.GetEvaluationId() != evalID {
			continue
		}
		scored := isNumeric(r.GetCriteriaType())
		row := ResponseRow{
			OutcomeCriteriaID: r.GetOutcomeCriteriaId(),
			Label:             r.GetCriteriaLabel(),
			Weight:            r.GetCriteriaWeight(),
			TypeLabel:         criteriaTypeLabel(r.GetCriteriaType()),
			Answer:            answerString(r, l),
			Scored:            scored,
			TestID:            "eval-response-row-" + r.GetOutcomeCriteriaId(),
		}
		rows = append(rows, row)
		if scored {
			w := r.GetCriteriaWeight()
			if w == 0 {
				w = 1
			}
			weightedSum += r.GetNumericValue() * w
			weightTotal += w
		}
	}
	if weightTotal == 0 {
		return rows, l.Scores.OverallNil
	}
	return rows, fmt.Sprintf("%.2f", weightedSum/weightTotal)
}

func isNumeric(t enums.CriteriaType) bool {
	return t == enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE ||
		t == enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE
}

func answerString(r *resppb.EvaluationResponse, l evaluation.Labels) string {
	switch r.GetCriteriaType() {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE,
		enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE:
		return fmt.Sprintf("%.2f", r.GetNumericValue())
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		if r.GetPassFailValue() {
			return l.Dimension.Pass
		}
		return l.Dimension.Fail
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return r.GetCategoricalValue()
	default:
		return r.GetTextValue()
	}
}

func criteriaTypeLabel(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "Numeric Score"
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE:
		return "Numeric Range"
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "Pass/Fail"
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return "Categorical"
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return "Text"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "Multi-Check"
	default:
		return ""
	}
}

// buildTabItems returns the 4-tab strip: Info / Scores / Sign Off / Audit (§D).
func buildTabItems(l evaluation.Labels, id string, routes evaluation.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "scores", Label: l.Tabs.Scores, Href: base + "?tab=scores", HxGet: action + "scores", Icon: "icon-bar-chart"},
		{Key: "signoff", Label: l.Tabs.SignOff, Href: base + "?tab=signoff", HxGet: action + "signoff", Icon: "icon-check-circle"},
		{Key: "audit", Label: l.Tabs.Audit, Href: base + "?tab=audit", HxGet: action + "audit", Icon: "icon-clock"},
	}
}

func statusString(s evalpb.EvaluationStatus) string {
	switch s {
	case evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT:
		return "draft"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED:
		return "submitted"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF:
		return "signed_off"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_ARCHIVED:
		return "archived"
	default:
		return "draft"
	}
}

func statusVariant(s evalpb.EvaluationStatus) string {
	switch s {
	case evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT:
		return "default"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED:
		return "info"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF:
		return "success"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_ARCHIVED:
		return "secondary"
	default:
		return "default"
	}
}

func overallScore(e *evalpb.Evaluation) string {
	if e.GetStatus() == evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT || e.GetOverallScore() == 0 {
		return "—"
	}
	return fmt.Sprintf("%.2f", e.GetOverallScore())
}

func stringOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
