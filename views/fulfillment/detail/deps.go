package detail

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// DetailViewDeps holds view dependencies for the fulfillment detail views.
type DetailViewDeps struct {
	Routes       fayna.FulfillmentRoutes
	Labels       fayna.FulfillmentLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Enriched read: fulfillment + items + status events + returns + allowed events
	GetFulfillmentItemPageData func(ctx context.Context, req *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error)
}
