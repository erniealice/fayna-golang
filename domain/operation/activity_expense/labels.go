package activity_expense

// activity_expense_labels.go — ActivityExpense label structs + DefaultActivityExpenseLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// ActivityExpenseLabels holds all translatable strings for the activity expense module.
// ActivityExpense is the charge detail for ENTRY_TYPE_EXPENSE job activities.
// TODO(P7 lyngua sweep): add lyngua JSON files for these labels.
type Labels struct {
	Page    PageLabels   `json:"page"`
	Buttons ButtonLabels `json:"buttons"`
	Columns ColumnLabels `json:"columns"`
	Empty   EmptyLabels  `json:"empty"`
	Form    FormLabels   `json:"form"`
	Detail  DetailLabels `json:"detail"`
	Errors  ErrorLabels  `json:"errors"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ButtonLabels struct {
	Add  string `json:"add"`
	Edit string `json:"edit"`
}

type ColumnLabels struct {
	ExpenseCategory string `json:"expenseCategory"`
	VendorRef       string `json:"vendorRef"`
	PaymentMethod   string `json:"paymentMethod"`
	MarkupPct       string `json:"markupPct"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
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

type DetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	TitlePrefix       string `json:"titlePrefix"`
	ExpenseCategory   string `json:"expenseCategory"`
	VendorRef         string `json:"vendorRef"`
	ReceiptURL        string `json:"receiptUrl"`
	PaymentMethod     string `json:"paymentMethod"`
	MarkupPctOverride string `json:"markupPctOverride"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultActivityExpenseLabels returns ActivityExpenseLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Expense Charges",
			Caption: "Expense cost entries per activity",
		},
		Buttons: ButtonLabels{
			Add:  "Add Expense",
			Edit: "Edit Expense",
		},
		Columns: ColumnLabels{
			ExpenseCategory: "Category",
			VendorRef:       "Vendor Ref",
			PaymentMethod:   "Payment Method",
			MarkupPct:       "Markup %",
		},
		Empty: EmptyLabels{
			Title:   "No expense entries",
			Message: "No expense charge recorded for this activity.",
		},
		Form: FormLabels{
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
		Detail: DetailLabels{
			PageTitle:         "Expense Charge",
			TitlePrefix:       "Expense: ",
			ExpenseCategory:   "Expense Category",
			VendorRef:         "Vendor Reference",
			ReceiptURL:        "Receipt URL",
			PaymentMethod:     "Payment Method",
			MarkupPctOverride: "Markup % Override",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Expense charge record not found",
			IDRequired:       "Activity ID is required",
		},
	}
}
