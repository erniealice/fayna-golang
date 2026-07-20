package document

import "testing"

// TestBuildSheetData_ManifestSeeding proves the blank-guard: EVERY manifest-
// declared root scalar and per-item scalar is present after BuildSheetData, even
// when the caller supplies nothing for it, while real values are overlaid intact.
func TestBuildSheetData_ManifestSeeding(t *testing.T) {
	// A header with two empty fields (SectionName, PrintedBy) + one student that
	// only carries a first-period final (period2 absent) — the exact "missing
	// cell" shapes the engine would otherwise leak.
	header := SheetHeader{
		Title:        "Grade Sheet",
		AcademicYear: "AY 2025-2026",
		NameLabel:    "Student",
		FinalLabel:   "Final",
		PrintedAt:    "2026-07-20 09:30",
	}
	periodLabels := []string{"Semester 1"} // only one period label supplied
	students := []SheetStudent{
		{Name: "Cruz, Juan", PhaseFinals: []string{"6"}, YearFinal: "6"},
	}
	data := BuildSheetData(header, periodLabels, students)

	// Every root scalar the manifest declares must exist (blank where unset).
	rootScalars := []string{
		"sheet_title", "section_name", "academic_year", "name_label",
		"period1_label", "period2_label", "final_label", "printed_by", "printed_at",
	}
	for _, k := range rootScalars {
		v, ok := data[k]
		if !ok {
			t.Errorf("root scalar %q absent (manifest not seeded)", k)
			continue
		}
		if _, isStr := v.(string); !isStr {
			t.Errorf("root scalar %q is %T, want string (%%v-only engine)", k, v)
		}
	}

	// Real values overlaid.
	if data["sheet_title"] != "Grade Sheet" {
		t.Errorf("sheet_title = %v, want Grade Sheet", data["sheet_title"])
	}
	if data["period1_label"] != "Semester 1" {
		t.Errorf("period1_label = %v, want Semester 1", data["period1_label"])
	}
	// Unsupplied fields blank-seeded, not missing.
	if data["section_name"] != "" {
		t.Errorf("section_name = %v, want blank", data["section_name"])
	}
	if data["printed_by"] != "" {
		t.Errorf("printed_by = %v, want blank", data["printed_by"])
	}
	// The unsupplied period2 header is blank-seeded (never a leaked token).
	if data["period2_label"] != "" {
		t.Errorf("period2_label = %v, want blank", data["period2_label"])
	}

	// The students loop: exactly one item, every manifest item scalar present.
	students0, ok := data["students"].([]any)
	if !ok {
		t.Fatalf("students is %T, want []any", data["students"])
	}
	if len(students0) != 1 {
		t.Fatalf("students has %d items, want 1", len(students0))
	}
	item, ok := students0[0].(map[string]any)
	if !ok {
		t.Fatalf("student item is %T, want map[string]any", students0[0])
	}
	for _, k := range []string{"name", "period1", "period2", "final"} {
		if _, ok := item[k]; !ok {
			t.Errorf("student item scalar %q absent (manifest not seeded)", k)
		}
	}
	if item["name"] != "Cruz, Juan" {
		t.Errorf("student name = %v, want Cruz, Juan", item["name"])
	}
	if item["period1"] != "6" {
		t.Errorf("student period1 = %v, want 6", item["period1"])
	}
	// period2 not supplied → blank-seeded.
	if item["period2"] != "" {
		t.Errorf("student period2 = %v, want blank", item["period2"])
	}
	if item["final"] != "6" {
		t.Errorf("student final = %v, want 6", item["final"])
	}
}

// TestBuildSheetData_PositionalMapping proves phase finals map positionally to
// period1/period2 in sequence order (slot 0 → period1).
func TestBuildSheetData_PositionalMapping(t *testing.T) {
	data := BuildSheetData(
		SheetHeader{Title: "GS"},
		[]string{"Sem 1", "Sem 2"},
		[]SheetStudent{{Name: "A", PhaseFinals: []string{"1", "2"}, YearFinal: "2"}},
	)
	if data["period1_label"] != "Sem 1" || data["period2_label"] != "Sem 2" {
		t.Errorf("period labels mis-mapped: p1=%v p2=%v", data["period1_label"], data["period2_label"])
	}
	item := data["students"].([]any)[0].(map[string]any)
	if item["period1"] != "1" || item["period2"] != "2" {
		t.Errorf("period cells mis-mapped: p1=%v p2=%v", item["period1"], item["period2"])
	}
}

// TestFmtValue covers the %v-only stringify contract helper.
func TestFmtValue(t *testing.T) {
	cases := map[string]struct {
		in   any
		want string
	}{
		"nil":    {nil, ""},
		"string": {"x", "x"},
		"int":    {7, "7"},
		"float":  {6.5, "6.5"},
	}
	for name, c := range cases {
		if got := fmtValue(c.in); got != c.want {
			t.Errorf("%s: fmtValue(%v) = %q, want %q", name, c.in, got, c.want)
		}
	}
}
