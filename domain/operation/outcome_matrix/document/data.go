package document

// data.go — buildSheetData: the grade-sheet (outcome-matrix) PDF data map.
//
// The document shape is the MMIS-parity composite gradesheet (Q1 LOCKED): a
// roster grid, one row per student, columns = per-period finals + the year Final.
// The artifact (grade-sheet-template-academic.docx) references a fixed placeholder
// contract — root scalars (title/subtitle/header cells/footer) + ONE table-row
// loop {{#students}} whose item scalars are {name, period<N>, final}. Column
// header WORDING is render-time data (period<N>_label, final_label, name_label),
// so a school's term names never bake into the artifact.
//
// The engine (fycha.md §1) stringifies with %v only and leaks nothing but blanks
// (its residual scrub). buildSheetData therefore (a) pre-formats every value to a
// string and (b) seeds EVERY manifest-declared path blank before overlaying the
// real values — the same manifest blank-guard the report-card block tree uses
// (outcome_summary/document/tree.go:92-153), cloned here against THIS package's
// embedded academic manifest. A student missing a period renders an empty cell,
// never a leaked {{period2}} token.

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

// SheetHeader carries the once-per-render root context: the pre-resolved lyngua
// title/labels + the delivery-group section, academic year, and the print stamp.
// Every field is a finished display string (the builder does no lookups).
type SheetHeader struct {
	Title        string // sheet_title (the page title label, e.g. "Grade Sheet")
	SectionName  string // section_name (delivery-group / section name)
	AcademicYear string // academic_year (price_schedule name)
	NameLabel    string // name_label (roster column header, e.g. "Student")
	FinalLabel   string // final_label (year-final column header, e.g. "Final")
	PrintedBy    string // printed_by (acting user's display name / id)
	PrintedAt    string // printed_at (formatted timestamp)
}

// SheetStudent is one roster row: the display name, the per-period final labels
// in sequence order (positional → period1, period2, …), and the stored year
// final (D8: read verbatim, never recomputed). Any slot may be "" → blank cell.
type SheetStudent struct {
	Name        string
	PhaseFinals []string
	YearFinal   string
}

// BuildSheetData assembles the grade-sheet data map. periodLabels are the ordered
// per-period column headers (sequence order); students carry per-period finals in
// the SAME order. The mapping is positional (slot i → period{i+1}); the academic
// artifact prints two period columns, so extra slots are harmless (the engine
// ignores unreferenced keys) and short slots are blank-seeded by the manifest.
// Pure and DB-free — the caller (export.go's PDF branch) does all the fetching.
func BuildSheetData(header SheetHeader, periodLabels []string, students []SheetStudent) map[string]any {
	data := map[string]any{
		"sheet_title":   header.Title,
		"section_name":  header.SectionName,
		"academic_year": header.AcademicYear,
		"name_label":    header.NameLabel,
		"final_label":   header.FinalLabel,
		"printed_by":    header.PrintedBy,
		"printed_at":    header.PrintedAt,
	}
	for i, lbl := range periodLabels {
		data[periodKey("period", i)+"_label"] = lbl
	}

	items := make([]any, 0, len(students))
	for _, s := range students {
		item := map[string]any{
			"name":  s.Name,
			"final": s.YearFinal,
		}
		for i, v := range s.PhaseFinals {
			item[periodKey("period", i)] = v
		}
		items = append(items, item)
	}
	data["students"] = items

	// Blank-seed every manifest-declared path so a missing scalar / short row
	// renders blank instead of leaking its {{token}} verbatim.
	applyAcademicManifest(data)
	return data
}

// periodKey builds the positional per-period key: prefix + "1" for slot 0, etc.
// (period1_label / period1, period2_label / period2, …) — the same 1-based naming
// the generator emits.
func periodKey(prefix string, slot int) string {
	return prefix + strconv.Itoa(slot+1)
}

// --- manifest blank-guard (cloned from outcome_summary/document/tree.go) -----

// manifestNode is one loop scope in the sheet manifest: the item-relative scalar
// paths the artifact references and its own nested loops. The root has the same
// shape (its scalars are absolute paths).
type manifestNode struct {
	Scalars []string                `json:"scalars"`
	Loops   map[string]manifestNode `json:"loops"`
}

var (
	sheetManifestOnce   sync.Once
	sheetManifestParsed manifestNode
)

// loadAcademicManifest parses the embedded academic manifest exactly once. On a
// parse failure it logs and leaves an empty manifest (the blank-guard degrades to
// a no-op rather than panicking a render).
func loadAcademicManifest() manifestNode {
	sheetManifestOnce.Do(func() {
		raw := AcademicManifest()
		if len(raw) == 0 {
			return
		}
		if err := json.Unmarshal(raw, &sheetManifestParsed); err != nil {
			log.Printf("grade sheet doc: academic manifest parse: %v", err)
			sheetManifestParsed = manifestNode{}
		}
	})
	return sheetManifestParsed
}

// applyAcademicManifest seeds EVERY manifest-referenced path in data with a blank
// leaf (scalars) or an empty list (loops), building nested maps as needed, then
// recurses into each real loop item so per-item scalars are seeded too. Real
// values already present are never clobbered.
func applyAcademicManifest(data map[string]any) {
	conformToManifest(data, loadAcademicManifest())
}

func conformToManifest(m map[string]any, node manifestNode) {
	for _, path := range node.Scalars {
		ensureBlankLeaf(m, strings.Split(path, "."))
	}
	for loopPath, child := range node.Loops {
		list := ensureList(m, strings.Split(loopPath, "."))
		for _, it := range list {
			if im, ok := it.(map[string]any); ok {
				conformToManifest(im, child)
			}
		}
	}
}

// ensureBlankLeaf walks segs from m, creating intermediate maps, and sets a blank
// string leaf iff the final key is absent (never overwrites a real value).
func ensureBlankLeaf(m map[string]any, segs []string) {
	if len(segs) == 0 {
		return
	}
	if len(segs) == 1 {
		if _, ok := m[segs[0]]; !ok {
			m[segs[0]] = ""
		}
		return
	}
	child, ok := m[segs[0]].(map[string]any)
	if !ok {
		child = map[string]any{}
		m[segs[0]] = child
	}
	ensureBlankLeaf(child, segs[1:])
}

// ensureList walks segs from m (creating intermediate maps) and guarantees the
// final key holds a []any, seeding an empty list when absent. Returns the list so
// callers can recurse into real items.
func ensureList(m map[string]any, segs []string) []any {
	if len(segs) == 0 {
		return nil
	}
	cur := m
	for _, s := range segs[:len(segs)-1] {
		child, ok := cur[s].(map[string]any)
		if !ok {
			child = map[string]any{}
			cur[s] = child
		}
		cur = child
	}
	last := segs[len(segs)-1]
	lst, ok := cur[last].([]any)
	if !ok {
		lst = []any{}
		cur[last] = lst
	}
	return lst
}

// fmtValue is the %v-only stringify contract in one place (the engine does the
// same). Exposed for the caller to normalize any non-string cell before it enters
// the data map, keeping the "pre-formatted strings only" invariant explicit.
func fmtValue(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
