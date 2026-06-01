// Package fayna provides translatable label structs for the operations/work domain.
//
// This file contains all label types for job templates, job execution, and job activities.
// Labels are loaded from lyngua translation files and injected into views at startup.
package fayna

// ---------------------------------------------------------------------------
// Job Template labels
// ---------------------------------------------------------------------------

// JobTemplateLabels holds all translatable strings for the job template module.
type JobTemplateLabels struct {
	Page        JobTemplatePageLabels       `json:"page"`
	Buttons     JobTemplateButtonLabels     `json:"buttons"`
	Columns     JobTemplateColumnLabels     `json:"columns"`
	Empty       JobTemplateEmptyLabels      `json:"empty"`
	Form        JobTemplateFormLabels       `json:"form"`
	Actions     JobTemplateActionLabels     `json:"actions"`
	Detail      JobTemplateDetailLabels     `json:"detail"`
	Tabs        JobTemplateTabLabels        `json:"tabs"`
	Confirm     JobTemplateConfirmLabels    `json:"confirm"`
	Errors      JobTemplateErrorLabels      `json:"errors"`
	BulkActions JobTemplateBulkActionLabels `json:"bulkActions"`
}

type JobTemplatePageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type JobTemplateButtonLabels struct {
	AddJobTemplate string `json:"addJobTemplate"`
}

type JobTemplateColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type JobTemplateEmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type JobTemplateActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
	// Add is the CTA label on the Phases tab.
	Add string `json:"add"`
	// AddTask is the CTA label on the Tasks tab.
	AddTask string `json:"addTask"`
}

type JobTemplateErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
	InUse            string `json:"inUse"`
	InvalidForm      string `json:"invalidForm"`
}

type JobTemplateFormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"descriptionPlaceholder"`
	Active          string `json:"active"`
}

type JobTemplateDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
}

type JobTemplateTabLabels struct {
	Info        string `json:"info"`
	Phases      string `json:"phases"`
	Tasks       string `json:"tasks"`
	Standards   string `json:"standards"`
	Attachments string `json:"attachments"`
	AuditTrail  string `json:"auditTrail"`
	History     string `json:"history"`
}

type JobTemplateConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

// JobTemplateBulkActionLabels holds translatable strings for job template bulk-action controls.
type JobTemplateBulkActionLabels struct {
	Delete                 string `json:"delete"`
	BulkDelete             string `json:"bulkDelete"`
	BulkDeleteConfirmTitle string `json:"bulkDeleteConfirmTitle"`
	BulkDeleteConfirmMsg   string `json:"bulkDeleteConfirmMsg"`
	SelectAll              string `json:"selectAll"`
	SelectedCount          string `json:"selectedCount"`
	Cancel                 string `json:"cancel"`
}

// DefaultJobTemplateLabels returns JobTemplateLabels with sensible English defaults.
func DefaultJobTemplateLabels() JobTemplateLabels {
	return JobTemplateLabels{
		Page: JobTemplatePageLabels{
			Heading:         "Job Templates",
			HeadingActive:   "Active Job Templates",
			HeadingInactive: "Inactive Job Templates",
			Caption:         "Manage reusable execution plans",
			CaptionActive:   "Manage your active job templates",
			CaptionInactive: "View inactive or archived job templates",
		},
		Buttons: JobTemplateButtonLabels{
			AddJobTemplate: "Add Template",
		},
		Columns: JobTemplateColumnLabels{
			Name:        "Name",
			Description: "Description",
			Status:      "Status",
		},
		Empty: JobTemplateEmptyLabels{
			Title:           "No job templates found",
			Message:         "No job templates to display.",
			ActiveTitle:     "No active job templates",
			ActiveMessage:   "Create your first job template to get started.",
			InactiveTitle:   "No inactive job templates",
			InactiveMessage: "Deactivated templates will appear here.",
		},
		Form: JobTemplateFormLabels{
			Name:            "Template Name",
			NamePlaceholder: "Enter template name",
			Description:     "Description",
			DescPlaceholder: "Enter template description...",
			Active:          "Active",
		},
		Actions: JobTemplateActionLabels{
			View:    "View Template",
			Edit:    "Edit Template",
			Delete:  "Delete Template",
			Add:     "+ Add Phase",
			AddTask: "+ Add Task",
		},
		Detail: JobTemplateDetailLabels{
			PageTitle:    "Job Template Details",
			Name:         "Name",
			Description:  "Description",
			Status:       "Status",
			CreatedDate:  "Created",
			ModifiedDate: "Last Modified",
		},
		Tabs: JobTemplateTabLabels{
			Info:        "Information",
			Phases:      "Phases",
			Tasks:       "Tasks",
			Standards:   "Standards",
			Attachments: "Attachments",
			AuditTrail:  "Audit Trail",
			History:     "History",
		},
		Confirm: JobTemplateConfirmLabels{
			Delete:        "Delete Template",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobTemplateErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Job template not found",
			IDRequired:       "Job template ID is required",
			NoPermission:     "No permission",
		},
	}
}

// ---------------------------------------------------------------------------
// Job Activity labels (cross-job timesheet / activity log)
// ---------------------------------------------------------------------------

// JobActivityLabels holds all translatable strings for the job activity module.
type JobActivityLabels struct {
	Page        JobActivityPageLabels       `json:"page"`
	Buttons     JobActivityButtonLabels     `json:"buttons"`
	Columns     JobActivityColumnLabels     `json:"columns"`
	Empty       JobActivityEmptyLabels      `json:"empty"`
	Form        JobActivityFormLabels       `json:"form"`
	Actions     JobActivityActionLabels     `json:"actions"`
	Detail      JobActivityDetailLabels     `json:"detail"`
	Status      JobActivityStatusLabels     `json:"status"`
	Tabs        JobActivityTabLabels        `json:"tabs"`
	Errors      JobActivityErrorLabels      `json:"errors"`
	BulkActions JobActivityBulkActionLabels `json:"bulkActions"`
	// Charge holds labels for the charge tab (job-activity-tab-charge):
	// labor/material/expense subtype field labels, edit CTAs, and empty states.
	// 2026-06-01 Wave 4.3 label sweep.
	Charge JobActivityChargeLabels `json:"charge"`
}

// JobActivityChargeLabels holds translatable strings for the charge tab
// (job_activity/templates/detail.html job-activity-tab-charge).
// 2026-06-01 Wave 4.3 label sweep.
type JobActivityChargeLabels struct {
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

// JobActivityTabLabels holds tab labels for the job activity detail page.
type JobActivityTabLabels struct {
	Info string `json:"info"`
	// Charge is the charge tab label (shows subtype detail: labor/material/expense).
	Charge      string `json:"charge"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type JobActivityPageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type JobActivityButtonLabels struct {
	AddActivity string `json:"addActivity"`
}

type JobActivityColumnLabels struct {
	Date        string `json:"date"`
	Job         string `json:"job"`
	EntryType   string `json:"entryType"`
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
	Amount      string `json:"amount"`
	Status      string `json:"status"`
}

type JobActivityEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type JobActivityFormLabels struct {
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

type JobActivityActionLabels struct {
	View    string `json:"view"`
	Edit    string `json:"edit"`
	Delete  string `json:"delete"`
	Submit  string `json:"submit"`
	Approve string `json:"approve"`
	Reject  string `json:"reject"`
	Post    string `json:"post"`
	Reverse string `json:"reverse"`
}

type JobActivityDetailLabels struct {
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

type JobActivityStatusLabels struct {
	Draft     string `json:"draft"`
	Submitted string `json:"submitted"`
	Approved  string `json:"approved"`
	Rejected  string `json:"rejected"`
}

type JobActivityErrorLabels struct {
	PermissionDenied     string `json:"permissionDenied"`
	NotFound             string `json:"notFound"`
	IDRequired           string `json:"idRequired"`
	InUse                string `json:"inUse"`
	InvalidForm          string `json:"invalidForm"`
	NoActivitiesSelected string `json:"noActivitiesSelected"`
	InvoiceGenFailed     string `json:"invoiceGenFailed"`
}

// JobActivityBulkActionLabels holds translatable strings for job activity bulk-action controls.
type JobActivityBulkActionLabels struct {
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

// ---------------------------------------------------------------------------
// Activity Labor labels
// ---------------------------------------------------------------------------

// ActivityLaborLabels holds all translatable strings for the activity labor module.
// ActivityLabor is the charge detail for ENTRY_TYPE_LABOR job activities.
// TODO(P7 lyngua sweep): add lyngua JSON files for these labels.
type ActivityLaborLabels struct {
	Page    ActivityLaborPageLabels   `json:"page"`
	Buttons ActivityLaborButtonLabels `json:"buttons"`
	Columns ActivityLaborColumnLabels `json:"columns"`
	Empty   ActivityLaborEmptyLabels  `json:"empty"`
	Form    ActivityLaborFormLabels   `json:"form"`
	Detail  ActivityLaborDetailLabels `json:"detail"`
	Errors  ActivityLaborErrorLabels  `json:"errors"`
}

type ActivityLaborPageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ActivityLaborButtonLabels struct {
	Add  string `json:"add"`
	Edit string `json:"edit"`
}

type ActivityLaborColumnLabels struct {
	Staff     string `json:"staff"`
	Hours     string `json:"hours"`
	RateType  string `json:"rateType"`
	TimeStart string `json:"timeStart"`
	TimeEnd   string `json:"timeEnd"`
}

type ActivityLaborEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ActivityLaborFormLabels struct {
	SectionCharge    string `json:"sectionCharge"`
	Staff            string `json:"staff"`
	Hours            string `json:"hours"`
	RateType         string `json:"rateType"`
	TimeStart        string `json:"timeStart"`
	TimeEnd          string `json:"timeEnd"`
	RateTypeStandard string `json:"rateTypeStandard"`
	RateTypeOvertime string `json:"rateTypeOvertime"`
	RateTypePremium  string `json:"rateTypePremium"`
	SubmitCreate     string `json:"submitCreate"`
	SubmitUpdate     string `json:"submitUpdate"`
}

type ActivityLaborDetailLabels struct {
	PageTitle   string `json:"pageTitle"`
	TitlePrefix string `json:"titlePrefix"`
	Staff       string `json:"staff"`
	Hours       string `json:"hours"`
	RateType    string `json:"rateType"`
	TimeStart   string `json:"timeStart"`
	TimeEnd     string `json:"timeEnd"`
}

type ActivityLaborErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// ---------------------------------------------------------------------------
// Activity Material labels
// ---------------------------------------------------------------------------

// ActivityMaterialLabels holds all translatable strings for the activity material module.
// ActivityMaterial is the charge detail for ENTRY_TYPE_MATERIAL job activities.
// TODO(P7 lyngua sweep): add lyngua JSON files for these labels.
type ActivityMaterialLabels struct {
	Page    ActivityMaterialPageLabels   `json:"page"`
	Buttons ActivityMaterialButtonLabels `json:"buttons"`
	Columns ActivityMaterialColumnLabels `json:"columns"`
	Empty   ActivityMaterialEmptyLabels  `json:"empty"`
	Form    ActivityMaterialFormLabels   `json:"form"`
	Detail  ActivityMaterialDetailLabels `json:"detail"`
	Errors  ActivityMaterialErrorLabels  `json:"errors"`
}

type ActivityMaterialPageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ActivityMaterialButtonLabels struct {
	Add  string `json:"add"`
	Edit string `json:"edit"`
}

type ActivityMaterialColumnLabels struct {
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unitOfMeasure"`
	LotNumber     string `json:"lotNumber"`
	Location      string `json:"location"`
}

type ActivityMaterialEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ActivityMaterialFormLabels struct {
	SectionMaterial string `json:"sectionMaterial"`
	Product         string `json:"product"`
	UnitOfMeasure   string `json:"unitOfMeasure"`
	LotNumber       string `json:"lotNumber"`
	Location        string `json:"location"`
	SubmitCreate    string `json:"submitCreate"`
	SubmitUpdate    string `json:"submitUpdate"`
}

type ActivityMaterialDetailLabels struct {
	PageTitle     string `json:"pageTitle"`
	TitlePrefix   string `json:"titlePrefix"`
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unitOfMeasure"`
	LotNumber     string `json:"lotNumber"`
	Location      string `json:"location"`
}

type ActivityMaterialErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultActivityMaterialLabels returns ActivityMaterialLabels with sensible English defaults.
func DefaultActivityMaterialLabels() ActivityMaterialLabels {
	return ActivityMaterialLabels{
		Page: ActivityMaterialPageLabels{
			Heading: "Material Charges",
			Caption: "Material usage entries per activity",
		},
		Buttons: ActivityMaterialButtonLabels{
			Add:  "Add Material",
			Edit: "Edit Material",
		},
		Columns: ActivityMaterialColumnLabels{
			Product:       "Product",
			UnitOfMeasure: "Unit",
			LotNumber:     "Lot #",
			Location:      "Location",
		},
		Empty: ActivityMaterialEmptyLabels{
			Title:   "No material entries",
			Message: "No material charge recorded for this activity.",
		},
		Form: ActivityMaterialFormLabels{
			SectionMaterial: "Material",
			Product:         "Product",
			UnitOfMeasure:   "Unit of Measure",
			LotNumber:       "Lot Number",
			Location:        "Location",
			SubmitCreate:    "Save",
			SubmitUpdate:    "Update",
		},
		Detail: ActivityMaterialDetailLabels{
			PageTitle:     "Material Charge",
			TitlePrefix:   "Material: ",
			Product:       "Product",
			UnitOfMeasure: "Unit of Measure",
			LotNumber:     "Lot Number",
			Location:      "Location",
		},
		Errors: ActivityMaterialErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Material charge record not found",
			IDRequired:       "Activity ID is required",
		},
	}
}

// ---------------------------------------------------------------------------
// Activity Expense labels
// ---------------------------------------------------------------------------

// ActivityExpenseLabels holds all translatable strings for the activity expense module.
// ActivityExpense is the charge detail for ENTRY_TYPE_EXPENSE job activities.
// TODO(P7 lyngua sweep): add lyngua JSON files for these labels.
type ActivityExpenseLabels struct {
	Page    ActivityExpensePageLabels   `json:"page"`
	Buttons ActivityExpenseButtonLabels `json:"buttons"`
	Columns ActivityExpenseColumnLabels `json:"columns"`
	Empty   ActivityExpenseEmptyLabels  `json:"empty"`
	Form    ActivityExpenseFormLabels   `json:"form"`
	Detail  ActivityExpenseDetailLabels `json:"detail"`
	Errors  ActivityExpenseErrorLabels  `json:"errors"`
}

type ActivityExpensePageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ActivityExpenseButtonLabels struct {
	Add  string `json:"add"`
	Edit string `json:"edit"`
}

type ActivityExpenseColumnLabels struct {
	ExpenseCategory string `json:"expenseCategory"`
	VendorRef       string `json:"vendorRef"`
	PaymentMethod   string `json:"paymentMethod"`
	MarkupPct       string `json:"markupPct"`
}

type ActivityExpenseEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ActivityExpenseFormLabels struct {
	SectionExpense           string `json:"sectionExpense"`
	ExpenseCategory          string `json:"expenseCategory"`
	VendorRef                string `json:"vendorRef"`
	ReceiptURL               string `json:"receiptUrl"`
	PaymentMethod            string `json:"paymentMethod"`
	PaymentMethodEmployee    string `json:"paymentMethodEmployee"`
	PaymentMethodCompanyCard string `json:"paymentMethodCompanyCard"`
	PaymentMethodVendorBill  string `json:"paymentMethodVendorBill"`
	MarkupPctOverride        string `json:"markupPctOverride"`
	SubmitCreate             string `json:"submitCreate"`
	SubmitUpdate             string `json:"submitUpdate"`
}

type ActivityExpenseDetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	TitlePrefix       string `json:"titlePrefix"`
	ExpenseCategory   string `json:"expenseCategory"`
	VendorRef         string `json:"vendorRef"`
	ReceiptURL        string `json:"receiptUrl"`
	PaymentMethod     string `json:"paymentMethod"`
	MarkupPctOverride string `json:"markupPctOverride"`
}

type ActivityExpenseErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultActivityExpenseLabels returns ActivityExpenseLabels with sensible English defaults.
func DefaultActivityExpenseLabels() ActivityExpenseLabels {
	return ActivityExpenseLabels{
		Page: ActivityExpensePageLabels{
			Heading: "Expense Charges",
			Caption: "Expense cost entries per activity",
		},
		Buttons: ActivityExpenseButtonLabels{
			Add:  "Add Expense",
			Edit: "Edit Expense",
		},
		Columns: ActivityExpenseColumnLabels{
			ExpenseCategory: "Category",
			VendorRef:       "Vendor Ref",
			PaymentMethod:   "Payment Method",
			MarkupPct:       "Markup %",
		},
		Empty: ActivityExpenseEmptyLabels{
			Title:   "No expense entries",
			Message: "No expense charge recorded for this activity.",
		},
		Form: ActivityExpenseFormLabels{
			SectionExpense:           "Expense",
			ExpenseCategory:          "Expense Category",
			VendorRef:                "Vendor Reference",
			ReceiptURL:               "Receipt URL",
			PaymentMethod:            "Payment Method",
			PaymentMethodEmployee:    "Employee (out-of-pocket)",
			PaymentMethodCompanyCard: "Company Card",
			PaymentMethodVendorBill:  "Vendor Bill",
			MarkupPctOverride:        "Markup % Override",
			SubmitCreate:             "Save",
			SubmitUpdate:             "Update",
		},
		Detail: ActivityExpenseDetailLabels{
			PageTitle:         "Expense Charge",
			TitlePrefix:       "Expense: ",
			ExpenseCategory:   "Expense Category",
			VendorRef:         "Vendor Reference",
			ReceiptURL:        "Receipt URL",
			PaymentMethod:     "Payment Method",
			MarkupPctOverride: "Markup % Override",
		},
		Errors: ActivityExpenseErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Expense charge record not found",
			IDRequired:       "Activity ID is required",
		},
	}
}

// DefaultActivityLaborLabels returns ActivityLaborLabels with sensible English defaults.
func DefaultActivityLaborLabels() ActivityLaborLabels {
	return ActivityLaborLabels{
		Page: ActivityLaborPageLabels{
			Heading: "Labor Charges",
			Caption: "Labor time entries per activity",
		},
		Buttons: ActivityLaborButtonLabels{
			Add:  "Add Labor",
			Edit: "Edit Labor",
		},
		Columns: ActivityLaborColumnLabels{
			Staff:     "Staff",
			Hours:     "Hours",
			RateType:  "Rate Type",
			TimeStart: "Start",
			TimeEnd:   "End",
		},
		Empty: ActivityLaborEmptyLabels{
			Title:   "No labor entries",
			Message: "No labor charge recorded for this activity.",
		},
		Form: ActivityLaborFormLabels{
			SectionCharge:    "Charge",
			Staff:            "Staff",
			Hours:            "Hours",
			RateType:         "Rate Type",
			TimeStart:        "Start Time",
			TimeEnd:          "End Time",
			RateTypeStandard: "Standard",
			RateTypeOvertime: "Overtime",
			RateTypePremium:  "Premium",
			SubmitCreate:     "Save",
			SubmitUpdate:     "Update",
		},
		Detail: ActivityLaborDetailLabels{
			PageTitle:   "Labor Charge",
			TitlePrefix: "Labor: ",
			Staff:       "Staff",
			Hours:       "Hours",
			RateType:    "Rate Type",
			TimeStart:   "Start Time",
			TimeEnd:     "End Time",
		},
		Errors: ActivityLaborErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Labor charge record not found",
			IDRequired:       "Activity ID is required",
		},
	}
}

// DefaultJobActivityLabels returns JobActivityLabels with sensible English defaults.
func DefaultJobActivityLabels() JobActivityLabels {
	return JobActivityLabels{
		Page: JobActivityPageLabels{
			Heading: "Activities",
			Caption: "Cross-job timesheet and activity log",
		},
		Buttons: JobActivityButtonLabels{
			AddActivity: "Add Activity",
		},
		Columns: JobActivityColumnLabels{
			Date:        "Date",
			Job:         "Job",
			EntryType:   "Type",
			Description: "Description",
			Quantity:    "Hrs/Qty",
			Amount:      "Amount",
			Status:      "Status",
		},
		Empty: JobActivityEmptyLabels{
			Title:   "No activities found",
			Message: "No activity entries to display.",
		},
		Form: JobActivityFormLabels{
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
		Actions: JobActivityActionLabels{
			View:    "View Activity",
			Edit:    "Edit Activity",
			Delete:  "Delete Activity",
			Submit:  "Submit for Approval",
			Approve: "Approve",
			Reject:  "Reject",
			Post:    "Post",
			Reverse: "Reverse",
		},
		Detail: JobActivityDetailLabels{
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
		Status: JobActivityStatusLabels{
			Draft:     "Draft",
			Submitted: "Submitted",
			Approved:  "Approved",
			Rejected:  "Rejected",
		},
		Tabs: JobActivityTabLabels{
			Info:        "Information",
			Charge:      "Charge",
			Attachments: "Attachments",
		},
		Errors: JobActivityErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Activity not found",
			IDRequired:       "Activity ID is required",
		},
		// 2026-06-01 Wave 4.3 label sweep — charge tab subtype detail.
		Charge: JobActivityChargeLabels{
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

// ---------------------------------------------------------------------------
// Job (operational activity) labels
// ---------------------------------------------------------------------------

// JobLabels holds all translatable strings for the job module.
type JobLabels struct {
	Page        JobPageLabels       `json:"page"`
	Buttons     JobButtonLabels     `json:"buttons"`
	Columns     JobColumnLabels     `json:"columns"`
	Empty       JobEmptyLabels      `json:"empty"`
	Form        JobFormLabels       `json:"form"`
	Actions     JobActionLabels     `json:"actions"`
	Detail      JobDetailLabels     `json:"detail"`
	Tabs        JobTabLabels        `json:"tabs"`
	Confirm     JobConfirmLabels    `json:"confirm"`
	Errors      JobErrorLabels      `json:"errors"`
	BulkActions JobBulkActionLabels `json:"bulkActions"`
	// Dashboard labels for the Job live dashboard
	// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	Dashboard JobDashboardLabels `json:"dashboard"`
}

// JobDashboardLabels holds translatable strings for the Job live dashboard.
// (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
type JobDashboardLabels struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// Stats
	StatActive    string `json:"statActive"`
	StatDoneMonth string `json:"statDoneMonth"`
	StatOverdue   string `json:"statOverdue"`
	StatHoursWeek string `json:"statHoursWeek"`
	// Widgets
	HoursPerWeek      string `json:"hoursPerWeek"`
	UpcomingDeadlines string `json:"upcomingDeadlines"`
	RecentActivity    string `json:"recentActivity"`
	NoDeadlines       string `json:"noDeadlines"`
	NoActivity        string `json:"noActivity"`
	// Quick actions
	QuickNewJob      string `json:"quickNewJob"`
	QuickNewTemplate string `json:"quickNewTemplate"`
	QuickLogHours    string `json:"quickLogHours"`
	QuickJobCalendar string `json:"quickJobCalendar"`
	// Common
	ViewAll   string `json:"viewAll"`
	AxisHours string `json:"axisHours"`
}

type JobPageLabels struct {
	Heading          string `json:"heading"`
	Caption          string `json:"caption"`
	HeadingDraft     string `json:"headingDraft"`
	CaptionDraft     string `json:"captionDraft"`
	HeadingActive    string `json:"headingActive"`
	CaptionActive    string `json:"captionActive"`
	HeadingCompleted string `json:"headingCompleted"`
	CaptionCompleted string `json:"captionCompleted"`
	HeadingClosed    string `json:"headingClosed"`
	CaptionClosed    string `json:"captionClosed"`
	HeadingPlanned   string `json:"headingPlanned"`
	CaptionPlanned   string `json:"captionPlanned"`
	HeadingReleased  string `json:"headingReleased"`
	CaptionReleased  string `json:"captionReleased"`
	HeadingOnHold    string `json:"headingOnHold"`
	CaptionOnHold    string `json:"captionOnHold"`
}

type JobButtonLabels struct {
	AddJob string `json:"addJob"`
}

type JobColumnLabels struct {
	Name     string `json:"name"`
	Client   string `json:"client"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Location string `json:"location"`
}

type JobEmptyLabels struct {
	Title             string `json:"title"`
	Message           string `json:"message"`
	PhasesTitle       string `json:"phasesTitle"`
	PhasesMessage     string `json:"phasesMessage"`
	ActivitiesTitle   string `json:"activitiesTitle"`
	ActivitiesMessage string `json:"activitiesMessage"`
	SettlementTitle   string `json:"settlementTitle"`
	SettlementMessage string `json:"settlementMessage"`
	OutcomesTitle     string `json:"outcomesTitle"`
	OutcomesMessage   string `json:"outcomesMessage"`

	// 2026-06-01 Wave 4.3 label sweep — budget & actuals tab empty states
	// (job/templates/detail.html job-tab-budget / job-tab-actuals).
	BudgetTitle           string `json:"budgetTitle"`
	BudgetMessage         string `json:"budgetMessage"`
	BudgetNoPhasesTitle   string `json:"budgetNoPhasesTitle"`
	BudgetNoPhasesMessage string `json:"budgetNoPhasesMessage"`
	BudgetNoTasks         string `json:"budgetNoTasks"`
	ActualsTitle          string `json:"actualsTitle"`
	ActualsMessage        string `json:"actualsMessage"`
}

type JobFormLabels struct {
	NamePlaceholder     string `json:"namePlaceholder"`
	ClientPlaceholder   string `json:"clientPlaceholder"`
	LocationPlaceholder string `json:"locationPlaceholder"`
	NameInfo            string `json:"nameInfo"`
	ClientInfo          string `json:"clientInfo"`
	LocationInfo        string `json:"locationInfo"`
}

type JobActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type JobDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	TitlePrefix  string `json:"titlePrefix"`
	SectionInfo  string `json:"sectionInfo"`
	Approval     string `json:"approval"`
	Description  string `json:"description"`
	Quantity     string `json:"quantity"`
	UnitCost     string `json:"unitCost"`
	TotalCost    string `json:"totalCost"`
	EntryDate    string `json:"entryDate"`
	EntryType    string `json:"entryType"`
	PhaseName    string `json:"phaseName"`
	PhaseOrder   string `json:"phaseOrder"`
	PhaseStatus  string `json:"phaseStatus"`
	TaskName     string `json:"taskName"`
	TaskOrder    string `json:"taskOrder"`
	TaskStatus   string `json:"taskStatus"`
	AssignedTo   string `json:"assignedTo"`
	TargetType   string `json:"targetType"`
	TargetID     string `json:"targetId"`
	AllocatedAmt string `json:"allocatedAmount"`
	SettleDate   string `json:"settlementDate"`
	SettleStatus string `json:"settlementStatus"`

	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 / §9 — breadcrumb
	// link rendered when Job.origin_type = ORIGIN_TYPE_SUBSCRIPTION pointing
	// back to the centymo subscription detail page.
	OriginSubscriptionLink string `json:"originSubscriptionLink"`

	// 2026-04-29 milestone-billing plan §4 — phase set-status surface on the
	// Job detail Phases tab. PhasesSectionTitle heads the per-phase mini-table;
	// PhaseMarkComplete is the row CTA; PhaseStatusPending / PhaseStatusActive /
	// PhaseStatusCompleted render the status badge.
	PhasesSectionTitle   string `json:"phasesSectionTitle"`
	PhaseMarkComplete    string `json:"phaseMarkComplete"`
	PhaseStatusPending   string `json:"phaseStatusPending"`
	PhaseStatusActive    string `json:"phaseStatusActive"`
	PhaseStatusCompleted string `json:"phaseStatusCompleted"`

	// 2026-06-01 Wave 4.3 label sweep — Budget tab (job-tab-budget).
	BudgetSectionTitle   string `json:"budgetSectionTitle"`
	BudgetTask           string `json:"budgetTask"`
	BudgetHours          string `json:"budgetHours"`
	BudgetSubtotalSuffix string `json:"budgetSubtotalSuffix"`
	BudgetTotalHours     string `json:"budgetTotalHours"`

	// 2026-06-01 Wave 4.3 label sweep — Actuals tab (job-tab-actuals).
	ActualsSectionTitle  string `json:"actualsSectionTitle"`
	ActualsCount         string `json:"actualsCount"`
	ActualsGrandTotal    string `json:"actualsGrandTotal"`
	VarianceSectionTitle string `json:"varianceSectionTitle"`
	VarianceBudgetHours  string `json:"varianceBudgetHours"`
	VarianceActualsCost  string `json:"varianceActualsCost"`
	VarianceNote         string `json:"varianceNote"`
}

type JobTabLabels struct {
	Info        string `json:"info"`
	Phases      string `json:"phases"`
	Activities  string `json:"activities"`
	Settlement  string `json:"settlement"`
	Outcomes    string `json:"outcomes"`
	Attachments string `json:"attachments"`
	Budget      string `json:"budget"`
	Actuals     string `json:"actuals"`
	History     string `json:"history"`
}

type JobConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type JobErrorLabels struct {
	NotFound             string `json:"notFound"`
	PermissionDenied     string `json:"permissionDenied"`
	InUse                string `json:"inUse"`
	IDRequired           string `json:"idRequired"`
	InvalidForm          string `json:"invalidForm"`
	NoIDs                string `json:"noIds"`
	StatusRequired       string `json:"statusRequired"`
	TargetStatusRequired string `json:"targetStatusRequired"`
}

// JobBulkActionLabels holds translatable strings for job bulk-action controls.
type JobBulkActionLabels struct {
	Delete                 string `json:"delete"`
	BulkDelete             string `json:"bulkDelete"`
	BulkDeleteConfirmTitle string `json:"bulkDeleteConfirmTitle"`
	BulkDeleteConfirmMsg   string `json:"bulkDeleteConfirmMsg"`
	SetStatus              string `json:"setStatus"`
	BulkSetStatus          string `json:"bulkSetStatus"`
	SelectAll              string `json:"selectAll"`
	SelectedCount          string `json:"selectedCount"`
	Cancel                 string `json:"cancel"`
}

// ---------------------------------------------------------------------------
// JobPhase standalone module labels
// ---------------------------------------------------------------------------

// JobPhaseLabels holds all translatable strings for the job_phase module.
type JobPhaseLabels struct {
	Page    JobPhasePageLabels    `json:"page"`
	Buttons JobPhaseButtonLabels  `json:"buttons"`
	Columns JobPhaseColumnLabels  `json:"columns"`
	Empty   JobPhaseEmptyLabels   `json:"empty"`
	Form    JobPhaseFormLabels    `json:"form"`
	Actions JobPhaseActionLabels  `json:"actions"`
	Detail  JobPhaseDetailLabels  `json:"detail"`
	Tabs    JobPhaseTabLabels     `json:"tabs"`
	Confirm JobPhaseConfirmLabels `json:"confirm"`
	Errors  JobPhaseErrorLabels   `json:"errors"`
}

type JobPhasePageLabels struct {
	Heading          string `json:"heading"`
	Caption          string `json:"caption"`
	HeadingPending   string `json:"headingPending"`
	HeadingActive    string `json:"headingActive"`
	HeadingCompleted string `json:"headingCompleted"`
}

type JobPhaseButtonLabels struct {
	AddPhase string `json:"addPhase"`
}

type JobPhaseColumnLabels struct {
	Name         string `json:"name"`
	Job          string `json:"job"`
	PhaseOrder   string `json:"phaseOrder"`
	Status       string `json:"status"`
	PlannedStart string `json:"plannedStart"`
	PlannedEnd   string `json:"plannedEnd"`
}

type JobPhaseEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type JobPhaseFormLabels struct {
	SectionPhase             string `json:"sectionPhase"`
	SectionResourceTiming    string `json:"sectionResourceTiming"`
	Name                     string `json:"name"`
	NamePlaceholder          string `json:"namePlaceholder"`
	PhaseOrder               string `json:"phaseOrder"`
	Status                   string `json:"status"`
	PlannedStart             string `json:"plannedStart"`
	PlannedEnd               string `json:"plannedEnd"`
	ActualStart              string `json:"actualStart"`
	ActualEnd                string `json:"actualEnd"`
	TemplatePhasePlaceholder string `json:"templatePhasePlaceholder"`
	Resource                 string `json:"resource"`
	ResourcePlaceholder      string `json:"resourcePlaceholder"`
	SetupMinutes             string `json:"setupMinutes"`
	RunMinutesPerUnit        string `json:"runMinutesPerUnit"`
	PredecessorPhase         string `json:"predecessorPhase"`
	PredecessorPlaceholder   string `json:"predecessorPlaceholder"`
}

type JobPhaseActionLabels struct {
	View         string `json:"view"`
	Edit         string `json:"edit"`
	Delete       string `json:"delete"`
	MarkComplete string `json:"markComplete"`
}

type JobPhaseDetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	Name              string `json:"name"`
	Job               string `json:"job"`
	Status            string `json:"status"`
	PhaseOrder        string `json:"phaseOrder"`
	PlannedStart      string `json:"plannedStart"`
	PlannedEnd        string `json:"plannedEnd"`
	ActualStart       string `json:"actualStart"`
	ActualEnd         string `json:"actualEnd"`
	Resource          string `json:"resource"`
	SetupMinutes      string `json:"setupMinutes"`
	RunMinutesPerUnit string `json:"runMinutesPerUnit"`
}

type JobPhaseTabLabels struct {
	Info        string `json:"info"`
	Tasks       string `json:"tasks"`
	Activities  string `json:"activities"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type JobPhaseConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type JobPhaseErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultJobPhaseLabels returns JobPhaseLabels with sensible English defaults.
func DefaultJobPhaseLabels() JobPhaseLabels {
	return JobPhaseLabels{
		Page: JobPhasePageLabels{
			Heading:          "Job Phases",
			Caption:          "Manage execution phases across jobs",
			HeadingPending:   "Pending Phases",
			HeadingActive:    "Active Phases",
			HeadingCompleted: "Completed Phases",
		},
		Buttons: JobPhaseButtonLabels{
			AddPhase: "Add Phase",
		},
		Columns: JobPhaseColumnLabels{
			Name:         "Name",
			Job:          "Job",
			PhaseOrder:   "#",
			Status:       "Status",
			PlannedStart: "Planned Start",
			PlannedEnd:   "Planned End",
		},
		Empty: JobPhaseEmptyLabels{
			Title:   "No phases found",
			Message: "No job phases to display.",
		},
		Form: JobPhaseFormLabels{
			SectionPhase:             "Phase",
			SectionResourceTiming:    "Resource & Timing",
			Name:                     "Phase Name",
			NamePlaceholder:          "Enter phase name",
			PhaseOrder:               "Order",
			Status:                   "Status",
			PlannedStart:             "Planned Start",
			PlannedEnd:               "Planned End",
			ActualStart:              "Actual Start",
			ActualEnd:                "Actual End",
			TemplatePhasePlaceholder: "Search template phase...",
			Resource:                 "Resource",
			ResourcePlaceholder:      "Search resource...",
			SetupMinutes:             "Setup (min)",
			RunMinutesPerUnit:        "Run (min/unit)",
			PredecessorPhase:         "Predecessor Phase",
			PredecessorPlaceholder:   "Select predecessor...",
		},
		Actions: JobPhaseActionLabels{
			View:         "View Phase",
			Edit:         "Edit Phase",
			Delete:       "Delete Phase",
			MarkComplete: "Mark Complete",
		},
		Detail: JobPhaseDetailLabels{
			PageTitle:         "Phase Details",
			Name:              "Name",
			Job:               "Job",
			Status:            "Status",
			PhaseOrder:        "Order",
			PlannedStart:      "Planned Start",
			PlannedEnd:        "Planned End",
			ActualStart:       "Actual Start",
			ActualEnd:         "Actual End",
			Resource:          "Resource",
			SetupMinutes:      "Setup (min)",
			RunMinutesPerUnit: "Run (min/unit)",
		},
		Tabs: JobPhaseTabLabels{
			Info:        "Information",
			Tasks:       "Tasks",
			Activities:  "Activities",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: JobPhaseConfirmLabels{
			Delete:        "Delete Phase",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobPhaseErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Phase not found",
			IDRequired:       "Phase ID is required",
		},
	}
}

// ---------------------------------------------------------------------------
// JobTemplatePhase drawer-only module labels
// ---------------------------------------------------------------------------

// JobTemplatePhaseLabels holds all translatable strings for the job_template_phase
// drawer-only view module. This module has no list page, no sidebar entry, and no
// standalone detail page — operators reach it only via the JobTemplate detail Phases tab.
type JobTemplatePhaseLabels struct {
	Columns JobTemplatePhaseColumnLabels `json:"columns"`
	Form    JobTemplatePhaseFormLabels   `json:"form"`
	Actions JobTemplatePhaseActionLabels `json:"actions"`
	Errors  JobTemplatePhaseErrorLabels  `json:"errors"`
}

type JobTemplatePhaseColumnLabels struct {
	Name        string `json:"name"`
	PhaseOrder  string `json:"phaseOrder"`
	EstDuration string `json:"estDuration"`
}

type JobTemplatePhaseFormLabels struct {
	SectionPhase           string `json:"sectionPhase"`
	SectionResource        string `json:"sectionResource"`
	SectionDependencies    string `json:"sectionDependencies"`
	Name                   string `json:"name"`
	NamePlaceholder        string `json:"namePlaceholder"`
	PhaseOrder             string `json:"phaseOrder"`
	EstDurationMinutes     string `json:"estDurationMinutes"`
	Resource               string `json:"resource"`
	ResourcePlaceholder    string `json:"resourcePlaceholder"`
	PredecessorPhase       string `json:"predecessorPhase"`
	PredecessorPlaceholder string `json:"predecessorPlaceholder"`
}

type JobTemplatePhaseActionLabels struct {
	Add    string `json:"add"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type JobTemplatePhaseErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultJobTemplatePhaseLabels returns JobTemplatePhaseLabels with sensible English defaults.
func DefaultJobTemplatePhaseLabels() JobTemplatePhaseLabels {
	return JobTemplatePhaseLabels{
		Columns: JobTemplatePhaseColumnLabels{
			Name:        "Name",
			PhaseOrder:  "#",
			EstDuration: "Est. Duration (min)",
		},
		Form: JobTemplatePhaseFormLabels{
			SectionPhase:           "Phase",
			SectionResource:        "Resource",
			SectionDependencies:    "Dependencies",
			Name:                   "Phase Name",
			NamePlaceholder:        "Enter phase name",
			PhaseOrder:             "Order",
			EstDurationMinutes:     "Estimated Duration (min)",
			Resource:               "Resource",
			ResourcePlaceholder:    "Search resource...",
			PredecessorPhase:       "Predecessor Phase",
			PredecessorPlaceholder: "Select predecessor...",
		},
		Actions: JobTemplatePhaseActionLabels{
			Add:    "+ Add Phase",
			Edit:   "Edit Phase",
			Delete: "Delete Phase",
		},
		Errors: JobTemplatePhaseErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Template phase not found",
			IDRequired:       "Template phase ID is required",
		},
	}
}

// ---------------------------------------------------------------------------
// JobTemplateTask drawer-only module labels
// ---------------------------------------------------------------------------

// JobTemplateTaskLabels holds all translatable strings for the job_template_task
// drawer-only view module. This module has no list page, no sidebar entry, and no
// standalone detail page — operators reach it only via the JobTemplate detail Tasks tab.
type JobTemplateTaskLabels struct {
	Columns JobTemplateTaskColumnLabels `json:"columns"`
	Form    JobTemplateTaskFormLabels   `json:"form"`
	Actions JobTemplateTaskActionLabels `json:"actions"`
	Errors  JobTemplateTaskErrorLabels  `json:"errors"`
}

type JobTemplateTaskColumnLabels struct {
	Name        string `json:"name"`
	StepOrder   string `json:"stepOrder"`
	EstDuration string `json:"estDuration"`
	Phase       string `json:"phase"`
}

type JobTemplateTaskFormLabels struct {
	SectionTask         string `json:"sectionTask"`
	SectionResource     string `json:"sectionResource"`
	Name                string `json:"name"`
	NamePlaceholder     string `json:"namePlaceholder"`
	StepOrder           string `json:"stepOrder"`
	EstDurationMinutes  string `json:"estDurationMinutes"`
	Resource            string `json:"resource"`
	ResourcePlaceholder string `json:"resourcePlaceholder"`
}

type JobTemplateTaskActionLabels struct {
	Add    string `json:"add"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type JobTemplateTaskErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultJobTemplateTaskLabels returns JobTemplateTaskLabels with sensible English defaults.
func DefaultJobTemplateTaskLabels() JobTemplateTaskLabels {
	return JobTemplateTaskLabels{
		Columns: JobTemplateTaskColumnLabels{
			Name:        "Name",
			StepOrder:   "#",
			EstDuration: "Est. Duration (min)",
			Phase:       "Phase",
		},
		Form: JobTemplateTaskFormLabels{
			SectionTask:         "Task",
			SectionResource:     "Resource",
			Name:                "Task Name",
			NamePlaceholder:     "Enter task name",
			StepOrder:           "Order",
			EstDurationMinutes:  "Estimated Duration (min)",
			Resource:            "Resource",
			ResourcePlaceholder: "Search resource...",
		},
		Actions: JobTemplateTaskActionLabels{
			Add:    "+ Add Task",
			Edit:   "Edit Task",
			Delete: "Delete Task",
		},
		Errors: JobTemplateTaskErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Template task not found",
			IDRequired:       "Template task ID is required",
		},
	}
}

// ---------------------------------------------------------------------------
// JobTask standalone module labels
// ---------------------------------------------------------------------------

// JobTaskLabels holds all translatable strings for the job_task module.
type JobTaskLabels struct {
	Page    JobTaskPageLabels    `json:"page"`
	Buttons JobTaskButtonLabels  `json:"buttons"`
	Columns JobTaskColumnLabels  `json:"columns"`
	Empty   JobTaskEmptyLabels   `json:"empty"`
	Form    JobTaskFormLabels    `json:"form"`
	Actions JobTaskActionLabels  `json:"actions"`
	Detail  JobTaskDetailLabels  `json:"detail"`
	Tabs    JobTaskTabLabels     `json:"tabs"`
	Confirm JobTaskConfirmLabels `json:"confirm"`
	Errors  JobTaskErrorLabels   `json:"errors"`
}

type JobTaskPageLabels struct {
	Heading           string `json:"heading"`
	Caption           string `json:"caption"`
	HeadingPending    string `json:"headingPending"`
	HeadingInProgress string `json:"headingInProgress"`
	HeadingCompleted  string `json:"headingCompleted"`
}

type JobTaskButtonLabels struct {
	AddTask string `json:"addTask"`
}

type JobTaskColumnLabels struct {
	Name              string `json:"name"`
	Phase             string `json:"phase"`
	StepOrder         string `json:"stepOrder"`
	Status            string `json:"status"`
	AssignedTo        string `json:"assignedTo"`
	PercentComplete   string `json:"percentComplete"`
	PlannedQuantity   string `json:"plannedQuantity"`
	CompletedQuantity string `json:"completedQuantity"`
}

type JobTaskEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type JobTaskFormLabels struct {
	SectionTask               string `json:"sectionTask"`
	SectionAssignmentResource string `json:"sectionAssignmentResource"`
	SectionSchedule           string `json:"sectionSchedule"`
	SectionActuals            string `json:"sectionActuals"`
	Name                      string `json:"name"`
	NamePlaceholder           string `json:"namePlaceholder"`
	StepOrder                 string `json:"stepOrder"`
	Status                    string `json:"status"`
	IsAdHoc                   string `json:"isAdHoc"`
	AssignedTo                string `json:"assignedTo"`
	AssignedToPlaceholder     string `json:"assignedToPlaceholder"`
	ResourceID                string `json:"resourceId"`
	ResourcePlaceholder       string `json:"resourcePlaceholder"`
	TemplateTaskID            string `json:"templateTaskId"`
	TemplateTaskPlaceholder   string `json:"templateTaskPlaceholder"`
	PlannedQuantity           string `json:"plannedQuantity"`
	CompletedQuantity         string `json:"completedQuantity"`
	PercentComplete           string `json:"percentComplete"`
	AllowParallel             string `json:"allowParallel"`
	ActualStart               string `json:"actualStart"`
	ActualEnd                 string `json:"actualEnd"`
}

type JobTaskActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type JobTaskDetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	Name              string `json:"name"`
	Phase             string `json:"phase"`
	StepOrder         string `json:"stepOrder"`
	Status            string `json:"status"`
	IsAdHoc           string `json:"isAdHoc"`
	AssignedTo        string `json:"assignedTo"`
	Resource          string `json:"resource"`
	TemplateTask      string `json:"templateTask"`
	PlannedQuantity   string `json:"plannedQuantity"`
	CompletedQuantity string `json:"completedQuantity"`
	PercentComplete   string `json:"percentComplete"`
	AllowParallel     string `json:"allowParallel"`
	ActualStart       string `json:"actualStart"`
	ActualEnd         string `json:"actualEnd"`
}

type JobTaskTabLabels struct {
	Info        string `json:"info"`
	Activities  string `json:"activities"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type JobTaskConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type JobTaskErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultJobTaskLabels returns JobTaskLabels with sensible English defaults.
func DefaultJobTaskLabels() JobTaskLabels {
	return JobTaskLabels{
		Page: JobTaskPageLabels{
			Heading:           "Job Tasks",
			Caption:           "Manage execution tasks across job phases",
			HeadingPending:    "Pending Tasks",
			HeadingInProgress: "In-Progress Tasks",
			HeadingCompleted:  "Completed Tasks",
		},
		Buttons: JobTaskButtonLabels{
			AddTask: "Add Task",
		},
		Columns: JobTaskColumnLabels{
			Name:              "Name",
			Phase:             "Phase",
			StepOrder:         "#",
			Status:            "Status",
			AssignedTo:        "Assigned To",
			PercentComplete:   "% Done",
			PlannedQuantity:   "Planned Qty",
			CompletedQuantity: "Completed Qty",
		},
		Empty: JobTaskEmptyLabels{
			Title:   "No tasks found",
			Message: "No job tasks to display.",
		},
		Form: JobTaskFormLabels{
			SectionTask:               "Task",
			SectionAssignmentResource: "Assignment & Resource",
			SectionSchedule:           "Schedule",
			SectionActuals:            "Actuals",
			Name:                      "Task Name",
			NamePlaceholder:           "Enter task name",
			StepOrder:                 "Order",
			Status:                    "Status",
			IsAdHoc:                   "Ad Hoc",
			AssignedTo:                "Assigned To",
			AssignedToPlaceholder:     "Search staff...",
			ResourceID:                "Resource",
			ResourcePlaceholder:       "Search resource...",
			TemplateTaskID:            "Template Task",
			TemplateTaskPlaceholder:   "Search template task...",
			PlannedQuantity:           "Planned Qty",
			CompletedQuantity:         "Completed Qty",
			PercentComplete:           "% Complete",
			AllowParallel:             "Allow Parallel",
			ActualStart:               "Actual Start",
			ActualEnd:                 "Actual End",
		},
		Actions: JobTaskActionLabels{
			View:   "View Task",
			Edit:   "Edit Task",
			Delete: "Delete Task",
		},
		Detail: JobTaskDetailLabels{
			PageTitle:         "Task Details",
			Name:              "Name",
			Phase:             "Phase",
			StepOrder:         "Order",
			Status:            "Status",
			IsAdHoc:           "Ad Hoc",
			AssignedTo:        "Assigned To",
			Resource:          "Resource",
			TemplateTask:      "Template Task",
			PlannedQuantity:   "Planned Qty",
			CompletedQuantity: "Completed Qty",
			PercentComplete:   "% Complete",
			AllowParallel:     "Allow Parallel",
			ActualStart:       "Actual Start",
			ActualEnd:         "Actual End",
		},
		Tabs: JobTaskTabLabels{
			Info:        "Information",
			Activities:  "Activities",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: JobTaskConfirmLabels{
			Delete:        "Delete Task",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobTaskErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Task not found",
			IDRequired:       "Task ID is required",
		},
	}
}

// ---------------------------------------------------------------------------
// Outcome Criteria labels (criteria library)
// ---------------------------------------------------------------------------

// OutcomeCriteriaLabels holds all translatable strings for the outcome criteria module.
type OutcomeCriteriaLabels struct {
	Page    OutcomeCriteriaPageLabels    `json:"page"`
	Buttons OutcomeCriteriaButtonLabels  `json:"buttons"`
	Columns OutcomeCriteriaColumnLabels  `json:"columns"`
	Empty   OutcomeCriteriaEmptyLabels   `json:"empty"`
	Form    OutcomeCriteriaFormLabels    `json:"form"`
	Actions OutcomeCriteriaActionLabels  `json:"actions"`
	Detail  OutcomeCriteriaDetailLabels  `json:"detail"`
	Tabs    OutcomeCriteriaTabLabels     `json:"tabs"`
	Confirm OutcomeCriteriaConfirmLabels `json:"confirm"`
	Errors  OutcomeCriteriaErrorLabels   `json:"errors"`
}

type OutcomeCriteriaPageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type OutcomeCriteriaButtonLabels struct {
	AddCriterion string `json:"addCriterion"`
}

type OutcomeCriteriaColumnLabels struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Scope   string `json:"scope"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type OutcomeCriteriaEmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type OutcomeCriteriaFormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Type            string `json:"type"`
	Scope           string `json:"scope"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"descriptionPlaceholder"`
	Required        string `json:"required"`
	Weight          string `json:"weight"`
	TypeInfo        string `json:"typeInfo"`
	ScopeInfo       string `json:"scopeInfo"`
	WeightInfo      string `json:"weightInfo"`
}

type OutcomeCriteriaActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type OutcomeCriteriaDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Scope        string `json:"scope"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Required     string `json:"required"`
	Weight       string `json:"weight"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
}

type OutcomeCriteriaTabLabels struct {
	Info        string `json:"info"`
	Thresholds  string `json:"thresholds"`
	Options     string `json:"options"`
	Templates   string `json:"templates"`
	Versions    string `json:"versions"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type OutcomeCriteriaConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type OutcomeCriteriaErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
}

// DefaultOutcomeCriteriaLabels returns OutcomeCriteriaLabels with sensible English defaults.
func DefaultOutcomeCriteriaLabels() OutcomeCriteriaLabels {
	return OutcomeCriteriaLabels{
		Page: OutcomeCriteriaPageLabels{
			Heading:         "Criteria Library",
			HeadingActive:   "Active Criteria",
			HeadingInactive: "Inactive Criteria",
			Caption:         "Manage reusable outcome evaluation criteria",
			CaptionActive:   "Manage your active outcome criteria",
			CaptionInactive: "View inactive or deprecated criteria",
		},
		Buttons: OutcomeCriteriaButtonLabels{
			AddCriterion: "Add Criterion",
		},
		Columns: OutcomeCriteriaColumnLabels{
			Name:    "Name",
			Type:    "Type",
			Scope:   "Scope",
			Version: "Version",
			Status:  "Status",
		},
		Empty: OutcomeCriteriaEmptyLabels{
			Title:           "No criteria found",
			Message:         "No outcome criteria to display.",
			ActiveTitle:     "No active criteria",
			ActiveMessage:   "Create your first outcome criterion to get started.",
			InactiveTitle:   "No inactive criteria",
			InactiveMessage: "Deactivated criteria will appear here.",
		},
		Form: OutcomeCriteriaFormLabels{
			Name:            "Name",
			NamePlaceholder: "Enter criterion name",
			Type:            "Criteria Type",
			Scope:           "Scope",
			Description:     "Description",
			DescPlaceholder: "Enter criterion description...",
			Required:        "Required",
			Weight:          "Weight",
			TypeInfo:        "The evaluation method used to measure this criterion (e.g. numeric score, pass/fail).",
			ScopeInfo:       "Whether this criterion applies at the task, phase, or job level.",
			WeightInfo:      "Relative importance of this criterion when computing an aggregate score.",
		},
		Actions: OutcomeCriteriaActionLabels{
			View:   "View Criterion",
			Edit:   "Edit Criterion",
			Delete: "Delete Criterion",
		},
		Detail: OutcomeCriteriaDetailLabels{
			PageTitle:    "Criterion Details",
			Name:         "Name",
			Description:  "Description",
			Type:         "Criteria Type",
			Scope:        "Scope",
			Version:      "Version",
			Status:       "Status",
			Required:     "Required",
			Weight:       "Weight",
			CreatedDate:  "Created",
			ModifiedDate: "Last Modified",
		},
		Tabs: OutcomeCriteriaTabLabels{
			Info:        "Information",
			Thresholds:  "Thresholds",
			Options:     "Options",
			Templates:   "Templates",
			Versions:    "Versions",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: OutcomeCriteriaConfirmLabels{
			Delete:        "Delete Criterion",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: OutcomeCriteriaErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Outcome criterion not found",
			IDRequired:       "Criterion ID is required",
			NoPermission:     "No permission",
		},
	}
}

// ---------------------------------------------------------------------------
// Task Outcome labels (outcome recording on job tasks)
// ---------------------------------------------------------------------------

// TaskOutcomeLabels holds all translatable strings for the task outcome module.
type TaskOutcomeLabels struct {
	Page    TaskOutcomePageLabels    `json:"page"`
	Buttons TaskOutcomeButtonLabels  `json:"buttons"`
	Columns TaskOutcomeColumnLabels  `json:"columns"`
	Empty   TaskOutcomeEmptyLabels   `json:"empty"`
	Form    TaskOutcomeFormLabels    `json:"form"`
	Actions TaskOutcomeActionLabels  `json:"actions"`
	Detail  TaskOutcomeDetailLabels  `json:"detail"`
	Tabs    TaskOutcomeTabLabels     `json:"tabs"`
	Confirm TaskOutcomeConfirmLabels `json:"confirm"`
	Errors  TaskOutcomeErrorLabels   `json:"errors"`
}

// TaskOutcomeTabLabels holds tab labels for the task outcome detail page.
type TaskOutcomeTabLabels struct {
	Info        string `json:"info"`
	Attachments string `json:"attachments"`
}

type TaskOutcomePageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type TaskOutcomeButtonLabels struct {
	RecordOutcome string `json:"recordOutcome"`
}

type TaskOutcomeColumnLabels struct {
	Task          string `json:"task"`
	Criteria      string `json:"criteria"`
	Value         string `json:"value"`
	Determination string `json:"determination"`
	RecordedBy    string `json:"recordedBy"`
	Date          string `json:"date"`
}

type TaskOutcomeEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type TaskOutcomeFormLabels struct {
	Task             string `json:"task"`
	Criteria         string `json:"criteria"`
	Value            string `json:"value"`
	Notes            string `json:"notes"`
	NotesPlaceholder string `json:"notesPlaceholder"`
	Determination    string `json:"determination"`
}

type TaskOutcomeActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type TaskOutcomeDetailLabels struct {
	PageTitle           string `json:"pageTitle"`
	Task                string `json:"task"`
	Criteria            string `json:"criteria"`
	CriteriaType        string `json:"criteriaType"`
	Value               string `json:"value"`
	Determination       string `json:"determination"`
	DeterminationSource string `json:"determinationSource"`
	DeterminationNote   string `json:"determinationNote"`
	RecordedBy          string `json:"recordedBy"`
	RecordedDate        string `json:"recordedDate"`
	RevisionNumber      string `json:"revisionNumber"`
	CreatedDate         string `json:"createdDate"`
}

type TaskOutcomeConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type TaskOutcomeErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultTaskOutcomeLabels returns TaskOutcomeLabels with sensible English defaults.
func DefaultTaskOutcomeLabels() TaskOutcomeLabels {
	return TaskOutcomeLabels{
		Page: TaskOutcomePageLabels{
			Heading: "Outcome Recording",
			Caption: "Record and review task outcome evaluations",
		},
		Buttons: TaskOutcomeButtonLabels{
			RecordOutcome: "Record Outcome",
		},
		Columns: TaskOutcomeColumnLabels{
			Task:          "Task",
			Criteria:      "Criteria",
			Value:         "Value",
			Determination: "Determination",
			RecordedBy:    "Recorded By",
			Date:          "Date",
		},
		Empty: TaskOutcomeEmptyLabels{
			Title:   "No outcomes found",
			Message: "No task outcome records to display.",
		},
		Form: TaskOutcomeFormLabels{
			Task:             "Task",
			Criteria:         "Criteria",
			Value:            "Value",
			Notes:            "Notes",
			NotesPlaceholder: "Enter outcome notes...",
			Determination:    "Determination",
		},
		Actions: TaskOutcomeActionLabels{
			View:   "View Outcome",
			Edit:   "Edit Outcome",
			Delete: "Delete Outcome",
		},
		Detail: TaskOutcomeDetailLabels{
			PageTitle:           "Outcome Details",
			Task:                "Task",
			Criteria:            "Criteria",
			CriteriaType:        "Criteria Type",
			Value:               "Value",
			Determination:       "Determination",
			DeterminationSource: "Determination Source",
			DeterminationNote:   "Note",
			RecordedBy:          "Recorded By",
			RecordedDate:        "Recorded Date",
			RevisionNumber:      "Revision",
			CreatedDate:         "Created",
		},
		Tabs: TaskOutcomeTabLabels{
			Info:        "Information",
			Attachments: "Attachments",
		},
		Confirm: TaskOutcomeConfirmLabels{
			Delete:        "Delete Outcome",
			DeleteMessage: "Are you sure you want to delete this outcome record? This action cannot be undone.",
		},
		Errors: TaskOutcomeErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Outcome record not found",
			IDRequired:       "Outcome ID is required",
		},
	}
}

// ---------------------------------------------------------------------------
// Outcome Summary labels (job/phase report cards)
// ---------------------------------------------------------------------------

// OutcomeSummaryLabels holds all translatable strings for the outcome summary module.
type OutcomeSummaryLabels struct {
	Page    OutcomeSummaryPageLabels   `json:"page"`
	Buttons OutcomeSummaryButtonLabels `json:"buttons"`
	Columns OutcomeSummaryColumnLabels `json:"columns"`
	Empty   OutcomeSummaryEmptyLabels  `json:"empty"`
	Detail  OutcomeSummaryDetailLabels `json:"detail"`
	Errors  OutcomeSummaryErrorLabels  `json:"errors"`
}

type OutcomeSummaryColumnLabels struct {
	Job           string `json:"job"`
	Determination string `json:"determination"`
	Score         string `json:"score"`
	ScoringMethod string `json:"scoringMethod"`
	Total         string `json:"total"`
	Pass          string `json:"pass"`
	Fail          string `json:"fail"`
	IssuedBy      string `json:"issuedBy"`
}

type OutcomeSummaryEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type OutcomeSummaryPageLabels struct {
	JobHeading   string `json:"jobHeading"`
	JobCaption   string `json:"jobCaption"`
	PhaseHeading string `json:"phaseHeading"`
	PhaseCaption string `json:"phaseCaption"`
}

type OutcomeSummaryButtonLabels struct {
	GenerateSummary string `json:"generateSummary"`
}

type OutcomeSummaryDetailLabels struct {
	OverallDetermination string `json:"overallDetermination"`
	PhaseDetermination   string `json:"phaseDetermination"`
	Score                string `json:"score"`
	ScoringMethod        string `json:"scoringMethod"`
	TotalCriteria        string `json:"totalCriteria"`
	PassCount            string `json:"passCount"`
	FailCount            string `json:"failCount"`
	ConditionalCount     string `json:"conditionalCount"`
	DeferredCount        string `json:"deferredCount"`
	NaCount              string `json:"naCount"`
	Narrative            string `json:"narrative"`
	IssuedBy             string `json:"issuedBy"`
	IssuedDate           string `json:"issuedDate"`
	ValidUntilDate       string `json:"validUntilDate"`
}

type OutcomeSummaryErrorLabels struct {
	NotFound         string `json:"notFound"`
	PermissionDenied string `json:"permissionDenied"`
}

// DefaultOutcomeSummaryLabels returns OutcomeSummaryLabels with sensible English defaults.
func DefaultOutcomeSummaryLabels() OutcomeSummaryLabels {
	return OutcomeSummaryLabels{
		Page: OutcomeSummaryPageLabels{
			JobHeading:   "Outcome Summary",
			JobCaption:   "Job-level outcome report card",
			PhaseHeading: "Phase Outcome Summary",
			PhaseCaption: "Phase-level outcome report card",
		},
		Buttons: OutcomeSummaryButtonLabels{
			GenerateSummary: "Generate Summary",
		},
		Columns: OutcomeSummaryColumnLabels{
			Job:           "Job",
			Determination: "Determination",
			Score:         "Score",
			ScoringMethod: "Scoring Method",
			Total:         "Total",
			Pass:          "Pass",
			Fail:          "Fail",
			IssuedBy:      "Issued By",
		},
		Empty: OutcomeSummaryEmptyLabels{
			Title:   "No summaries",
			Message: "No outcome summaries have been generated yet.",
		},
		Detail: OutcomeSummaryDetailLabels{
			OverallDetermination: "Overall Determination",
			PhaseDetermination:   "Phase Determination",
			Score:                "Score",
			ScoringMethod:        "Scoring Method",
			TotalCriteria:        "Total Criteria",
			PassCount:            "Pass",
			FailCount:            "Fail",
			ConditionalCount:     "Conditional",
			DeferredCount:        "Deferred",
			NaCount:              "N/A",
			Narrative:            "Narrative",
			IssuedBy:             "Issued By",
			IssuedDate:           "Issued Date",
			ValidUntilDate:       "Valid Until",
		},
		Errors: OutcomeSummaryErrorLabels{
			NotFound:         "Outcome summary not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
	}
}

// ---------------------------------------------------------------------------
// Fulfillment labels
// ---------------------------------------------------------------------------

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

// DefaultJobLabels returns JobLabels with sensible English defaults.
func DefaultJobLabels() JobLabels {
	return JobLabels{
		Page: JobPageLabels{
			Heading:          "Jobs",
			Caption:          "Manage operational jobs",
			HeadingDraft:     "Draft Jobs",
			CaptionDraft:     "Review jobs that are still being prepared",
			HeadingActive:    "Active Jobs",
			CaptionActive:    "Track work currently in progress",
			HeadingCompleted: "Completed Jobs",
			CaptionCompleted: "Review jobs that have been completed",
			HeadingClosed:    "Closed Jobs",
			CaptionClosed:    "View archived or closed jobs",
		},
		Buttons: JobButtonLabels{
			AddJob: "Add Job",
		},
		Columns: JobColumnLabels{
			Name:     "Name",
			Client:   "Client",
			Status:   "Status",
			Created:  "Created",
			Location: "Location",
		},
		Empty: JobEmptyLabels{
			Title:             "No jobs found",
			Message:           "No jobs to display.",
			PhasesTitle:       "No phases",
			PhasesMessage:     "This job has no phases defined yet.",
			ActivitiesTitle:   "No activities",
			ActivitiesMessage: "No activity entries have been recorded for this job yet.",
			SettlementTitle:   "No settlements",
			SettlementMessage: "No cost allocations have been settled for this job yet.",
			OutcomesTitle:     "No outcomes",
			OutcomesMessage:   "No outcome evaluations have been recorded for this job yet.",
			// 2026-06-01 Wave 4.3 label sweep — budget & actuals tab empty states.
			BudgetTitle:           "No budget available",
			BudgetMessage:         "No template attached. Budget unavailable until a JobTemplate is linked to this matter.",
			BudgetNoPhasesTitle:   "No phases defined",
			BudgetNoPhasesMessage: "The linked template has no phases or tasks. Add phases to the template to see the budget breakdown.",
			BudgetNoTasks:         "No tasks",
			ActualsTitle:          "No actuals recorded",
			ActualsMessage:        "No activity entries have been posted for this job yet.",
		},
		Form: JobFormLabels{
			NamePlaceholder:     "Enter job name",
			ClientPlaceholder:   "Select client",
			LocationPlaceholder: "Select location",
			NameInfo:            "The name of the job as it appears in lists and documents.",
			ClientInfo:          "The client this job is being performed for.",
			LocationInfo:        "The location or site where this job takes place.",
		},
		Actions: JobActionLabels{
			View:   "View Job",
			Edit:   "Edit Job",
			Delete: "Delete Job",
		},
		Detail: JobDetailLabels{
			PageTitle:    "Job Details",
			TitlePrefix:  "Job ",
			SectionInfo:  "Job Information",
			Approval:     "Approval",
			Description:  "Description",
			Quantity:     "Quantity",
			UnitCost:     "Unit Cost",
			TotalCost:    "Total Cost",
			EntryDate:    "Date",
			EntryType:    "Entry Type",
			PhaseName:    "Phase",
			PhaseOrder:   "Order",
			PhaseStatus:  "Status",
			TaskName:     "Task",
			TaskOrder:    "Task Order",
			TaskStatus:   "Task Status",
			AssignedTo:   "Assigned To",
			TargetType:   "Target Type",
			TargetID:     "Target ID",
			AllocatedAmt: "Allocated Amount",
			SettleDate:   "Settlement Date",
			SettleStatus: "Settlement Status",
			// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 / §9.
			OriginSubscriptionLink: "Spawned from Subscription",
			// 2026-04-29 milestone-billing plan §4 — phase set-status surface.
			PhasesSectionTitle:   "Phases",
			PhaseMarkComplete:    "Mark Complete",
			PhaseStatusPending:   "Pending",
			PhaseStatusActive:    "Active",
			PhaseStatusCompleted: "Completed",
			// 2026-06-01 Wave 4.3 label sweep — Budget tab.
			BudgetSectionTitle:   "Estimated Hours by Phase",
			BudgetTask:           "Task",
			BudgetHours:          "Hours",
			BudgetSubtotalSuffix: "subtotal",
			BudgetTotalHours:     "Total estimated hours",
			// 2026-06-01 Wave 4.3 label sweep — Actuals tab.
			ActualsSectionTitle:  "Cost by Entry Type",
			ActualsCount:         "Count",
			ActualsGrandTotal:    "Grand Total",
			VarianceSectionTitle: "Budget vs Actuals",
			VarianceBudgetHours:  "Budget (estimated hours)",
			VarianceActualsCost:  "Actuals (total cost)",
			VarianceNote:         "Full money-vs-money variance available after Wave 3 (resource bill rates).",
		},
		Tabs: JobTabLabels{
			Info:        "Information",
			Phases:      "Phases",
			Activities:  "Activities",
			Settlement:  "Settlement",
			Outcomes:    "Outcomes",
			Attachments: "Attachments",
			Budget:      "Budget",
			Actuals:     "Actuals",
			History:     "History",
		},
		Confirm: JobConfirmLabels{
			Delete:        "Delete Job",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobErrorLabels{
			NotFound:         "Job not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
		Dashboard: JobDashboardLabels{
			Title:             "Jobs Dashboard",
			Subtitle:          "Active workload, upcoming deadlines, hours logged, and risk",
			StatActive:        "Active Jobs",
			StatDoneMonth:     "Done This Month",
			StatOverdue:       "Overdue",
			StatHoursWeek:     "Hours This Week",
			HoursPerWeek:      "Hours per Week",
			UpcomingDeadlines: "Upcoming Deadlines",
			RecentActivity:    "Recent Activity",
			NoDeadlines:       "No upcoming deadlines",
			NoActivity:        "No recent activity",
			QuickNewJob:       "New Job",
			QuickNewTemplate:  "New Job Template",
			QuickLogHours:     "Log Hours",
			QuickJobCalendar:  "Job Calendar",
			ViewAll:           "View All",
			AxisHours:         "Hours",
		},
	}
}
