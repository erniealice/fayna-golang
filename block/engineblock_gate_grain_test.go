package block

import (
	"strings"
	"testing"

	consumerapp "github.com/erniealice/espyna-golang/consumer/app"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
)

// TestEngineBlock_UnknownGateGrainIsBootError pins the boot-error contract of
// DocumentOptions.GateGrain: an unrecognized grain must refuse to mount the
// engine block — never a runtime guess, never a silent fallback to either
// grain (the deliberate divergence from the option structs' fail-safe ignore
// grammar). The validation runs BEFORE anything else, so even a bare
// AppContext observes it.
func TestEngineBlock_UnknownGateGrainIsBootError(t *testing.T) {
	t.Parallel()

	opt := EngineBlock(WithOutcomeSummaryOptions(outcome_summary.Options{
		Document: outcome_summary.DocumentOptions{GateGrain: "sideways"},
	}))
	err := opt(&consumerapp.AppContext{})
	if err == nil {
		t.Fatal("an unrecognized GateGrain must be a boot error, got nil")
	}
	if !strings.Contains(err.Error(), "GateGrain") {
		t.Fatalf("boot error must identify the GateGrain option, got: %v", err)
	}
}

// TestEngineBlock_RecognizedGateGrainsPassValidation proves both recognized
// grains get PAST the grain validation: with a bare AppContext the next
// failure is the UseCases assertion — a different, later error — so the grain
// check demonstrably did not trip.
func TestEngineBlock_RecognizedGateGrainsPassValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		grain string
	}{
		{"zero_value", ""},
		{"subscription_group", outcome_summary.GateGrainSubscriptionGroup},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opt := EngineBlock(WithOutcomeSummaryOptions(outcome_summary.Options{
				Document: outcome_summary.DocumentOptions{GateGrain: c.grain},
			}))
			err := opt(&consumerapp.AppContext{})
			if err == nil {
				t.Fatal("a bare AppContext must still fail RequireUseCases, got nil")
			}
			if strings.Contains(err.Error(), "GateGrain") {
				t.Fatalf("a recognized grain must not trip the grain validation, got: %v", err)
			}
		})
	}
}
