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
	// DefaultSlot unset: kind 0 must land on the least-capable legacy variant.
	v, slot, ok := o.VariantFor(0)
	if !ok || slot != SlotStaff || len(v.KnownWidgets()) != 2 {
		t.Fatalf("fail-closed resolution must pick the lowest-capability variant: ok=%v slot=%d widgets=%d", ok, slot, len(v.KnownWidgets()))
	}

	// DefaultSlot pointing at an UNDECLARED slot → same fail-closed path.
	o.DefaultSlot = 99
	_, slot, _ = o.VariantFor(0)
	if slot != SlotStaff {
		t.Fatalf("absent DefaultSlot must fall through to lowest-capability, got slot=%d", slot)
	}
}

func TestVariantForSectionOnlyLeastCapability(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotOperatorOwner: {
			Sections: []Section{
				{Key: SectionOverview, Widgets: []WidgetRef{WidgetCompletionSummary, WidgetApprovalProgress}},
			},
		},
		SlotStaff: {
			Sections: []Section{
				{Key: SectionOverview, Widgets: []WidgetRef{WidgetCompletionSummary}},
			},
		},
		SlotOperatorStaff: {
			Sections: []Section{
				{Key: SectionPulse, Widgets: []WidgetRef{WidgetCompletionSummary, WidgetCompletionSummary}},
				{Key: SectionPerformance, Widgets: []WidgetRef{WidgetCompletionSummary}},
			},
		},
	}}
	_, slot, ok := o.VariantFor(0)
	if !ok || slot != SlotStaff {
		t.Fatalf("section-only fallback must pick least-capable effective section set: ok=%v slot=%d", ok, slot)
	}
}

func TestSectionForExactConfiguredSection(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotStaff: {
			Sections: []Section{
				{Key: SectionOverview, Widgets: []WidgetRef{WidgetCompletionSummary}},
				{Key: SectionPulse, Widgets: []WidgetRef{WidgetAttention}},
				{Key: SectionPerformance, Widgets: []WidgetRef{WidgetApprovalProgress}},
			},
		},
	}}
	section, slot, ok := o.SectionFor(SlotStaff, SectionPerformance)
	if !ok || slot != SlotStaff {
		t.Fatalf("exact section lookup must resolve: ok=%v slot=%d", ok, slot)
	}
	if section.Key != SectionPerformance {
		t.Fatalf("exact configured section should be returned: got=%s", section.Key)
	}
	if len(section.Widgets) != 1 || section.Widgets[0] != WidgetApprovalProgress {
		t.Fatalf("configured section should preserve widgets: %#v", section.Widgets)
	}
}

func TestSectionForDefaultsToFirstSectionWhenRequestedKeyIsEmpty(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotStaff: {
			Sections: []Section{
				{Key: SectionPulse, Widgets: []WidgetRef{WidgetQuickLinks}},
				{Key: SectionOverview, Widgets: []WidgetRef{WidgetCompletionSummary}},
			},
		},
	}}
	section, slot, ok := o.SectionFor(SlotStaff, "")
	if !ok || slot != SlotStaff {
		t.Fatalf("empty key must still resolve default section: ok=%v slot=%d", ok, slot)
	}
	if section.Key != SectionPulse {
		t.Fatalf("empty/absent key must return variant default section: got=%s", section.Key)
	}
}

func TestSectionForUnknownSectionFallsBackToFirst(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotStaff: {
			Sections: []Section{
				{Key: SectionPulse, Widgets: []WidgetRef{WidgetQuickLinks}},
				{Key: SectionPerformance, Widgets: []WidgetRef{WidgetApprovalProgress}},
			},
		},
	}}
	section, slot, ok := o.SectionFor(SlotStaff, SectionAttention)
	if !ok || slot != SlotStaff {
		t.Fatalf("unknown key must fail-safe to first section: ok=%v slot=%d", ok, slot)
	}
	if section.Key != SectionPulse {
		t.Fatalf("unknown section key must fallback to first section: got=%s", section.Key)
	}
}

func TestSectionForLegacyWidgetsOnlyUsesImplicitOverview(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotStaff: {
			Widgets: []WidgetRef{
				WidgetCompletionSummary,
				WidgetQuickLinks,
			},
		},
	}}
	section, slot, ok := o.SectionFor(SlotStaff, SectionOverview)
	if !ok || slot != SlotStaff {
		t.Fatalf("legacy widget-only variant must resolve: ok=%v slot=%d", ok, slot)
	}
	if section.Key != SectionOverview {
		t.Fatalf("legacy variant must synthesize overview section: got=%s", section.Key)
	}
	if len(section.Widgets) != 2 {
		t.Fatalf("legacy widget-only section should include legacy widgets: %#v", section.Widgets)
	}
	sectionFromUnknown, _, ok := o.SectionFor(SlotStaff, SectionPerformance)
	if !ok {
		t.Fatal("legacy widget-only variant should resolve unknown section by fallback")
	}
	if sectionFromUnknown.Key != SectionOverview {
		t.Fatalf("legacy widget-only section fallback must still be overview: got=%s", sectionFromUnknown.Key)
	}
}

func TestSectionForInertReturnsNoSection(t *testing.T) {
	var o Options
	_, _, ok := o.SectionFor(0, SectionOverview)
	if ok {
		t.Fatal("inert options must return ok=false for SectionFor")
	}
}

func TestVariantForSectionUnknownWidgetsAreFilteredForCapability(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotStaff: {
			Sections: []Section{
				{Key: SectionOverview, Widgets: []WidgetRef{WidgetRef("made_up_widget"), WidgetCompletionSummary}},
			},
		},
		SlotOperatorStaff: {
			Sections: []Section{
				{Key: SectionOverview, Widgets: []WidgetRef{WidgetRef("made_up_widget"), WidgetCompletionSummary, WidgetRef("also_unknown")}},
			},
		},
	}}
	sections, slot, ok := o.SectionsFor(0)
	if !ok || slot != SlotStaff {
		t.Fatalf("section-only capability ranking must filter unknowns: ok=%v slot=%d", ok, slot)
	}
	// both variants are filtered to exactly one effective widget; tie breaks toward
	// the higher slot.
	if len(sections[0].Widgets) != 1 || sections[0].Widgets[0] != WidgetCompletionSummary {
		t.Fatalf("unknown widget refs must be filtered before ranking: got=%v", sections[0].Widgets)
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

func TestWidgetAndSectionsLegacyCompatibility(t *testing.T) {
	o := Options{Variants: map[Slot]Variant{
		SlotStaff: {
			Widgets:  []WidgetRef{WidgetCompletionSummary, WidgetQuickLinks, WidgetRef("made_up_widget")},
			Sections: []Section{},
		},
	}}
	v, _, ok := o.VariantFor(SlotStaff)
	if !ok {
		t.Fatal("legacy widget-only variant must resolve exactly")
	}
	if got := v.KnownWidgets(); len(got) != 2 || got[0] != WidgetCompletionSummary || got[1] != WidgetQuickLinks {
		t.Fatalf("legacy widget-only variant must stay compatible: %v", got)
	}
	sections, slot, ok := o.SectionsFor(SlotStaff)
	if !ok {
		t.Fatal("legacy widget-only variant must still resolve sections")
	}
	if slot != SlotStaff {
		t.Fatalf("legacy widget-only variant must retain slot: got=%d", slot)
	}
	if len(sections) != 1 || sections[0].Key != SectionOverview || len(sections[0].Widgets) != 2 {
		t.Fatalf("legacy widget-only variant must synthesize overview section for compatibility: %v", sections)
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
	for _, ref := range []WidgetRef{
		WidgetCompletionSummary,
		WidgetApprovalProgress,
		WidgetAttention,
		WidgetReviewPipeline,
		WidgetPerformanceTable,
	} {
		b := WidgetBundle(ref)
		if len(b) != 1 || b[0] != "outcome_completion:read" {
			t.Fatalf("%s bundle must be exactly the dedicated Q7 code, got %v", ref, b)
		}
	}
	if len(WidgetBundle(WidgetQuickLinks)) != 0 {
		t.Fatal("quick_links performs no data read — bundle must be empty")
	}
}
