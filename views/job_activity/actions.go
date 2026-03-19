package job_activity

import (
	"context"
	"log"
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
				FormAction:   deps.Routes.CreateURL,
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

		_, err := deps.CreateJobActivity(ctx, &jobactivitypb.CreateJobActivityRequest{
			Data: &jobactivitypb.JobActivity{
				Id:             id,
				JobId:          r.FormValue("job_id"),
				EntryType:      entryType,
				Quantity:        quantity,
				UnitCost:        unitCost * 100, // store in cents
				TotalCost:       quantity * unitCost * 100,
				Currency:        r.FormValue("currency"),
				Description:     &description,
				BillableStatus:  parseBillableStatus(r.FormValue("billable_status")),
				ApprovalStatus:  jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_DRAFT,
				Active:          true,
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
				FormAction:     route.ResolveURL(deps.Routes.UpdateURL, "id", id),
				IsEdit:         true,
				ID:             id,
				JobID:          record.GetJobId(),
				EntryType:      record.GetEntryType().String(),
				Description:    desc,
				Quantity:       record.GetQuantity(),
				UnitCost:       record.GetUnitCost() / 100,
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

		_, err := deps.UpdateJobActivity(ctx, &jobactivitypb.UpdateJobActivityRequest{
			Data: &jobactivitypb.JobActivity{
				Id:             id,
				JobId:          r.FormValue("job_id"),
				EntryType:      entryType,
				Quantity:        quantity,
				UnitCost:        unitCost * 100,
				TotalCost:       quantity * unitCost * 100,
				Currency:        r.FormValue("currency"),
				Description:     &description,
				BillableStatus:  parseBillableStatus(r.FormValue("billable_status")),
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

// parseEntryType converts a form string to the EntryType enum.
func parseEntryType(s string) jobactivitypb.EntryType {
	switch s {
	case "LABOR", "labor":
		return jobactivitypb.EntryType_ENTRY_TYPE_LABOR
	case "MATERIAL", "material":
		return jobactivitypb.EntryType_ENTRY_TYPE_MATERIAL
	case "EXPENSE", "expense":
		return jobactivitypb.EntryType_ENTRY_TYPE_EXPENSE
	default:
		return jobactivitypb.EntryType_ENTRY_TYPE_UNSPECIFIED
	}
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
