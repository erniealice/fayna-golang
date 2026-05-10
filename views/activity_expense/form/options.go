package form

// PaymentMethodLabels carries the display labels used by BuildPaymentMethodOptions.
// Populated from fayna.ActivityExpenseLabels.Form.PaymentMethod* fields by the action layer.
type PaymentMethodLabels struct {
	Employee    string
	CompanyCard string
	VendorBill  string
}

// BuildPaymentMethodOptions returns a []Option for the payment method select picker.
// The current value is pre-selected. Accepts both the snake_case API form and a
// human-readable form for round-trip safety.
func BuildPaymentMethodOptions(labels PaymentMethodLabels, current string) []Option {
	type row struct {
		value string
		label string
	}
	rows := []row{
		{"employee", labels.Employee},
		{"company_card", labels.CompanyCard},
		{"vendor_bill", labels.VendorBill},
	}

	opts := make([]Option, 0, len(rows))
	for _, r := range rows {
		opts = append(opts, Option{
			Value:    r.value,
			Label:    r.label,
			Selected: current == r.value,
		})
	}
	return opts
}
