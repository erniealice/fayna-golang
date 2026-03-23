package action

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// Deps holds dependencies for fulfillment action handlers.
type Deps struct {
	Routes fayna.FulfillmentRoutes
	Labels fayna.FulfillmentLabels

	CreateFulfillment func(ctx context.Context, req *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error)
	UpdateFulfillment func(ctx context.Context, req *fulfillmentpb.UpdateFulfillmentRequest) (*fulfillmentpb.UpdateFulfillmentResponse, error)
	DeleteFulfillment func(ctx context.Context, req *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error)

	// GetFulfillmentItemPageData is used by edit GET to fetch current field values.
	GetFulfillmentItemPageData func(ctx context.Context, req *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error)

	TransitionStatus        func(ctx context.Context, req *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error)
	CreateFulfillmentReturn func(ctx context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error)
}

// strPtr returns a pointer to a string value.
func strPtr(s string) *string {
	return &s
}
