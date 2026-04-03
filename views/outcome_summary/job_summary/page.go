package job_summary

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
)

// Deps holds view dependencies for the job outcome summary view.
type Deps struct {
	Routes       fayna.OutcomeSummaryRoutes
	Labels       fayna.OutcomeSummaryLabels
	CommonLabels pyeza.CommonLabels

	// Job outcome summary read
	GetJobOutcomeSummaryByJob func(ctx context.Context, req *jobsumpb.GetJobOutcomeSummaryByJobRequest) (*jobsumpb.GetJobOutcomeSummaryByJobResponse, error)
}

// PageData holds the data for the job outcome summary page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Summary         map[string]any
	Labels          fayna.OutcomeSummaryLabels
}

// summaryToMap converts a JobOutcomeSummary protobuf to a map[string]any for template use.
func summaryToMap(s *jobsumpb.JobOutcomeSummary) map[string]any {
	return map[string]any{
		"id":                    s.GetId(),
		"job_id":                s.GetJobId(),
		"summary_type":          s.GetSummaryType().String(),
		"overall_determination": overallDeterminationString(s.GetOverallDetermination()),
		"determination_variant": overallDeterminationVariant(s.GetOverallDetermination()),
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
		"valid_until_date":      s.GetValidUntilDate(),
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

// NewView creates the job outcome summary view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.GetJobOutcomeSummaryByJob(ctx, &jobsumpb.GetJobOutcomeSummaryByJobRequest{
			JobId: id,
		})
		if err != nil {
			log.Printf("Failed to get job outcome summary for job %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load outcome summary: %w", err))
		}
		if resp.GetJobOutcomeSummary() == nil {
			log.Printf("Job outcome summary for job %s not found", id)
			return view.Error(fmt.Errorf("outcome summary not found"))
		}
		summary := summaryToMap(resp.GetJobOutcomeSummary())

		l := deps.Labels

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.JobHeading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    l.Page.JobHeading,
				HeaderSubtitle: l.Page.JobCaption,
				HeaderIcon:     "icon-bar-chart",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-outcome-summary-content",
			Summary:         summary,
			Labels:          l,
		}

		return view.OK("job-outcome-summary", pageData)
	})
}
