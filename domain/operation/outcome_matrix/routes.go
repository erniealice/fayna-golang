package outcome_matrix

// routes.go — OutcomeMatrix route constants and Routes config struct.
//
// The matrix is scoped to one job_template ({id} == job_template_id), mirroring
// grade_sheet.go's subjectFilter. MatrixURL serves both the full page and the
// HTMX content partial; RecordURL is the batch-save POST target.
//
// SECTION-SCOPED SIBLINGS (20260725): a template can span several sections — the
// AY2026-27 grade-scoped generation fans out up to 4 ways, so "Arts — Grade 10"
// renders Palladium + Platinum + Tantalum as one 87-student sheet and all three
// courses-list rows link to the same URL. The Group* routes add the missing
// (template, section) grain. They are ADDITIVE: the template-scoped routes above
// keep their exact current behaviour, so every existing bookmark, redirect and
// e2e selector still resolves and no golden entry mutates.
//
// {group_id} == subscription_group_id. The placeholder name is IDENTICAL across
// the generic and education tiers — route.json overrides the whole path string,
// and the mux registers the overridden string while the handler reads
// r.PathValue("group_id"). Education renames only the visible segment
// ("subscription-group" -> "section"), exactly as outcome_summary's ClientCardURL
// renames "client" -> "student" while keeping {client_id}.

const (
	MatrixURL = "/outcome-matrix/{id}"
	RecordURL = "/action/outcome-matrix/{id}/record"

	// GroupMatrixURL narrows MatrixURL to a single section. The handler MUST
	// validate the (template, section) PAIR, not merely apply the filter: the
	// adapter's section predicate carries no academic-year term and students hold
	// active membership in a section in every year, so a foreign-AY section id
	// yields a plausible non-empty PARTIAL roster rather than an empty one.
	GroupMatrixURL = "/outcome-matrix/{id}/subscription-group/{group_id}"

	// Sheet-level CSV download. A GET sibling of MatrixURL (NOT under /action/*
	// — safe method, no state change, so neither the CSRF hook nor the
	// action-workspace signature guard applies; the raw-handler registration
	// wraps it with the ViewAdapter's RBAC context, same as outcome_summary's
	// SectionExportURL). Honors the same ?scope= and ?hide= as the HTML view.
	ExportURL = "/outcome-matrix/{id}/export"

	// Section-scoped siblings of ExportURL / DownloadDrawerURL. Same trust model
	// as their template-scoped originals (see each below).
	//
	// There is deliberately NO section form of RecordURL or NarrativeURL. Those
	// are cell-grain POSTs: the cells they touch are identified by outcome id and
	// already gated on the acting staff's ownership by resolveCellAuthority, so a
	// section page posting to the template-scoped action is neither wider nor
	// weaker — it is the same authority check over the same cells. Adding section
	// forms would only duplicate the gate and give it a second place to rot.
	GroupExportURL         = "/outcome-matrix/{id}/subscription-group/{group_id}/export"
	GroupDownloadDrawerURL = "/action/outcome-matrix/{id}/subscription-group/{group_id}/download"

	// Export drawer (20260720 Q3): the GET view rendering the download form
	// (period + format selects + hidden scope/hide) into #sheetContent. It sits
	// under /action/* for slug tidiness, but it is a GET (safe method): the CSRF
	// hook and the action-workspace signature guard constrain NON-safe methods
	// only, so a GET drawer needs no signed form (same trust model as the raw
	// GET ExportURL). Registered with r.GET, not the action-POST path.
	DownloadDrawerURL = "/action/outcome-matrix/{id}/download"

	// Per-phase approval transition POST targets ({id} == job_template_id; the
	// job_template_phase_id rides in the form body). Each is a real, query-free
	// HTMX POST form signed by {{actionForm}} over its exact resolved path (the
	// landed Q7 rowActionTokens table path does NOT cover these bar buttons —
	// codex fresh finding). All live under /action/* so the CSRF + action-
	// workspace guards apply exactly as for RecordURL.
	SubmitURL  = "/action/outcome-matrix/{id}/submit"
	VerifyURL  = "/action/outcome-matrix/{id}/verify"
	PublishURL = "/action/outcome-matrix/{id}/publish"
	ReturnURL  = "/action/outcome-matrix/{id}/return"

	// NarrativeURL is the per-cell narrative drawer (N-1 LOCKED 2026-07-23): a
	// dedicated route serving GET (render the drawer form, editability resolved
	// SERVER-SIDE by the shared authority core) + POST (save the cell's
	// determination_note through the task_outcome:update use case). ONE path, two
	// verbs — the TemplateUploadURL precedent, NOT an extension of the value
	// record protocol. {id} == job_template_id (same sheet scope as RecordURL);
	// the target cell's outcome_id rides in ?outcome_id= on the GET and in the
	// signed form body on the POST (like the approval bar's job_template_phase_id).
	//
	// Path shape (C8 ServeMux): the trailing literal segment "narrative" keeps it
	// unambiguous against every sibling route — it is a distinct final segment
	// from the {id}/{record,submit,verify,publish,return} verbs, and the /action/
	// prefix + {id} placeholder never collides with the literal
	// /action/outcome-matrix/templates/* family (there is no templates/narrative).
	// Under /action/* so the POST inherits the CSRF validator + signed
	// action-workspace guard exactly as RecordURL does; the GET is a safe method
	// (those guards constrain non-safe methods only), same trust model as the
	// DownloadDrawerURL GET. Education overrides the display slug via route.json.
	NarrativeURL = "/action/outcome-matrix/{id}/narrative"

	// TemplateSettingsURL is the standalone grade-sheet template management page
	// (Wave C / P4): a list of workspace→document_template bindings (scoped by
	// job_category + optional price_schedule) + upload/publish/delete. A dedicated
	// settings surface (Q5), NOT a tab on the grid. GET only — no mutation, so it
	// stays OUTSIDE /action/ (safe method). Education overrides it to
	// /grade-sheet/templates via education/route.json. Mirrors the JOSDT
	// outcome_summary.TemplateSettingsURL shape.
	TemplateSettingsURL = "/outcome-matrix/templates"
	// The three template MUTATIONS live under /action/* so they inherit the CSRF
	// validator + signed workspace-form guard (both default-scoped to /action/ in
	// espyna's middleware chain). Registering them elsewhere silently bypasses both
	// guards (JOSDT B4 precedent). Education keeps its /grade-sheet/templates
	// display vocabulary UNDER the /action/ prefix (see education/route.json).
	//
	// TemplateUploadURL is the upload drawer (GET = form, POST = create a DRAFT
	// binding + its document_template artifact). Education →
	// /action/grade-sheet/templates/upload.
	TemplateUploadURL = "/action/outcome-matrix/templates/upload"
	// TemplatePublishURL publishes a DRAFT binding (id in ?id= query, appended by
	// the table row-action JS) via the controlled publish transaction. Flat (no
	// path param) so it composes with the generic "activate" row action.
	// Education → /action/grade-sheet/templates/publish.
	TemplatePublishURL = "/action/outcome-matrix/templates/publish"
	// TemplateDeleteURL deletes a DRAFT binding (POST, id in form). Education →
	// /action/grade-sheet/templates/delete.
	TemplateDeleteURL = "/action/outcome-matrix/templates/delete"
)

// Routes holds all route paths for the outcome matrix view.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	MatrixURL         string `json:"matrix_url"`
	RecordURL         string `json:"record_url"`
	ExportURL         string `json:"export_url"`
	DownloadDrawerURL string `json:"download_drawer_url"`

	// Section-scoped siblings — same handlers, one extra path parameter. The
	// approval transitions below deliberately have NO section form: approval
	// state lives on job_phase (per student job), so section-grain approval is
	// representable, but narrowing it needs the four RPCs to carry a section and
	// six espyna helpers to stop keying on (templateID, phaseID). Until then the
	// section page renders its approval band READ-ONLY rather than offering a
	// Submit that silently flips every section on the template.
	GroupMatrixURL         string `json:"group_matrix_url"`
	GroupExportURL         string `json:"group_export_url"`
	GroupDownloadDrawerURL string `json:"group_download_drawer_url"`

	// Per-phase approval transition routes.
	SubmitURL  string `json:"submit_url"`
	VerifyURL  string `json:"verify_url"`
	PublishURL string `json:"publish_url"`
	ReturnURL  string `json:"return_url"`

	// Per-cell narrative drawer (GET form + POST save on one path).
	NarrativeURL string `json:"narrative_url"`

	// Grade-sheet template settings (P4 management surface).
	TemplateSettingsURL string `json:"template_settings_url"`
	TemplateUploadURL   string `json:"template_upload_url"`
	TemplatePublishURL  string `json:"template_publish_url"`
	TemplateDeleteURL   string `json:"template_delete_url"`
}

// DefaultRoutes returns a Routes populated from the package-level route
// constants. ActiveNav "job" matches task_outcome / grade_sheet.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "outcome-matrix",

		MatrixURL:         MatrixURL,
		RecordURL:         RecordURL,
		ExportURL:         ExportURL,
		DownloadDrawerURL: DownloadDrawerURL,

		GroupMatrixURL:         GroupMatrixURL,
		GroupExportURL:         GroupExportURL,
		GroupDownloadDrawerURL: GroupDownloadDrawerURL,

		SubmitURL:  SubmitURL,
		VerifyURL:  VerifyURL,
		PublishURL: PublishURL,
		ReturnURL:  ReturnURL,

		NarrativeURL: NarrativeURL,

		TemplateSettingsURL: TemplateSettingsURL,
		TemplateUploadURL:   TemplateUploadURL,
		TemplatePublishURL:  TemplatePublishURL,
		TemplateDeleteURL:   TemplateDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_matrix.matrix":          r.MatrixURL,
		"outcome_matrix.record":          r.RecordURL,
		"outcome_matrix.export":          r.ExportURL,
		"outcome_matrix.download_drawer": r.DownloadDrawerURL,
		"outcome_matrix.submit":          r.SubmitURL,
		"outcome_matrix.verify":          r.VerifyURL,
		"outcome_matrix.publish":         r.PublishURL,
		"outcome_matrix.return":          r.ReturnURL,
		"outcome_matrix.narrative":       r.NarrativeURL,

		"outcome_matrix.group_matrix":          r.GroupMatrixURL,
		"outcome_matrix.group_export":          r.GroupExportURL,
		"outcome_matrix.group_download_drawer": r.GroupDownloadDrawerURL,

		"outcome_matrix.template_settings": r.TemplateSettingsURL,
		"outcome_matrix.template_upload":   r.TemplateUploadURL,
		"outcome_matrix.template_publish":  r.TemplatePublishURL,
		"outcome_matrix.template_delete":   r.TemplateDeleteURL,
	}
}
