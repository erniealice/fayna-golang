package list

import (
	fulfillment "github.com/erniealice/fayna-golang/domain/fulfillment/fulfillment"
	"github.com/erniealice/pyeza-golang/types"
)

// fulfillmentColumns returns the column definitions for the fulfillment list table.
func fulfillmentColumns(l fulfillment.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "delivery_mode", Label: l.Columns.DeliveryMode},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-5xl"},
		{Key: "supplier_name", Label: l.Columns.SupplierName, NoSort: true},
		{Key: "item_count", Label: l.Columns.ItemCount, NoSort: true, WidthClass: "col-lg"},
	}
}
