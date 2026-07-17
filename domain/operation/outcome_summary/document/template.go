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
//     KEPT (registered version); superseded as the operative block artifact by
//     v3.
//   - report-card-template-v3.docx (v3, gen_template_v3.py) — the SAME block
//     layout as v2 against an earlier generic placeholder contract
//     ({{#primary_jobs}}, {{phase_N_scaled_label}}, …). FROZEN: its emissions
//     were superseded, so no selector renders it (see TemplateV3, Deprecated).
//   - outcome-summary-template-block.docx (gen_template_block.py) — the block
//     layout against the CONVERGED generic, proto-grounded placeholder contract
//     (one body loop {{#job_categories.<code>.jobs}}, per-phase leaves
//     {{job_template_phases.<phase>.…}}, dot-path {{client_attributes.<code>}},
//     and the wired attendance grid). This is the embedded artifact the block
//     TemplateVariant now selects. Its companion
//     outcome-summary-template-block.manifest.json enumerates every scalar and
//     loop path the artifact references, so the builder can seed the referenced
//     tree blank before overlay (the engine leaks unresolved leaves verbatim).
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

//go:embed report-card-template-v3.docx
var reportCardTemplateV3 []byte

//go:embed outcome-summary-template-block.docx
var reportCardTemplateBlock []byte

//go:embed outcome-summary-template-block.manifest.json
var reportCardTemplateBlockManifest []byte

// Template returns the ORIGINAL v1 summary-layout template bytes — the
// package-wide zero-option fallback (tiers that configure nothing keep their
// exact prior document), and the registered version-1 artifact for the
// existing binding row.
func Template() []byte { return reportCardTemplateV1 }

// TemplateV1 is the explicit-name alias for the original v1 artifact.
func TemplateV1() []byte { return reportCardTemplateV1 }

// TemplateV2 returns the v2 faithful block-layout template bytes. KEPT as the
// registered v2 artifact; no longer the block-variant fallback (v3 supersedes
// it). Its content is school-specific operator material — it must never become
// another tier's implicit fallback.
func TemplateV2() []byte { return reportCardTemplateV2 }

// Deprecated: the v3 placeholder emissions were superseded by the converged
// contract, and no selector renders these bytes any longer (the block
// TemplateVariant selects TemplateBlock). The frozen bytes and this accessor are
// retained only as the registered v3 artifact; do not wire new callers to it.
func TemplateV3() []byte { return reportCardTemplateV3 }

// TemplateBlock returns the converged generic-variable block-layout template
// bytes — one body loop {{#job_categories.<code>.jobs}}, per-phase leaves, the
// singleton homeroom projection, and the wired attendance grid. Selected as the
// embedded artifact where the app opts in via DocumentOptions.TemplateVariant ==
// TemplateVariantBlock. Its content is school-specific operator material — it
// must never become another tier's implicit fallback.
func TemplateBlock() []byte { return reportCardTemplateBlock }

// ManifestBlock returns the blank-guard manifest JSON that accompanies
// TemplateBlock: every scalar path and loop path the artifact references,
// generated from the same profile as the DOCX. The builder consumes it to seed
// the referenced tree blank before overlay (the engine leaks unresolved leaves
// verbatim), so no handwritten placeholder inventory is maintained.
func ManifestBlock() []byte { return reportCardTemplateBlockManifest }
