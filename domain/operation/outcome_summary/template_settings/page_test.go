package template_settings

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary_document_template"
	phasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	subscriptiongroupdocumentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"

	subscriptiongroupdocument "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/subscription_group_document"
)

// makeZip builds an in-memory ZIP from name→content entries. Entry order is not
// guaranteed (map iteration), which is fine for these structural assertions.
func makeZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %q: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("zip write %q: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// minimalDocx is the smallest archive that passes validateDocxArchive: both
// mandatory OOXML parts, safe paths, well under every cap.
func minimalDocx(t *testing.T) []byte {
	t.Helper()
	return makeZip(t, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0"?><Types/>`,
		"_rels/.rels":         `<?xml version="1.0"?><Relationships/>`,
		"word/document.xml":   `<?xml version="1.0"?><w:document/>`,
	})
}

func TestValidateDocxArchive_Valid(t *testing.T) {
	if err := validateDocxArchive(minimalDocx(t)); err != nil {
		t.Fatalf("expected a well-formed docx to pass, got: %v", err)
	}
}

func TestPeriodOptions_ByAcademicYearAndFailClosed(t *testing.T) {
	deps := &Deps{ListPhaseCodesByPriceSchedule: func(_ context.Context, req *phasepb.ListPhaseCodesByPriceScheduleRequest) (*phasepb.ListPhaseCodesByPriceScheduleResponse, error) {
		if req.GetPriceScheduleId() != "ay-1" {
			t.Fatalf("schedule = %q", req.GetPriceScheduleId())
		}
		return &phasepb.ListPhaseCodesByPriceScheduleResponse{Success: true, Options: []*phasepb.PhaseCodeOption{{Code: "s1", Names: []string{"Term 1"}}}}, nil
	}}
	opts, err := periodOptions(context.Background(), deps, "ay-1", "Full year")
	if err != nil || len(opts) != 2 || opts[0].Value != "" || opts[0].Label != "Full year" || opts[1].Value != "s1" || opts[1].Label != "Term 1" {
		t.Fatalf("options = %#v, err = %v", opts, err)
	}
	if !hasPeriodCode(opts, "s1") || hasPeriodCode(opts, "forged") {
		t.Fatal("phase membership check failed")
	}
	if _, err := periodOptions(context.Background(), &Deps{}, "ay-1", "Full year"); err == nil {
		t.Fatal("missing lookup should fail closed")
	}
}

func TestUploadAction_RejectsPhaseOutsideAcademicYearBeforeArtifact(t *testing.T) {
	rec := &uploadRecorder{}
	deps := rec.deps()
	deps.ListPhaseCodesByPriceSchedule = func(_ context.Context, _ *phasepb.ListPhaseCodesByPriceScheduleRequest) (*phasepb.ListPhaseCodesByPriceScheduleResponse, error) {
		return &phasepb.ListPhaseCodesByPriceScheduleResponse{Success: true, Options: []*phasepb.PhaseCodeOption{{Code: "s1", Names: []string{"Term 1"}}}}, nil
	}
	res := NewUploadAction(deps).Handle(permsCtx(bindingPermissionEntity+":create"), uploadPostWithFields(t, minimalDocx(t), map[string]string{"price_schedule_id": "ay-1", "job_template_phase_code": "forged"}))
	if res.StatusCode == http.StatusOK || len(rec.order) != 0 {
		t.Fatalf("invalid phase reached artifact create: status=%d, calls=%v", res.StatusCode, rec.order)
	}
}

func TestUploadAction_PeriodOptionsFragment(t *testing.T) {
	deps := &Deps{ListPhaseCodesByPriceSchedule: func(_ context.Context, _ *phasepb.ListPhaseCodesByPriceScheduleRequest) (*phasepb.ListPhaseCodesByPriceScheduleResponse, error) {
		return &phasepb.ListPhaseCodesByPriceScheduleResponse{Success: true, Options: []*phasepb.PhaseCodeOption{{Code: "s2", Names: []string{"Term 2"}}}}, nil
	}}
	deps.Labels.TemplateSettings.PeriodFullYear = "Full year"
	request := httptest.NewRequest(http.MethodGet, "/templates/upload?period_options=1&price_schedule_id=ay-1", nil)
	res := NewUploadAction(deps).Handle(permsCtx(bindingPermissionEntity+":create"), &view.ViewContext{Request: request})
	if res.StatusCode != http.StatusOK || res.Template != "outcome-summary-template-period-options" {
		t.Fatalf("fragment result: status=%d template=%q", res.StatusCode, res.Template)
	}
	form, ok := res.Data.(*UploadFormData)
	if !ok || len(form.PeriodOptions) != 2 || form.PeriodOptions[1].Value != "s2" {
		t.Fatalf("fragment options: %#v", res.Data)
	}
}

func TestValidateDocxArchive_NotAZip(t *testing.T) {
	if err := validateDocxArchive([]byte("this is definitely not a zip archive")); err == nil {
		t.Fatal("expected non-zip bytes to be rejected")
	}
}

func TestValidateDocxArchive_Empty(t *testing.T) {
	if err := validateDocxArchive(nil); err == nil {
		t.Fatal("expected empty bytes to be rejected")
	}
}

func TestValidateDocxArchive_MissingDocument(t *testing.T) {
	z := makeZip(t, map[string]string{
		"[Content_Types].xml": `<Types/>`,
		"word/styles.xml":     `<styles/>`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected an archive missing word/document.xml to be rejected")
	}
}

func TestValidateDocxArchive_MissingContentTypes(t *testing.T) {
	z := makeZip(t, map[string]string{
		"word/document.xml": `<w:document/>`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected an archive missing [Content_Types].xml to be rejected")
	}
}

func TestValidateDocxArchive_PathTraversal(t *testing.T) {
	z := makeZip(t, map[string]string{
		"[Content_Types].xml":  `<Types/>`,
		"word/document.xml":    `<w:document/>`,
		"../../etc/passwd.xml": `nope`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected a '..' traversal entry to be rejected")
	}
}

func TestValidateDocxArchive_AbsolutePath(t *testing.T) {
	z := makeZip(t, map[string]string{
		"[Content_Types].xml": `<Types/>`,
		"word/document.xml":   `<w:document/>`,
		"/abs/evil.xml":       `nope`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected an absolute-path entry to be rejected")
	}
}

func TestValidateDocxArchive_TooManyEntries(t *testing.T) {
	entries := map[string]string{
		"[Content_Types].xml": `<Types/>`,
		"word/document.xml":   `<w:document/>`,
	}
	for i := 0; i < maxArchiveEntries+1; i++ {
		entries[fmt.Sprintf("word/media/e%d.bin", i)] = "x"
	}
	if err := validateDocxArchive(makeZip(t, entries)); err == nil {
		t.Fatalf("expected an archive with > %d entries to be rejected", maxArchiveEntries)
	}
}

// ── Q4: permission-family alignment + upload-orphan cleanup ─────────────────

// permsCtx returns a request context carrying exactly the given permission
// codes (the same seam the ViewAdapter uses in production).
func permsCtx(codes ...string) context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(codes))
}

// uploadRecorder wires a Deps whose closures record call order + arguments and
// fail on demand, so the tests can pin the bytes-LAST ordering and the
// compensation contract.
type uploadRecorder struct {
	order []string

	failCreateDoc     bool
	failCreateBinding bool
	failUpload        bool
	listErr           error

	createdDocID      string
	createdBindingID  string
	createdSupersedes string
	deletedBindingID  string
	deletedDocID      string
	createdContainer  string
	createdStorageKey string
	uploadedContainer string
	uploadedKey       string

	bindings  []*bindingpb.JobOutcomeSummaryDocumentTemplate
	documents []*documenttemplatepb.DocumentTemplate
}

func (r *uploadRecorder) deps() *Deps {
	return &Deps{
		UploadTemplate: func(_ context.Context, container, key string, _ []byte, _ string) error {
			r.order = append(r.order, "upload")
			if r.failUpload {
				return errors.New("upload boom")
			}
			r.uploadedContainer = container
			r.uploadedKey = key
			return nil
		},
		CreateDocumentTemplate: func(_ context.Context, req *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error) {
			r.order = append(r.order, "create_doc")
			if r.failCreateDoc {
				return nil, errors.New("create doc boom")
			}
			r.createdDocID = req.GetData().GetId()
			r.createdContainer = req.GetData().GetStorageContainer()
			r.createdStorageKey = req.GetData().GetStorageKey()
			return &documenttemplatepb.CreateDocumentTemplateResponse{Success: true}, nil
		},
		DeleteDocumentTemplate: func(_ context.Context, req *documenttemplatepb.DeleteDocumentTemplateRequest) (*documenttemplatepb.DeleteDocumentTemplateResponse, error) {
			r.order = append(r.order, "delete_doc")
			r.deletedDocID = req.GetData().GetId()
			return &documenttemplatepb.DeleteDocumentTemplateResponse{Success: true}, nil
		},
		CreateTemplateBinding: func(_ context.Context, req *bindingpb.CreateJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.CreateJobOutcomeSummaryDocumentTemplateResponse, error) {
			r.order = append(r.order, "create_binding")
			if r.failCreateBinding {
				return nil, errors.New("create binding boom")
			}
			r.createdBindingID = "b-created"
			r.createdSupersedes = req.GetData().GetSupersedesBindingId()
			return &bindingpb.CreateJobOutcomeSummaryDocumentTemplateResponse{
				Data:    []*bindingpb.JobOutcomeSummaryDocumentTemplate{{Id: "b-created"}},
				Success: true,
			}, nil
		},
		DeleteTemplateBinding: func(_ context.Context, req *bindingpb.DeleteJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.DeleteJobOutcomeSummaryDocumentTemplateResponse, error) {
			r.order = append(r.order, "delete_binding")
			r.deletedBindingID = req.GetData().GetId()
			return &bindingpb.DeleteJobOutcomeSummaryDocumentTemplateResponse{Success: true}, nil
		},
		ListTemplateBindings: func(_ context.Context, _ *bindingpb.ListJobOutcomeSummaryDocumentTemplatesRequest) (*bindingpb.ListJobOutcomeSummaryDocumentTemplatesResponse, error) {
			if r.listErr != nil {
				return nil, r.listErr
			}
			return &bindingpb.ListJobOutcomeSummaryDocumentTemplatesResponse{Data: r.bindings, Success: true}, nil
		},
		ListDocumentTemplates: func(_ context.Context, _ *documenttemplatepb.ListDocumentTemplatesRequest) (*documenttemplatepb.ListDocumentTemplatesResponse, error) {
			return &documenttemplatepb.ListDocumentTemplatesResponse{Data: r.documents, Success: true}, nil
		},
	}
}

func uploadPost(t *testing.T, content []byte) *view.ViewContext {
	return uploadPostWithFields(t, content, nil)
}

func uploadPostWithFields(t *testing.T, content []byte, extra map[string]string) *view.ViewContext {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("name", "Test Template"); err != nil {
		t.Fatalf("write name field: %v", err)
	}
	for key, value := range extra {
		if err := mw.WriteField(key, value); err != nil {
			t.Fatalf("write %q field: %v", key, err)
		}
	}
	fw, err := mw.CreateFormFile("template_file", "report.docx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/settings/report-card-templates/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return &view.ViewContext{Request: req}
}

// TestListView_GatesOnBindingFamily locks the split-role alignment: the page
// gate cites the SAME code the list use case enforces
// (job_outcome_summary_document_template:list). Holding only the PARENT
// entity's code is no longer enough — and holding the binding code without the
// parent code is.
func TestListView_GatesOnBindingFamily(t *testing.T) {
	v := NewListView((&uploadRecorder{}).deps())

	res := v.Handle(permsCtx("job_outcome_summary:list"), &view.ViewContext{})
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("parent-entity-only role must be forbidden, got %d", res.StatusCode)
	}
	res = v.Handle(permsCtx("job_outcome_summary_document_template:list"), &view.ViewContext{})
	if res.StatusCode != http.StatusOK {
		t.Errorf("binding-family list role must see the page, got %d", res.StatusCode)
	}
}

// TestPublishAction_GatesOnBindingFamily: publish cites :update of the binding
// entity (the publish use case's Gatekeeper code), not the parent's.
func TestPublishAction_GatesOnBindingFamily(t *testing.T) {
	published := ""
	deps := (&uploadRecorder{}).deps()
	deps.PublishTemplateBinding = func(_ context.Context, req *bindingpb.PublishJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.PublishJobOutcomeSummaryDocumentTemplateResponse, error) {
		published = req.GetId()
		return &bindingpb.PublishJobOutcomeSummaryDocumentTemplateResponse{Success: true}, nil
	}
	v := NewPublishAction(deps)
	vc := &view.ViewContext{Request: httptest.NewRequest(http.MethodPost, "/x?id=b-1", nil)}

	if res := v.Handle(permsCtx("job_outcome_summary:update"), vc); res.StatusCode == http.StatusOK || published != "" {
		t.Errorf("parent-entity-only role must not publish (status %d, published %q)", res.StatusCode, published)
	}
	if res := v.Handle(permsCtx("job_outcome_summary_document_template:update"), vc); res.StatusCode != http.StatusOK || published != "b-1" {
		t.Errorf("binding-family update role must publish (status %d, published %q)", res.StatusCode, published)
	}
}

func TestListRows_ReplacementEditAndDraftActions(t *testing.T) {
	rec := &uploadRecorder{bindings: []*bindingpb.JobOutcomeSummaryDocumentTemplate{
		{Id: "published", DocumentTemplateId: "dt-published", VersionStatus: enums.VersionStatus_VERSION_STATUS_PUBLISHED, DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "dt-published", Name: "Published", DocumentPurpose: documentPurpose}},
		{Id: "draft", DocumentTemplateId: "dt-draft", VersionStatus: enums.VersionStatus_VERSION_STATUS_DRAFT, DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "dt-draft", Name: "Draft", DocumentPurpose: documentPurpose}},
	}}
	deps := rec.deps()
	deps.Routes.TemplateUploadURL = "/templates/upload"
	deps.Routes.TemplatePublishURL = "/templates/publish"
	deps.Routes.TemplateDeleteURL = "/templates/delete"
	deps.CommonLabels.Actions.Edit = "Edit"
	perms := types.NewUserPermissions([]string{
		bindingPermissionEntity + ":create",
		bindingPermissionEntity + ":update",
		bindingPermissionEntity + ":delete",
	})

	rows := buildBindingRows(context.Background(), deps, perms)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	if len(rows[0].Actions) != 1 || rows[0].Actions[0].Type != "edit" || rows[0].Actions[0].URL != deps.Routes.TemplateUploadURL {
		t.Fatalf("published actions = %#v, want replacement edit only", rows[0].Actions)
	}
	if len(rows[1].Actions) != 2 || rows[1].Actions[0].Type != "activate" || rows[1].Actions[1].Type != "delete" {
		t.Fatalf("draft actions = %#v, want publish + delete", rows[1].Actions)
	}
}

func TestUploadAction_ReplacementDrawerPrefillsSource(t *testing.T) {
	source := &bindingpb.JobOutcomeSummaryDocumentTemplate{
		Id: "published", DocumentTemplateId: "dt-published", VersionStatus: enums.VersionStatus_VERSION_STATUS_PUBLISHED,
		DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "dt-published", Name: "Existing report card", DocumentPurpose: documentPurpose},
	}
	deps := (&uploadRecorder{bindings: []*bindingpb.JobOutcomeSummaryDocumentTemplate{source}}).deps()
	deps.CommonLabels.Actions.Edit = "Edit"
	req := httptest.NewRequest(http.MethodGet, "/templates/upload?id=published", nil)
	res := NewUploadAction(deps).Handle(permsCtx(bindingPermissionEntity+":create"), &view.ViewContext{Request: req})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("replacement drawer status = %d, want 200", res.StatusCode)
	}
	form, ok := res.Data.(*UploadFormData)
	if !ok {
		t.Fatalf("drawer data = %T, want *UploadFormData", res.Data)
	}
	if !form.IsEdit || form.SourceBindingID != source.GetId() || form.Name != "Existing report card" || form.FormTitle != "Edit" {
		t.Fatalf("replacement drawer was not prefilled: %#v", form)
	}
}

func TestUploadAction_ReplacementCreatesSuccessorWithoutMutatingSource(t *testing.T) {
	source := &bindingpb.JobOutcomeSummaryDocumentTemplate{
		Id: "published", DocumentTemplateId: "dt-published", VersionStatus: enums.VersionStatus_VERSION_STATUS_PUBLISHED,
		DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "dt-published", Name: "Existing report card", DocumentPurpose: documentPurpose},
	}
	rec := &uploadRecorder{bindings: []*bindingpb.JobOutcomeSummaryDocumentTemplate{source}}
	res := NewUploadAction(rec.deps()).Handle(
		permsCtx(bindingPermissionEntity+":create"),
		uploadPostWithFields(t, minimalDocx(t), map[string]string{"source_binding_id": source.GetId()}),
	)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("replacement upload status = %d, want 200", res.StatusCode)
	}
	if rec.createdSupersedes != source.GetId() {
		t.Errorf("supersedes_binding_id = %q, want %q", rec.createdSupersedes, source.GetId())
	}
	if rec.createdDocID == "" || rec.createdDocID == source.GetDocumentTemplateId() {
		t.Errorf("replacement must create a new artifact, got %q", rec.createdDocID)
	}
	if rec.deletedBindingID != "" || rec.deletedDocID != "" || source.GetVersionStatus() != enums.VersionStatus_VERSION_STATUS_PUBLISHED {
		t.Errorf("source history was mutated or deleted: binding=%q doc=%q status=%s", rec.deletedBindingID, rec.deletedDocID, source.GetVersionStatus())
	}

	draft := &bindingpb.JobOutcomeSummaryDocumentTemplate{Id: "draft-source", DocumentTemplateId: "dt-draft", VersionStatus: enums.VersionStatus_VERSION_STATUS_DRAFT, DocumentTemplate: &documenttemplatepb.DocumentTemplate{Id: "dt-draft", DocumentPurpose: documentPurpose}}
	rec = &uploadRecorder{bindings: []*bindingpb.JobOutcomeSummaryDocumentTemplate{draft}}
	res = NewUploadAction(rec.deps()).Handle(
		permsCtx(bindingPermissionEntity+":create"),
		uploadPostWithFields(t, minimalDocx(t), map[string]string{"source_binding_id": draft.GetId()}),
	)
	if res.StatusCode == http.StatusOK || len(rec.order) != 0 {
		t.Fatalf("draft lineage source must fail before writes: status=%d order=%v", res.StatusCode, rec.order)
	}

	rec = &uploadRecorder{}
	res = NewUploadAction(rec.deps()).Handle(
		permsCtx(bindingPermissionEntity+":create"),
		uploadPostWithFields(t, minimalDocx(t), map[string]string{"source_binding_id": "forged-source"}),
	)
	if res.StatusCode == http.StatusOK || len(rec.order) != 0 {
		t.Fatalf("forged lineage source must fail before writes: status=%d order=%v", res.StatusCode, rec.order)
	}
}

// TestUploadAction_BytesLastOrdering locks the Q4 orphan fix: the storage
// write happens LAST, only after both permission-gated creates succeed.
func TestUploadAction_BytesLastOrdering(t *testing.T) {
	rec := &uploadRecorder{}
	v := NewUploadAction(rec.deps())

	res := v.Handle(permsCtx("job_outcome_summary_document_template:create"), uploadPost(t, minimalDocx(t)))
	if res.StatusCode != http.StatusOK {
		t.Fatalf("upload should succeed, got %d", res.StatusCode)
	}
	want := []string{"create_doc", "create_binding", "upload"}
	if fmt.Sprint(rec.order) != fmt.Sprint(want) {
		t.Errorf("bytes must be written LAST: order = %v, want %v", rec.order, want)
	}
	if rec.uploadedKey == "" || rec.createdDocID == "" {
		t.Errorf("expected a stored object + doc row (key %q, doc %q)", rec.uploadedKey, rec.createdDocID)
	}
	if rec.createdContainer != storageContainerFallback || rec.uploadedContainer != storageContainerFallback {
		t.Errorf("new local/mock locator must use fallback %q before composition resolution (created %q, upload %q)", storageContainerFallback, rec.createdContainer, rec.uploadedContainer)
	}
	if rec.createdStorageKey != rec.uploadedKey || !strings.HasPrefix(rec.uploadedKey, storagePrefix+"/") {
		t.Errorf("storage key = %q (created %q), want prefix %q and identical persisted/uploaded keys", rec.uploadedKey, rec.createdStorageKey, storagePrefix+"/")
	}
}

// TestUploadAction_DeniedCreateLeavesNoStorageOrphan: a failed (e.g. denied)
// document_template create must leave NOTHING — in particular no storage
// object (the exact orphan the live pass found in tmp/storage).
func TestUploadAction_DeniedCreateLeavesNoStorageOrphan(t *testing.T) {
	rec := &uploadRecorder{failCreateDoc: true}
	v := NewUploadAction(rec.deps())

	res := v.Handle(permsCtx("job_outcome_summary_document_template:create"), uploadPost(t, minimalDocx(t)))
	if res.StatusCode == http.StatusOK {
		t.Fatal("upload must fail when the doc-template create fails")
	}
	for _, step := range rec.order {
		if step == "upload" {
			t.Error("no storage byte write may happen after a failed create (orphan)")
		}
	}
}

// TestUploadAction_BindingCreateFailureCompensatesDocRow: a failed binding
// create deletes the just-created artifact row and never writes bytes.
func TestUploadAction_BindingCreateFailureCompensatesDocRow(t *testing.T) {
	rec := &uploadRecorder{failCreateBinding: true}
	v := NewUploadAction(rec.deps())

	res := v.Handle(permsCtx("job_outcome_summary_document_template:create"), uploadPost(t, minimalDocx(t)))
	if res.StatusCode == http.StatusOK {
		t.Fatal("upload must fail when the binding create fails")
	}
	for _, step := range rec.order {
		if step == "upload" {
			t.Error("no storage byte write may happen after a failed binding create")
		}
	}
	if rec.deletedDocID == "" || rec.deletedDocID != rec.createdDocID {
		t.Errorf("the orphaned doc row must be compensated (created %q, deleted %q)", rec.createdDocID, rec.deletedDocID)
	}
}

// TestUploadAction_ByteWriteFailureCompensatesBothRows: a failed byte write
// (creates already committed) deletes the draft binding AND the artifact row.
func TestUploadAction_ByteWriteFailureCompensatesBothRows(t *testing.T) {
	rec := &uploadRecorder{failUpload: true}
	v := NewUploadAction(rec.deps())

	res := v.Handle(permsCtx("job_outcome_summary_document_template:create"), uploadPost(t, minimalDocx(t)))
	if res.StatusCode == http.StatusOK {
		t.Fatal("upload must fail when the byte write fails")
	}
	if rec.deletedBindingID != "b-created" {
		t.Errorf("the draft binding must be compensated, deleted %q", rec.deletedBindingID)
	}
	if rec.deletedDocID == "" || rec.deletedDocID != rec.createdDocID {
		t.Errorf("the artifact row must be compensated (created %q, deleted %q)", rec.createdDocID, rec.deletedDocID)
	}
}

// TestDeleteAction_ReapsUnreferencedArtifact: deleting the LAST binding that
// references an artifact row reaps the row; any remaining reference — or an
// incomplete reference scan — leaves it in place (fail-safe).
func TestDeleteAction_ReapsUnreferencedArtifact(t *testing.T) {
	ctx := permsCtx("job_outcome_summary_document_template:delete")
	vc := &view.ViewContext{Request: httptest.NewRequest(http.MethodPost, "/x?id=b-1", nil)}

	// Sole reference → reaped.
	rec := &uploadRecorder{bindings: []*bindingpb.JobOutcomeSummaryDocumentTemplate{
		{Id: "b-1", DocumentTemplateId: "dt-1"},
	}}
	if res := NewDeleteAction(rec.deps()).Handle(ctx, vc); res.StatusCode != http.StatusOK {
		t.Fatalf("delete should succeed, got %d", res.StatusCode)
	}
	if rec.deletedBindingID != "b-1" || rec.deletedDocID != "dt-1" {
		t.Errorf("sole-reference delete must reap the artifact (binding %q, doc %q)", rec.deletedBindingID, rec.deletedDocID)
	}

	// Still referenced by another binding → never reaped.
	rec = &uploadRecorder{bindings: []*bindingpb.JobOutcomeSummaryDocumentTemplate{
		{Id: "b-1", DocumentTemplateId: "dt-1"},
		{Id: "b-2", DocumentTemplateId: "dt-1"},
	}}
	if res := NewDeleteAction(rec.deps()).Handle(ctx, vc); res.StatusCode != http.StatusOK {
		t.Fatalf("delete should succeed, got %d", res.StatusCode)
	}
	if rec.deletedDocID != "" {
		t.Errorf("a still-referenced artifact must never be reaped, deleted %q", rec.deletedDocID)
	}

	// Reference scan fails → fail-safe, nothing reaped.
	rec = &uploadRecorder{listErr: errors.New("list boom")}
	if res := NewDeleteAction(rec.deps()).Handle(ctx, vc); res.StatusCode != http.StatusOK {
		t.Fatalf("delete should still succeed on a scan failure, got %d", res.StatusCode)
	}
	if rec.deletedDocID != "" {
		t.Errorf("an incomplete reference scan must never reap, deleted %q", rec.deletedDocID)
	}

	// Parent-entity-only role → denied, nothing deleted.
	rec = &uploadRecorder{bindings: []*bindingpb.JobOutcomeSummaryDocumentTemplate{{Id: "b-1", DocumentTemplateId: "dt-1"}}}
	if res := NewDeleteAction(rec.deps()).Handle(permsCtx("job_outcome_summary:update"), vc); res.StatusCode == http.StatusOK || rec.deletedBindingID != "" {
		t.Errorf("parent-entity-only role must not delete (deleted %q)", rec.deletedBindingID)
	}
}

// ── R2/C3 (DEC-1a): phase uploads reuse the subscription_group_document_template manifest validator ──

// clientPhaseValidator wires the SAME exported validator the Section
// Templates upload path (subscription_group_document_template_settings)
// applies to RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1
// bindings, mirroring how outcome_summary_module.go wires it in production.
func clientPhaseValidator() func([]byte) error {
	return func(content []byte) error {
		return subscriptiongroupdocument.ValidateTemplate(
			subscriptiongroupdocumentpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1,
			content,
		)
	}
}

// jhsClientPhaseCandidate reads the SAME committed candidate DOCX the
// subscription_group_document package's own manifest test validates
// (TestJHSClientPhaseAuthoringTemplateMatchesManifest) — a real,
// manifest-conforming client-phase report card — rather than hand-authoring a
// second copy of the token contract in this package.
func jhsClientPhaseCandidate(t *testing.T) []byte {
	t.Helper()
	docx, err := os.ReadFile("../../../../../../docs/plan/20260923-individual-report-card-downloads/artifacts/JHS Progress Report - MMIS Template v7.docx")
	if err != nil {
		t.Fatalf("read candidate docx: %v", err)
	}
	return docx
}

func phaseUploadFields(priceScheduleID, phaseCode string) map[string]string {
	fields := map[string]string{}
	if priceScheduleID != "" {
		fields["price_schedule_id"] = priceScheduleID
	}
	if phaseCode != "" {
		fields["job_template_phase_code"] = phaseCode
	}
	return fields
}

func withPhaseCodeOption(deps *Deps, phaseCode string) {
	deps.ListPhaseCodesByPriceSchedule = func(_ context.Context, _ *phasepb.ListPhaseCodesByPriceScheduleRequest) (*phasepb.ListPhaseCodesByPriceScheduleResponse, error) {
		return &phasepb.ListPhaseCodesByPriceScheduleResponse{Success: true, Options: []*phasepb.PhaseCodeOption{{Code: phaseCode, Names: []string{"Term 1"}}}}, nil
	}
}

// TestUploadAction_PhaseUpload_ValidCandidatePasses: a phase-scoped upload
// (job_template_phase_code non-empty) whose DOCX satisfies the client-phase
// manifest contract succeeds.
func TestUploadAction_PhaseUpload_ValidCandidatePasses(t *testing.T) {
	rec := &uploadRecorder{}
	deps := rec.deps()
	deps.ValidatePhaseTemplate = clientPhaseValidator()
	withPhaseCodeOption(deps, "term1")

	res := NewUploadAction(deps).Handle(
		permsCtx(bindingPermissionEntity+":create"),
		uploadPostWithFields(t, jhsClientPhaseCandidate(t), phaseUploadFields("ay-1", "term1")),
	)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("valid phase-scoped candidate must pass, got status %d", res.StatusCode)
	}
	if rec.createdDocID == "" || rec.uploadedKey == "" {
		t.Errorf("valid phase upload must reach artifact create + storage write (doc %q, key %q)", rec.createdDocID, rec.uploadedKey)
	}
}

// TestUploadAction_PhaseUpload_NonConformingDocxRejected: a phase-scoped
// upload whose DOCX passes the generic OOXML archive check but has NONE of
// the client-phase manifest's required tokens is rejected before any create.
func TestUploadAction_PhaseUpload_NonConformingDocxRejected(t *testing.T) {
	rec := &uploadRecorder{}
	deps := rec.deps()
	deps.ValidatePhaseTemplate = clientPhaseValidator()
	withPhaseCodeOption(deps, "term1")

	res := NewUploadAction(deps).Handle(
		permsCtx(bindingPermissionEntity+":create"),
		uploadPostWithFields(t, minimalDocx(t), phaseUploadFields("ay-1", "term1")),
	)
	if res.StatusCode == http.StatusOK {
		t.Fatal("a phase-scoped upload missing the manifest's required tokens must be rejected")
	}
	if len(rec.order) != 0 {
		t.Errorf("a rejected phase manifest must never reach artifact create, calls = %v", rec.order)
	}
}

// TestUploadAction_WholeYearUpload_SameNonConformingDocxStillPasses: the SAME
// non-conforming bytes that fail the phase check above pass unchanged for a
// whole-year upload (empty phase code) — the Year Final contract is
// validateDocxArchive only, exactly as before this change.
func TestUploadAction_WholeYearUpload_SameNonConformingDocxStillPasses(t *testing.T) {
	rec := &uploadRecorder{}
	deps := rec.deps()
	deps.ValidatePhaseTemplate = clientPhaseValidator()

	res := NewUploadAction(deps).Handle(
		permsCtx(bindingPermissionEntity+":create"),
		uploadPostWithFields(t, minimalDocx(t), phaseUploadFields("ay-1", "")),
	)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("a whole-year upload must keep today's behavior (validateDocxArchive only), got status %d", res.StatusCode)
	}
}

// TestUploadAction_PhaseUpload_NilValidatorFailsClosed: a phase-scoped upload
// with no ValidatePhaseTemplate closure configured is rejected rather than
// silently skipping the manifest check.
func TestUploadAction_PhaseUpload_NilValidatorFailsClosed(t *testing.T) {
	rec := &uploadRecorder{}
	deps := rec.deps() // ValidatePhaseTemplate left nil
	withPhaseCodeOption(deps, "term1")

	res := NewUploadAction(deps).Handle(
		permsCtx(bindingPermissionEntity+":create"),
		uploadPostWithFields(t, jhsClientPhaseCandidate(t), phaseUploadFields("ay-1", "term1")),
	)
	if res.StatusCode == http.StatusOK {
		t.Fatal("a phase-scoped upload with no configured validator must fail closed")
	}
	if len(rec.order) != 0 {
		t.Errorf("a fail-closed nil-validator rejection must never reach artifact create, calls = %v", rec.order)
	}
}
