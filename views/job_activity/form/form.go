package form

import fayna "github.com/erniealice/fayna-golang"

// Data is the template-facing data shape for the job activity drawer form.
// Used by both Create (FormAction = AddURL, IsEdit = false) and
// Update (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fayna.JobActivityLabels verbatim — templates
// read .Labels.Columns.*, .Labels.Form.*, .Labels.Errors.* directly.
//
// Monetary amounts (UnitCost, BillRate, BillAmount, Amount, HourlyRate) are
// in major units in this struct (display / form layer). Handlers ×100 on submit
// to convert to centavos before the use case call.
type Data struct {
	FormAction     string
	IsEdit         bool
	ID             string
	JobID          string
	EntryType      string
	Description    string
	Quantity       float64
	UnitCost       float64
	Currency       string
	BillableStatus string
	// BillRate and BillAmount drive the BILLABLE (T&M overage) path.
	// Stored on JobActivity proto in centavos; UI carries major units.
	BillRate   float64
	BillAmount float64
	// Labor fields
	Hours      float64
	HourlyRate float64
	StaffID    string
	// Material fields
	ProductID     string
	UnitOfMeasure string
	LotNumber     string
	Amount        float64
	// Expense fields
	ExpenseCategory string
	VendorRef       string
	Labels          fayna.JobActivityLabels
	CommonLabels    any
}
