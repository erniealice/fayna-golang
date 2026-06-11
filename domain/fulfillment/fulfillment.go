package fulfillment

// fulfillment.go — facade: re-exports every fulfillment entity-local type with
// the "Fulfillment" prefix so that consumers (block/, service-admin) keep writing
// fulfillment.FulfillmentLabels, fulfillment.DefaultFulfillmentRoutes(), etc.
//
// Rule D: this file may NOT import the domain facade from inside an entity
// package (that would create a cycle). Entity packages import nothing from
// domain/fulfillment/; only the domain/fulfillment/ packages import entity packages.

import (
	ffpkg "github.com/erniealice/fayna-golang/domain/fulfillment/fulfillment"
)

// ---------------------------------------------------------------------------
// Fulfillment label type aliases
// ---------------------------------------------------------------------------

type FulfillmentLabels = ffpkg.Labels
type FulfillmentStatusLabels = ffpkg.StatusLabels
type DeliveryModeLabels = ffpkg.DeliveryModeLabels
type FulfillmentColumnLabels = ffpkg.ColumnLabels
type FulfillmentTabLabels = ffpkg.TabLabels
type FulfillmentActionLabels = ffpkg.ActionLabels
type FulfillmentButtonLabels = ffpkg.ButtonLabels
type FulfillmentEmptyLabels = ffpkg.EmptyLabels
type FulfillmentErrorLabels = ffpkg.ErrorLabels
type FulfillmentDashboardLabels = ffpkg.DashboardLabels

func DefaultFulfillmentLabels() FulfillmentLabels { return ffpkg.DefaultLabels() }

// ---------------------------------------------------------------------------
// Fulfillment route type alias
// ---------------------------------------------------------------------------

type FulfillmentRoutes = ffpkg.Routes

func DefaultFulfillmentRoutes() FulfillmentRoutes { return ffpkg.DefaultRoutes() }

