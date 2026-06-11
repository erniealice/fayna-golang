package job_activity

// job_activity_labels.go — JobActivity label structs + DefaultJobActivityLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobActivityLabels holds all translatable strings for the job activity module.
type Labels struct {
	Page        PageLabels       `json:"page"`
	Buttons     ButtonLabels     `json:"buttons"`
	Columns     ColumnLabels     `json:"columns"`
	Empty       EmptyLabels      `json:"empty"`
	Form        FormLabels       `json:"form"`
	Actions     ActionLabels     `json:"actions"`
	Detail      DetailLabels     `json:"detail"`
	Status      StatusLabels     `json:"status"`
	Tabs        TabLabels        `json:"tabs"`
	Errors      ErrorLabels      `json:"errors"`
	BulkActions BulkActionLabels `json:"bulkActions"`
	// Charge holds labels for the charge tab (job-activity-tab-charge):
	// labor/material/expense subtype field labels, edit CTAs, and empty states.
	// 2026-06-01 Wave 4.3 label sweep.
	Charge ChargeLabels `json:"charge"`
}

// ChargeLabels holds translatable strings for the charge tab
// (job_activity/templates/detail.html job-activity-tab-charge).
// 2026-06-01 Wave 4.3 label sweep.
type ChargeLabels struct {
	// Labor subtype
	EditLabor      string `json:"editLabor"`
	EditLaborTitle string `json:"editLaborTitle"`
	EmptyLabor     string `json:"emptyLabor"`
	// Material subtype
	EditMaterial      string `json:"editMaterial"`
	EditMaterialTitle string `json:"editMaterialTitle"`
	EmptyMaterial     string `json:"emptyMaterial"`
	// Expense subtype
	EditExpense      string `json:"editExpense"`
	EditExpenseTitle string `json:"editExpenseTitle"`
	EmptyExpense     string `json:"emptyExpense"`
	VendorRef        string `json:"vendorRef"`
	ReceiptURL       string `json:"receiptUrl"`
	PaymentMethod    string `json:"paymentMethod"`
	MarkupPct        string `json:"markupPct"`
	// Fallback for equipment/subcontract/unspecified entry types.
	Unavailable string `json:"unavailable"`
}

// TabLabels holds tab labels for the job activity detail page.
type TabLabels struct {
	Info string `json:"info"`
	// Charge is the charge tab label (shows subtype detail: labor/material/expense).
	Charge      string `json:"charge"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ButtonLabels struct {
	AddActivity string `json:"addActivity"`
}

type ColumnLabels struct {
	Date        string `json:"date"`
	Job         string `json:"job"`
	EntryType   string `json:"entryType"`
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
	Amount      string `json:"amount"`
	Status      string `json:"status"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
	Job                string `json:"job"`
	Task               string `json:"task"`
	EntryType          string `json:"entryType"`
	Description        string `json:"description"`
	BillableStatus     string `json:"billableStatus"`
	Hours              string `json:"hours"`
	HourlyRate         string `json:"hourlyRate"`
	Product            string `json:"product"`
	Quantity           string `json:"quantity"`
	UnitCost           string `json:"unitCost"`
	Amount             string `json:"amount"`
	Category           string `json:"category"`
	EntryTypeInfo      string `json:"entryTypeInfo"`
	BillableStatusInfo string `json:"billableStatusInfo"`
	QuantityInfo       string `json:"quantityInfo"`
	UnitCostInfo       string `json:"unitCostInfo"`

	// 2026-04-29 milestone-billing plan §5/§6 — operator-facing JobActivity
	// drawer fields and tab CTAs. The selectors driving phase5 specs 09/11
	// reference these labels via lyngua.
	AddCta                    string `json:"addCta"`
	BillRate                  string `json:"billRate"`
	BillAmount                string `json:"billAmount"`
	BillableStatusIncluded    string `json:"billableStatusIncluded"`
	BillableStatusBillable    string `json:"billableStatusBillable"`
	BillableStatusNonBillable string `json:"billableStatusNonBillable"`
	BillableStatusWriteOff    string `json:"billableStatusWriteOff"`
	EntryTypeLabor            string `json:"entryTypeLabor"`
	EntryTypeMaterial         string `json:"entryTypeMaterial"`
	EntryTypeExpense          string `json:"entryTypeExpense"`
	EntryTypeEquipment        string `json:"entryTypeEquipment"`
	EntryTypeSubcontract      string `json:"entryTypeSubcontract"`
	SubmitCreate              string `json:"submitCreate"`
	SubmitUpdate              string `json:"submitUpdate"`
}

type ActionLabels struct {
	View    string `json:"view"`
	Edit    string `json:"edit"`
	Delete  string `json:"delete"`
	Submit  string `json:"submit"`
	Approve string `json:"approve"`
	Reject  string `json:"reject"`
	Post    string `json:"post"`
	Reverse string `json:"reverse"`
}

type DetailLabels struct {
	PageTitle      string `json:"pageTitle"`
	TitlePrefix    string `json:"titlePrefix"`
	Job            string `json:"job"`
	EntryType      string `json:"entryType"`
	EntryDate      string `json:"entryDate"`
	Description    string `json:"description"`
	Quantity       string `json:"quantity"`
	UnitCost       string `json:"unitCost"`
	TotalCost      string `json:"totalCost"`
	Currency       string `json:"currency"`
	BillableStatus string `json:"billableStatus"`
	ApprovalStatus string `json:"approvalStatus"`
	PostingStatus  string `json:"postingStatus"`
	CreatedDate    string `json:"createdDate"`
	// Labor subtype
	Staff     string `json:"staff"`
	Hours     string `json:"hours"`
	RateType  string `json:"rateType"`
	TimeStart string `json:"timeStart"`
	TimeEnd   string `json:"timeEnd"`
	// Material subtype
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unitOfMeasure"`
	LotNumber     string `json:"lotNumber"`
	Location      string `json:"location"`
	// Expense subtype
	ExpenseCategory string `json:"expenseCategory"`
	Vendor          string `json:"vendor"`
	Receipt         string `json:"receipt"`
	Reimbursable    string `json:"reimbursable"`
}

type StatusLabels struct {
	Draft     string `json:"draft"`
	Submitted string `json:"submitted"`
	Approved  string `json:"approved"`
	Rejected  string `json:"rejected"`
}

type ErrorLabels struct {
	PermissionDenied     string `json:"permissionDenied"`
	NotFound             string `json:"notFound"`
	IDRequired           string `json:"idRequired"`
	InUse                string `json:"inUse"`
	InvalidForm          string `json:"invalidForm"`
	NoActivitiesSelected string `json:"noActivitiesSelected"`
	InvoiceGenFailed     string `json:"invoiceGenFailed"`
}

// BulkActionLabels holds translatable strings for job activity bulk-action controls.
type BulkActionLabels struct {
	Delete                 string `json:"delete"`
	BulkDelete             string `json:"bulkDelete"`
	BulkDeleteConfirmTitle string `json:"bulkDeleteConfirmTitle"`
	BulkDeleteConfirmMsg   string `json:"bulkDeleteConfirmMsg"`
	GenerateInvoice        string `json:"generateInvoice"`
	GenerateInvoiceConfirm string `json:"generateInvoiceConfirm"`
	SelectAll              string `json:"selectAll"`
	SelectedCount          string `json:"selectedCount"`
	Cancel                 string `json:"cancel"`
}

// DefaultJobActivityLabels returns JobActivityLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Activities",
			Caption: "Cross-job timesheet and activity log",
		},
		Buttons: ButtonLabels{
			AddActivity: "Add Activity",
		},
		Columns: ColumnLabels{
			Date:        "Date",
			Job:         "Job",
			EntryType:   "Type",
			Description: "Description",
			Quantity:    "Hrs/Qty",
			Amount:      "Amount",
			Status:      "Status",
		},
		Empty: EmptyLabels{
			Title:   "No activities found",
			Message: "No activity entries to display.",
		},
		Form: FormLabels{
			Job:                "Job",
			Task:               "Task",
			EntryType:          "Entry Type",
			Description:        "Description",
			BillableStatus:     "Billable Status",
			Hours:              "Hours",
			HourlyRate:         "Hourly Rate",
			Product:            "Product",
			Quantity:           "Quantity",
			UnitCost:           "Unit Cost",
			Amount:             "Amount",
			Category:           "Category",
			EntryTypeInfo:      "Labor = time-based; Material = goods used; Expense = cost incurred.",
			BillableStatusInfo: "Whether this activity is charged to the client.",
			QuantityInfo:       "Number of units or hours recorded for this activity entry.",
			UnitCostInfo:       "Cost per unit or per hour for this activity entry.",
			// 2026-04-29 milestone-billing plan §5/§6.
			AddCta:                    "+ Add Activity",
			BillRate:                  "Bill Rate",
			BillAmount:                "Bill Amount",
			BillableStatusIncluded:    "Included",
			BillableStatusBillable:    "Billable (T&M)",
			BillableStatusNonBillable: "Non-billable",
			BillableStatusWriteOff:    "Write-off",
			EntryTypeLabor:            "Labor",
			EntryTypeMaterial:         "Material",
			EntryTypeExpense:          "Expense",
			EntryTypeEquipment:        "Equipment",
			EntryTypeSubcontract:      "Subcontract",
			SubmitCreate:              "Save",
			SubmitUpdate:              "Update",
		},
		Actions: ActionLabels{
			View:    "View Activity",
			Edit:    "Edit Activity",
			Delete:  "Delete Activity",
			Submit:  "Submit for Approval",
			Approve: "Approve",
			Reject:  "Reject",
			Post:    "Post",
			Reverse: "Reverse",
		},
		Detail: DetailLabels{
			PageTitle:       "Activity Details",
			TitlePrefix:     "Activity ",
			Job:             "Job",
			EntryType:       "Entry Type",
			EntryDate:       "Date",
			Description:     "Description",
			Quantity:        "Quantity",
			UnitCost:        "Unit Cost",
			TotalCost:       "Total Cost",
			Currency:        "Currency",
			BillableStatus:  "Billable Status",
			ApprovalStatus:  "Approval Status",
			PostingStatus:   "Posting Status",
			CreatedDate:     "Created",
			Staff:           "Staff",
			Hours:           "Hours",
			RateType:        "Rate Type",
			TimeStart:       "Start Time",
			TimeEnd:         "End Time",
			Product:         "Product",
			UnitOfMeasure:   "Unit of Measure",
			LotNumber:       "Lot Number",
			Location:        "Location",
			ExpenseCategory: "Expense Category",
			Vendor:          "Vendor",
			Receipt:         "Receipt",
			Reimbursable:    "Reimbursable",
		},
		Status: StatusLabels{
			Draft:     "Draft",
			Submitted: "Submitted",
			Approved:  "Approved",
			Rejected:  "Rejected",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Charge:      "Charge",
			Attachments: "Attachments",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Activity not found",
			IDRequired:       "Activity ID is required",
		},
		// 2026-06-01 Wave 4.3 label sweep — charge tab subtype detail.
		Charge: ChargeLabels{
			EditLabor:         "Edit Labor Charge",
			EditLaborTitle:    "Edit labor charge",
			EmptyLabor:        "No labor charge recorded.",
			EditMaterial:      "Edit Material Charge",
			EditMaterialTitle: "Edit material charge",
			EmptyMaterial:     "No material charge recorded.",
			EditExpense:       "Edit Expense Charge",
			EditExpenseTitle:  "Edit expense charge",
			EmptyExpense:      "No expense charge recorded.",
			VendorRef:         "Vendor Ref",
			ReceiptURL:        "Receipt URL",
			PaymentMethod:     "Payment Method",
			MarkupPct:         "Markup %",
			Unavailable:       "Charge detail not available for this entry type.",
		},
	}
}
