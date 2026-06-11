// Package action holds the HTTP/HTMX handlers for the activity expense CRUD lifecycle.
package action

import (
	"context"
	"fmt"
	"strconv"

	operation "github.com/erniealice/fayna-golang/domain/operation"
	activityexpenseform "github.com/erniealice/fayna-golang/domain/operation/views/activity_expense/form"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
)

// Deps holds the dependencies for the activity expense action handlers.
type Deps struct {
	Routes operation.ActivityExpenseRoutes
	Labels operation.ActivityExpenseLabels

	// Use case functions — all injected by block.go via reflection-based wiring.
	// Nil-safe: handlers return an HTMXError when the required function is nil.
	CreateActivityExpense func(ctx context.Context, req *activityexpensepb.CreateActivityExpenseRequest) (*activityexpensepb.CreateActivityExpenseResponse, error)
	ReadActivityExpense   func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)
	UpdateActivityExpense func(ctx context.Context, req *activityexpensepb.UpdateActivityExpenseRequest) (*activityexpensepb.UpdateActivityExpenseResponse, error)
	DeleteActivityExpense func(ctx context.Context, req *activityexpensepb.DeleteActivityExpenseRequest) (*activityexpensepb.DeleteActivityExpenseResponse, error)
}

// buildPaymentMethodLabels extracts the payment method labels for option building.
func buildPaymentMethodLabels(labels operation.ActivityExpenseLabels) activityexpenseform.PaymentMethodLabels {
	return activityexpenseform.PaymentMethodLabels{
		Employee:    labels.Form.PaymentMethodEmployee,
		CompanyCard: labels.Form.PaymentMethodCompanyCard,
		VendorBill:  labels.Form.PaymentMethodVendorBill,
	}
}

// buildFormData populates a form.Data from an ActivityExpense proto record.
// Used by the Edit GET handler.
func buildFormData(record *activityexpensepb.ActivityExpense, routes operation.ActivityExpenseRoutes, labels operation.ActivityExpenseLabels) *activityexpenseform.Data {
	pmLabels := buildPaymentMethodLabels(labels)
	current := record.GetPaymentMethod()
	return &activityexpenseform.Data{
		IsEdit:                   true,
		ActivityID:               record.GetActivityId(),
		ExpenseCategoryID:        record.GetExpenseCategoryId(),
		VendorRef:                record.GetVendorRef(),
		ReceiptURL:               record.GetReceiptUrl(),
		PaymentMethod:            current,
		PaymentMethodOptions:     activityexpenseform.BuildPaymentMethodOptions(pmLabels, current),
		MarkupPctOverride:        record.GetMarkupPctOverride(),
		ExpenseCategorySearchURL: routes.ExpenseCategorySearchURL,
		Labels:                   labels,
	}
}

// buildEmptyFormData populates a form.Data for the Add GET handler.
// activityID is pre-filled from the ?activity_id query param.
func buildEmptyFormData(activityID string, routes operation.ActivityExpenseRoutes, labels operation.ActivityExpenseLabels) *activityexpenseform.Data {
	pmLabels := buildPaymentMethodLabels(labels)
	return &activityexpenseform.Data{
		IsEdit:                   false,
		ActivityID:               activityID,
		PaymentMethod:            "employee",
		PaymentMethodOptions:     activityexpenseform.BuildPaymentMethodOptions(pmLabels, "employee"),
		ExpenseCategorySearchURL: routes.ExpenseCategorySearchURL,
		Labels:                   labels,
	}
}

// parseFormFloat parses a float64 from a form value, returning 0.0 on error.
func parseFormFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// addFormAction returns the URL for the Add form's hx-post attribute.
func addFormAction(routes operation.ActivityExpenseRoutes) string {
	return routes.AddURL
}

// editFormAction returns the URL for the Edit form's hx-post attribute for a given activity ID.
func editFormAction(routes operation.ActivityExpenseRoutes, activityID string) string {
	if activityID == "" {
		return routes.EditURL
	}
	const placeholder = "{id}"
	url := routes.EditURL
	for i := 0; i <= len(url)-len(placeholder); i++ {
		if url[i:i+len(placeholder)] == placeholder {
			return url[:i] + activityID + url[i+len(placeholder):]
		}
	}
	return fmt.Sprintf("%s/%s", routes.EditURL, activityID)
}
