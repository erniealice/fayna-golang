package home

import "testing"

// The option grammar's fail-safe contract (plan.md §4.2, mirroring
// job.Options): zero value inert; unknown refs skipped; DefaultSlot
// fail-closed to the lowest-capability declared variant.

func TestOptionsZeroValueIsInert(t *testing.T) {
	var o Options
	if o.Enabled() {
		t.Fatal("zero-value Options must be inert (Enabled() == false)")
	}
	if _, _, ok := o.VariantFor(SlotStaff); ok {
		t.Fatal("zero-value Options must resolve no variant")
	}
}

func TestVariantForExactMatch(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotOperatorStaff: {Widgets: []WidgetRef{WidgetCompletionSummary, WidgetAttention}},
		SlotStaff:         {Widgets: []WidgetRef{WidgetCompletionSummary}},
	}}
	v, slot, ok := o.VariantFor(SlotStaff)
	if !ok || slot != SlotStaff || len(v.Widgets) != 1 {
		t.Fatalf("exact match failed: ok=%v slot=%d widgets=%d", ok, slot, len(v.Widgets))
	}
}

func TestVariantForDefaultSlot(t *testing.T) {
	o := Options{
		Variants: map[Slot]Variant{
			SlotOperatorStaff: {Widgets: []WidgetRef{WidgetCompletionSummary, WidgetAttention}},
			SlotStaff:         {Widgets: []WidgetRef{WidgetCompletionSummary}},
		},
		DefaultSlot: SlotStaff,
	}
	// Unresolved kind-0 session → DefaultSlot.
	_, slot, ok := o.VariantFor(0)
	if !ok || slot != SlotStaff {
		t.Fatalf("kind 0 must resolve DefaultSlot: ok=%v slot=%d", ok, slot)
	}
	// Unknown kind → DefaultSlot too.
	_, slot, _ = o.VariantFor(42)
	if slot != SlotStaff {
		t.Fatalf("unknown kind must resolve DefaultSlot, got slot=%d", slot)
	}
}

func TestVariantForFailClosedLowestCapability(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotOperatorOwner: {Widgets: []WidgetRef{WidgetCompletionSummary, WidgetApprovalProgress, WidgetAttention}},
		SlotOperatorStaff: {Widgets: []WidgetRef{WidgetCompletionSummary, WidgetApprovalProgress, WidgetAttention, WidgetQuickLinks}},
		SlotStaff:         {Widgets: []WidgetRef{WidgetCompletionSummary, WidgetQuickLinks}},
	}}
	// DefaultSlot unset: kind 0 must land on the FEWEST-widget variant.
	v, slot, ok := o.VariantFor(0)
	if !ok || slot != SlotStaff || len(v.Widgets) != 2 {
		t.Fatalf("fail-closed resolution must pick the lowest-capability variant: ok=%v slot=%d widgets=%d", ok, slot, len(v.Widgets))
	}

	// DefaultSlot pointing at an UNDECLARED slot → same fail-closed path.
	o.DefaultSlot = 99
	_, slot, _ = o.VariantFor(0)
	if slot != SlotStaff {
		t.Fatalf("absent DefaultSlot must fall through to lowest-capability, got slot=%d", slot)
	}
}

func TestVariantForLowestCapabilityTieBreaksToHighestSlot(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotOperatorOwner: {Widgets: []WidgetRef{WidgetCompletionSummary}},
		SlotStaff:         {Widgets: []WidgetRef{WidgetQuickLinks}},
	}}
	_, slot, _ := o.VariantFor(0)
	if slot != SlotStaff {
		t.Fatalf("equal-size rosters must tie toward the highest (least-privileged) slot, got %d", slot)
	}
}

func TestKnownWidgetsSkipsUnknownRefs(t *testing.T) {
	v := Variant{Widgets: []WidgetRef{
		WidgetCompletionSummary,
		WidgetRef("made_up_widget"),
		WidgetQuickLinks,
	}}
	got := v.KnownWidgets()
	if len(got) != 2 || got[0] != WidgetCompletionSummary || got[1] != WidgetQuickLinks {
		t.Fatalf("unknown refs must be skipped fail-safe, got %v", got)
	}
}

func TestTabOptionsGrammar(t *testing.T) {
	if (TabOptions{}).Enabled() {
		t.Fatal("zero TabOptions must disable tabs")
	}
	if (TabOptions{GroupByField: "something_else"}).Enabled() {
		t.Fatal("unrecognized GroupByField must disable tabs (fail-safe)")
	}
	if !(TabOptions{GroupByField: TabEntityJobCategory}).Enabled() {
		t.Fatal("job_category ref must enable tabs")
	}
}

func TestWidgetBundles(t *testing.T) {
	for _, ref := range []WidgetRef{WidgetCompletionSummary, WidgetApprovalProgress, WidgetAttention} {
		b := WidgetBundle(ref)
		if len(b) != 1 || b[0] != "outcome_completion:read" {
			t.Fatalf("%s bundle must be exactly the dedicated Q7 code, got %v", ref, b)
		}
	}
	if len(WidgetBundle(WidgetQuickLinks)) != 0 {
		t.Fatal("quick_links performs no data read — bundle must be empty")
	}
}
