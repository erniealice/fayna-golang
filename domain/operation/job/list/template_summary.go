package list

// template_summary.go — the education-tier ("Classes") template-grain
// delivery summary that replaces the per-job table on the job List view.
//
// docs/plan/20260710-staff-class-list/tasks.md S6 (LOCKED, O5): the view's
// data assembly is ONE server-side aggregate call — espyna's
// service/operation/job_template_summary ListJobTemplateSummaries — which does
// the GROUP-BY, resolver-scoping, and every column resolution
// (group/deliverer/schedule/product) in a single query. This replaces the
// former ~76-fetch Go aggregation (page-loop jobs + seats + group/plan/staff
// lookups). One row per job_template that has >=1 (resolver-scoped) job for the
// URL {status} segment. Columns: template name, delivery group name, deliverer
// (staff of record), item count (DISTINCT scoped jobs), schedule name. Row link
// -> outcome_matrix.matrix ("/outcome-matrix/{id}", id=job_template_id).
//
// Row scoping is entirely resolver-level (espyna principalscope inside the
// adapter — STAFF principals see only their reachable jobs, the seat tier
// widens that to the class grain; non-staff see all). This file passes NO
// staff_id request filter anywhere.
//
// NAMING: every identifier here is generic (JobTemplate / SubscriptionGroup /
// Staff — the real espyna entity names). Education vocabulary
// ("Classes"/"Section"/"Teacher"/"Students"/"Academic Year") enters ONLY via
// lyngua (packages/lyngua/translations/en/education/job.json) — never as a Go
// identifier, filename, or default label here.

import (
	"context"
	"log"
	"sort"
	"strconv"
	"strings"

	job "github.com/erniealice/fayna-golang/domain/operation/job"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"

	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// templateSummaryRow is one job_template's aggregated delivery-summary row.
type templateSummaryRow struct {
	TemplateID    string
	TemplateName  string
	GroupName     string
	DelivererName string
	ItemCount     int
	ScheduleName  string
}

// buildDeliverySummaryTable builds the template-grain TableConfig for the
// education tier from the single ListJobTemplateSummaries call. The row/column
// building, labels, links (outcome-matrix per row), and empty state are
// unchanged from the pre-S6 view-side compose; only the data-assembly layer
// moved server-side.
func buildDeliverySummaryTable(ctx context.Context, deps *ListViewDeps, status string, perms *types.UserPermissions) (*types.TableConfig, error) {
	rows := buildTemplateSummaryRows(ctx, deps, status)

	l := deps.Labels
	columns := templateSummaryColumns(l)
	tableRows := make([]types.TableRow, 0, len(rows))
	for _, r := range rows {
		matrixURL := route.ResolveURL(deps.MatrixDetailURL, "id", r.TemplateID)
		tableRows = append(tableRows, types.TableRow{
			ID:   r.TemplateID,
			Href: matrixURL,
			Cells: []types.TableCell{
				{Type: "text", Value: r.TemplateName},
				{Type: "text", Value: r.GroupName},
				{Type: "text", Value: r.DelivererName},
				{Type: "number", Value: strconv.Itoa(r.ItemCount)},
				{Type: "text", Value: r.ScheduleName},
			},
			DataAttrs: map[string]string{
				"name":      r.TemplateName,
				"group":     r.GroupName,
				"deliverer": r.DelivererName,
				"schedule":  r.ScheduleName,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: matrixURL},
			},
		})
	}
	types.ApplyColumnStyles(columns, tableRows)

	tableConfig := &types.TableConfig{
		ID:                   "job-template-summary-table",
		Columns:              columns,
		Rows:                 tableRows,
		ShowSearch:           true,
		ShowActions:          true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "group",
		DefaultSortDirection: "asc",
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}
	types.ApplyTableSettings(tableConfig)
	return tableConfig, nil
}

// templateSummaryColumns declares the template-grain columns. Go defaults
// (job.Labels.Columns.{Group,Deliverer,Items,Schedule}) stay generic
// ("Group"/"Deliverer"/"Items"/"Schedule"); education overrides them to
// "Section"/"Teacher"/"Students"/"Academic Year" via lyngua.
func templateSummaryColumns(l job.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "group", Label: l.Columns.Group},
		{Key: "deliverer", Label: l.Columns.Deliverer},
		{Key: "items", Label: l.Columns.Items},
		{Key: "schedule", Label: l.Columns.Schedule},
	}
}

// buildTemplateSummaryRows issues the ONE server-side aggregate call (espyna
// service/operation/job_template_summary) for the {status} segment and maps
// each returned JobTemplateSummary to a view row. All aggregation, resolver-
// scoping, status filtering, and group/deliverer/schedule/product resolution
// happen server-side in a single GROUP-BY query. Rows already arrive ordered by
// group then template name; we re-sort defensively to preserve the LOCKED order
// even when a non-postgres provider returns them unordered.
func buildTemplateSummaryRows(ctx context.Context, deps *ListViewDeps, status string) []templateSummaryRow {
	if deps.ListJobTemplateSummaries == nil {
		return nil
	}
	resp, err := deps.ListJobTemplateSummaries(ctx, &summarypb.ListJobTemplateSummariesRequest{
		Status: jobStatusFilterValue(status),
	})
	if err != nil {
		log.Printf("Failed to list job template summaries: %v", err)
		return nil
	}

	summaries := resp.GetSummaries()
	rows := make([]templateSummaryRow, 0, len(summaries))
	for _, s := range summaries {
		rows = append(rows, templateSummaryRow{
			TemplateID:    s.GetJobTemplateId(),
			TemplateName:  s.GetJobTemplateName(),
			GroupName:     s.GetSubscriptionGroupName(),
			DelivererName: joinDelivererNames(s.GetDeliverers()),
			ItemCount:     int(s.GetJobCount()),
			ScheduleName:  s.GetPriceScheduleName(),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].GroupName != rows[j].GroupName {
			return rows[i].GroupName < rows[j].GroupName
		}
		return rows[i].TemplateName < rows[j].TemplateName
	})
	return rows
}

// joinDelivererNames renders a template's deliverer column. A template can have
// MORE THAN ONE deliverer (a merged deliverable delivered by several staff — e.g.
// a Section's two rotation-strand Teachers); their names render comma-joined
// ("A. Purisima, D. Cabornay"). The names are sorted for a STABLE display order
// even when a non-postgres provider returns the deliverers unordered (the postgres
// adapter already emits them in staff_name order; sorting here is defensive, the
// same discipline as the row re-sort above). Blank names are dropped.
func joinDelivererNames(deliverers []*summarypb.Deliverer) string {
	names := make([]string, 0, len(deliverers))
	for _, d := range deliverers {
		if n := d.GetStaffName(); n != "" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
