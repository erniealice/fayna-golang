// Package detail provides the ActivityLabor single-record detail page.
// V1: single-tab info view (no tab strip needed). Deep-linked from the
// JobActivity detail charge tab Edit CTA.
package detail

import (
	"context"
	"fmt"
	"log"
	"time"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
)

// DetailViewDeps holds the dependencies for the activity labor detail page.
type DetailViewDeps struct {
	Routes       fayna.ActivityLaborRoutes
	Labels       fayna.ActivityLaborLabels
	CommonLabels pyeza.CommonLabels

	// ReadActivityLabor fetches a single record by activity_id.
	// Nil-safe: renders a "not wired" placeholder when absent.
	ReadActivityLabor func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
}

// PageData holds the data for the activity labor detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	ActivityID      string
	StaffID         string
	Hours           string
	RateType        string
	TimeStart       string
	TimeEnd         string
	EditURL         string
	Labels          fayna.ActivityLaborLabels
}

// NewView creates the activity labor detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a.
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_labor", "read") {
			return view.Forbidden("activity_labor:read")
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
				HeaderIcon:     "icon-clock",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "activity-labor-detail-content",
			ActivityID:      activityID,
			Labels:          l,
		}

		// Resolve the Edit URL for the CTA.
		pageData.EditURL = resolveEditURL(deps.Routes.EditURL, activityID)

		if deps.ReadActivityLabor == nil {
			// Stub: render with empty fields + gap notice.
			return view.OK("activity-labor-detail", pageData)
		}

		resp, err := deps.ReadActivityLabor(ctx, &activitylaborpb.ReadActivityLaborRequest{
			Data: &activitylaborpb.ActivityLabor{ActivityId: activityID},
		})
		if err != nil {
			log.Printf("Failed to read activity labor %s: %v", activityID, err)
			return view.Error(fmt.Errorf("failed to load labor record: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("labor record not found for activity %s", activityID))
		}

		record := data[0]
		pageData.StaffID = record.GetStaffId()
		pageData.Hours = fmt.Sprintf("%.2f", record.GetHours())
		pageData.RateType = record.GetRateType().String()
		pageData.TimeStart = formatTimestamp(record.TimeStart)
		pageData.TimeEnd = formatTimestamp(record.TimeEnd)

		return view.OK("activity-labor-detail", pageData)
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

// formatTimestamp converts a Unix int64 pointer to a human-readable string.
func formatTimestamp(ts *int64) string {
	if ts == nil || *ts == 0 {
		return ""
	}
	return time.Unix(*ts, 0).Format("2006-01-02 15:04")
}
