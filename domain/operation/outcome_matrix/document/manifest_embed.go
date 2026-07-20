// Package document builds the grade-sheet (outcome-matrix) PDF data map — the
// MMIS-parity composite gradesheet (roster × per-period finals + year Final,
// decisions.html Q1 LOCKED). It renders through the fycha `doctemplate` engine,
// but ONLY via injected closures (generateDoc / generatePDF) threaded through the
// block Infra — this package, like the rest of fayna, never imports fycha in
// non-test code (the architecture boundary; see block/engineblock.go).
//
// Asymmetry with outcome_summary/document (report cards), BY DESIGN (Q1 /
// entities.html §5): the report-card package go:embed's its .docx artifacts and
// falls back to them when no operator binding resolves. The grade sheet does the
// OPPOSITE — a resolver miss FAILS LOUD ("no template configured"), so the .docx
// is an authoring asset operators upload via the Grade Sheet Templates settings
// page and is deliberately NOT embedded here. ONLY the manifest is embedded: it
// is a small, static blank-seed contract (every placeholder path the academic
// artifact references), generated alongside the .docx by gen_template_sheet.py.
// The builder seeds every manifest path blank before overlay so a missing cell
// renders blank rather than leaking a token (the engine's residual scrub is the
// backstop; the manifest is the guarantee — fycha.md §1/§4).
package document

import _ "embed"

//go:embed grade-sheet-template-academic.manifest.json
var academicSheetManifest []byte

// AcademicManifest returns the blank-guard manifest JSON for the academic
// grade-sheet artifact: every root scalar path, the one {{#students}} loop, and
// its per-item scalar paths, generated from the same profile as the .docx. The
// builder consumes it to seed the referenced tree blank before overlay, so no
// handwritten placeholder inventory is maintained (idiom 2, fycha.md §4).
func AcademicManifest() []byte { return academicSheetManifest }
