package fulfillment

// Labels holds all labels for fulfillment views.
type Labels struct {
	PageTitle string `json:"page_title"`
	AppLabel  string `json:"app_label"`
	Title     string `json:"title"`

	Status  StatusLabels       `json:"status"`
	Type    DeliveryModeLabels `json:"type"`
	Columns ColumnLabels       `json:"columns"`
	Tabs    TabLabels          `json:"tabs"`
	Actions ActionLabels       `json:"actions"`
	Buttons ButtonLabels       `json:"buttons"`
	Empty   EmptyLabels        `json:"empty"`
	Errors  ErrorLabels        `json:"errors"`
	// Dashboard labels for the Fulfillment live dashboard.
	// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	Dashboard DashboardLabels `json:"dashboard"`
}

// DashboardLabels holds translatable strings for the Fulfillment
// live dashboard. (Phase 3 — Pyeza dashboard block + per-app live dashboards
// plan).
type DashboardLabels struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// Stats
	StatPending        string `json:"stat_pending"`
	StatInTransit      string `json:"stat_in_transit"`
	StatDeliveredToday string `json:"stat_delivered_today"`
	StatExceptions     string `json:"stat_exceptions"`
	// Widgets
	DailyDelivered   string `json:"daily_delivered"`
	StatusMix        string `json:"status_mix"`
	RecentExceptions string `json:"recent_exceptions"`
	NoExceptions     string `json:"no_exceptions"`
	// Quick actions
	QuickNewFulfillment string `json:"quick_new_fulfillment"`
	QuickPickPack       string `json:"quick_pick_pack"`
	QuickProcessReturn  string `json:"quick_process_return"`
	QuickMarkDelivered  string `json:"quick_mark_delivered"`
	// Common
	ViewAll            string `json:"view_all"`
	AxisCount          string `json:"axis_count"`
	AvgFulfillmentDays string `json:"avg_fulfillment_days"`
}

type StatusLabels struct {
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

type ColumnLabels struct {
	DeliveryMode string `json:"delivery_mode"`
	Status       string `json:"status"`
	SupplierName string `json:"supplier_name"`
	ScheduledAt  string `json:"scheduled_at"`
	ItemCount    string `json:"item_count"`
	Notes        string `json:"notes"`
}

type TabLabels struct {
	Info        string `json:"info"`
	Items       string `json:"items"`
	History     string `json:"history"`
	Returns     string `json:"returns"`
	Attachments string `json:"attachments"`
}

type ActionLabels struct {
	MarkReady      string `json:"mark_ready"`
	Dispatch       string `json:"dispatch"`
	Deliver        string `json:"deliver"`
	DeliverPartial string `json:"deliver_partial"`
	MarkFailed     string `json:"mark_failed"`
	Cancel         string `json:"cancel"`
	Retry          string `json:"retry"`
}

type ButtonLabels struct {
	AddFulfillment string `json:"add_fulfillment"`
	Edit           string `json:"edit"`
	Delete         string `json:"delete"`
	Transition     string `json:"transition"`
	Return         string `json:"return"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	LoadFailed       string `json:"load_failed"`
	TransitionFailed string `json:"transition_failed"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		PageTitle: "Fulfillment",
		AppLabel:  "Fulfillment",
		Title:     "Fulfillments",
		Status: StatusLabels{
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
		Columns: ColumnLabels{
			DeliveryMode: "Method",
			Status:       "Status",
			SupplierName: "Supplier",
			ScheduledAt:  "Scheduled",
			ItemCount:    "Items",
			Notes:        "Notes",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Items:       "Items",
			History:     "History",
			Returns:     "Returns",
			Attachments: "Attachments",
		},
		Actions: ActionLabels{
			MarkReady:      "Mark Ready",
			Dispatch:       "Dispatch",
			Deliver:        "Deliver",
			DeliverPartial: "Partial Delivery",
			MarkFailed:     "Mark Failed",
			Cancel:         "Cancel",
			Retry:          "Retry",
		},
		Buttons: ButtonLabels{
			AddFulfillment: "Add Fulfillment",
			Edit:           "Edit",
			Delete:         "Delete",
			Transition:     "Update Status",
			Return:         "Create Return",
		},
		Empty: EmptyLabels{
			Title:   "No fulfillments found",
			Message: "No fulfillments to display.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			LoadFailed:       "Failed to load fulfillment data",
			TransitionFailed: "Failed to update fulfillment status",
		},
	}
}
