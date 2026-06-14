package performance

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
)

// deps.go — performance panel dependencies + the view-local projection.
//
// DESIGN (decoupling): the espyna read use case GetPerformancePanelData returns a
// Go-shaped projection that lives in espyna `internal/` (PanelData{Rows[]{Seat,
// LatestRating}}) — fayna CANNOT import it. So the panel defines its OWN
// presentation-shaped projection here (PanelData / PanelRow / GroupKey / CycleBanner)
// and the Deps closure returns THAT. The Integrator/block writes the thin adapter
// closure (espyna PanelData → this PanelData), and is the ONLY place that touches
// espyna: it also classifies each row into a GroupKey (Overdue/Due/UpToDate) from
// the seat cadence + latest evaluation period, and supplies the CycleBanner X-of-Y
// read projection (§F.2). The view itself calls NO espyna code — CR-5 servicing
// gating happens inside GetPerformancePanelData at the espyna layer.

// GroupKey is the cycle-status matrix bucket a panel row falls into (pages.md §G.1).
// The string values double as the `perf-group-{group}` testid suffix.
type GroupKey string

const (
	GroupOverdue  GroupKey = "overdue"
	GroupDue      GroupKey = "due"
	GroupUpToDate GroupKey = "up-to-date"
)

// PanelRow is one associate/seat row of the matrix. The block populates it from
// the espyna PanelRow (Seat + LatestRating) plus the cadence-derived Group.
type PanelRow struct {
	SeatID         string   // subscription_seat.id → perf-row-{sr_id} / CTAs
	StaffID        string   // subject_staff_id → associate-rating-{staff_id}
	AssociateName  string   // resolved staff display name (block joins staff)
	ClientName     string   // resolved client display name
	LatestRating   *float64 // nil → "Not yet rated", blank pill
	LatestEvalID   string   // last evaluation id for the View-last-review CTA ("" → hide)
	Group          GroupKey // matrix bucket
}

// CycleBanner is the shared "X of Y" read projection (§F.2) rendered atop the
// panel. Frozen denominator Y = COUNT(member); X = COUNT(member ⋈ SUBMITTED/
// SIGNED_OFF). Nil when no open cycle scopes the panel.
type CycleBanner struct {
	CycleID         string
	Name            string
	Completed       int    // X
	Total           int    // Y
	SignOffDueLabel string // pre-formatted date (block formats; "" hides)
	CloseLabel      string // pre-formatted date ("" hides)
}

// PanelData is the assembled panel projection the view consumes.
type PanelData struct {
	Rows   []PanelRow
	Banner *CycleBanner // nil when no scoping cycle
}

// ListViewDeps holds the dependencies for the performance panel page.
type ListViewDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// GetPanelData is the block-supplied adapter over espyna's servicing-gated
	// GetPerformancePanelData (CR-5). Returns the view-local PanelData. The view
	// passes no client_id/subscription_id — scoping is server-side (deny-before-SQL
	// on the espyna side). A nil PanelData is treated as empty.
	GetPanelData func(ctx context.Context) (*PanelData, error)
}

// ModuleDeps mirrors the entity-package deps surface for the block to construct
// the unit. The performance panel has a single view, so this aliases ListViewDeps
// fields; kept distinct for parity with the other operation entities.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	GetPanelData func(ctx context.Context) (*PanelData, error)
}
