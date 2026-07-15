// Package document renders a student's report card as a .docx through the
// fycha `doctemplate` engine (injected as a GenerateDoc closure — fayna does
// NOT import fycha). The template is authored by gen_template.py (co-located)
// and embedded here as the built-in fallback, mirroring the centymo invoice
// download precedent (invoice_download.go embeds its template and treats a
// storage-backed template as an optional override).
//
// The render honors the split-source contract (plan 20260713-report-card-documents
// §Render Contract): semester bands are the recompute-equivalent
// phase_outcome_summary.scaled_label; the MYP Overall / year-final is the STORED
// job_outcome_summary.scaled_label (never recomputed). See data.go.
package document

import _ "embed"

//go:embed report-card-template.docx
var reportCardTemplate []byte

// Template returns the embedded report-card .docx template bytes.
func Template() []byte { return reportCardTemplate }
