package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// PageData holds the data for the job activity detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Activity        map[string]any
	SubtypeData     map[string]any
	Labels          fayna.JobActivityLabels

	// Convenience fields for template rendering
	EntryType      string
	ApprovalStatus string
	StatusVariant  string
	Amount         types.TableCell
	Currency       string
}

// activityToMap converts a JobActivity protobuf to a map[string]any for template use.
func activityToMap(a *jobactivitypb.JobActivity) map[string]any {
	jobName := ""
	if j := a.GetJob(); j != nil {
		jobName = j.GetName()
	}

	currency := a.GetCurrency()
	unitCostCell := types.MoneyCell(float64(a.GetUnitCost()), currency, true)
	totalCostCell := types.MoneyCell(float64(a.GetTotalCost()), currency, true)

	return map[string]any{
		"id":              a.GetId(),
		"job_id":          a.GetJobId(),
		"job":             jobName,
		"job_task_id":     a.GetJobTaskId(),
		"entry_type":      entryTypeString(a.GetEntryType()),
		"entry_date":      a.GetEntryDateString(),
		"description":     a.GetDescription(),
		"quantity":        fmt.Sprintf("%.2f", a.GetQuantity()),
		"unit_cost":       unitCostCell,
		"total_cost":      totalCostCell,
		"currency":        currency,
		"billable_status": billableStatusString(a.GetBillableStatus()),
		"approval_status": approvalStatusString(a.GetApprovalStatus()),
		"posting_status":  a.GetPostingStatus().String(),
		"created_by":      a.GetCreatedBy(),
		"date_created":    a.GetDateCreatedString(),
		"active":          a.GetActive(),
	}
}

// NewView creates the job activity detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadJobActivity(ctx, &jobactivitypb.ReadJobActivityRequest{
			Data: &jobactivitypb.JobActivity{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job activity %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load activity: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job activity %s not found", id)
			return view.Error(fmt.Errorf("activity not found"))
		}
		record := data[0]
		activity := activityToMap(record)

		currency := record.GetCurrency()
		entryType := entryTypeString(record.GetEntryType())
		approvalStatus := approvalStatusString(record.GetApprovalStatus())

		l := deps.Labels
		headerTitle := l.Detail.TitlePrefix + id

		subtypeData := loadSubtypeData(ctx, deps, id, record.GetEntryType())

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "job",
				ActiveSubNav:   "activities",
				HeaderTitle:    headerTitle,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-clock",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-activity-detail-content",
			Activity:        activity,
			SubtypeData:     subtypeData,
			Labels:          l,
			EntryType:       entryType,
			ApprovalStatus:  approvalStatus,
			StatusVariant:   approvalStatusVariant(approvalStatus),
			Amount:          types.MoneyCell(float64(record.GetTotalCost()), currency, true),
			Currency:        currency,
		}

		return view.OK("job-activity-detail", pageData)
	})
}

// loadSubtypeData fetches entry-type-specific detail data.
func loadSubtypeData(ctx context.Context, deps *DetailViewDeps, id string, entryType jobactivitypb.EntryType) map[string]any {
	result := map[string]any{}

	switch entryType {
	case jobactivitypb.EntryType_ENTRY_TYPE_LABOR:
		if deps.ReadActivityLabor == nil {
			return result
		}
		resp, err := deps.ReadActivityLabor(ctx, &activitylaborpb.ReadActivityLaborRequest{
			Data: &activitylaborpb.ActivityLabor{ActivityId: id},
		})
		if err != nil {
			log.Printf("Failed to read labor detail for activity %s: %v", id, err)
			return result
		}
		data := resp.GetData()
		if len(data) > 0 {
			labor := data[0]
			result["staff_id"] = labor.GetStaffId()
			result["hours"] = fmt.Sprintf("%.2f", labor.GetHours())
			result["rate_type"] = labor.GetRateType().String()
			result["time_start"] = labor.GetTimeStartString()
			result["time_end"] = labor.GetTimeEndString()
		}

	case jobactivitypb.EntryType_ENTRY_TYPE_MATERIAL:
		if deps.ReadActivityMaterial == nil {
			return result
		}
		resp, err := deps.ReadActivityMaterial(ctx, &activitymaterialpb.ReadActivityMaterialRequest{
			Data: &activitymaterialpb.ActivityMaterial{ActivityId: id},
		})
		if err != nil {
			log.Printf("Failed to read material detail for activity %s: %v", id, err)
			return result
		}
		data := resp.GetData()
		if len(data) > 0 {
			mat := data[0]
			productName := ""
			if p := mat.GetProduct(); p != nil {
				productName = p.GetName()
			}
			result["product"] = productName
			result["product_id"] = mat.GetProductId()
			result["unit_of_measure"] = mat.GetUnitOfMeasure()
			result["lot_number"] = mat.GetLotNumber()
			result["location_id"] = mat.GetLocationId()
		}

	case jobactivitypb.EntryType_ENTRY_TYPE_EXPENSE:
		if deps.ReadActivityExpense == nil {
			return result
		}
		resp, err := deps.ReadActivityExpense(ctx, &activityexpensepb.ReadActivityExpenseRequest{
			Data: &activityexpensepb.ActivityExpense{ActivityId: id},
		})
		if err != nil {
			log.Printf("Failed to read expense detail for activity %s: %v", id, err)
			return result
		}
		data := resp.GetData()
		if len(data) > 0 {
			exp := data[0]
			result["expense_category"] = exp.GetExpenseCategory()
			result["vendor_ref"] = exp.GetVendorRef()
			result["receipt_url"] = exp.GetReceiptUrl()
			result["reimbursable"] = exp.GetReimbursable()
		}

	case jobactivitypb.EntryType_ENTRY_TYPE_EQUIPMENT:
		// Wave 3 will add ActivityEquipment table
		return result

	case jobactivitypb.EntryType_ENTRY_TYPE_SUBCONTRACT:
		// Wave 3 will add ActivitySubcontract table; for now ActivityExpense.is_subcontract flag covers this case
		return result
	}

	return result
}

func entryTypeString(t jobactivitypb.EntryType) string {
	switch t {
	case jobactivitypb.EntryType_ENTRY_TYPE_LABOR:
		return "labor"
	case jobactivitypb.EntryType_ENTRY_TYPE_MATERIAL:
		return "material"
	case jobactivitypb.EntryType_ENTRY_TYPE_EXPENSE:
		return "expense"
	case jobactivitypb.EntryType_ENTRY_TYPE_EQUIPMENT:
		return "Equipment"
	case jobactivitypb.EntryType_ENTRY_TYPE_SUBCONTRACT:
		return "Subcontract"
	default:
		return "unspecified"
	}
}

func approvalStatusString(s jobactivitypb.ActivityApprovalStatus) string {
	switch s {
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_DRAFT:
		return "draft"
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_SUBMITTED:
		return "submitted"
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_APPROVED:
		return "approved"
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_REJECTED:
		return "rejected"
	default:
		return "draft"
	}
}

func approvalStatusVariant(status string) string {
	switch status {
	case "draft":
		return "default"
	case "submitted":
		return "warning"
	case "approved":
		return "success"
	case "rejected":
		return "danger"
	default:
		return "default"
	}
}

func billableStatusString(s jobactivitypb.BillableStatus) string {
	switch s {
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_BILLABLE:
		return "billable"
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_NON_BILLABLE:
		return "non_billable"
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_INCLUDED:
		return "included"
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_WRITE_OFF:
		return "write_off"
	default:
		return "unspecified"
	}
}
