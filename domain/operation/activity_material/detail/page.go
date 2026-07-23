// Package detail provides the ActivityMaterial single-record detail page.
// V1: single-tab info view (no tab strip needed). Deep-linked from the
// JobActivity detail charge tab Edit CTA.
package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/activity_material"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
)

// DetailViewDeps holds the dependencies for the activity material detail page.
type DetailViewDeps struct {
	Routes       activity_material.Routes
	Labels       activity_material.Labels
	CommonLabels pyeza.CommonLabels

	// ReadActivityMaterial fetches a single record by activity_id.
	// Nil-safe: renders a "not wired" placeholder when absent.
	ReadActivityMaterial func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
}

// PageData holds the data for the activity material detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	ActivityID      string
	ProductID       string
	ProductName     string
	UnitOfMeasure   string
	LotNumber       string
	LocationID      string
	LocationName    string
	EditURL         string
	Labels          activity_material.Labels
}

// NewView creates the activity material detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a.
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_material", "read") {
			return view.Forbidden("activity_material:read")
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
				HeaderIcon:     "icon-package",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "activity-material-detail-content",
			ActivityID:      activityID,
			Labels:          l,
		}

		// Resolve the Edit URL for the CTA.
		pageData.EditURL = resolveEditURL(deps.Routes.EditURL, activityID)

		if deps.ReadActivityMaterial == nil {
			// Stub: render with empty fields + gap notice.
			return view.OK("activity-material-detail", pageData)
		}

		resp, err := deps.ReadActivityMaterial(ctx, &activitymaterialpb.ReadActivityMaterialRequest{
			Data: &activitymaterialpb.ActivityMaterial{ActivityId: activityID},
		})
		if err != nil {
			log.Printf("Failed to read activity material %s: %v", activityID, err)
			return view.Error(fmt.Errorf("failed to load material record: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("material record not found for activity %s", activityID))
		}

		record := data[0]
		pageData.ProductID = record.GetProductId()
		if p := record.GetProduct(); p != nil {
			pageData.ProductName = p.GetName()
		}
		pageData.UnitOfMeasure = record.GetUnitOfMeasure()
		pageData.LotNumber = record.GetLotNumber()
		pageData.LocationID = record.GetLocationId()
		if loc := record.GetLocation(); loc != nil {
			pageData.LocationName = loc.GetName()
		}

		return view.OK("activity-material-detail", pageData)
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
