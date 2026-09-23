package client_card

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	espynaports "github.com/erniealice/espyna-golang/ports"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func narrowReportViewEnabled(ctx context.Context, deps *Deps, perms *types.UserPermissions) bool {
	return !outcome_summary.CanLegacyDetail(perms) &&
		outcome_summary.CanExplicitExport(perms, ctx, deps.ResolvePrincipalKind) &&
		deps.Options.List.SubscriptionGroups() &&
		deps.Options.SubscriptionGroupExportEnabled() &&
		deps.GetSubscriptionGroupClientReportCard != nil
}

func renderReportView(ctx context.Context, viewCtx *view.ViewContext, deps *Deps, subscriptionGroupID, clientID string) view.ViewResult {
	if deps.GetSubscriptionGroupClientReportCard == nil {
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	if strings.TrimSpace(subscriptionGroupID) == "" || strings.TrimSpace(clientID) == "" {
		return view.Forbidden("subscription_group_outcome_export:read")
	}

	response, err := deps.GetSubscriptionGroupClientReportCard(ctx, &exportpb.GetSubscriptionGroupClientReportCardRequest{
		SubscriptionGroupId:  subscriptionGroupID,
		ClientId:             clientID,
		ClientAttributeCodes: append([]string(nil), deps.ClientAttributeCodes...),
	})
	if err != nil {
		log.Printf("client report view: scoped projection read failed: %v", err)
		if errors.Is(err, espynaports.ErrClientReportNotFound) {
			return view.ViewResult{Error: fmt.Errorf("client report not found"), StatusCode: http.StatusNotFound}
		}
		return view.ViewResult{Error: fmt.Errorf("client report projection is unavailable"), StatusCode: http.StatusServiceUnavailable}
	}
	if response == nil || !response.GetSuccess() {
		log.Printf("client report view: scoped projection returned an incomplete response")
		return view.ViewResult{Error: fmt.Errorf("client report projection is unavailable"), StatusCode: http.StatusServiceUnavailable}
	}
	projection := response.GetReportCard()
	if !clientProjectionValid(projection, subscriptionGroupID, clientID) {
		return view.ViewResult{Error: fmt.Errorf("client report not found"), StatusCode: http.StatusNotFound}
	}

	table := buildProjectedClientTable(deps, projection, deps.Options.ClientCard.BandByCategory())
	clientName := projectedClientName(projection.GetClient())
	return clientProjectionPage(ctx, viewCtx, deps, projection, clientName, table)
}

func clientProjectionValid(projection *exportpb.ClientReportCardProjection, groupID, clientID string) bool {
	return projection != nil &&
		projection.GetContext() != nil && projection.GetContext().GetSubscriptionGroupId() == groupID &&
		projection.GetClient() != nil && projection.GetClient().GetClientId() == clientID &&
		len(projection.GetClientSubscriptionIds()) > 0
}

type phaseColumn struct {
	code     string
	label    string
	sequence int32
}

func buildProjectedClientTable(deps *Deps, projection *exportpb.ClientReportCardProjection, bandByCategory bool) *types.TableConfig {
	if projection == nil || len(projection.GetJobs()) == 0 {
		return nil
	}

	jobs := append([]*jobpb.Job(nil), projection.GetJobs()...)
	jobTemplateByID := make(map[string]*jobtemplatepb.JobTemplate, len(projection.GetJobTemplates()))
	for _, template := range projection.GetJobTemplates() {
		if template != nil && strings.TrimSpace(template.GetId()) != "" {
			jobTemplateByID[template.GetId()] = template
		}
	}
	categoryByID := make(map[string]*jobcategorypb.JobCategory, len(projection.GetJobCategories()))
	for _, category := range projection.GetJobCategories() {
		if category != nil && strings.TrimSpace(category.GetId()) != "" {
			categoryByID[category.GetId()] = category
		}
	}
	templatePhaseByID := make(map[string]*jobtemplatephasepb.JobTemplatePhase, len(projection.GetJobTemplatePhases()))
	phaseColumnsByCode := make(map[string]phaseColumn)
	for _, phase := range projection.GetJobTemplatePhases() {
		if phase == nil {
			continue
		}
		templatePhaseByID[phase.GetId()] = phase
		code := strings.TrimSpace(phase.GetCode())
		if code == "" {
			code = phase.GetId()
		}
		if code == "" {
			continue
		}
		label := strings.TrimSpace(phase.GetName())
		if label == "" {
			label = code
		}
		current, exists := phaseColumnsByCode[code]
		candidate := phaseColumn{code: code, label: label, sequence: phase.GetPhaseOrder()}
		if !exists || candidate.sequence < current.sequence || current.sequence == 0 {
			phaseColumnsByCode[code] = candidate
		}
	}
	phaseColumns := make([]phaseColumn, 0, len(phaseColumnsByCode))
	for _, phase := range phaseColumnsByCode {
		phaseColumns = append(phaseColumns, phase)
	}
	sort.SliceStable(phaseColumns, func(i, j int) bool {
		if phaseColumns[i].sequence != phaseColumns[j].sequence {
			return phaseColumns[i].sequence < phaseColumns[j].sequence
		}
		if phaseColumns[i].label != phaseColumns[j].label {
			return phaseColumns[i].label < phaseColumns[j].label
		}
		return phaseColumns[i].code < phaseColumns[j].code
	})

	jobByID := make(map[string]*jobpb.Job, len(jobs))
	for _, job := range jobs {
		if job != nil {
			jobByID[job.GetId()] = job
		}
	}
	phasesByJob := make(map[string]map[string]*jobphasepb.JobPhase)
	for _, phase := range projection.GetJobPhases() {
		if phase == nil || phase.GetJobId() == "" || jobByID[phase.GetJobId()] == nil {
			continue
		}
		templatePhase := templatePhaseByID[phase.GetTemplatePhaseId()]
		code := ""
		if templatePhase != nil {
			code = strings.TrimSpace(templatePhase.GetCode())
		}
		if code == "" {
			code = strings.TrimSpace(phase.GetName())
		}
		if code == "" {
			code = phase.GetId()
		}
		if phasesByJob[phase.GetJobId()] == nil {
			phasesByJob[phase.GetJobId()] = make(map[string]*jobphasepb.JobPhase)
		}
		phasesByJob[phase.GetJobId()][code] = phase
	}
	phaseSummaryByID := make(map[string]*phasesumpb.PhaseOutcomeSummary, len(projection.GetPhaseOutcomeSummaries()))
	for _, summary := range projection.GetPhaseOutcomeSummaries() {
		if summary != nil && summary.GetActive() && summary.GetJobPhaseId() != "" {
			phaseSummaryByID[summary.GetJobPhaseId()] = summary
		}
	}
	yearSummaryByJob := make(map[string]*jobsumpb.JobOutcomeSummary, len(projection.GetJobOutcomeSummaries()))
	for _, summary := range projection.GetJobOutcomeSummaries() {
		if summary != nil && summary.GetActive() && summary.GetJobId() != "" {
			yearSummaryByJob[summary.GetJobId()] = summary
		}
	}

	sort.SliceStable(jobs, func(i, j int) bool {
		left := subjectName(jobTemplateByID[jobs[i].GetJobTemplateId()], jobs[i].GetId())
		right := subjectName(jobTemplateByID[jobs[j].GetJobTemplateId()], jobs[j].GetId())
		if left != right {
			return left < right
		}
		return jobs[i].GetId() < jobs[j].GetId()
	})

	labels := deps.Labels
	columns := make([]types.TableColumn, 0, len(phaseColumns)+2)
	columns = append(columns, types.TableColumn{Key: "subject", Label: labels.Client.SubjectColumn, MinWidth: "14rem", NoSort: true})
	for _, phase := range phaseColumns {
		columns = append(columns, types.TableColumn{Key: "phase-" + phase.code, Label: phase.label, Align: "center", MinWidth: "5rem", NoSort: true})
	}
	finalLabel := strings.TrimSpace(labels.Client.YearColumn)
	if text := strings.TrimSpace(labels.Client.FinalColumn); text != "" {
		if finalLabel != "" {
			finalLabel += " "
		}
		finalLabel += text
	}
	columns = append(columns, types.TableColumn{Key: "year-final", Label: finalLabel, Align: "center", MinWidth: "5rem", NoSort: true})

	rows := make([]types.TableRow, 0, len(jobs))
	rowsByCategory := make(map[string][]types.TableRow)
	for _, job := range jobs {
		if job == nil || strings.TrimSpace(job.GetId()) == "" {
			continue
		}
		template := jobTemplateByID[job.GetJobTemplateId()]
		cells := []types.TableCell{{Type: "text", Value: subjectName(template, job.GetId()), CSVValue: subjectName(template, job.GetId())}}
		phaseRows := phasesByJob[job.GetId()]
		for _, phaseColumn := range phaseColumns {
			phase := phaseRows[phaseColumn.code]
			grade := ""
			if phase != nil {
				grade = phaseOutcomeValue(phaseSummaryByID[phase.GetId()], deps.Labels.SubscriptionGroup.RatingEmpty)
			}
			cells = append(cells, types.TableCell{Type: "text", Value: grade, CSVValue: grade})
		}
		year := deps.Labels.SubscriptionGroup.RatingEmpty
		if year == "" {
			year = "—"
		}
		if summary := yearSummaryByJob[job.GetId()]; summary != nil {
			year = outcomeValue(summary.ScaledLabel, summary.ScaledScore, summary.SummaryScore, summary.Narrative, deps.Labels.SubscriptionGroup.RatingEmpty)
		}
		cells = append(cells, types.TableCell{Type: "text", Value: year, CSVValue: year})
		categoryID := ""
		if template != nil {
			categoryID = template.GetJobCategoryId()
		}
		categoryCode := "uncategorized"
		if category := categoryByID[categoryID]; category != nil {
			categoryCode = strings.TrimSpace(category.GetCode())
			if categoryCode == "" {
				categoryCode = category.GetId()
			}
		}
		row := types.TableRow{
			ID:        job.GetId(),
			DataAttrs: map[string]string{"testid": "rc-subject-" + short(job.GetId()), "job-category": categoryCode},
			Cells:     cells,
		}
		rows = append(rows, row)
		rowsByCategory[categoryID] = append(rowsByCategory[categoryID], row)
	}

	table := &types.TableConfig{
		ID:          "report-cards-client",
		Columns:     columns,
		Rows:        rows,
		ShowSearch:  false,
		ShowColumns: true,
		ShowDensity: true,
		ShowExport:  true,
		ShowEntries: true,
		Labels:      deps.TableLabels,
		Caption:     labels.Client.Title,
		EmptyState:  types.TableEmptyState{Title: labels.Empty.Title, Message: labels.SubscriptionGroup.NotComputedBanner},
	}
	if bandByCategory {
		table.Rows = nil
		table.Groups = buildProjectionGroups(projection.GetJobCategories(), rowsByCategory, categoryByID, labels.Client.UncategorizedBand)
		for index := range table.Groups {
			types.ApplyColumnStyles(table.Columns, table.Groups[index].Rows)
		}
	} else {
		types.ApplyColumnStyles(table.Columns, table.Rows)
	}
	return table
}

func buildProjectionGroups(categories []*jobcategorypb.JobCategory, rowsByCategory map[string][]types.TableRow, categoryByID map[string]*jobcategorypb.JobCategory, uncategorizedTitle string) []types.TableRowGroup {
	groups := make([]types.TableRowGroup, 0, len(rowsByCategory))
	ordered := append([]*jobcategorypb.JobCategory(nil), categories...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
		if left.GetSortOrder() != right.GetSortOrder() {
			return left.GetSortOrder() < right.GetSortOrder()
		}
		if left.GetName() != right.GetName() {
			return left.GetName() < right.GetName()
		}
		return left.GetId() < right.GetId()
	})
	seen := make(map[string]bool)
	for _, category := range ordered {
		if category == nil || len(rowsByCategory[category.GetId()]) == 0 {
			continue
		}
		code := strings.TrimSpace(category.GetCode())
		if code == "" {
			code = category.GetId()
		}
		groups = append(groups, types.TableRowGroup{
			ID:        "rc-band-" + safeKey(code),
			Title:     category.GetName(),
			Rows:      rowsByCategory[category.GetId()],
			DataAttrs: map[string]string{"testid": "rc-band-" + safeKey(code)},
		})
		seen[category.GetId()] = true
	}
	uncategorized := make([]types.TableRow, 0)
	for categoryID, rows := range rowsByCategory {
		if categoryID == "" || !seen[categoryID] || categoryByID[categoryID] == nil {
			uncategorized = append(uncategorized, rows...)
		}
	}
	if len(uncategorized) > 0 {
		sort.SliceStable(uncategorized, func(i, j int) bool { return uncategorized[i].ID < uncategorized[j].ID })
		if strings.TrimSpace(uncategorizedTitle) == "" {
			uncategorizedTitle = "Uncategorized"
		}
		groups = append(groups, types.TableRowGroup{ID: "rc-band-uncategorized", Title: uncategorizedTitle, Rows: uncategorized, DataAttrs: map[string]string{"testid": "rc-band-uncategorized"}})
	}
	return groups
}

func phaseOutcomeValue(summary *phasesumpb.PhaseOutcomeSummary, empty string) string {
	if summary == nil {
		return empty
	}
	return outcomeValue(summary.ScaledLabel, summary.ScaledScore, summary.SummaryScore, summary.Narrative, empty)
}

func outcomeValue(label *string, scaled, summary *float64, narrative *string, empty string) string {
	if label != nil && strings.TrimSpace(*label) != "" {
		return strings.TrimSpace(*label)
	}
	if scaled != nil {
		return fmt.Sprintf("%g", *scaled)
	}
	if summary != nil {
		return fmt.Sprintf("%g", *summary)
	}
	if narrative != nil && strings.TrimSpace(*narrative) != "" {
		return strings.TrimSpace(*narrative)
	}
	return empty
}

func subjectName(template *jobtemplatepb.JobTemplate, fallback string) string {
	if template != nil {
		if name := strings.TrimSpace(template.GetName()); name != "" {
			return name
		}
	}
	return fallback
}

func projectedClientName(client *exportpb.ClientReportCardClient) string {
	if client == nil {
		return ""
	}
	if name := strings.TrimSpace(client.GetName()); name != "" {
		return name
	}
	first, last := strings.TrimSpace(client.GetFirstName()), strings.TrimSpace(client.GetLastName())
	if first != "" && last != "" {
		return last + ", " + first
	}
	return strings.TrimSpace(strings.Join([]string{first, last}, " "))
}

func safeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if b.Len() > 0 {
			b.WriteByte('-')
		}
	}
	if key := strings.Trim(b.String(), "-"); key != "" {
		return key
	}
	return "uncategorized"
}

func clientProjectionPage(ctx context.Context, viewCtx *view.ViewContext, deps *Deps, projection *exportpb.ClientReportCardProjection, clientName string, table *types.TableConfig) view.ViewResult {
	labels := deps.Labels
	context := projection.GetContext()
	groupName := strings.TrimSpace(context.GetSubscriptionGroupName())
	headerTitle := clientName
	if groupName != "" {
		headerTitle += " — " + groupName
	}
	if table != nil {
		configureClientToolbar(ctx, table, deps, context.GetSubscriptionGroupId(), projection.GetClient().GetClientId())
	}
	pageData := &PageData{
		PageData: types.PageData{
			CacheVersion:        viewCtx.CacheVersion,
			Title:               labels.Client.Title,
			CurrentPath:         viewCtx.CurrentPath,
			ActiveNav:           deps.Routes.ActiveNav,
			ActiveSubNav:        "report-cards",
			HeaderBreadcrumb:    labels.SubscriptionGroup.Title,
			HeaderBreadcrumbURL: route.ResolveURL(deps.Routes.SubscriptionGroupURL, "id", context.GetSubscriptionGroupId()),
			HeaderTitle:         headerTitle,
			HeaderSubtitle:      labels.Client.Subtitle,
			HeaderIcon:          "icon-award",
			CommonLabels:        deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-client-content",
		Table:           table,
		NotComputed:     table == nil,
		Banner:          labels.SubscriptionGroup.NotComputedBanner,
	}
	return view.OK("outcome-summary-client", pageData)
}

func configureClientToolbar(ctx context.Context, table *types.TableConfig, deps *Deps, groupID, clientID string) {
	if table == nil {
		return
	}
	labels := deps.Labels
	if deps.ClientDocumentMounted && deps.Options.SubscriptionGroupExportEnabled() &&
		deps.Routes.ClientDocumentURL != "" && deps.Routes.ClientDownloadDrawerURL != "" &&
		outcome_summary.CanExplicitExport(view.GetUserPermissions(ctx), ctx, deps.ResolvePrincipalKind) {
		actionURL := route.ResolveURL(deps.Routes.ClientDownloadDrawerURL, "id", groupID, "client_id", clientID)
		table.PrimaryAction = &types.PrimaryAction{Label: labels.Client.DownloadAction, ActionURL: actionURL, SheetTitle: labels.Client.DownloadAction, TestID: "rc-download-pdf"}
	}
}
