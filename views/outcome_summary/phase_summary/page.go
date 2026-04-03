package phase_summary

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
)

// Deps holds view dependencies for the phase outcome summary view.
type Deps struct {
	Routes       fayna.OutcomeSummaryRoutes
	Labels       fayna.OutcomeSummaryLabels
	CommonLabels pyeza.CommonLabels

	// Phase outcome summary read
	GetPhaseOutcomeSummaryByJobPhase func(ctx context.Context, req *phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error)
}

// PageData holds the data for the phase outcome summary page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Summary         map[string]any
	Labels          fayna.OutcomeSummaryLabels
}

// summaryToMap converts a PhaseOutcomeSummary protobuf to a map[string]any for template use.
func summaryToMap(s *phasesumpb.PhaseOutcomeSummary) map[string]any {
	return map[string]any{
		"id":                    s.GetId(),
		"job_phase_id":          s.GetJobPhaseId(),
		"job_id":                s.GetJobId(),
		"summary_type":          s.GetSummaryType().String(),
		"phase_determination":   overallDeterminationString(s.GetPhaseDetermination()),
		"determination_variant": overallDeterminationVariant(s.GetPhaseDetermination()),
		"scoring_method":        s.GetScoringMethod().String(),
		"summary_score":         fmt.Sprintf("%.2f", s.GetSummaryScore()),
		"total_criteria_count":  s.GetTotalCriteriaCount(),
		"pass_count":            s.GetPassCount(),
		"fail_count":            s.GetFailCount(),
		"conditional_count":     s.GetConditionalCount(),
		"deferred_count":        s.GetDeferredCount(),
		"na_count":              s.GetNaCount(),
		"narrative":             s.GetNarrative(),
		"issued_by":             s.GetIssuedBy(),
		"active":                s.GetActive(),
		"date_created_string":   s.GetDateCreatedString(),
		"date_modified_string":  s.GetDateModifiedString(),
	}
}

func overallDeterminationString(d enums.OverallDetermination) string {
	switch d {
	case enums.OverallDetermination_OVERALL_DETERMINATION_ACCEPTED:
		return "accepted"
	case enums.OverallDetermination_OVERALL_DETERMINATION_CONDITIONALLY_ACCEPTED:
		return "conditional"
	case enums.OverallDetermination_OVERALL_DETERMINATION_REJECTED:
		return "rejected"
	case enums.OverallDetermination_OVERALL_DETERMINATION_IN_PROGRESS:
		return "in_progress"
	case enums.OverallDetermination_OVERALL_DETERMINATION_DEFERRED:
		return "deferred"
	default:
		return "unspecified"
	}
}

func overallDeterminationVariant(d enums.OverallDetermination) string {
	switch d {
	case enums.OverallDetermination_OVERALL_DETERMINATION_ACCEPTED:
		return "success"
	case enums.OverallDetermination_OVERALL_DETERMINATION_CONDITIONALLY_ACCEPTED:
		return "warning"
	case enums.OverallDetermination_OVERALL_DETERMINATION_REJECTED:
		return "danger"
	case enums.OverallDetermination_OVERALL_DETERMINATION_IN_PROGRESS:
		return "info"
	case enums.OverallDetermination_OVERALL_DETERMINATION_DEFERRED:
		return "default"
	default:
		return "default"
	}
}

// NewView creates the phase outcome summary view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		phaseID := viewCtx.Request.PathValue("phase_id")

		resp, err := deps.GetPhaseOutcomeSummaryByJobPhase(ctx, &phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest{
			JobPhaseId: phaseID,
		})
		if err != nil {
			log.Printf("Failed to get phase outcome summary for phase %s (job %s): %v", phaseID, id, err)
			return view.Error(fmt.Errorf("failed to load phase outcome summary: %w", err))
		}
		if resp.GetPhaseOutcomeSummary() == nil {
			log.Printf("Phase outcome summary for phase %s not found", phaseID)
			return view.Error(fmt.Errorf("phase outcome summary not found"))
		}
		summary := summaryToMap(resp.GetPhaseOutcomeSummary())

		l := deps.Labels

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.PhaseHeading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    l.Page.PhaseHeading,
				HeaderSubtitle: l.Page.PhaseCaption,
				HeaderIcon:     "icon-bar-chart",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "phase-outcome-summary-content",
			Summary:         summary,
			Labels:          l,
		}

		return view.OK("phase-outcome-summary", pageData)
	})
}
