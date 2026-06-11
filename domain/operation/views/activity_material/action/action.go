// Package action holds the HTTP/HTMX handlers for the activity material CRUD lifecycle.
package action

import (
	"context"
	"fmt"

	operation "github.com/erniealice/fayna-golang/domain/operation"
	activitymaterialform "github.com/erniealice/fayna-golang/domain/operation/views/activity_material/form"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
)

// Deps holds the dependencies for the activity material action handlers.
type Deps struct {
	Routes operation.ActivityMaterialRoutes
	Labels operation.ActivityMaterialLabels

	// Use case functions — all injected by block.go via reflection-based wiring.
	// Nil-safe: handlers return an HTMXError when the required function is nil.
	CreateActivityMaterial func(ctx context.Context, req *activitymaterialpb.CreateActivityMaterialRequest) (*activitymaterialpb.CreateActivityMaterialResponse, error)
	ReadActivityMaterial   func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	UpdateActivityMaterial func(ctx context.Context, req *activitymaterialpb.UpdateActivityMaterialRequest) (*activitymaterialpb.UpdateActivityMaterialResponse, error)
	DeleteActivityMaterial func(ctx context.Context, req *activitymaterialpb.DeleteActivityMaterialRequest) (*activitymaterialpb.DeleteActivityMaterialResponse, error)
}

// buildFormData populates a form.Data from an ActivityMaterial proto record.
// Used by the Edit GET handler.
func buildFormData(record *activitymaterialpb.ActivityMaterial, routes operation.ActivityMaterialRoutes, labels operation.ActivityMaterialLabels) *activitymaterialform.Data {
	productName := ""
	if p := record.GetProduct(); p != nil {
		productName = p.GetName()
	}
	locationName := ""
	if l := record.GetLocation(); l != nil {
		locationName = l.GetName()
	}
	return &activitymaterialform.Data{
		IsEdit:            true,
		ActivityID:        record.GetActivityId(),
		ProductID:         record.GetProductId(),
		ProductName:       productName,
		UnitOfMeasure:     record.GetUnitOfMeasure(),
		LotNumber:         record.GetLotNumber(),
		LocationID:        record.GetLocationId(),
		LocationName:      locationName,
		ProductSearchURL:  routes.ProductSearchURL,
		LocationSearchURL: routes.LocationSearchURL,
		Labels:            labels,
	}
}

// buildEmptyFormData populates a form.Data for the Add GET handler.
// activityID is pre-filled from the ?activity_id query param.
func buildEmptyFormData(activityID string, routes operation.ActivityMaterialRoutes, labels operation.ActivityMaterialLabels) *activitymaterialform.Data {
	return &activitymaterialform.Data{
		IsEdit:            false,
		ActivityID:        activityID,
		ProductSearchURL:  routes.ProductSearchURL,
		LocationSearchURL: routes.LocationSearchURL,
		Labels:            labels,
	}
}

// addFormAction returns the URL for the Add form's hx-post attribute.
func addFormAction(routes operation.ActivityMaterialRoutes) string {
	return routes.AddURL
}

// editFormAction returns the URL for the Edit form's hx-post attribute for a given activity ID.
func editFormAction(routes operation.ActivityMaterialRoutes, activityID string) string {
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
