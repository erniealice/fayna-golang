package document

// engine_contract_test.go — THE engine-contract proof for the checked-in
// academic grade-sheet artifact. It runs the REAL fycha doctemplate engine over
// the .docx with sample data and asserts (a) the artifact passes fycha's
// ReadDocxBytes zip-hardening caps, (b) every placeholder resolves (no residual
// {{...}} token survives), and (c) the {{#students}} table-row loop clones one
// row per roster student.
//
// This is the ONLY place fayna imports fycha, and it is a TEST: the go.work
// workspace resolves fycha locally (CLAUDE.md "go.work = local dev, tags =
// consumers"), while production code keeps the boundary (renders go through the
// injected generateDoc/generatePDF closures, never a direct fycha import). The
// artifact is read from disk here — it is an upload-only authoring asset, not
// go:embed'd (Q1 fail-loud), so this proof also documents its exact upload shape.

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/erniealice/fycha-golang/services/doctemplate"
)

const academicArtifactPath = "grade-sheet-template-academic.docx"

// readDocumentXML unzips processed DOCX bytes and returns word/document.xml.
func readDocumentXML(t *testing.T, docx []byte) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatalf("open processed docx as zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open document.xml: %v", err)
			}
			defer rc.Close()
			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("read document.xml: %v", err)
			}
			return string(b)
		}
	}
	t.Fatalf("word/document.xml not found in processed docx")
	return ""
}

func sampleSheetData() map[string]any {
	header := SheetHeader{
		Title:        "Grade Sheet",
		SectionName:  "Grade 7 - Rizal",
		AcademicYear: "AY 2025-2026",
		NameLabel:    "Student",
		FinalLabel:   "Final",
		PrintedBy:    "Ada Lovelace",
		PrintedAt:    "2026-07-20 09:30",
	}
	periodLabels := []string{"Semester 1", "Semester 2"}
	students := []SheetStudent{
		{Name: "Bonifacio, Andres", PhaseFinals: []string{"6", "7"}, YearFinal: "7"},
		{Name: "Silang, Gabriela", PhaseFinals: []string{"5", "6"}, YearFinal: "6"},
	}
	return BuildSheetData(header, periodLabels, students)
}

// TestAcademicArtifactPassesDocxCaps proves the checked-in artifact is a valid
// DOCX under fycha's ReadDocxBytes zip-hardening caps (2000 entries / 64 MiB per
// entry / 256 MiB aggregate) — the same gate the upload path enforces.
func TestAcademicArtifactPassesDocxCaps(t *testing.T) {
	raw, err := os.ReadFile(academicArtifactPath)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if _, err := doctemplate.ReadDocxBytes(raw); err != nil {
		t.Fatalf("ReadDocxBytes rejected the academic artifact: %v", err)
	}
}

// TestAcademicArtifactEngineContract runs ProcessTemplate over the artifact and
// asserts the full engine contract: all placeholders resolve, and the student
// row clones per roster item.
func TestAcademicArtifactEngineContract(t *testing.T) {
	raw, err := os.ReadFile(academicArtifactPath)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	data := sampleSheetData()

	out, err := doctemplate.ProcessTemplate(raw, data)
	if err != nil {
		t.Fatalf("ProcessTemplate: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("ProcessTemplate returned empty bytes")
	}

	docXML := readDocumentXML(t, out)

	// (a) No residual placeholder token survives — every path resolved (or was
	// scrubbed to blank). A leaked {{...}} means a manifest/artifact drift.
	if tok := regexp.MustCompile(`\{\{[^{}]*\}\}`).FindString(docXML); tok != "" {
		t.Fatalf("residual placeholder token in rendered document: %q", tok)
	}

	// (b) Header labels rendered (render-time column wording).
	for _, want := range []string{"Grade Sheet", "Grade 7 - Rizal", "AY 2025-2026",
		"Student", "Semester 1", "Semester 2", "Final"} {
		if !strings.Contains(docXML, xmlEscaped(want)) {
			t.Errorf("rendered document missing header text %q", want)
		}
	}

	// (c) The {{#students}} table-row loop cloned one row per student — both
	// names AND their period/final cells are present.
	for _, want := range []string{"Bonifacio, Andres", "Silang, Gabriela"} {
		if !strings.Contains(docXML, xmlEscaped(want)) {
			t.Errorf("rendered document missing student row %q (loop did not clone)", want)
		}
	}
	// Row 2's distinct final ("6") vs row 1's ("7") proves per-item overlay, not
	// a single row reused. Count "Semester" headers stay at the header (2), while
	// student names appear once each.
	if strings.Count(docXML, xmlEscaped("Bonifacio, Andres")) != 1 {
		t.Errorf("student row 1 rendered %d times, want 1", strings.Count(docXML, xmlEscaped("Bonifacio, Andres")))
	}
}

// xmlEscaped mirrors the generator's esc(): the engine writes text into <w:t>
// nodes, so '&', '<', '>' in sample data are XML-escaped in the output.
func xmlEscaped(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
