// Package document renders a student's report card as a .docx through the
// fycha `doctemplate` engine (injected as a GenerateDoc closure — fayna does
// NOT import fycha). Two template artifacts ship embedded:
//
//   - report-card-template.docx (v1, gen_template.py) — the ORIGINAL compact
//     summary layout. Kept intact as the registered version-1 artifact (the
//     education1 binding row rc-tmpl-education1-001 points at these bytes);
//     never deleted or regenerated.
//   - report-card-template-v2.docx (v2, gen_template_v2.py) — the FAITHFUL
//     block layout rebuilt page-accurate from the operator's live printed card
//     (spec: docs/plan/20260713-report-card-documents/codex-pdf-spec.md).
//     This is the embedded DEFAULT the download handler renders.
//
// An operator-uploaded, AY-scoped binding template (ResolveTemplateBytes)
// still overrides the embedded default, mirroring the centymo invoice
// precedent (invoice_download.go).
//
// The render honors the split-source contract (plan 20260713-report-card-documents
// §Render Contract): semester bands are the recompute-equivalent
// phase_outcome_summary.scaled_label; the MYP Overall / year-final is the STORED
// job_outcome_summary.scaled_label (never recomputed). See data.go.
package document

import _ "embed"

//go:embed report-card-template.docx
var reportCardTemplateV1 []byte

//go:embed report-card-template-v2.docx
var reportCardTemplateV2 []byte

// Template returns the embedded default report-card .docx template bytes
// (the v2 faithful block layout).
func Template() []byte { return reportCardTemplateV2 }

// TemplateV1 returns the ORIGINAL v1 summary-layout template bytes — the
// registered version-1 artifact, kept for the existing binding row and for
// operators who re-publish the original layout.
func TemplateV1() []byte { return reportCardTemplateV1 }
