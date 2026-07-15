// Package template_settings renders the TB3 report-card template management
// surface: a standalone settings page (D3 — a dedicated sidebar entry AFTER the
// reports item, NOT a landing tab) listing the operator-uploaded, AY-scoped
// document-template BINDINGS (job_outcome_summary_document_template), with
// upload (create a DRAFT binding + its .docx artifact), publish (the controlled
// publish transaction), and delete.
//
// Reuse note: this mirrors the hybra views/template + pyeza
// attachment-upload-drawer-form patterns (list table + upload drawer + row
// actions), but targets the BINDING entity (version/validity/publish lifecycle)
// rather than a bare document_template, so it is a bespoke view.
//
// Security: the page GET gates on job_outcome_summary:list; every mutation
// (upload/publish/delete) gates on job_outcome_summary:update. Tenant isolation
// for the binding CRUD/publish is enforced in the espyna adapter from trusted
// context — the view supplies no workspace_id. The uploaded artifact is pinned
// to .docx by extension.
package template_settings

import (
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

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary_document_template"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// documentPurpose is the canonical generic discriminator the render resolver +
// upload path share (D6). Vertical vocabulary ("Report Card") lives only in
// lyngua values, never here.
const documentPurpose = "report_card"

const (
	storageBucket  = "templates"
	storagePrefix  = "templates/report_card"
	docxExt        = ".docx"
	docxContentTyp = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	maxUploadBytes = 10 << 20
	tableID        = "report-card-templates-table"
	dateLayout     = "2006-01-02"
)

// Deps holds the template-settings view dependencies. The doc-template artifact
// + storage closures are app-AppContext-derived (like GenerateDoc); the binding
// lifecycle closures come from the espyna binding use cases via the fayna block
// seam. All optional/nil-safe: a nil write closure degrades the surface to a
// "not configured" error rather than a panic.
type Deps struct {
	Routes       outcome_summary.Routes
	Labels       outcome_summary.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Schedule dropdown source (reuses the report-cards price_schedule list).
	ListPriceSchedules func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)

	// document_template artifact (bytes in storage + metadata row).
	UploadTemplate         func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListDocumentTemplates  func(ctx context.Context, req *documenttemplatepb.ListDocumentTemplatesRequest) (*documenttemplatepb.ListDocumentTemplatesResponse, error)
	CreateDocumentTemplate func(ctx context.Context, req *documenttemplatepb.CreateDocumentTemplateRequest) (*documenttemplatepb.CreateDocumentTemplateResponse, error)

	// Binding lifecycle (espyna use cases via the block seam).
	ListTemplateBindings   func(ctx context.Context, req *bindingpb.ListJobOutcomeSummaryDocumentTemplatesRequest) (*bindingpb.ListJobOutcomeSummaryDocumentTemplatesResponse, error)
	CreateTemplateBinding  func(ctx context.Context, req *bindingpb.CreateJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.CreateJobOutcomeSummaryDocumentTemplateResponse, error)
	DeleteTemplateBinding  func(ctx context.Context, req *bindingpb.DeleteJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.DeleteJobOutcomeSummaryDocumentTemplateResponse, error)
	PublishTemplateBinding func(ctx context.Context, req *bindingpb.PublishJobOutcomeSummaryDocumentTemplateRequest) (*bindingpb.PublishJobOutcomeSummaryDocumentTemplateResponse, error)
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
	Labels          outcome_summary.TemplateSettingsLabels
	CommonLabels    any
	ScheduleOptions []types.SelectOption
	AcceptTypes     string
}

// NewListView builds the settings list view (GET). Fail-closed on
// job_outcome_summary:list.
func NewListView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "list") {
			return view.Forbidden("job_outcome_summary:list")
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
				Label:           l.UploadAction,
				ActionURL:       deps.Routes.TemplateUploadURL,
				Icon:            "icon-upload",
				Disabled:        !perms.Can("job_outcome_summary", "update"),
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
				ActiveSubNav:   "report-card-templates",
				HeaderTitle:    l.Title,
				HeaderSubtitle: l.Subtitle,
				HeaderIcon:     "icon-file-text",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "outcome-summary-template-settings-content",
			Table:           tableConfig,
		}
		return view.OK("outcome-summary-template-settings", pageData)
	})
}

// NewUploadAction is the upload drawer (GET = form, POST = create draft binding).
func NewUploadAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "update") {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		l := deps.Labels.TemplateSettings

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("outcome-summary-template-upload-drawer-form", &UploadFormData{
				FormAction:      deps.Routes.TemplateUploadURL,
				Labels:          l,
				CommonLabels:    deps.CommonLabels,
				ScheduleOptions: scheduleOptions(ctx, deps, l.ScheduleFallback),
				AcceptTypes:     docxExt,
			})
		}

		// POST — create the document_template artifact + a DRAFT binding.
		if deps.UploadTemplate == nil || deps.CreateDocumentTemplate == nil || deps.CreateTemplateBinding == nil {
			return view.HTMXError(l.NotConfigured)
		}
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

		docID := newID()
		objectKey := fmt.Sprintf("%s/%s%s", storagePrefix, docID, docxExt)
		if err := deps.UploadTemplate(ctx, storageBucket, objectKey, content, docxContentTyp); err != nil {
			log.Printf("report-card template upload: store bytes: %v", err)
			return view.HTMXError(l.UploadFailed)
		}

		bucket := storageBucket
		key := objectKey
		orig := header.Filename
		size := header.Size
		if _, err := deps.CreateDocumentTemplate(ctx, &documenttemplatepb.CreateDocumentTemplateRequest{
			Data: &documenttemplatepb.DocumentTemplate{
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
			},
		}); err != nil {
			log.Printf("report-card template upload: create document_template: %v", err)
			return view.HTMXError(l.UploadFailed)
		}

		// Create the DRAFT binding. Server forces DRAFT/version/publish audit
		// (create use case) + assigns workspace_id from trusted context; the view
		// never sends workspace_id. price_schedule_id + validity are operator-set.
		binding := &bindingpb.JobOutcomeSummaryDocumentTemplate{
			DocumentTemplateId: docID,
		}
		if ps := strings.TrimSpace(viewCtx.Request.FormValue("price_schedule_id")); ps != "" {
			binding.PriceScheduleId = &ps
		}
		if ts := parseDate(viewCtx.Request.FormValue("validity_start")); ts != nil {
			binding.ValidityStart = ts
		}
		if ts := parseDate(viewCtx.Request.FormValue("validity_end")); ts != nil {
			binding.ValidityEnd = ts
		}
		if _, err := deps.CreateTemplateBinding(ctx, &bindingpb.CreateJobOutcomeSummaryDocumentTemplateRequest{Data: binding}); err != nil {
			log.Printf("report-card template upload: create binding: %v", err)
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

// NewPublishAction publishes a DRAFT binding (POST; id via ?id=).
func NewPublishAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "update") {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		if deps.PublishTemplateBinding == nil {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		id := bindingID(viewCtx)
		if id == "" {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		if _, err := deps.PublishTemplateBinding(ctx, &bindingpb.PublishJobOutcomeSummaryDocumentTemplateRequest{Id: id}); err != nil {
			log.Printf("report-card template publish %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess(tableID)
	})
}

// NewDeleteAction deletes a binding (POST; id via ?id=).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "update") {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		if deps.DeleteTemplateBinding == nil {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		id := bindingID(viewCtx)
		if id == "" {
			return view.HTMXError(deps.Labels.TemplateSettings.NotConfigured)
		}
		if _, err := deps.DeleteTemplateBinding(ctx, &bindingpb.DeleteJobOutcomeSummaryDocumentTemplateRequest{
			Data: &bindingpb.JobOutcomeSummaryDocumentTemplate{Id: id},
		}); err != nil {
			log.Printf("report-card template delete %s: %v", id, err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess(tableID)
	})
}

// --- list assembly -------------------------------------------------------

func bindingColumns(l outcome_summary.TemplateSettingsLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.NameColumn},
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
	resp, err := deps.ListTemplateBindings(ctx, &bindingpb.ListJobOutcomeSummaryDocumentTemplatesRequest{})
	if err != nil {
		log.Printf("report-card template settings: list bindings: %v", err)
		return nil
	}
	l := deps.Labels.TemplateSettings
	docNames := docTemplateNames(ctx, deps)
	schedNames := scheduleNames(ctx, deps)
	canUpdate := perms.Can("job_outcome_summary", "update")

	rows := make([]types.TableRow, 0, len(resp.GetData()))
	for _, b := range resp.GetData() {
		id := b.GetId()
		name := bindingTemplateName(b, docNames)
		schedule := bindingScheduleName(b, schedNames, l.ScheduleFallback)
		version := fmt.Sprintf("v%d", b.GetVersion())
		statusLabel, statusVariant := statusBadge(b.GetVersionStatus(), l)
		validity := formatValidity(b, l)

		actions := []types.TableAction{}
		if b.GetVersionStatus() == enums.VersionStatus_VERSION_STATUS_DRAFT {
			actions = append(actions, types.TableAction{
				Type: "activate", Label: l.PublishAction, Action: "activate",
				URL:             deps.Routes.TemplatePublishURL,
				ItemName:        name,
				ConfirmTitle:    l.PublishAction,
				ConfirmMessage:  l.PublishConfirm,
				Disabled:        !canUpdate,
				DisabledTooltip: l.NotConfigured,
			})
		}
		actions = append(actions, types.TableAction{
			Type: "delete", Label: l.DeleteAction, Action: "delete",
			URL:             deps.Routes.TemplateDeleteURL,
			ItemName:        name,
			ConfirmTitle:    l.DeleteAction,
			ConfirmMessage:  l.DeleteConfirm,
			Disabled:        !canUpdate,
			DisabledTooltip: l.NotConfigured,
		})

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: schedule},
				{Type: "text", Value: version},
				{Type: "badge", Value: statusLabel, Variant: statusVariant},
				{Type: "text", Value: validity},
			},
			DataAttrs: map[string]string{"testid": "rct-row-" + short(id)},
			Actions:   actions,
		})
	}
	return rows
}

func bindingTemplateName(b *bindingpb.JobOutcomeSummaryDocumentTemplate, docNames map[string]string) string {
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

func bindingScheduleName(b *bindingpb.JobOutcomeSummaryDocumentTemplate, schedNames map[string]string, fallback string) string {
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

func statusBadge(s enums.VersionStatus, l outcome_summary.TemplateSettingsLabels) (string, string) {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return l.StatusPublished, "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return l.StatusDeprecated, "warning"
	default:
		return l.StatusDraft, "default"
	}
}

func formatValidity(b *bindingpb.JobOutcomeSummaryDocumentTemplate, l outcome_summary.TemplateSettingsLabels) string {
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

// docTemplateNames maps document_template_id → name (report_card purpose only),
// used when the binding List does not hydrate the nested document_template.
func docTemplateNames(ctx context.Context, deps *Deps) map[string]string {
	out := map[string]string{}
	if deps.ListDocumentTemplates == nil {
		return out
	}
	resp, err := deps.ListDocumentTemplates(ctx, &documenttemplatepb.ListDocumentTemplatesRequest{})
	if err != nil {
		log.Printf("report-card template settings: list document templates: %v", err)
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

// scheduleOptions builds the upload drawer schedule <select> options: a blank
// "workspace default" entry first, then one per price_schedule (name ASC).
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
			log.Printf("report-card template settings: list price schedules: %v", err)
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

// parseDate turns a yyyy-mm-dd form value into a UTC Timestamp, or nil.
func parseDate(v string) *timestamppb.Timestamp {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return nil
	}
	return timestamppb.New(t.UTC())
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
