// Package action holds the HTTP/HTMX handlers for the activity labor CRUD lifecycle.
package action

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/erniealice/fayna-golang/domain/operation/activity_labor"
	activitylaborform "github.com/erniealice/fayna-golang/domain/operation/activity_labor/form"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
)

// Deps holds the dependencies for the activity labor action handlers.
type Deps struct {
	Routes activity_labor.Routes
	Labels activity_labor.Labels

	// Use case functions — all injected by block.go via reflection-based wiring.
	// Nil-safe: handlers return an HTMXError when the required function is nil.
	CreateActivityLabor func(ctx context.Context, req *activitylaborpb.CreateActivityLaborRequest) (*activitylaborpb.CreateActivityLaborResponse, error)
	ReadActivityLabor   func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	UpdateActivityLabor func(ctx context.Context, req *activitylaborpb.UpdateActivityLaborRequest) (*activitylaborpb.UpdateActivityLaborResponse, error)
	DeleteActivityLabor func(ctx context.Context, req *activitylaborpb.DeleteActivityLaborRequest) (*activitylaborpb.DeleteActivityLaborResponse, error)
}

// parseFormFloat parses a float64 from a form value, returning 0.0 on error.
func parseFormFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseDatetimeLocal converts a datetime-local input value ("2006-01-02T15:04")
// to a Unix int64 timestamp. Returns nil when the input is empty or unparseable.
func parseDatetimeLocal(s string) *int64 {
	if s == "" {
		return nil
	}
	// datetime-local format is "2006-01-02T15:04" (no seconds, no timezone).
	// Treat as local time — the server timezone applies.
	t, err := time.Parse("2006-01-02T15:04", s)
	if err != nil {
		// Try with seconds included (some browsers/tests include them).
		t, err = time.Parse("2006-01-02T15:04:05", s)
		if err != nil {
			return nil
		}
	}
	ts := t.Unix()
	return &ts
}

// formatDatetimeLocal converts a Unix int64 timestamp to the
// "2006-01-02T15:04" format expected by datetime-local HTML inputs.
// Returns "" when ts is 0 or nil.
func formatDatetimeLocal(ts *int64) string {
	if ts == nil || *ts == 0 {
		return ""
	}
	return time.Unix(*ts, 0).Format("2006-01-02T15:04")
}

// buildFormData populates a form.Data from an ActivityLabor proto record.
// Used by the Edit GET handler.
func buildFormData(record *activitylaborpb.ActivityLabor, routes activity_labor.Routes, labels activity_labor.Labels) *activitylaborform.Data {
	rateTypeLabels := activitylaborform.RateTypeLabels{
		Standard: labels.Form.RateTypeStandard,
		Overtime: labels.Form.RateTypeOvertime,
		Premium:  labels.Form.RateTypePremium,
	}
	currentRateType := activitylaborform.RateTypeToString(record.GetRateType())

	return &activitylaborform.Data{
		IsEdit:          true,
		ActivityID:      record.GetActivityId(),
		StaffID:         record.GetStaffId(),
		Hours:           record.GetHours(),
		RateType:        currentRateType,
		RateTypeOptions: activitylaborform.BuildRateTypeOptions(rateTypeLabels, currentRateType),
		TimeStart:       formatDatetimeLocal(record.TimeStart),
		TimeEnd:         formatDatetimeLocal(record.TimeEnd),
		StaffSearchURL:  routes.StaffSearchURL,
		Labels:          labels,
	}
}

// buildEmptyFormData populates a form.Data for the Add GET handler.
// activityID is pre-filled from the ?activity_id query param.
func buildEmptyFormData(activityID string, routes activity_labor.Routes, labels activity_labor.Labels) *activitylaborform.Data {
	rateTypeLabels := activitylaborform.RateTypeLabels{
		Standard: labels.Form.RateTypeStandard,
		Overtime: labels.Form.RateTypeOvertime,
		Premium:  labels.Form.RateTypePremium,
	}
	return &activitylaborform.Data{
		IsEdit:          false,
		ActivityID:      activityID,
		RateType:        "RATE_TYPE_STANDARD",
		RateTypeOptions: activitylaborform.BuildRateTypeOptions(rateTypeLabels, "RATE_TYPE_STANDARD"),
		StaffSearchURL:  routes.StaffSearchURL,
		Labels:          labels,
	}
}

// addFormAction returns the URL for the Add form's hx-post attribute.
func addFormAction(routes activity_labor.Routes) string {
	return routes.AddURL
}

// editFormAction returns the URL for the Edit form's hx-post attribute for a given activity ID.
func editFormAction(routes activity_labor.Routes, activityID string) string {
	if activityID == "" {
		return routes.EditURL
	}
	// routes.EditURL pattern: "/action/activity-labor/edit/{id}"
	// Use simple string replacement to resolve the pattern.
	const placeholder = "{id}"
	url := routes.EditURL
	for i := 0; i < len(url)-len(placeholder)+1; i++ {
		if url[i:i+len(placeholder)] == placeholder {
			return url[:i] + activityID + url[i+len(placeholder):]
		}
	}
	return fmt.Sprintf("%s/%s", routes.EditURL, activityID)
}
