package list

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

// Deps holds view dependencies.
type Deps struct {
	Routes                 fayna.OutcomeSummaryRoutes
	ListJobOutcomeSummarys func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)
	Labels                 fayna.OutcomeSummaryLabels
	CommonLabels           pyeza.CommonLabels
	TableLabels            types.TableLabels
}

// PageData holds the data for the outcome summary list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the outcome summary list view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		resp, err := deps.ListJobOutcomeSummarys(ctx, &jobsumpb.ListJobOutcomeSummarysRequest{})
		if err != nil {
			log.Printf("Failed to list outcome summaries: %v", err)
			return view.Error(fmt.Errorf("failed to load outcome summaries: %w", err))
		}

		l := deps.Labels
		columns := summaryColumns(l)
		rows := buildTableRows(resp.GetData(), l, deps.Routes)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "outcome-summary-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "job",
			DefaultSortDirection: "desc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
		}

		// List page highlights "report-cards" in sidebar, not "jobs"
		activeSubNav := "report-cards"

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.JobHeading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   activeSubNav,
				HeaderTitle:    l.Page.JobHeading,
				HeaderSubtitle: l.Page.JobCaption,
				HeaderIcon:     "icon-award",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "outcome-summary-list-content",
			Table:           tableConfig,
		}

		return view.OK("outcome-summary-list", pageData)
	})
}

func summaryColumns(l fayna.OutcomeSummaryLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "job", Label: l.Columns.Job, Sortable: true, MinWidth: "140px"},
		{Key: "determination", Label: l.Detail.OverallDetermination, Sortable: true, MinWidth: "120px"},
		{Key: "score", Label: l.Detail.Score, Sortable: true, MinWidth: "80px"},
		{Key: "scoring_method", Label: l.Detail.ScoringMethod, Sortable: true, MinWidth: "120px"},
		{Key: "total", Label: l.Detail.TotalCriteria, Sortable: true, MinWidth: "60px"},
		{Key: "pass", Label: l.Detail.PassCount, Sortable: true, MinWidth: "60px"},
		{Key: "fail", Label: l.Detail.FailCount, Sortable: true, MinWidth: "60px"},
		{Key: "issued_by", Label: l.Detail.IssuedBy, Sortable: true, MinWidth: "100px"},
	}
}

func buildTableRows(summaries []*jobsumpb.JobOutcomeSummary, l fayna.OutcomeSummaryLabels, routes fayna.OutcomeSummaryRoutes) []types.TableRow {
	var rows []types.TableRow
	for _, s := range summaries {
		determination := overallDeterminationString(s.GetOverallDetermination())
		variant := overallDeterminationVariant(s.GetOverallDetermination())

		row := types.TableRow{
			ID: s.GetId(),
			Cells: []types.TableCell{
				{Value: s.GetJobId()},
				{Type: "badge", Value: determination, Variant: variant},
				{Value: fmt.Sprintf("%.2f", s.GetSummaryScore())},
				{Value: s.GetScoringMethod().String()},
				{Value: fmt.Sprintf("%d", s.GetTotalCriteriaCount())},
				{Value: fmt.Sprintf("%d", s.GetPassCount())},
				{Value: fmt.Sprintf("%d", s.GetFailCount())},
				{Value: s.GetIssuedBy()},
			},
			Actions: []types.TableAction{
				{Type: "view", Label: "View Summary", Href: fmt.Sprintf("/app/outcomes/summary/job/%s", s.GetJobId())},
			},
		}
		rows = append(rows, row)
	}
	return rows
}

func overallDeterminationString(d enums.OverallDetermination) string {
	switch d {
	case enums.OverallDetermination_OVERALL_DETERMINATION_ACCEPTED:
		return "Accepted"
	case enums.OverallDetermination_OVERALL_DETERMINATION_CONDITIONALLY_ACCEPTED:
		return "Conditional"
	case enums.OverallDetermination_OVERALL_DETERMINATION_REJECTED:
		return "Rejected"
	case enums.OverallDetermination_OVERALL_DETERMINATION_IN_PROGRESS:
		return "In Progress"
	case enums.OverallDetermination_OVERALL_DETERMINATION_DEFERRED:
		return "Deferred"
	default:
		return "Unspecified"
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
