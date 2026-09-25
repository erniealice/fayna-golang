package subscription_group

import (
	"context"
	"log"
	"net/url"
	"strconv"
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
		return reportPage(viewCtx, deps, options.GetContext(), nil, tabs, activeTab, reportCategoryID(category), deps.Labels.SubscriptionGroup.NotComputedBanner)
	}

	categoryID := strings.TrimSpace(category.GetJobCategoryId())
	if categoryID == "" {
		return view.Forbidden("subscription_group_outcome_export:read")
	}
	if _, _, bandConfigured, _ := deps.Options.ExportRowBandConfig(); bandConfigured {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("attribute", "list") {
			return view.Forbidden("attribute:list")
		}
		if !perms.Can("client_attribute", "list") {
			return view.Forbidden("client_attribute:list")
		}
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
	// Same per-row action gating as the static grid (page.go buildGroupTable):
	// the explicit drawer download takes priority over the legacy PDF anchor,
	// and neither requires the view-client-card action, which is always shown.
	perms := view.GetUserPermissions(ctx)
	clientDocumentMounted := deps.ClientDocumentMounted && deps.Routes.ClientDocumentURL != ""
	canDownloadClientDocument := clientDocumentMounted && deps.Options.SubscriptionGroupExportEnabled() &&
		deps.Routes.ClientDownloadDrawerURL != "" && outcome_summary.CanExplicitExport(perms, ctx, deps.ResolvePrincipalKind)
	canDownloadLegacyDocument := clientDocumentMounted && outcome_summary.CanLegacyDetail(perms)
	table := buildReportTable(deps, normalized, subscriptionGroupID, canDownloadClientDocument, canDownloadLegacyDocument)
	return reportPage(viewCtx, deps, options.GetContext(), table, tabs, activeTab, reportCategoryID(category), "")
}

func normalizeReportMatrix(ctx context.Context, deps *Deps, response *exportpb.GetSubscriptionGroupOutcomeExportResponse) (*explicitMatrix, error) {
	// Keep the configured band and sort projection used by the group exporter.
	// orderExplicitRows prefixes names for CSV; restore the source names here so
	// the group grid can apply its own final, continuous row numbering.
	originalNames := make(map[string]string, len(response.GetClientRows()))
	for _, row := range response.GetClientRows() {
		if row != nil {
			originalNames[row.GetClientId()] = row.GetClientName()
		}
	}
	matrix, err := normalizeExplicitMatrix(ctx, deps, response)
	if err != nil {
		return nil, err
	}
	for i := range matrix.rows {
		if !matrix.rows[i].band && !matrix.rows[i].blank {
			matrix.rows[i].name = originalNames[matrix.rows[i].clientID]
		}
	}
	return matrix, nil
}

func reportCategoryID(category *exportpb.JobCategoryOption) string {
	if category == nil {
		return ""
	}
	return strings.TrimSpace(category.GetJobCategoryId())
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

func buildReportTable(deps *Deps, matrix *explicitMatrix, subscriptionGroupID string, canDownloadClientDocument, canDownloadLegacyDocument bool) *types.TableConfig {
	columns := make([]types.TableColumn, 0, len(matrix.columns)+2)
	// Client column is the first FROZEN column (TableClass
	// "data-table-freeze2", set below). Width must equal the CSS --freeze2-c1
	// default (14rem) so the second frozen column's sticky left offset lines
	// up — same contract as the static grid (page.go buildColumns).
	columns = append(columns, types.TableColumn{
		Key:      "client",
		Label:    deps.Labels.SubscriptionGroup.ClientColumn,
		Width:    "14rem",
		MinWidth: "14rem",
		NoSort:   true,
	})
	// Second frozen column: the per-row actions (view client card + download
	// drawer / legacy PDF). Blank header mirrors the static grid's action
	// column. Excluded from CSV/Excel export by key (actionsColumnKey).
	columns = append(columns, types.TableColumn{Key: actionsColumnKey, Label: "", Width: "5rem", MinWidth: "5rem", Align: "center", NoSort: true})
	for _, column := range matrix.columns {
		columns = append(columns, types.TableColumn{
			Key:      "tmpl-" + column.GetJobTemplateId(),
			Label:    explicitColumnName(column),
			MinWidth: "6.25rem",
			Align:    "center",
			NoSort:   true,
		})
	}

	empty := deps.Labels.SubscriptionGroup.RatingEmpty
	tableRows := make([]types.TableRow, 0, len(matrix.rows))
	bandValues := make(map[string]string)
	for _, row := range matrix.rows {
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
		cells = append(cells, rowActionsCell(subscriptionGroupID, row.clientID, deps.Routes, deps.Labels, canDownloadClientDocument, canDownloadLegacyDocument))
		for index := range matrix.columns {
			value := outcome_summary.FormatReportCell(row.cells[matrix.columns[index].GetJobTemplateId()], empty, deps.Options.ReportCellFormat())
			cells = append(cells, types.TableCell{Type: "text", Value: value, CSVValue: value})
		}
		tableRows = append(tableRows, types.TableRow{
			ID:        row.clientID,
			DataAttrs: map[string]string{"testid": "rc-row-" + short(row.clientID)},
			Cells:     cells,
		})
		bandValues[row.clientID] = row.group
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
		// The per-row download is the frozen SECOND column, not a trailing
		// actions cell — so no trailing actions column (matches page.go).
		ShowActions: false,
		// Freeze the first two columns (client + actions) while the subject
		// columns scroll horizontally — same contract as the static grid.
		TableClass: "data-table-freeze2",
		Labels:     deps.TableLabels,
		Caption:    deps.Labels.SubscriptionGroup.Title,
		EmptyState: types.TableEmptyState{
			Title:   deps.Labels.Empty.Title,
			Message: deps.Labels.SubscriptionGroup.NotComputedBanner,
		},
	}
	if _, _, grouped, _ := deps.Options.ExportRowBandConfig(); grouped {
		table.Groups = types.GroupRowsByValue(tableRows, bandValues, types.GroupRowsByValueOptions{
			LeadingOrder: deps.Options.Row.GroupValueOrder,
			GroupID:      func(value string) string { return "rc-band-" + slug(value) },
		})
	} else {
		table.Rows = tableRows
	}
	numberRows(table)
	types.ApplyColumnStyles(table.Columns, allRows(table))
	return table
}

func reportClientName(row explicitRow) string {
	lastName := strings.TrimSpace(row.lastName)
	firstName := strings.TrimSpace(row.firstName)
	if lastName != "" && firstName != "" {
		return lastName + ", " + firstName
	}
	if name := strings.TrimSpace(row.name); name != "" {
		return stripExporterSequence(name)
	}
	return strings.TrimSpace(strings.Join([]string{firstName, lastName}, " "))
}

func stripExporterSequence(name string) string {
	if !strings.HasPrefix(name, "[") {
		return name
	}
	close := strings.IndexByte(name, ']')
	if close < 2 || close+1 >= len(name) || name[close+1] != ' ' {
		return name
	}
	if _, err := strconv.Atoi(name[1:close]); err != nil {
		return name
	}
	return strings.TrimSpace(name[close+2:])
}

func reportPage(viewCtx *view.ViewContext, deps *Deps, exportContext *exportpb.SubscriptionGroupOutcomeExportContext, table *types.TableConfig, tabs []pyeza.TabItem, activeTab, category, banner string) view.ViewResult {
	l := deps.Labels
	drawerURL := ""
	if category != "" && strings.TrimSpace(deps.Routes.SubscriptionGroupDownloadDrawerURL) != "" {
		drawerURL = route.ResolveURL(deps.Routes.SubscriptionGroupDownloadDrawerURL, "id", exportContext.GetSubscriptionGroupId()) + "?mode=fixed&job_category_id=" + url.QueryEscape(category)
	}
	if table != nil && drawerURL != "" {
		table.PrimaryAction = &types.PrimaryAction{
			Label:      l.SubscriptionGroupExport.DownloadAction,
			SheetTitle: l.SubscriptionGroupExport.DrawerTitle,
			ActionURL:  drawerURL,
			TestID:     "rc-subscription-group-download-open",
		}
	}
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
		ContentTemplate:   "outcome-summary-subscription-group-content",
		Table:             table,
		NotComputed:       table == nil,
		Banner:            banner,
		TabItems:          tabs,
		ActiveTab:         activeTab,
		TabsAria:          l.SubscriptionGroup.CategoryTabsAriaLabel,
		DownloadDrawerURL: drawerURL,
		DownloadButton:    l.SubscriptionGroupExport.DownloadAction,
		DownloadTitle:     l.SubscriptionGroupExport.DrawerTitle,
	}
	return view.OK("outcome-summary-subscription-group", pageData)
}
