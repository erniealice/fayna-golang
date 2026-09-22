package client_card

import (
	"context"
	"errors"
	"log"
	"sort"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

const maxNarrowMatrixCalls = 12

func narrowReportViewEnabled(ctx context.Context, deps *Deps, perms *types.UserPermissions) bool {
	return !outcome_summary.CanLegacyDetail(perms) &&
		outcome_summary.CanExplicitExport(perms, ctx, deps.ResolvePrincipalKind) &&
		deps.Options.List.SubscriptionGroups() &&
		deps.Options.SubscriptionGroupExportEnabled() &&
		deps.GetSubscriptionGroupOutcomeExport != nil
}

type reportMatrixReader struct {
	ctx       context.Context
	deps      *Deps
	callCount int
	capLogged bool
}

var errNarrowMatrixCallCap = errors.New("narrow matrix call cap reached")

func (r *reportMatrixReader) logCap() {
	if r.capLogged {
		return
	}
	r.capLogged = true
	log.Printf("client report view: matrix call cap reached; truncating remaining reads")
}

func (r *reportMatrixReader) read(request *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
	if r.callCount >= maxNarrowMatrixCalls {
		r.logCap()
		return nil, errNarrowMatrixCallCap
	}
	r.callCount++
	response, err := r.deps.GetSubscriptionGroupOutcomeExport(r.ctx, request)
	if err != nil {
		log.Printf("client report view: matrix read failed: %v", err)
		return nil, err
	}
	return response, nil
}

func clientReportMatrixValid(response *exportpb.GetSubscriptionGroupOutcomeExportResponse, subscriptionGroupID string) bool {
	return response != nil && response.GetSuccess() && response.GetContext() != nil && response.GetContext().GetSubscriptionGroupId() == subscriptionGroupID
}

func renderReportView(ctx context.Context, viewCtx *view.ViewContext, deps *Deps, subscriptionGroupID, clientID string) view.ViewResult {
	if deps.GetSubscriptionGroupOutcomeExport == nil {
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	options, err := deps.GetSubscriptionGroupOutcomeExport(ctx, &exportpb.GetSubscriptionGroupOutcomeExportRequest{
		SubscriptionGroupId: subscriptionGroupID,
	})
	if err != nil || !clientReportMatrixValid(options, subscriptionGroupID) {
		return view.Forbidden("subscription_group_outcome_export:read")
	}

	reader := &reportMatrixReader{ctx: ctx, deps: deps}
	categories := options.GetJobCategories()
	rawCategoryID := strings.TrimSpace(viewCtx.Request.URL.Query().Get("jc"))
	selected := reportCategoryByID(categories, rawCategoryID)
	finalByCategory := make(map[string]*exportpb.GetSubscriptionGroupOutcomeExportResponse)
	finalAttempted := make(map[string]bool)
	var finalResponse *exportpb.GetSubscriptionGroupOutcomeExportResponse
	if selected != nil {
		if selected.GetFinalOutcomeAvailable() {
			categoryID := selected.GetJobCategoryId()
			var readErr error
			finalResponse, readErr = reader.read(clientFinalRequest(subscriptionGroupID, categoryID))
			if readErr != nil || !clientReportMatrixValid(finalResponse, subscriptionGroupID) {
				return view.Forbidden("subscription_group_outcome_export:read")
			}
			finalByCategory[categoryID] = finalResponse
			finalAttempted[categoryID] = true
		}
	} else {
		for _, category := range categories {
			if category == nil || !category.GetFinalOutcomeAvailable() || strings.TrimSpace(category.GetJobCategoryId()) == "" {
				continue
			}
			categoryID := category.GetJobCategoryId()
			response, readErr := reader.read(clientFinalRequest(subscriptionGroupID, categoryID))
			if errors.Is(readErr, errNarrowMatrixCallCap) {
				break
			}
			if readErr != nil || !clientReportMatrixValid(response, subscriptionGroupID) {
				return view.Forbidden("subscription_group_outcome_export:read")
			}
			finalByCategory[categoryID] = response
			finalAttempted[categoryID] = true
			if clientRowFromResponse(response, clientID) != nil {
				selected = category
				finalResponse = response
				break
			}
		}
	}
	if selected == nil {
		selected = firstReportCategory(categories)
	}
	if selected == nil {
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	categoryID := selected.GetJobCategoryId()
	if finalResponse == nil {
		finalResponse = finalByCategory[categoryID]
	}
	if !finalAttempted[categoryID] && selected.GetFinalOutcomeAvailable() {
		var readErr error
		finalResponse, readErr = reader.read(clientFinalRequest(subscriptionGroupID, categoryID))
		if readErr != nil || !clientReportMatrixValid(finalResponse, subscriptionGroupID) {
			return view.Forbidden("subscription_group_outcome_export:read")
		}
		finalByCategory[categoryID] = finalResponse
		finalAttempted[categoryID] = true
	}

	phases := reportPhases(selected)
	remaining := maxNarrowMatrixCalls - reader.callCount
	if remaining < len(phases) {
		reader.logCap()
		if remaining < 0 {
			remaining = 0
		}
		phases = phases[:remaining]
	}
	phaseResponses := make([]clientPhaseResponse, 0, len(phases))
	for _, phase := range phases {
		response, readErr := reader.read(clientPhaseRequest(subscriptionGroupID, categoryID, phase.GetCode()))
		if readErr != nil || !clientReportMatrixValid(response, subscriptionGroupID) {
			return view.Forbidden("subscription_group_outcome_export:read")
		}
		phaseResponses = append(phaseResponses, clientPhaseResponse{phase: phase, response: response})
	}

	finalRow := clientRowFromResponse(finalResponse, clientID)
	phaseRows := make([]clientPhaseResponse, 0, len(phaseResponses))
	for _, phaseResponse := range phaseResponses {
		if clientRowFromResponse(phaseResponse.response, clientID) != nil {
			phaseRows = append(phaseRows, phaseResponse)
		}
	}
	if finalRow == nil && len(phaseRows) == 0 {
		return view.Forbidden("subscription_group_outcome_export:read")
	}

	subjectColumns := reportColumns(finalResponse, phaseResponses)
	if len(subjectColumns) == 0 {
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	clientName := clientReportName(finalRow)
	if clientName == "" {
		for _, phaseResponse := range phaseRows {
			if row := clientRowFromResponse(phaseResponse.response, clientID); row != nil {
				clientName = clientReportName(row)
				break
			}
		}
	}
	if clientName == "" {
		clientName = clientID
	}
	table := buildClientReportTable(deps, subjectColumns, phases, finalResponse, phaseResponses, clientID)
	return clientReportPage(viewCtx, deps, options.GetContext(), clientName, table)
}

type clientPhaseResponse struct {
	phase    *exportpb.JobTemplatePhaseOption
	response *exportpb.GetSubscriptionGroupOutcomeExportResponse
}

func clientFinalRequest(subscriptionGroupID, categoryID string) *exportpb.GetSubscriptionGroupOutcomeExportRequest {
	return &exportpb.GetSubscriptionGroupOutcomeExportRequest{
		SubscriptionGroupId: subscriptionGroupID,
		JobCategoryId:       stringPointer(categoryID),
		OutcomeSelector:     &exportpb.GetSubscriptionGroupOutcomeExportRequest_FinalOutcome{FinalOutcome: true},
	}
}

func clientPhaseRequest(subscriptionGroupID, categoryID, phaseCode string) *exportpb.GetSubscriptionGroupOutcomeExportRequest {
	return &exportpb.GetSubscriptionGroupOutcomeExportRequest{
		SubscriptionGroupId: subscriptionGroupID,
		JobCategoryId:       stringPointer(categoryID),
		OutcomeSelector:     &exportpb.GetSubscriptionGroupOutcomeExportRequest_JobTemplatePhaseCode{JobTemplatePhaseCode: phaseCode},
	}
}

func reportCategoryByID(categories []*exportpb.JobCategoryOption, id string) *exportpb.JobCategoryOption {
	if id == "" {
		return nil
	}
	for _, category := range categories {
		if category != nil && category.GetJobCategoryId() == id {
			return category
		}
	}
	return nil
}

func firstReportCategory(categories []*exportpb.JobCategoryOption) *exportpb.JobCategoryOption {
	for _, category := range categories {
		if category != nil {
			return category
		}
	}
	return nil
}

func reportPhases(category *exportpb.JobCategoryOption) []*exportpb.JobTemplatePhaseOption {
	if category == nil {
		return nil
	}
	phases := make([]*exportpb.JobTemplatePhaseOption, 0, len(category.GetJobTemplatePhases()))
	for _, phase := range category.GetJobTemplatePhases() {
		if phase != nil && !phase.GetAmbiguous() && strings.TrimSpace(phase.GetCode()) != "" {
			phases = append(phases, phase)
		}
	}
	sort.SliceStable(phases, func(i, j int) bool {
		if phases[i].GetSequenceOrder() != phases[j].GetSequenceOrder() {
			return phases[i].GetSequenceOrder() < phases[j].GetSequenceOrder()
		}
		return phases[i].GetCode() < phases[j].GetCode()
	})
	return phases
}

func clientRowFromResponse(response *exportpb.GetSubscriptionGroupOutcomeExportResponse, clientID string) *exportpb.SubscriptionGroupOutcomeClientRow {
	if response == nil || !response.GetSuccess() {
		return nil
	}
	for _, row := range response.GetClientRows() {
		if row != nil && row.GetClientId() == clientID {
			return row
		}
	}
	return nil
}

func reportColumns(finalResponse *exportpb.GetSubscriptionGroupOutcomeExportResponse, phaseResponses []clientPhaseResponse) []*exportpb.JobTemplateColumn {
	if finalResponse != nil && len(finalResponse.GetJobTemplateColumns()) > 0 {
		return finalResponse.GetJobTemplateColumns()
	}
	for _, phaseResponse := range phaseResponses {
		if phaseResponse.response != nil && len(phaseResponse.response.GetJobTemplateColumns()) > 0 {
			return phaseResponse.response.GetJobTemplateColumns()
		}
	}
	return nil
}

func buildClientReportTable(deps *Deps, subjectColumns []*exportpb.JobTemplateColumn, phases []*exportpb.JobTemplatePhaseOption, finalResponse *exportpb.GetSubscriptionGroupOutcomeExportResponse, phaseResponses []clientPhaseResponse, clientID string) *types.TableConfig {
	l := deps.Labels
	columns := make([]types.TableColumn, 0, len(phases)+2)
	columns = append(columns, types.TableColumn{Key: "subject", Label: l.Client.SubjectColumn, MinWidth: "14rem", NoSort: true})
	for _, phase := range phases {
		label := strings.TrimSpace(phase.GetName())
		if label == "" {
			label = phase.GetCode()
		}
		columns = append(columns, types.TableColumn{Key: "phase-" + phase.GetCode(), Label: label, Align: "center", MinWidth: "5rem", NoSort: true})
	}
	finalLabel := strings.TrimSpace(l.Client.YearColumn)
	if finalText := strings.TrimSpace(l.Client.FinalColumn); finalText != "" {
		if finalLabel != "" {
			finalLabel += " "
		}
		finalLabel += finalText
	}
	columns = append(columns, types.TableColumn{Key: "year-final", Label: finalLabel, Align: "center", MinWidth: "5rem", NoSort: true})

	finalRow := clientRowFromResponse(finalResponse, clientID)
	rows := make([]types.TableRow, 0, len(subjectColumns))
	for _, column := range subjectColumns {
		cells := []types.TableCell{{Value: explicitColumnLabel(column)}}
		for _, phaseResponse := range phaseResponses {
			phaseRow := clientRowFromResponse(phaseResponse.response, clientID)
			value := reportValueForTemplate(phaseRow, column.GetJobTemplateId(), l.SubscriptionGroup.RatingEmpty)
			cells = append(cells, types.TableCell{Type: "text", Value: value, CSVValue: value})
		}
		value := reportValueForTemplate(finalRow, column.GetJobTemplateId(), l.SubscriptionGroup.RatingEmpty)
		cells = append(cells, types.TableCell{Type: "text", Value: value, CSVValue: value})
		rows = append(rows, types.TableRow{
			ID:        column.GetJobTemplateId(),
			DataAttrs: map[string]string{"testid": "rc-subject-" + short(column.GetJobTemplateId())},
			Cells:     cells,
		})
	}
	types.ApplyColumnStyles(columns, rows)
	return &types.TableConfig{
		ID:          "report-cards-client",
		Columns:     columns,
		Rows:        rows,
		ShowSearch:  true,
		ShowColumns: true,
		ShowDensity: true,
		ShowExport:  true,
		ShowEntries: true,
		Labels:      deps.TableLabels,
		Caption:     l.Client.Title,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.SubscriptionGroup.NotComputedBanner,
		},
	}
}

func reportValueForTemplate(row *exportpb.SubscriptionGroupOutcomeClientRow, templateID, empty string) string {
	if row == nil {
		return empty
	}
	for _, cell := range row.GetCells() {
		if cell != nil && cell.GetJobTemplateId() == templateID {
			return outcome_summary.ExportCellValue(cell, empty)
		}
	}
	return empty
}

func explicitColumnLabel(column *exportpb.JobTemplateColumn) string {
	if column == nil {
		return ""
	}
	if label := strings.TrimSpace(column.GetDisplayName()); label != "" {
		return label
	}
	return column.GetJobTemplateId()
}

func clientReportName(row *exportpb.SubscriptionGroupOutcomeClientRow) string {
	if row == nil {
		return ""
	}
	lastName := strings.TrimSpace(row.GetClientLastName())
	firstName := strings.TrimSpace(row.GetClientFirstName())
	if lastName != "" && firstName != "" {
		return lastName + ", " + firstName
	}
	if name := strings.TrimSpace(row.GetClientName()); name != "" {
		return name
	}
	return strings.TrimSpace(strings.Join([]string{firstName, lastName}, " "))
}

func clientReportPage(viewCtx *view.ViewContext, deps *Deps, exportContext *exportpb.SubscriptionGroupOutcomeExportContext, clientName string, table *types.TableConfig) view.ViewResult {
	l := deps.Labels
	headerTitle := clientName
	if groupName := strings.TrimSpace(exportContext.GetSubscriptionGroupName()); groupName != "" {
		if headerTitle != "" {
			headerTitle += " — "
		}
		headerTitle += groupName
	}
	pageData := &PageData{
		PageData: types.PageData{
			CacheVersion:        viewCtx.CacheVersion,
			Title:               l.Client.Title,
			CurrentPath:         viewCtx.CurrentPath,
			ActiveNav:           deps.Routes.ActiveNav,
			ActiveSubNav:        "report-cards",
			HeaderBreadcrumb:    l.SubscriptionGroup.Title,
			HeaderBreadcrumbURL: route.ResolveURL(deps.Routes.SubscriptionGroupURL, "id", exportContext.GetSubscriptionGroupId()),
			HeaderTitle:         headerTitle,
			HeaderSubtitle:      l.Client.Subtitle,
			HeaderIcon:          "icon-award",
			CommonLabels:        deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-client-content",
		Table:           table,
		NotComputed:     table == nil,
		Banner:          l.SubscriptionGroup.NotComputedBanner,
	}
	return view.OK("outcome-summary-client", pageData)
}

func stringPointer(value string) *string { return &value }
