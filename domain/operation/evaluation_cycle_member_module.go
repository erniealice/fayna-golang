package operation

import (
	memberpkg "github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle_member"

	"github.com/erniealice/pyeza-golang/view"
)

// EvaluationCycleMemberModule is the cycle-member module assembler.
//
// STR-1: evaluation_cycle_member gets the full 14-layer cascade but NO standalone
// list page and NO routes. It surfaces only via the evaluation_cycle detail
// Members tab (members-tab.html, rendered by the cycle detail view) + the X-of-Y
// banner projection. This assembler therefore registers no routes — it exists to
// complete the cascade and expose the member templates FS + labels to the
// Integrator's catalog wiring.
type EvaluationCycleMemberModule struct {
	Labels memberpkg.Labels
}

// NewEvaluationCycleMemberModule constructs the (route-less) member module.
func NewEvaluationCycleMemberModule() *EvaluationCycleMemberModule {
	return &EvaluationCycleMemberModule{
		Labels: memberpkg.DefaultLabels(),
	}
}

// RegisterRoutes is a no-op (STR-1: no standalone member routes).
func (m *EvaluationCycleMemberModule) RegisterRoutes(r view.RouteRegistrar) {}
