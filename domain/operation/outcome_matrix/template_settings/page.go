// Package template_settings renders the grade-sheet (outcome-matrix) template
// management surface (Wave C / P4): a standalone settings page — a dedicated
// sidebar entry, NOT a landing tab — listing the operator-uploaded,
// category/schedule-scoped document-template BINDINGS
// (job_template_document_template), with upload (create a DRAFT binding + its
// .docx artifact), publish (the controlled publish transaction), and delete.
//
// It is the JOSDT (outcome_summary/template_settings) sibling: the binding
// entity carries the version/validity/publish lifecycle. The ONE axis change
// versus JOSDT is the additive job_category dimension (the sheet-shape axis —
// academic vs deportment sheets have different column shapes), exposed as the
// category <select> on the upload drawer + a Category list column.
//
// Security (Q4 — permission-family alignment): the view gates on the SAME
// permission family the invoked use cases enforce — the binding entity's own
// codes (job_template_document_template:list/create/update/delete), matching the
// espyna Gatekeeper checks (list→list, upload→create, publish→update,
// delete→delete). Gating the view on the PARENT entity showed split-role users
// enabled controls that then failed downstream, and hid controls from users who
// actually held rights. Tenant isolation for the binding CRUD/publish is
// enforced in the espyna adapter from trusted context — the view supplies no
// workspace_id. The uploaded artifact is pinned to .docx by extension.
package template_settings

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	outcome_matrix "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_document_template"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
)

// documentPurpose is the canonical generic discriminator the render resolver +
// upload path share (Q6). Vertical vocabulary ("Grade Sheet") lives only in
// lyngua values, never here. Filters the settings list AND (belt-and-suspenders)
// the espyna FindApplicable resolver so the sheet + report-card template
// families never co-mingle.
const documentPurpose = "outcome_matrix"

// bindingPermissionEntity is the permission-family entity every gate in this
// view cites — the binding's OWN entity, byte-identical to the entity the
// espyna use cases hand the ActionGatekeeper
// (entityid.JobTemplateDocumentTemplate). View and use case MUST agree (Q4): a
// divergent family yields enabled-but-failing or hidden-but-permitted controls
// for split-role users.
const bindingPermissionEntity = "job_template_document_template"

const (
	storageBucket  = "templates"
	storagePrefix  = "templates/outcome_matrix"
	docxExt        = ".docx"
	docxContentTyp = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	maxUploadBytes = 10 << 20
	// maxRequestBytes bounds the whole multipart request body (the .docx cap plus
	// headroom for the multipart envelope + text fields) so ParseMultipartForm can
	// never be steered into unbounded reads off the wire.
	maxRequestBytes = maxUploadBytes + (1 << 20)
	tableID         = "outcome-matrix-template-table"
	dateLayout      = "2006-01-02"
)

// DOCX archive hardening caps. A .docx is an OOXML ZIP; validate its structure
// and bound its declared expansion before trusting the bytes downstream (the
// renderer reads every entry). These caps are generous versus a real grade-sheet
// template (tens of KB) yet refuse a zip-bomb or a mislabeled archive.
const (
	maxArchiveEntries        = 2000
	maxEntryUncompressed     = 64 << 20  // 64 MiB per entry
	maxAggregateUncompressed = 256 << 20 // 256 MiB total declared
)

// Deps holds the template-settings view dependencies. The doc-template artifact
// + storage closures are app-AppContext-derived (like GenerateDoc); the binding
// lifecycle closures come from the espyna binding use cases via the fayna block
// seam. All optional/nil-safe: a nil write closure degrades the surface to a
// "not configured" error rather than a panic.
type Deps struct {
	Routes       outcome_matrix.Routes
	Labels       outcome_matrix.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Schedule dropdown source (reuses the price_schedule list).
	ListPriceSchedules func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	// Category dropdown source (the NEW sheet-shape axis). The upload drawer
	// offers one option per job_category (plus a blank "Any category").
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)

	// document_template artifact (bytes in storage + metadata row).
	UploadTemplate         func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListDocumentTemplates  func(ctx context.Context, req *documenttemplatepb.ListDocumentTemplatesRequest) (*documenttemplatepb.ListDocumentTemplatesResponse, error)
	CreateDocumentTemplate func(ctx context.Context, req *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error)

	// DeleteDocumentTemplate soft-deletes a document_template row. Used by the
	// Q4 orphan cleanup: compensation when a later upload step fails, and
	// reaping the artifact row after its last referencing binding is deleted.
	// Optional/nil-safe — when unwired the orphan is logged and left in place
	// (render-safe: nothing resolves an unbound or inactive artifact row).
	DeleteDocumentTemplate func(ctx context.Context, req *documenttemplatepb.DeleteDocumentTemplateRequest) (*documenttemplatepb.DeleteDocumentTemplateResponse, error)

	// Binding lifecycle (espyna use cases via the block seam).
	ListTemplateBindings   func(ctx context.Context, req *bindingpb.ListJobTemplateDocumentTemplatesRequest) (*bindingpb.ListJobTemplateDocumentTemplatesResponse, error)
	CreateTemplateBinding  func(ctx context.Context, req *bindingpb.CreateJobTemplateDocumentTemplateRequest) (*bindingpb.CreateJobTemplateDocumentTemplateResponse, error)
	DeleteTemplateBinding  func(ctx context.Context, req *bindingpb.DeleteJobTemplateDocumentTemplateRequest) (*bindingpb.DeleteJobTemplateDocumentTemplateResponse, error)
	PublishTemplateBinding func(ctx context.Context, req *bindingpb.PublishJobTemplateDocumentTemplateRequest) (*bindingpb.PublishJobTemplateDocumentTemplateResponse, error)
}

// PageData is the settings-page data.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// UploadFormData is the upload drawer form data. WorkspaceID is injected by the
// ViewAdapter for the action_workspace_guard (same as the hybra upload form).
type UploadFormData struct {
	FormAction      string
	WorkspaceID     string
	Labels          outcome_matrix.TemplateSettingsLabels
	CommonLabels    any
	CategoryOptions []types.SelectOption
	ScheduleOptions []types.SelectOption
	AcceptTypes     string
}

// NewListView builds the settings list view (GET). Fail-closed on
// job_template_document_template:list — the SAME code the invoked list use case
// enforces (Q4). The list is FILTERED to documentPurpose='outcome_matrix' so it
// never shows report-card templates (and vice versa).
func NewListView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can(bindingPermissionEntity, "list") {
			return view.Forbidden(bindingPermissionEntity + ":list")
		}

		l := deps.Labels.TemplateSettings
		rows := buildBindingRows(ctx, deps, perms)

		tableConfig := &types.TableConfig{
			ID:          tableID,
			RefreshURL:  deps.Routes.TemplateSettingsURL,
			Columns:     bindingColumns(l),
			Rows:        rows,
			ShowSearch:  true,
			ShowActions: true,
			ShowEntries: true,
			Labels:      deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.EmptyTitle,
				Message: l.EmptyMessage,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:     l.UploadAction,
				ActionURL: deps.Routes.TemplateUploadURL,
				Icon:      "icon-upload",
				TestID:    "outcome-matrix-template-upload",
				// Upload CREATES a draft binding — cite the create code the
				// upload path's binding use case enforces (Q4).
				Disabled:        !perms.Can(bindingPermissionEntity, "create"),
				DisabledTooltip: l.NotConfigured,
			},
		}
		types.ApplyColumnStyles(tableConfig.Columns, tableConfig.Rows)
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Title,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   "outcome-matrix-templates",
				HeaderTitle:    l.Title,
				HeaderSubtitle: l.Subtitle,
				HeaderIcon:     "icon-file-text",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "outcome-matrix-template-settings-content",
			Table:           tableConfig,
		}
		return view.OK("outcome-matrix-template-settings", pageData)
	})
}

// NewUploadAction is the upload drawer (GET = form, POST = create draft
// binding). Gates on job_template_document_template:create — the code the
// CreateTemplateBinding use case enforces (Q4); the document_template artifact
// create additionally enforces document_template:create downstream.
func NewUploadAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can(bindingPermissionEntity, "create") {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		l := deps.Labels.TemplateSettings

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("outcome-matrix-template-upload-drawer-form", &UploadFormData{
				FormAction:      deps.Routes.TemplateUploadURL,
				Labels:          l,
				CommonLabels:    deps.CommonLabels,
				CategoryOptions: categoryOptions(ctx, deps, l.CategoryPlaceholder),
				ScheduleOptions: scheduleOptions(ctx, deps, l.SchedulePlaceholder),
				AcceptTypes:     docxExt,
			})
		}

		// POST — create the document_template artifact + a DRAFT binding.
		if deps.UploadTemplate == nil || deps.CreateDocumentTemplate == nil || deps.CreateTemplateBinding == nil {
			return view.HTMXError(l.NotConfigured)
		}
		// Bound the request body BEFORE parsing so a large/streamed multipart body
		// cannot exhaust memory during ParseMultipartForm. (A nil ResponseWriter is
		// safe: MaxBytesReader only type-asserts it to signal early connection
		// close, and the assertion no-ops on nil.)
		viewCtx.Request.Body = http.MaxBytesReader(nil, viewCtx.Request.Body, maxRequestBytes)
		if err := viewCtx.Request.ParseMultipartForm(maxUploadBytes); err != nil {
			return view.HTMXError(l.UploadFailed)
		}

		name := strings.TrimSpace(viewCtx.Request.FormValue("name"))
		if name == "" {
			return view.HTMXError(l.NameLabel)
		}

		fh, header, err := viewCtx.Request.FormFile("template_file")
		if err != nil {
			return view.HTMXError(l.InvalidFile)
		}
		defer fh.Close()

		// Pin intent: .docx only (extension check). A non-.docx upload is rejected.
		if strings.ToLower(filepath.Ext(header.Filename)) != docxExt {
			return view.HTMXError(l.InvalidFile)
		}
		if header.Size > maxUploadBytes {
			return view.HTMXError(l.UploadFailed)
		}
		content, err := io.ReadAll(io.LimitReader(fh, maxUploadBytes+1))
		if err != nil || len(content) == 0 || int64(len(content)) > maxUploadBytes {
			return view.HTMXError(l.UploadFailed)
		}

		// Structure hardening: a .docx must be a well-formed OOXML ZIP with the two
		// mandatory parts and safe, bounded entries. Reject anything else (a
		// mislabeled/renamed file, a zip-bomb, a traversal-crafted archive) before
		// it reaches storage or the renderer.
		if err := validateDocxArchive(content); err != nil {
			log.Printf("grade-sheet template upload: reject archive: %v", err)
			return view.HTMXError(l.InvalidFile)
		}

		// Q4 upload-orphan cleanup — ORDERING IS LOAD-BEARING. The storage bytes
		// are written LAST, after BOTH permission-gated creates succeed: a denied
		// or failed document_template create must leave NOTHING at all; a failed
		// binding create is compensated by deleting the just-created artifact row;
		// and a failed byte write is compensated by deleting the draft binding +
		// artifact row. Any residue on a FAILED compensation is rows pointing at a
		// missing object — render-safe (the resolver's dt.storage_key join filters
		// it out) and logged.
		docID := newID()
		objectKey := fmt.Sprintf("%s/%s%s", storagePrefix, docID, docxExt)

		bucket := storageBucket
		key := objectKey
		orig := header.Filename
		size := header.Size
		docReq := &documenttemplatepb.DocumentTemplate{
			Id:               docID,
			Name:             name,
			TemplateType:     "docx",
			DocumentPurpose:  documentPurpose,
			StorageContainer: &bucket,
			StorageKey:       &key,
			OriginalFilename: &orig,
			FileSizeBytes:    &size,
			Status:           "active",
			Active:           true,
		}
		if notes := strings.TrimSpace(viewCtx.Request.FormValue("description")); notes != "" {
			docReq.Description = &notes
		}
		if _, err := deps.CreateDocumentTemplate(ctx, &documenttemplatepb.CreateDocumentTemplateRequest{
			Data: docReq,
		}); err != nil {
			log.Printf("grade-sheet template upload: create document_template: %v", err)
			return view.HTMXError(l.UploadFailed)
		}

		// Create the DRAFT binding. Server forces DRAFT/version/publish audit
		// (create use case) + assigns workspace_id from trusted context; the view
		// never sends workspace_id. job_category_id + price_schedule_id are the
		// operator-set scope axes (both optional — a blank "Any" leaves them NULL).
		binding := &bindingpb.JobTemplateDocumentTemplate{
			DocumentTemplateId: docID,
		}
		if jc := strings.TrimSpace(viewCtx.Request.FormValue("job_category_id")); jc != "" {
			binding.JobCategoryId = &jc
		}
		if ps := strings.TrimSpace(viewCtx.Request.FormValue("price_schedule_id")); ps != "" {
			binding.PriceScheduleId = &ps
		}
		createResp, err := deps.CreateTemplateBinding(ctx, &bindingpb.CreateJobTemplateDocumentTemplateRequest{Data: binding})
		if err != nil {
			log.Printf("grade-sheet template upload: create binding: %v", err)
			cleanupDocumentTemplate(ctx, deps, docID)
			return view.HTMXError(l.UploadFailed)
		}

		// Bytes LAST. On failure, compensate both created rows (best effort).
		if err := deps.UploadTemplate(ctx, storageBucket, objectKey, content, docxContentTyp); err != nil {
			log.Printf("grade-sheet template upload: store bytes: %v", err)
			if createResp != nil && len(createResp.GetData()) > 0 {
				bindingID := createResp.GetData()[0].GetId()
				if _, derr := deps.DeleteTemplateBinding(ctx, &bindingpb.DeleteJobTemplateDocumentTemplateRequest{
					Data: &bindingpb.JobTemplateDocumentTemplate{Id: bindingID},
				}); derr != nil {
					log.Printf("grade-sheet template upload: compensate delete binding %s: %v", bindingID, derr)
				}
			}
			cleanupDocumentTemplate(ctx, deps, docID)
			return view.HTMXError(l.UploadFailed)
		}

		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true}`,
				"HX-Redirect": deps.Routes.TemplateSettingsURL,
			},
		}
	})
}

// NewPublishAction publishes a DRAFT binding (POST; id via ?id=). Gates on
// job_template_document_template:update — the code the publish use case enforces
// (Q4).
func NewPublishAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can(bindingPermissionEntity, "update") {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		if deps.PublishTemplateBinding == nil {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		id := bindingID(viewCtx)
		if id == "" {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		if _, err := deps.PublishTemplateBinding(ctx, &bindingpb.PublishJobTemplateDocumentTemplateRequest{Id: id}); err != nil {
			log.Printf("grade-sheet template publish %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess(tableID)
	})
}

// NewDeleteAction deletes a binding (POST; id via ?id=). Gates on
// job_template_document_template:delete — the code the delete use case enforces
// (Q4). The delete itself stays draft-only (use case + adapter guard); after it
// succeeds the now-unreferenced document_template artifact row is reaped (Q4
// upload-orphan cleanup).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can(bindingPermissionEntity, "delete") {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		if deps.DeleteTemplateBinding == nil {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		id := bindingID(viewCtx)
		if id == "" {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		// Capture the binding's artifact reference BEFORE the delete (the row is
		// soft-deleted and invisible to the default list afterwards).
		docTemplateID := bindingDocTemplateID(ctx, deps, id)
		if _, err := deps.DeleteTemplateBinding(ctx, &bindingpb.DeleteJobTemplateDocumentTemplateRequest{
			Data: &bindingpb.JobTemplateDocumentTemplate{Id: id},
		}); err != nil {
			log.Printf("grade-sheet template delete %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		reapUnreferencedDocTemplate(ctx, deps, docTemplateID, id)
		return view.HTMXSuccess(tableID)
	})
}

// --- Q4 upload-orphan cleanup helpers ------------------------------------

// cleanupDocumentTemplate best-effort soft-deletes an orphaned
// document_template row (upload compensation / post-delete reap). Nil-safe:
// without the closure the orphan row is logged and left in place — render-safe,
// since the resolver only reaches ACTIVE artifact rows through a PUBLISHED
// binding join. Never touches storage: the storage provider contract exposes
// no object delete, so object reaping is a provider-lifecycle concern.
func cleanupDocumentTemplate(ctx context.Context, deps *Deps, docID string) {
	if docID == "" {
		return
	}
	if deps.DeleteDocumentTemplate == nil {
		log.Printf("grade-sheet template cleanup: document_template %s left in place (delete closure not wired)", docID)
		return
	}
	if _, err := deps.DeleteDocumentTemplate(ctx, &documenttemplatepb.DeleteDocumentTemplateRequest{
		Data: &documenttemplatepb.DocumentTemplate{Id: docID},
	}); err != nil {
		log.Printf("grade-sheet template cleanup: delete document_template %s: %v", docID, err)
	}
}

// bindingDocTemplateID resolves a binding's document_template_id via the list
// closure (active pass + explicit inactive pass), "" when unresolvable —
// cleanup then degrades to a no-op (fail-safe).
func bindingDocTemplateID(ctx context.Context, deps *Deps, bindingID string) string {
	all, ok := listAllBindings(ctx, deps)
	if !ok {
		return ""
	}
	for _, b := range all {
		if b.GetId() == bindingID {
			return b.GetDocumentTemplateId()
		}
	}
	return ""
}

// reapUnreferencedDocTemplate soft-deletes the artifact row a just-deleted
// binding referenced, IFF no OTHER binding — any lifecycle status, active or
// soft-deleted — still references it. FAIL-SAFE: any remaining reference, or a
// reference scan that could not COMPLETELY resolve (list error), skips the
// reap; we never remove an artifact that anything might still point at.
// Storage bytes are intentionally left (no object-delete API in the storage
// contract).
func reapUnreferencedDocTemplate(ctx context.Context, deps *Deps, docTemplateID, deletedBindingID string) {
	if docTemplateID == "" {
		return
	}
	all, ok := listAllBindings(ctx, deps)
	if !ok {
		log.Printf("grade-sheet template cleanup: reference scan incomplete — leaving document_template %s in place", docTemplateID)
		return
	}
	for _, b := range all {
		if b.GetId() == deletedBindingID {
			continue // the binding just soft-deleted — not a live reference
		}
		if b.GetDocumentTemplateId() == docTemplateID {
			return // still referenced — never reap
		}
	}
	cleanupDocumentTemplate(ctx, deps, docTemplateID)
}

// listAllBindings returns every binding (active + soft-deleted) via the list
// closure's two-pass active/inactive merge (the generic List defaults to
// active-only unless an explicit active filter is supplied). ok=false when the
// scan is incomplete (missing closure or any pass errored) — callers making
// destructive decisions MUST treat that as "unknown references" (fail-safe).
func listAllBindings(ctx context.Context, deps *Deps) ([]*bindingpb.JobTemplateDocumentTemplate, bool) {
	if deps.ListTemplateBindings == nil {
		return nil, false
	}
	out := make([]*bindingpb.JobTemplateDocumentTemplate, 0, 8)
	seen := map[string]bool{}
	requests := []*bindingpb.ListJobTemplateDocumentTemplatesRequest{
		{},
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{{
			Field:      "active",
			FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}},
		}}}},
	}
	for _, req := range requests {
		resp, err := deps.ListTemplateBindings(ctx, req)
		if err != nil {
			log.Printf("grade-sheet template cleanup: list bindings: %v", err)
			return nil, false
		}
		for _, b := range resp.GetData() {
			if id := b.GetId(); id != "" && !seen[id] {
				seen[id] = true
				out = append(out, b)
			}
		}
	}
	return out, true
}

// --- list assembly -------------------------------------------------------

func bindingColumns(l outcome_matrix.TemplateSettingsLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.NameColumn},
		{Key: "category", Label: l.CategoryColumn, WidthClass: "col-2xl"},
		{Key: "schedule", Label: l.ScheduleColumn, WidthClass: "col-3xl"},
		{Key: "version", Label: l.VersionColumn, WidthClass: "col-lg"},
		{Key: "status", Label: l.StatusColumn, WidthClass: "col-2xl"},
		{Key: "validity", Label: l.ValidityColumn, WidthClass: "col-3xl"},
	}
}

func buildBindingRows(ctx context.Context, deps *Deps, perms *types.UserPermissions) []types.TableRow {
	if deps.ListTemplateBindings == nil {
		return nil
	}
	resp, err := deps.ListTemplateBindings(ctx, &bindingpb.ListJobTemplateDocumentTemplatesRequest{})
	if err != nil {
		log.Printf("grade-sheet template settings: list bindings: %v", err)
		return nil
	}
	l := deps.Labels.TemplateSettings
	docNames := docTemplateNames(ctx, deps)
	schedNames := scheduleNames(ctx, deps)
	catNames := categoryNames(ctx, deps)
	// Q4: each row action cites the code its use case enforces — publish gates
	// :update, delete gates :delete (a split role may hold one, not the other).
	canPublish := perms.Can(bindingPermissionEntity, "update")
	canDelete := perms.Can(bindingPermissionEntity, "delete")

	rows := make([]types.TableRow, 0, len(resp.GetData()))
	for _, b := range resp.GetData() {
		id := b.GetId()
		name := bindingTemplateName(b, docNames)
		category := bindingCategoryName(b, catNames, l.CategoryPlaceholder)
		schedule := bindingScheduleName(b, schedNames, l.SchedulePlaceholder)
		version := fmt.Sprintf("v%d", b.GetVersion())
		statusLabel, statusVariant := statusBadge(b.GetVersionStatus(), l)
		validity := formatValidity(b, l)

		// Publish + Delete are DRAFT-only. A PUBLISHED/DEPRECATED binding is part
		// of the immutable version history (historical as_of renders resolve it),
		// so it exposes no row action — mirroring the server-side draft-only gate
		// enforced in the publish + delete use cases and the persistence adapter.
		actions := []types.TableAction{}
		if b.GetVersionStatus() == enums.VersionStatus_VERSION_STATUS_DRAFT {
			actions = append(actions, types.TableAction{
				Type: "activate", Label: l.PublishAction, Action: "activate",
				URL:             deps.Routes.TemplatePublishURL,
				ItemName:        name,
				ConfirmTitle:    l.PublishAction,
				ConfirmMessage:  l.PublishConfirm,
				Disabled:        !canPublish,
				DisabledTooltip: l.NotConfigured,
			})
			actions = append(actions, types.TableAction{
				Type: "delete", Label: l.DeleteAction, Action: "delete",
				URL:             deps.Routes.TemplateDeleteURL,
				ItemName:        name,
				ConfirmTitle:    l.DeleteAction,
				ConfirmMessage:  l.DeleteConfirm,
				Disabled:        !canDelete,
				DisabledTooltip: l.NotConfigured,
			})
		}

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: category},
				{Type: "text", Value: schedule},
				{Type: "text", Value: version},
				{Type: "badge", Value: statusLabel, Variant: statusVariant},
				{Type: "text", Value: validity},
			},
			DataAttrs: map[string]string{"testid": "omt-row-" + short(id)},
			Actions:   actions,
		})
	}
	return rows
}

func bindingTemplateName(b *bindingpb.JobTemplateDocumentTemplate, docNames map[string]string) string {
	if dt := b.GetDocumentTemplate(); dt != nil {
		if n := strings.TrimSpace(dt.GetName()); n != "" {
			return n
		}
	}
	if n := strings.TrimSpace(docNames[b.GetDocumentTemplateId()]); n != "" {
		return n
	}
	return "—"
}

func bindingScheduleName(b *bindingpb.JobTemplateDocumentTemplate, schedNames map[string]string, fallback string) string {
	psID := strings.TrimSpace(b.GetPriceScheduleId())
	if psID == "" {
		return fallback
	}
	if ps := b.GetPriceSchedule(); ps != nil {
		if n := strings.TrimSpace(ps.GetName()); n != "" {
			return n
		}
	}
	if n := strings.TrimSpace(schedNames[psID]); n != "" {
		return n
	}
	return psID
}

// bindingCategoryName resolves the binding's job_category display name (the NEW
// sheet-shape axis). An unscoped binding (no job_category_id) shows the "Any
// category" fallback.
func bindingCategoryName(b *bindingpb.JobTemplateDocumentTemplate, catNames map[string]string, fallback string) string {
	jcID := strings.TrimSpace(b.GetJobCategoryId())
	if jcID == "" {
		return fallback
	}
	if jc := b.GetJobCategory(); jc != nil {
		if n := strings.TrimSpace(jc.GetName()); n != "" {
			return n
		}
	}
	if n := strings.TrimSpace(catNames[jcID]); n != "" {
		return n
	}
	return jcID
}

func statusBadge(s enums.VersionStatus, l outcome_matrix.TemplateSettingsLabels) (string, string) {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return l.StatusPublished, "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return l.StatusDeprecated, "warning"
	default:
		return l.StatusDraft, "default"
	}
}

func formatValidity(b *bindingpb.JobTemplateDocumentTemplate, l outcome_matrix.TemplateSettingsLabels) string {
	var start, end string
	if b.GetValidityStart() != nil {
		start = b.GetValidityStart().AsTime().Format(dateLayout)
	}
	if b.GetValidityEnd() != nil {
		end = b.GetValidityEnd().AsTime().Format(dateLayout)
	}
	switch {
	case start == "" && end == "":
		return "—"
	case start != "" && end != "":
		return start + " – " + end
	case start != "":
		return "≥ " + start
	default:
		return "< " + end
	}
}

// docTemplateNames maps document_template_id → name (outcome_matrix purpose
// only), used when the binding List does not hydrate the nested document_template.
func docTemplateNames(ctx context.Context, deps *Deps) map[string]string {
	out := map[string]string{}
	if deps.ListDocumentTemplates == nil {
		return out
	}
	resp, err := deps.ListDocumentTemplates(ctx, &documenttemplatepb.ListDocumentTemplatesRequest{})
	if err != nil {
		log.Printf("grade-sheet template settings: list document templates: %v", err)
		return out
	}
	for _, t := range resp.GetData() {
		if t.GetDocumentPurpose() != documentPurpose {
			continue
		}
		if id := t.GetId(); id != "" {
			out[id] = t.GetName()
		}
	}
	return out
}

// scheduleNames maps price_schedule_id → name for the display column + dropdown.
func scheduleNames(ctx context.Context, deps *Deps) map[string]string {
	out := map[string]string{}
	for _, ps := range listAllSchedules(ctx, deps) {
		if id := ps.GetId(); id != "" {
			out[id] = ps.GetName()
		}
	}
	return out
}

// categoryNames maps job_category_id → name for the Category display column.
func categoryNames(ctx context.Context, deps *Deps) map[string]string {
	out := map[string]string{}
	for _, jc := range listAllCategories(ctx, deps) {
		if id := jc.GetId(); id != "" {
			out[id] = jc.GetName()
		}
	}
	return out
}

// scheduleOptions builds the upload drawer schedule <select> options: a blank
// "any schedule" entry first, then one per price_schedule (name ASC).
func scheduleOptions(ctx context.Context, deps *Deps, fallback string) []types.SelectOption {
	opts := []types.SelectOption{{Value: "", Label: fallback}}
	schedules := listAllSchedules(ctx, deps)
	sort.SliceStable(schedules, func(i, j int) bool {
		return strings.ToLower(schedules[i].GetName()) < strings.ToLower(schedules[j].GetName())
	})
	for _, ps := range schedules {
		if id := ps.GetId(); id != "" {
			opts = append(opts, types.SelectOption{Value: id, Label: ps.GetName()})
		}
	}
	return opts
}

// categoryOptions builds the upload drawer category <select> options (the NEW
// sheet-shape axis): a blank "any category" entry first, then one per
// job_category (name ASC).
func categoryOptions(ctx context.Context, deps *Deps, fallback string) []types.SelectOption {
	opts := []types.SelectOption{{Value: "", Label: fallback}}
	categories := listAllCategories(ctx, deps)
	sort.SliceStable(categories, func(i, j int) bool {
		return strings.ToLower(categories[i].GetName()) < strings.ToLower(categories[j].GetName())
	})
	for _, jc := range categories {
		if id := jc.GetId(); id != "" {
			opts = append(opts, types.SelectOption{Value: id, Label: jc.GetName()})
		}
	}
	return opts
}

// listAllSchedules returns every price_schedule (active + inactive) so bindings
// scoped to a frozen historical AY still resolve a name. Mirrors the list view's
// two-call active/inactive merge. Nil-safe.
func listAllSchedules(ctx context.Context, deps *Deps) []*priceschedulepb.PriceSchedule {
	if deps.ListPriceSchedules == nil {
		return nil
	}
	out := make([]*priceschedulepb.PriceSchedule, 0, 4)
	seen := map[string]bool{}
	requests := []*priceschedulepb.ListPriceSchedulesRequest{
		{},
		{Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{{
			Field:      "active",
			FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: false}},
		}}}},
	}
	for _, req := range requests {
		resp, err := deps.ListPriceSchedules(ctx, req)
		if err != nil {
			log.Printf("grade-sheet template settings: list price schedules: %v", err)
			continue
		}
		for _, ps := range resp.GetData() {
			if id := ps.GetId(); id != "" && !seen[id] {
				seen[id] = true
				out = append(out, ps)
			}
		}
	}
	return out
}

// listAllCategories returns every job_category. Nil-safe (a nil closure or an
// error yields no categories → the drawer offers only "Any category").
func listAllCategories(ctx context.Context, deps *Deps) []*jobcategorypb.JobCategory {
	if deps.ListJobCategories == nil {
		return nil
	}
	resp, err := deps.ListJobCategories(ctx, &jobcategorypb.ListJobCategoriesRequest{})
	if err != nil {
		log.Printf("grade-sheet template settings: list job categories: %v", err)
		return nil
	}
	out := make([]*jobcategorypb.JobCategory, 0, len(resp.GetData()))
	seen := map[string]bool{}
	for _, jc := range resp.GetData() {
		if id := jc.GetId(); id != "" && !seen[id] {
			seen[id] = true
			out = append(out, jc)
		}
	}
	return out
}

// --- small helpers -------------------------------------------------------

// bindingID reads the target binding id from ?id= (the table row-action JS
// appends it) or the form body, fail-closed to "".
func bindingID(viewCtx *view.ViewContext) string {
	if id := strings.TrimSpace(viewCtx.Request.URL.Query().Get("id")); id != "" {
		return id
	}
	_ = viewCtx.Request.ParseForm()
	return strings.TrimSpace(viewCtx.Request.FormValue("id"))
}

// validateDocxArchive verifies content is a well-formed DOCX (OOXML) ZIP: it
// must open as a zip, contain the two mandatory OOXML parts ([Content_Types].xml
// + word/document.xml), use only safe relative entry paths (no absolute paths, no
// ".." traversal), and stay within the entry-count, per-entry, and aggregate
// declared-uncompressed-size caps. It inspects only the central-directory
// metadata — it never decompresses an entry — so a declared-size zip-bomb is
// rejected without expansion. Returns a non-nil error describing the first
// violation; nil means the archive is structurally acceptable.
func validateDocxArchive(content []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return fmt.Errorf("not a valid docx (zip) archive: %w", err)
	}
	if len(zr.File) > maxArchiveEntries {
		return fmt.Errorf("archive has too many entries: %d > %d", len(zr.File), maxArchiveEntries)
	}
	var (
		haveContentTypes bool
		haveDocument     bool
		aggregate        uint64
	)
	for _, f := range zr.File {
		name := strings.ReplaceAll(f.Name, "\\", "/")
		if strings.HasPrefix(name, "/") {
			return fmt.Errorf("archive entry has an absolute path: %q", f.Name)
		}
		for _, seg := range strings.Split(name, "/") {
			if seg == ".." {
				return fmt.Errorf("archive entry escapes its root: %q", f.Name)
			}
		}
		if f.UncompressedSize64 > maxEntryUncompressed {
			return fmt.Errorf("archive entry %q declares %d uncompressed bytes (> %d)", f.Name, f.UncompressedSize64, uint64(maxEntryUncompressed))
		}
		aggregate += f.UncompressedSize64
		if aggregate > maxAggregateUncompressed {
			return fmt.Errorf("archive declares too many uncompressed bytes (> %d)", uint64(maxAggregateUncompressed))
		}
		switch name {
		case "[Content_Types].xml":
			haveContentTypes = true
		case "word/document.xml":
			haveDocument = true
		}
	}
	if !haveContentTypes {
		return fmt.Errorf("docx missing [Content_Types].xml")
	}
	if !haveDocument {
		return fmt.Errorf("docx missing word/document.xml")
	}
	return nil
}

// newID returns a random 32-char hex id (stdlib; the doc-template Id + storage
// key). Falls back to a timestamp-derived id if crypto/rand is unavailable.
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func short(id string) string {
	if len(id) > 8 {
		return id[len(id)-8:]
	}
	return id
}
