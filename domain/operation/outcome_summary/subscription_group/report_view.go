package subscription_group

import (
	"context"
	"log"
	"sort"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func narrowReportViewEnabled(ctx context.Context, deps *Deps, perms *types.UserPermissions) bool {
	return !outcome_summary.CanLegacyDetail(perms) &&
		outcome_summary.CanExplicitExport(perms, ctx, deps.ResolvePrincipalKind) &&
		deps.Options.List.SubscriptionGroups() &&
		deps.Options.SubscriptionGroupExportEnabled() &&
		deps.GetSubscriptionGroupOutcomeExport != nil
}

func reportMatrixValid(response *exportpb.GetSubscriptionGroupOutcomeExportResponse, subscriptionGroupID string) bool {
	return response != nil && response.GetSuccess() && response.GetContext() != nil && response.GetContext().GetSubscriptionGroupId() == subscriptionGroupID
}

func renderReportView(ctx context.Context, viewCtx *view.ViewContext, deps *Deps, subscriptionGroupID string) view.ViewResult {
	if deps.GetSubscriptionGroupOutcomeExport == nil {
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	options, err := deps.GetSubscriptionGroupOutcomeExport(ctx, &exportpb.GetSubscriptionGroupOutcomeExportRequest{
		SubscriptionGroupId: subscriptionGroupID,
	})
	if err != nil || !reportMatrixValid(options, subscriptionGroupID) {
		return view.Forbidden("subscription_group_outcome_export:read")
	}

	category := selectReportCategory(options.GetJobCategories(), viewCtx.Request.URL.Query().Get("jc"))
	tabs, activeTab := reportCategoryTabs(deps, subscriptionGroupID, options.GetJobCategories(), category)
	if category == nil || !category.GetFinalOutcomeAvailable() {
		return reportPage(viewCtx, deps, options.GetContext(), nil, tabs, activeTab, deps.Labels.SubscriptionGroup.NotComputedBanner)
	}

	categoryID := strings.TrimSpace(category.GetJobCategoryId())
	if categoryID == "" {
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	matrix, err := deps.GetSubscriptionGroupOutcomeExport(ctx, &exportpb.GetSubscriptionGroupOutcomeExportRequest{
		SubscriptionGroupId: subscriptionGroupID,
		JobCategoryId:       &categoryID,
		OutcomeSelector:     &exportpb.GetSubscriptionGroupOutcomeExportRequest_FinalOutcome{FinalOutcome: true},
	})
	if err != nil || !reportMatrixValid(matrix, subscriptionGroupID) {
		return view.Forbidden("subscription_group_outcome_export:read")
	}

	normalized, err := normalizeReportMatrix(ctx, deps, matrix)
	if err != nil {
		log.Printf("subscription group report view: normalize matrix: %v", err)
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	table := buildReportTable(deps, normalized, subscriptionGroupID)
	return reportPage(viewCtx, deps, options.GetContext(), table, tabs, activeTab, "")
}

func normalizeReportMatrix(ctx context.Context, deps *Deps, response *exportpb.GetSubscriptionGroupOutcomeExportResponse) (*explicitMatrix, error) {
	isolated := *deps
	isolated.Options.Row.GroupByField = ""
	isolated.Options.Row.GroupValueOrder = nil
	isolated.Options.Row.SortField = ""
	isolated.Options.Row.SortDirection = ""
	return normalizeExplicitMatrix(ctx, &isolated, response)
}

func selectReportCategory(categories []*exportpb.JobCategoryOption, rawID string) *exportpb.JobCategoryOption {
	rawID = strings.TrimSpace(rawID)
	if rawID != "" {
		for _, category := range categories {
			if category != nil && category.GetJobCategoryId() == rawID {
				return category
			}
		}
	}
	for _, category := range categories {
		if category != nil && category.GetFinalOutcomeAvailable() {
			return category
		}
	}
	for _, category := range categories {
		if category != nil {
			return category
		}
	}
	return nil
}

func reportCategoryTabs(deps *Deps, subscriptionGroupID string, categories []*exportpb.JobCategoryOption, selected *exportpb.JobCategoryOption) ([]pyeza.TabItem, string) {
	base := route.ResolveURL(deps.Routes.SubscriptionGroupURL, "id", subscriptionGroupID)
	items := make([]pyeza.TabItem, 0, len(categories))
	for _, category := range categories {
		if category == nil || strings.TrimSpace(category.GetJobCategoryId()) == "" {
			continue
		}
		categoryID := category.GetJobCategoryId()
		label := strings.TrimSpace(category.GetName())
		if label == "" {
			label = categoryID
		}
		items = append(items, pyeza.TabItem{
			Key:   groupTabKey(categoryID),
			Label: label,
			Href:  groupCategoryURL(base, categoryID),
		})
	}
	active := ""
	if selected != nil {
		active = groupTabKey(selected.GetJobCategoryId())
	}
	return items, active
}

func buildReportTable(deps *Deps, matrix *explicitMatrix, subscriptionGroupID string) *types.TableConfig {
	columns := make([]types.TableColumn, 0, len(matrix.columns)+1)
	columns = append(columns, types.TableColumn{
		Key:      "client",
		Label:    deps.Labels.SubscriptionGroup.ClientColumn,
		MinWidth: "14rem",
		NoSort:   true,
	})
	for _, column := range matrix.columns {
		columns = append(columns, types.TableColumn{
			Key:      "tmpl-" + column.GetJobTemplateId(),
			Label:    explicitColumnName(column),
			MinWidth: "6.25rem",
			Align:    "center",
			NoSort:   true,
		})
	}

	rows := append([]explicitRow(nil), matrix.rows...)
	sort.SliceStable(rows, func(i, j int) bool {
		left, right := rows[i], rows[j]
		leftLast, rightLast := strings.ToLower(strings.TrimSpace(left.lastName)), strings.ToLower(strings.TrimSpace(right.lastName))
		if leftLast != rightLast {
			return leftLast < rightLast
		}
		leftFirst, rightFirst := strings.ToLower(strings.TrimSpace(left.firstName)), strings.ToLower(strings.TrimSpace(right.firstName))
		if leftFirst != rightFirst {
			return leftFirst < rightFirst
		}
		return left.clientID < right.clientID
	})

	empty := deps.Labels.SubscriptionGroup.RatingEmpty
	tableRows := make([]types.TableRow, 0, len(rows))
	for _, row := range rows {
		if row.band || row.blank {
			continue
		}
		name := reportClientName(row)
		cells := []types.TableCell{{
			Type:     "link",
			Value:    name,
			Href:     route.ResolveURL(deps.Routes.ClientCardURL, "id", subscriptionGroupID, "client_id", row.clientID),
			TestID:   "rc-client-" + short(row.clientID),
			CSVValue: name,
		}}
		for index := range matrix.columns {
			value := outcome_summary.ExportCellValue(row.cells[matrix.columns[index].GetJobTemplateId()], empty)
			cells = append(cells, types.TableCell{Type: "text", Value: value, CSVValue: value})
		}
		tableRows = append(tableRows, types.TableRow{
			ID:        row.clientID,
			DataAttrs: map[string]string{"testid": "rc-row-" + short(row.clientID)},
			Cells:     cells,
		})
	}

	table := &types.TableConfig{
		ID:          "report-cards-grid",
		Columns:     columns,
		Rows:        tableRows,
		ShowSearch:  true,
		ShowColumns: true,
		ShowDensity: true,
		ShowExport:  true,
		ShowEntries: true,
		ShowActions: false,
		Labels:      deps.TableLabels,
		Caption:     deps.Labels.SubscriptionGroup.Title,
		EmptyState: types.TableEmptyState{
			Title:   deps.Labels.Empty.Title,
			Message: deps.Labels.SubscriptionGroup.NotComputedBanner,
		},
	}
	types.ApplyColumnStyles(table.Columns, table.Rows)
	return table
}

func reportClientName(row explicitRow) string {
	lastName := strings.TrimSpace(row.lastName)
	firstName := strings.TrimSpace(row.firstName)
	if lastName != "" && firstName != "" {
		return lastName + ", " + firstName
	}
	if name := strings.TrimSpace(row.name); name != "" {
		return name
	}
	return strings.TrimSpace(strings.Join([]string{firstName, lastName}, " "))
}

func reportPage(viewCtx *view.ViewContext, deps *Deps, exportContext *exportpb.SubscriptionGroupOutcomeExportContext, table *types.TableConfig, tabs []pyeza.TabItem, activeTab, banner string) view.ViewResult {
	l := deps.Labels
	pageData := &PageData{
		PageData: types.PageData{
			CacheVersion:        viewCtx.CacheVersion,
			Title:               l.SubscriptionGroup.Title,
			CurrentPath:         viewCtx.CurrentPath,
			ActiveNav:           deps.Routes.ActiveNav,
			ActiveSubNav:        "report-cards",
			HeaderBreadcrumb:    l.SubscriptionGroup.Title,
			HeaderBreadcrumbURL: route.ResolveURL(deps.Routes.ListURL),
			HeaderTitle:         exportContext.GetSubscriptionGroupName(),
			HeaderIcon:          "icon-award",
			CommonLabels:        deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-subscription-group-content",
		Table:           table,
		NotComputed:     table == nil,
		Banner:          banner,
		TabItems:        tabs,
		ActiveTab:       activeTab,
		TabsAria:        l.SubscriptionGroup.CategoryTabsAriaLabel,
	}
	return view.OK("outcome-summary-subscription-group", pageData)
}
