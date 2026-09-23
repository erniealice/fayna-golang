package block

import (
	"context"
	"strings"
	"testing"
)

// TestAssertSeam_NilReturnsZero pins the nil-safe/unwired contract: a nil
// AppContext seam value must produce T's zero value (a nil closure), never a
// panic. Downstream nil checks (e.g. "if infra.ResolveTemplateBytes == nil")
// keep working unchanged.
func TestAssertSeam_NilReturnsZero(t *testing.T) {
	t.Parallel()

	got := assertSeam[func(context.Context, string, string) ([]byte, error)]("ResolveTemplateBytes", nil)
	if got != nil {
		t.Fatalf("assertSeam(nil) = non-nil closure, want nil")
	}
}

// TestAssertSeam_CorrectTypePassesThrough proves a value of the exact
// expected closure type is returned unchanged (comparing behavior, since Go
// funcs are not comparable with ==).
func TestAssertSeam_CorrectTypePassesThrough(t *testing.T) {
	t.Parallel()

	want := []byte("template-bytes")
	var v any = func(ctx context.Context, priceScheduleID, phaseCode string) ([]byte, error) {
		return want, nil
	}

	got := assertSeam[func(context.Context, string, string) ([]byte, error)]("ResolveTemplateBytes", v)
	if got == nil {
		t.Fatal("assertSeam(correct type) = nil, want the injected closure")
	}
	b, err := got(context.Background(), "sched-1", "phase-1")
	if err != nil || string(b) != string(want) {
		t.Fatalf("assertSeam(correct type) returned a closure that did not behave like the original: b=%q err=%v", b, err)
	}
}

// TestAssertSeam_WrongTypePanics is the regression pin for R1: a non-nil seam
// value of the WRONG concrete type (e.g. a consuming app still assigning the
// old 2-arg ResolveTemplateBytes closure against fayna's 3-arg contract) must
// fail loud at boot, not silently degrade to a nil/unwired seam via the bare
// `v, _ := v.(T)` idiom.
func TestAssertSeam_WrongTypePanics(t *testing.T) {
	t.Parallel()

	// A plausible but WRONG signature: 2-arg instead of the expected 3-arg.
	var v any = func(ctx context.Context, priceScheduleID string) ([]byte, error) {
		return nil, nil
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("assertSeam(wrong type) did not panic, want a fail-loud panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("assertSeam panic value is %T, want string", r)
		}
		if !strings.Contains(msg, "ResolveTemplateBytes") {
			t.Fatalf("panic message %q must name the seam", msg)
		}
		if !strings.Contains(msg, "want") {
			t.Fatalf("panic message %q must state the expected type", msg)
		}
	}()

	assertSeam[func(context.Context, string, string) ([]byte, error)]("ResolveTemplateBytes", v)
}
