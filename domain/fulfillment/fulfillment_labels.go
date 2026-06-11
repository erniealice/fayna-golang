package fulfillment

// FulfillmentLabels holds all labels for fulfillment views.
type FulfillmentLabels struct {
	PageTitle string `json:"page_title"`
	AppLabel  string `json:"app_label"`
	Title     string `json:"title"`

	Status  FulfillmentStatusLabels `json:"status"`
	Type    DeliveryModeLabels      `json:"type"`
	Columns FulfillmentColumnLabels `json:"columns"`
	Tabs    FulfillmentTabLabels    `json:"tabs"`
	Actions FulfillmentActionLabels `json:"actions"`
	Buttons FulfillmentButtonLabels `json:"buttons"`
	Empty   FulfillmentEmptyLabels  `json:"empty"`
	Errors  FulfillmentErrorLabels  `json:"errors"`
	// Dashboard labels for the Fulfillment live dashboard.
	// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	Dashboard FulfillmentDashboardLabels `json:"dashboard"`
}

// FulfillmentDashboardLabels holds translatable strings for the Fulfillment
// live dashboard. (Phase 3 — Pyeza dashboard block + per-app live dashboards
// plan).
type FulfillmentDashboardLabels struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// Stats
	StatPending        string `json:"statPending"`
	StatInTransit      string `json:"statInTransit"`
	StatDeliveredToday string `json:"statDeliveredToday"`
	StatExceptions     string `json:"statExceptions"`
	// Widgets
	DailyDelivered   string `json:"dailyDelivered"`
	StatusMix        string `json:"statusMix"`
	RecentExceptions string `json:"recentExceptions"`
	NoExceptions     string `json:"noExceptions"`
	// Quick actions
	QuickNewFulfillment string `json:"quickNewFulfillment"`
	QuickPickPack       string `json:"quickPickPack"`
	QuickProcessReturn  string `json:"quickProcessReturn"`
	QuickMarkDelivered  string `json:"quickMarkDelivered"`
	// Common
	ViewAll            string `json:"viewAll"`
	AxisCount          string `json:"axisCount"`
	AvgFulfillmentDays string `json:"avgFulfillmentDays"`
}

type FulfillmentStatusLabels struct {
	Pending            string `json:"pending"`
	Ready              string `json:"ready"`
	InTransit          string `json:"in_transit"`
	Delivered          string `json:"delivered"`
	PartiallyDelivered string `json:"partially_delivered"`
	Failed             string `json:"failed"`
	Cancelled          string `json:"cancelled"`
}

type DeliveryModeLabels struct {
	Instant      string `json:"instant"`
	Scheduled    string `json:"scheduled"`
	Shipped      string `json:"shipped"`
	Digital      string `json:"digital"`
	Project      string `json:"project"`
	Subscription string `json:"subscription"`
}

type FulfillmentColumnLabels struct {
	DeliveryMode string `json:"delivery_mode"`
	Status       string `json:"status"`
	SupplierName string `json:"supplier_name"`
	ScheduledAt  string `json:"scheduled_at"`
	ItemCount    string `json:"item_count"`
	Notes        string `json:"notes"`
}

type FulfillmentTabLabels struct {
	Info        string `json:"info"`
	Items       string `json:"items"`
	History     string `json:"history"`
	Returns     string `json:"returns"`
	Attachments string `json:"attachments"`
}

type FulfillmentActionLabels struct {
	MarkReady      string `json:"mark_ready"`
	Dispatch       string `json:"dispatch"`
	Deliver        string `json:"deliver"`
	DeliverPartial string `json:"deliver_partial"`
	MarkFailed     string `json:"mark_failed"`
	Cancel         string `json:"cancel"`
	Retry          string `json:"retry"`
}

type FulfillmentButtonLabels struct {
	AddFulfillment string `json:"add_fulfillment"`
	Edit           string `json:"edit"`
	Delete         string `json:"delete"`
	Transition     string `json:"transition"`
	Return         string `json:"return"`
}

type FulfillmentEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FulfillmentErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	LoadFailed       string `json:"load_failed"`
	TransitionFailed string `json:"transition_failed"`
}

// DefaultFulfillmentLabels returns FulfillmentLabels with sensible English defaults.
func DefaultFulfillmentLabels() FulfillmentLabels {
	return FulfillmentLabels{
		PageTitle: "Fulfillment",
		AppLabel:  "Fulfillment",
		Title:     "Fulfillments",
		Status: FulfillmentStatusLabels{
			Pending:            "Pending",
			Ready:              "Ready",
			InTransit:          "In Transit",
			Delivered:          "Delivered",
			PartiallyDelivered: "Partially Delivered",
			Failed:             "Failed",
			Cancelled:          "Cancelled",
		},
		Type: DeliveryModeLabels{
			Instant:      "Instant",
			Scheduled:    "Scheduled",
			Shipped:      "Shipped",
			Digital:      "Digital",
			Project:      "Project",
			Subscription: "Subscription",
		},
		Columns: FulfillmentColumnLabels{
			DeliveryMode: "Method",
			Status:       "Status",
			SupplierName: "Supplier",
			ScheduledAt:  "Scheduled",
			ItemCount:    "Items",
			Notes:        "Notes",
		},
		Tabs: FulfillmentTabLabels{
			Info:        "Information",
			Items:       "Items",
			History:     "History",
			Returns:     "Returns",
			Attachments: "Attachments",
		},
		Actions: FulfillmentActionLabels{
			MarkReady:      "Mark Ready",
			Dispatch:       "Dispatch",
			Deliver:        "Deliver",
			DeliverPartial: "Partial Delivery",
			MarkFailed:     "Mark Failed",
			Cancel:         "Cancel",
			Retry:          "Retry",
		},
		Buttons: FulfillmentButtonLabels{
			AddFulfillment: "Add Fulfillment",
			Edit:           "Edit",
			Delete:         "Delete",
			Transition:     "Update Status",
			Return:         "Create Return",
		},
		Empty: FulfillmentEmptyLabels{
			Title:   "No fulfillments found",
			Message: "No fulfillments to display.",
		},
		Errors: FulfillmentErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			LoadFailed:       "Failed to load fulfillment data",
			TransitionFailed: "Failed to update fulfillment status",
		},
	}
}
