// Package detail provides the ActivityExpense single-record detail page.
// V1: single-tab info view (no tab strip needed). Deep-linked from the
// JobActivity detail charge tab Edit CTA.
package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/activity_expense"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
)

// DetailViewDeps holds the dependencies for the activity expense detail page.
type DetailViewDeps struct {
	Routes       activity_expense.Routes
	Labels       activity_expense.Labels
	CommonLabels pyeza.CommonLabels

	// ReadActivityExpense fetches a single record by activity_id.
	// Nil-safe: renders a "not wired" placeholder when absent.
	ReadActivityExpense func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)
}

// PageData holds the data for the activity expense detail page.
type PageData struct {
	types.PageData
	ContentTemplate   string
	ActivityID        string
	ExpenseCategoryID string
	VendorRef         string
	ReceiptURL        string
	PaymentMethod     string
	MarkupPctOverride string
	EditURL           string
	Labels            activity_expense.Labels
}

// NewView creates the activity expense detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a.
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_expense", "read") {
			return view.Forbidden("activity_expense:read")
		}
		_ = perms

		activityID := viewCtx.Request.PathValue("id")
		if activityID == "" {
			return view.Error(fmt.Errorf("activity_id path value is required"))
		}

		l := deps.Labels
		headerTitle := l.Detail.TitlePrefix + activityID

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    headerTitle,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-receipt",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "activity-expense-detail-content",
			ActivityID:      activityID,
			Labels:          l,
		}

		// Resolve the Edit URL for the CTA.
		pageData.EditURL = resolveEditURL(deps.Routes.EditURL, activityID)

		if deps.ReadActivityExpense == nil {
			// Stub: render with empty fields + gap notice.
			return view.OK("activity-expense-detail", pageData)
		}

		resp, err := deps.ReadActivityExpense(ctx, &activityexpensepb.ReadActivityExpenseRequest{
			Data: &activityexpensepb.ActivityExpense{ActivityId: activityID},
		})
		if err != nil {
			log.Printf("Failed to read activity expense %s: %v", activityID, err)
			return view.Error(fmt.Errorf("failed to load expense record: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("expense record not found for activity %s", activityID))
		}

		record := data[0]
		pageData.ExpenseCategoryID = record.GetExpenseCategoryId()
		pageData.VendorRef = record.GetVendorRef()
		pageData.ReceiptURL = record.GetReceiptUrl()
		pageData.PaymentMethod = record.GetPaymentMethod()
		pageData.MarkupPctOverride = fmt.Sprintf("%.2f", record.GetMarkupPctOverride())

		return view.OK("activity-expense-detail", pageData)
	})
}

// resolveEditURL replaces the {id} placeholder in the edit URL pattern.
func resolveEditURL(editURLPattern, id string) string {
	const placeholder = "{id}"
	for i := 0; i <= len(editURLPattern)-len(placeholder); i++ {
		if editURLPattern[i:i+len(placeholder)] == placeholder {
			return editURLPattern[:i] + id + editURLPattern[i+len(placeholder):]
		}
	}
	return fmt.Sprintf("%s/%s", editURLPattern, id)
}
