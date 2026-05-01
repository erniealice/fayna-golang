package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// PageData holds the data for the fulfillment detail page.
type PageData struct {
	types.PageData
	ContentTemplate  string
	Fulfillment      map[string]any
	Labels           fayna.FulfillmentLabels
	ActiveTab        string
	TabItems         []pyeza.TabItem
	ItemsTable       *types.TableConfig
	HistoryTable     *types.TableConfig
	ReturnsTable     *types.TableConfig
	AllowedEvents    []string
	SupplierName     string
	RevenueReference string
	TransitionURL    string
}

// fulfillmentToMap converts a Fulfillment protobuf to map[string]any for template use.
func fulfillmentToMap(f *fulfillmentpb.Fulfillment) map[string]any {
	return map[string]any{
		"id":                 f.GetId(),
		"workspace_id":       f.GetWorkspaceId(),
		"revenue_id":         f.GetRevenueId(),
		"supplier_id":        f.GetSupplierId(),
		"delivery_mode": f.GetDeliveryMode(),
		"status":             f.GetStatus(),
		"status_variant":     fulfillmentStatusVariant(f.GetStatus()),
		"provider_status":    f.GetProviderStatus(),
		"provider_reference": f.GetProviderReference(),
		"delivery_cost":      types.MoneyCell(float64(f.GetDeliveryCost()), f.GetCurrency(), true),
		"currency":           f.GetCurrency(),
		"notes":              f.GetNotes(),
		"active":             f.GetActive(),
		"created_by":         f.GetCreatedBy(),
	}
}

func fulfillmentStatusVariant(status string) string {
	switch status {
	case "PENDING":
		return "warning"
	case "READY":
		return "info"
	case "IN_TRANSIT":
		return "info"
	case "DELIVERED":
		return "success"
	case "PARTIALLY_DELIVERED":
		return "warning"
	case "FAILED":
		return "danger"
	case "CANCELLED":
		return "default"
	default:
		return "default"
	}
}

// NewView creates the fulfillment detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.GetFulfillmentItemPageData(ctx, &fulfillmentpb.GetFulfillmentItemPageDataRequest{
			Id: id,
		})
		if err != nil {
			log.Printf("Failed to read fulfillment %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load fulfillment: %w", err))
		}

		f := resp.GetFulfillment()
		if f == nil {
			log.Printf("Fulfillment %s not found", id)
			return view.Error(fmt.Errorf("fulfillment not found"))
		}

		fulfillmentMap := fulfillmentToMap(f)
		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          fmt.Sprintf("Fulfillment %s", id),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "fulfillment",
				HeaderTitle:    fmt.Sprintf("Fulfillment %s", id),
				HeaderSubtitle: l.PageTitle,
				HeaderIcon:     "icon-package",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate:  "fulfillment-detail-content",
			Fulfillment:      fulfillmentMap,
			Labels:           l,
			ActiveTab:        activeTab,
			TabItems:         tabItems,
			AllowedEvents:    resp.GetAllowedEvents(),
			SupplierName:     resp.GetSupplierName(),
			RevenueReference: resp.GetRevenueReference(),
			TransitionURL:    route.ResolveURL(deps.Routes.TransitionURL, "id", id),
		}

		// Load tab-specific data
		loadTabData(ctx, deps, pageData, resp, activeTab)

		// KB help content
		if viewCtx.Translations != nil {
			if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "fulfillment-detail"); kb != nil {
					pageData.HasHelp = true
					pageData.HelpContent = kb.Body
				}
			}
		}

		return view.OK("fulfillment-detail", pageData)
	})
}

func buildTabItems(l fayna.FulfillmentLabels, id string, routes fayna.FulfillmentRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", Icon: "icon-info"},
		{Key: "items", Label: l.Tabs.Items, Href: base + "?tab=items", Icon: "icon-list"},
		{Key: "history", Label: l.Tabs.History, Href: base + "?tab=history", Icon: "icon-clock"},
		{Key: "returns", Label: l.Tabs.Returns, Href: base + "?tab=returns", Icon: "icon-refresh-ccw"},
	}
}

func loadTabData(
	_ context.Context,
	deps *DetailViewDeps,
	pageData *PageData,
	resp *fulfillmentpb.GetFulfillmentItemPageDataResponse,
	activeTab string,
) {
	l := deps.Labels
	switch activeTab {
	case "info":
		// Fulfillment map has all info fields.
	case "items":
		pageData.ItemsTable = buildItemsTable(resp.GetItems(), l, deps.TableLabels)
	case "history":
		pageData.HistoryTable = buildHistoryTable(resp.GetStatusEvents(), l, deps.TableLabels)
	case "returns":
		pageData.ReturnsTable = buildReturnsTable(resp.GetReturns(), l, deps.TableLabels)
	}
}

func buildItemsTable(items []*fulfillmentpb.FulfillmentItem, l fayna.FulfillmentLabels, tableLabels types.TableLabels) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "product_id", Label: "Product", NoSort: true},
		{Key: "delivery_mode", Label: l.Columns.DeliveryMode, NoSort: true},
		{Key: "quantity_ordered", Label: "Qty Ordered", NoSort: true, WidthClass: "col-2xl"},
		{Key: "quantity_delivered", Label: "Qty Delivered", NoSort: true, WidthClass: "col-3xl"},
		{Key: "status", Label: l.Columns.Status, NoSort: true, WidthClass: "col-2xl"},
	}

	rows := make([]types.TableRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, types.TableRow{
			ID: item.GetId(),
			Cells: []types.TableCell{
				{Type: "text", Value: item.GetProductId()},
				{Type: "text", Value: item.GetDeliveryMode()},
				{Type: "text", Value: fmt.Sprintf("%.2f", item.GetQuantityOrdered())},
				{Type: "text", Value: fmt.Sprintf("%.2f", item.GetQuantityDelivered())},
				{Type: "badge", Value: item.GetStatus(), Variant: "default"},
			},
		})
	}

	return &types.TableConfig{
		ID:      "fulfillment-items-table",
		Columns: columns,
		Rows:    rows,
		Labels:  tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No items",
			Message: "No fulfillment items found.",
		},
	}
}

func buildHistoryTable(events []*fulfillmentpb.FulfillmentStatusEvent, l fayna.FulfillmentLabels, tableLabels types.TableLabels) *types.TableConfig {
	_ = l // labels reserved for future use
	columns := []types.TableColumn{
		{Key: "from_status", Label: "From Status", NoSort: true},
		{Key: "to_status", Label: "To Status", NoSort: true},
		{Key: "reason", Label: "Reason", NoSort: true},
		{Key: "occurred_at", Label: "Date", NoSort: true, WidthClass: "col-4xl"},
	}

	rows := make([]types.TableRow, 0, len(events))
	for _, ev := range events {
		occurredAt := ""
		if ts := ev.GetOccurredAt(); ts != nil {
			occurredAt = ts.AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		rows = append(rows, types.TableRow{
			ID: fmt.Sprintf("%d", ev.GetId()),
			Cells: []types.TableCell{
				{Type: "badge", Value: ev.GetFromStatus(), Variant: fulfillmentStatusVariant(ev.GetFromStatus())},
				{Type: "badge", Value: ev.GetToStatus(), Variant: fulfillmentStatusVariant(ev.GetToStatus())},
				{Type: "text", Value: ev.GetReason()},
				types.DateTimeCell(occurredAt, types.DateReadable),
			},
		})
	}

	return &types.TableConfig{
		ID:      "fulfillment-history-table",
		Columns: columns,
		Rows:    rows,
		Labels:  tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No history",
			Message: "No status events recorded yet.",
		},
	}
}

func buildReturnsTable(returns []*fulfillmentpb.FulfillmentReturn, l fayna.FulfillmentLabels, tableLabels types.TableLabels) *types.TableConfig {
	_ = l // labels reserved for future use
	columns := []types.TableColumn{
		{Key: "status", Label: "Status", NoSort: true, WidthClass: "col-2xl"},
		{Key: "reason", Label: "Reason", NoSort: true},
		{Key: "notes", Label: "Notes", NoSort: true},
		{Key: "date_created", Label: "Date", NoSort: true, WidthClass: "col-4xl"},
	}

	rows := make([]types.TableRow, 0, len(returns))
	for _, ret := range returns {
		dateCreated := ""
		if dc := ret.GetDateCreated(); dc != 0 {
			dateCreated = fmt.Sprintf("%d", dc)
		}
		rows = append(rows, types.TableRow{
			ID: ret.GetId(),
			Cells: []types.TableCell{
				{Type: "badge", Value: ret.GetStatus(), Variant: "default"},
				{Type: "text", Value: ret.GetReason()},
				{Type: "text", Value: ret.GetNotes()},
				types.DateTimeCell(dateCreated, types.DateReadable),
			},
		})
	}

	return &types.TableConfig{
		ID:      "fulfillment-returns-table",
		Columns: columns,
		Rows:    rows,
		Labels:  tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No returns",
			Message: "No return requests recorded yet.",
		},
	}
}
