package operation

// activity_expense_labels.go — ActivityExpense label structs + DefaultActivityExpenseLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

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
