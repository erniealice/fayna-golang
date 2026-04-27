package job_activity

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// activityFormData is the template data for the job activity drawer form.
type activityFormData struct {
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

// parseFormFloat parses a float64 from a form value, returning 0 on error.
func parseFormFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// newCreateAction creates the job activity create action (GET = form, POST = create).
func newCreateAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("job-activity-drawer-form", &activityFormData{
				FormAction:   deps.Routes.AddURL,
				JobID:        viewCtx.Request.URL.Query().Get("job_id"),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		id := deps.NewID()

		entryTypeStr := r.FormValue("entry_type")
		entryType := parseEntryType(entryTypeStr)

		quantity := parseFormFloat(r.FormValue("quantity"))
		unitCost := parseFormFloat(r.FormValue("unit_cost"))

		description := r.FormValue("description")

		unitCostCentavos := int64(math.Round(unitCost * 100))
		totalCostCentavos := int64(math.Round(quantity * unitCost * 100))

		_, err := deps.CreateJobActivity(ctx, &jobactivitypb.CreateJobActivityRequest{
			Data: &jobactivitypb.JobActivity{
				Id:             id,
				JobId:          r.FormValue("job_id"),
				EntryType:      entryType,
				Quantity:       quantity,
				UnitCost:       unitCostCentavos,
				TotalCost:      totalCostCentavos,
				Currency:       r.FormValue("currency"),
				Description:    &description,
				BillableStatus: parseBillableStatus(r.FormValue("billable_status")),
				ApprovalStatus: jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_DRAFT,
				Active:         true,
			},
		})
		if err != nil {
			log.Printf("Failed to create job activity: %v", err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// newUpdateAction creates the job activity update action (GET = pre-filled form, POST = update).
func newUpdateAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			id := viewCtx.Request.URL.Query().Get("id")
			if id == "" {
				return fayna.HTMXError(deps.Labels.Errors.IDRequired)
			}

			readResp, err := deps.ReadJobActivity(ctx, &jobactivitypb.ReadJobActivityRequest{
				Data: &jobactivitypb.JobActivity{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job activity %s: %v", id, err)
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			desc := ""
			if record.Description != nil {
				desc = *record.Description
			}

			return view.OK("job-activity-drawer-form", &activityFormData{
				FormAction:     route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:         true,
				ID:             id,
				JobID:          record.GetJobId(),
				EntryType:      record.GetEntryType().String(),
				Description:    desc,
				Quantity:       record.GetQuantity(),
				UnitCost:       float64(record.GetUnitCost()) / 100,
				Currency:       record.GetCurrency(),
				BillableStatus: record.GetBillableStatus().String(),
				Labels:         deps.Labels,
				CommonLabels:   nil, // injected by ViewAdapter
			})
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		id := r.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		entryTypeStr := r.FormValue("entry_type")
		entryType := parseEntryType(entryTypeStr)

		quantity := parseFormFloat(r.FormValue("quantity"))
		unitCost := parseFormFloat(r.FormValue("unit_cost"))

		description := r.FormValue("description")

		unitCostCentavos := int64(math.Round(unitCost * 100))
		totalCostCentavos := int64(math.Round(quantity * unitCost * 100))

		_, err := deps.UpdateJobActivity(ctx, &jobactivitypb.UpdateJobActivityRequest{
			Data: &jobactivitypb.JobActivity{
				Id:             id,
				JobId:          r.FormValue("job_id"),
				EntryType:      entryType,
				Quantity:       quantity,
				UnitCost:       unitCostCentavos,
				TotalCost:      totalCostCentavos,
				Currency:       r.FormValue("currency"),
				Description:    &description,
				BillableStatus: parseBillableStatus(r.FormValue("billable_status")),
			},
		})
		if err != nil {
			log.Printf("Failed to update job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// newDeleteAction creates the job activity delete action (POST).
func newDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteJobActivity(ctx, &jobactivitypb.DeleteJobActivityRequest{
			Data: &jobactivitypb.JobActivity{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// newSubmitAction creates the submit-for-approval action (POST).
func newSubmitAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.SubmitForApproval(ctx, &jobactivitypb.SubmitForApprovalRequest{
			ActivityId: id,
		})
		if err != nil {
			log.Printf("Failed to submit job activity %s for approval: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// newApproveAction creates the approve-activity action (POST).
func newApproveAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "approve") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.ApproveActivity(ctx, &jobactivitypb.ApproveJobActivityRequest{
			ActivityId: id,
			ApprovedBy: viewCtx.Request.FormValue("approved_by"),
		})
		if err != nil {
			log.Printf("Failed to approve job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// newRejectAction creates the reject-activity action (POST).
func newRejectAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "approve") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.RejectActivity(ctx, &jobactivitypb.RejectJobActivityRequest{
			ActivityId: id,
			RejectedBy: viewCtx.Request.FormValue("rejected_by"),
			Reason:     viewCtx.Request.FormValue("reason"),
		})
		if err != nil {
			log.Printf("Failed to reject job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// newPostAction creates the post-activity action (POST).
func newPostAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "post") && !perms.Can("job_activity", "manage") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.PostActivity(ctx, &jobactivitypb.PostJobActivityRequest{
			ActivityId: id,
			PostedBy:   viewCtx.Request.FormValue("posted_by"),
		})
		if err != nil {
			log.Printf("Failed to post job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// newReverseAction creates the reverse-activity action (POST).
func newReverseAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "manage") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.ReverseActivity(ctx, &jobactivitypb.ReverseJobActivityRequest{
			ActivityId: id,
			ReversedBy: viewCtx.Request.FormValue("reversed_by"),
			Reason:     viewCtx.Request.FormValue("reason"),
		})
		if err != nil {
			log.Printf("Failed to reverse job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}

// parseEntryType converts a form string to the EntryType enum.
func parseEntryType(s string) jobactivitypb.EntryType {
	switch s {
	case "LABOR", "labor":
		return jobactivitypb.EntryType_ENTRY_TYPE_LABOR
	case "MATERIAL", "material":
		return jobactivitypb.EntryType_ENTRY_TYPE_MATERIAL
	case "EXPENSE", "expense":
		return jobactivitypb.EntryType_ENTRY_TYPE_EXPENSE
	case "EQUIPMENT", "equipment":
		return jobactivitypb.EntryType_ENTRY_TYPE_EQUIPMENT
	case "SUBCONTRACT", "subcontract":
		return jobactivitypb.EntryType_ENTRY_TYPE_SUBCONTRACT
	default:
		return jobactivitypb.EntryType_ENTRY_TYPE_UNSPECIFIED
	}
}

// newBulkGenerateInvoiceAction creates the bulk generate invoice action (POST).
// It receives a list of selected activity IDs via multipart form-data and calls
// GenerateInvoiceFromActivities, then redirects to the new revenue detail page.
func newBulkGenerateInvoiceAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		r := viewCtx.Request
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			// Fall back to regular form parse (non-multipart submissions)
			if err2 := r.ParseForm(); err2 != nil {
				return fayna.HTMXError("Invalid form data")
			}
		}

		ids := r.Form["id"]
		if len(ids) == 0 {
			return fayna.HTMXError("No activities selected")
		}

		if deps.GenerateInvoiceFromActivities == nil {
			return fayna.HTMXError("Invoice generation not available")
		}

		revenueID, err := deps.GenerateInvoiceFromActivities(ctx, ids, "", "", "PHP", "")
		if err != nil {
			log.Printf("Failed to generate invoice from activities: %v", err)
			return fayna.HTMXError(fmt.Sprintf("Failed to generate invoice: %v", err))
		}

		redirectURL := fmt.Sprintf("/app/revenue/detail/%s?tab=items", revenueID)
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Redirect": redirectURL,
				"HX-Trigger":  `{"formSuccess":true}`,
			},
		}
	})
}

// parseBillableStatus converts a form string to the BillableStatus enum.
func parseBillableStatus(s string) jobactivitypb.BillableStatus {
	switch s {
	case "billable":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_BILLABLE
	case "non_billable":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_NON_BILLABLE
	case "included":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_INCLUDED
	case "write_off":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_WRITE_OFF
	default:
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_UNSPECIFIED
	}
}
