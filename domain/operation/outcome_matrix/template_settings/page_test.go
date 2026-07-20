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
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_document_template"
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
	createdPurpose    string
	createdBindingID  string
	createdCategoryID string
	createdScheduleID string
	deletedBindingID  string
	deletedDocID      string
	uploadedKey       string

	bindings []*bindingpb.JobTemplateDocumentTemplate
}

func (r *uploadRecorder) deps() *Deps {
	return &Deps{
		UploadTemplate: func(_ context.Context, _, key string, _ []byte, _ string) error {
			r.order = append(r.order, "upload")
			if r.failUpload {
				return errors.New("upload boom")
			}
			r.uploadedKey = key
			return nil
		},
		CreateDocumentTemplate: func(_ context.Context, req *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error) {
			r.order = append(r.order, "create_doc")
			if r.failCreateDoc {
				return nil, errors.New("create doc boom")
			}
			r.createdDocID = req.GetData().GetId()
			r.createdPurpose = req.GetData().GetDocumentPurpose()
			return &documenttemplatepb.CreateDocumentTemplateResponse{Success: true}, nil
		},
		DeleteDocumentTemplate: func(_ context.Context, req *documenttemplatepb.DeleteDocumentTemplateRequest) (*documenttemplatepb.DeleteDocumentTemplateResponse, error) {
			r.order = append(r.order, "delete_doc")
			r.deletedDocID = req.GetData().GetId()
			return &documenttemplatepb.DeleteDocumentTemplateResponse{Success: true}, nil
		},
		CreateTemplateBinding: func(_ context.Context, req *bindingpb.CreateJobTemplateDocumentTemplateRequest) (*bindingpb.CreateJobTemplateDocumentTemplateResponse, error) {
			r.order = append(r.order, "create_binding")
			if r.failCreateBinding {
				return nil, errors.New("create binding boom")
			}
			r.createdBindingID = "b-created"
			r.createdCategoryID = req.GetData().GetJobCategoryId()
			r.createdScheduleID = req.GetData().GetPriceScheduleId()
			return &bindingpb.CreateJobTemplateDocumentTemplateResponse{
				Data:    []*bindingpb.JobTemplateDocumentTemplate{{Id: "b-created"}},
				Success: true,
			}, nil
		},
		DeleteTemplateBinding: func(_ context.Context, req *bindingpb.DeleteJobTemplateDocumentTemplateRequest) (*bindingpb.DeleteJobTemplateDocumentTemplateResponse, error) {
			r.order = append(r.order, "delete_binding")
			r.deletedBindingID = req.GetData().GetId()
			return &bindingpb.DeleteJobTemplateDocumentTemplateResponse{Success: true}, nil
		},
		ListTemplateBindings: func(_ context.Context, _ *bindingpb.ListJobTemplateDocumentTemplatesRequest) (*bindingpb.ListJobTemplateDocumentTemplatesResponse, error) {
			if r.listErr != nil {
				return nil, r.listErr
			}
			return &bindingpb.ListJobTemplateDocumentTemplatesResponse{Data: r.bindings, Success: true}, nil
		},
	}
}

// uploadPost builds a minimal multipart POST (name + file). extra adds/overrides
// text fields (e.g. job_category_id) so the axis-persistence tests can exercise
// the category select.
func uploadPost(t *testing.T, content []byte, extra map[string]string) *view.ViewContext {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("name", "Test Template"); err != nil {
		t.Fatalf("write name field: %v", err)
	}
	for k, v := range extra {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatalf("write %q field: %v", k, err)
		}
	}
	fw, err := mw.CreateFormFile("template_file", "sheet.docx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/action/outcome-matrix/templates/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return &view.ViewContext{Request: req}
}

// TestListView_GatesOnBindingFamily locks the split-role alignment: the page
// gate cites the SAME code the list use case enforces
// (job_template_document_template:list). Holding only the PARENT entity's code is
// not enough — and holding the binding code is.
func TestListView_GatesOnBindingFamily(t *testing.T) {
	v := NewListView((&uploadRecorder{}).deps())

	res := v.Handle(permsCtx("job_template:list"), &view.ViewContext{})
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("parent-entity-only role must be forbidden, got %d", res.StatusCode)
	}
	res = v.Handle(permsCtx("job_template_document_template:list"), &view.ViewContext{})
	if res.StatusCode != http.StatusOK {
		t.Errorf("binding-family list role must see the page, got %d", res.StatusCode)
	}
}

// TestPublishAction_GatesOnBindingFamily: publish cites :update of the binding
// entity (the publish use case's Gatekeeper code), not the parent's.
func TestPublishAction_GatesOnBindingFamily(t *testing.T) {
	published := ""
	deps := (&uploadRecorder{}).deps()
	deps.PublishTemplateBinding = func(_ context.Context, req *bindingpb.PublishJobTemplateDocumentTemplateRequest) (*bindingpb.PublishJobTemplateDocumentTemplateResponse, error) {
		published = req.GetId()
		return &bindingpb.PublishJobTemplateDocumentTemplateResponse{Success: true}, nil
	}
	v := NewPublishAction(deps)
	vc := &view.ViewContext{Request: httptest.NewRequest(http.MethodPost, "/x?id=b-1", nil)}

	if res := v.Handle(permsCtx("job_template:update"), vc); res.StatusCode == http.StatusOK || published != "" {
		t.Errorf("parent-entity-only role must not publish (status %d, published %q)", res.StatusCode, published)
	}
	if res := v.Handle(permsCtx("job_template_document_template:update"), vc); res.StatusCode != http.StatusOK || published != "b-1" {
		t.Errorf("binding-family update role must publish (status %d, published %q)", res.StatusCode, published)
	}
}

// TestUploadAction_BytesLastOrdering locks the Q4 orphan fix: the storage write
// happens LAST, only after both permission-gated creates succeed. It also pins
// the Q6 purpose stamp (DocumentPurpose='outcome_matrix') so the sheet family
// never co-mingles with report-card templates.
func TestUploadAction_BytesLastOrdering(t *testing.T) {
	rec := &uploadRecorder{}
	v := NewUploadAction(rec.deps())

	res := v.Handle(permsCtx("job_template_document_template:create"), uploadPost(t, minimalDocx(t), nil))
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
	if rec.createdPurpose != documentPurpose {
		t.Errorf("uploaded document_template must be stamped purpose=%q, got %q", documentPurpose, rec.createdPurpose)
	}
}

// TestUploadAction_PersistsCategoryAxis: the NEW sheet-shape axis. A submitted
// job_category_id rides onto the DRAFT binding; a blank category leaves it unset.
func TestUploadAction_PersistsCategoryAxis(t *testing.T) {
	rec := &uploadRecorder{}
	res := NewUploadAction(rec.deps()).Handle(
		permsCtx("job_template_document_template:create"),
		uploadPost(t, minimalDocx(t), map[string]string{"job_category_id": "cat-academic", "price_schedule_id": "ps-2026"}),
	)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("upload should succeed, got %d", res.StatusCode)
	}
	if rec.createdCategoryID != "cat-academic" {
		t.Errorf("the binding must carry the submitted job_category_id, got %q", rec.createdCategoryID)
	}
	if rec.createdScheduleID != "ps-2026" {
		t.Errorf("the binding must carry the submitted price_schedule_id, got %q", rec.createdScheduleID)
	}

	// Blank category → unset (workspace-wide fallback binding).
	rec = &uploadRecorder{}
	if res := NewUploadAction(rec.deps()).Handle(
		permsCtx("job_template_document_template:create"),
		uploadPost(t, minimalDocx(t), nil),
	); res.StatusCode != http.StatusOK {
		t.Fatalf("upload should succeed, got %d", res.StatusCode)
	}
	if rec.createdCategoryID != "" {
		t.Errorf("a blank category must leave job_category_id unset, got %q", rec.createdCategoryID)
	}
}

// TestUploadAction_DeniedCreateLeavesNoStorageOrphan: a failed (e.g. denied)
// document_template create must leave NOTHING — in particular no storage object.
func TestUploadAction_DeniedCreateLeavesNoStorageOrphan(t *testing.T) {
	rec := &uploadRecorder{failCreateDoc: true}
	v := NewUploadAction(rec.deps())

	res := v.Handle(permsCtx("job_template_document_template:create"), uploadPost(t, minimalDocx(t), nil))
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

	res := v.Handle(permsCtx("job_template_document_template:create"), uploadPost(t, minimalDocx(t), nil))
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

	res := v.Handle(permsCtx("job_template_document_template:create"), uploadPost(t, minimalDocx(t), nil))
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
	ctx := permsCtx("job_template_document_template:delete")
	vc := &view.ViewContext{Request: httptest.NewRequest(http.MethodPost, "/x?id=b-1", nil)}

	// Sole reference → reaped.
	rec := &uploadRecorder{bindings: []*bindingpb.JobTemplateDocumentTemplate{
		{Id: "b-1", DocumentTemplateId: "dt-1"},
	}}
	if res := NewDeleteAction(rec.deps()).Handle(ctx, vc); res.StatusCode != http.StatusOK {
		t.Fatalf("delete should succeed, got %d", res.StatusCode)
	}
	if rec.deletedBindingID != "b-1" || rec.deletedDocID != "dt-1" {
		t.Errorf("sole-reference delete must reap the artifact (binding %q, doc %q)", rec.deletedBindingID, rec.deletedDocID)
	}

	// Still referenced by another binding → never reaped.
	rec = &uploadRecorder{bindings: []*bindingpb.JobTemplateDocumentTemplate{
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
	rec = &uploadRecorder{bindings: []*bindingpb.JobTemplateDocumentTemplate{{Id: "b-1", DocumentTemplateId: "dt-1"}}}
	if res := NewDeleteAction(rec.deps()).Handle(permsCtx("job_template:update"), vc); res.StatusCode == http.StatusOK || rec.deletedBindingID != "" {
		t.Errorf("parent-entity-only role must not delete (deleted %q)", rec.deletedBindingID)
	}
}
