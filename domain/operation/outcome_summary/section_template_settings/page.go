// Package section_template_settings owns the subscription-group outcome
// document-template management surface. Vertical wording is supplied by
// Lyngua; code, routes, permissions, and persistence use canonical concepts.
package section_template_settings

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

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	sectiondocument "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/section_document"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	documenttemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/template"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	planpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/plan"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	bindingPermissionEntity        = "subscription_group_document_template"
	artifactPermissionEntity       = "document_template"
	documentPurpose                = "subscription_group_outcome_summary"
	storagePrefix                  = "templates/subscription_group"
	docxExt                        = ".docx"
	docxContentType                = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	maxUploadBytes           int64 = 10 << 20
	maxRequestBytes                = maxUploadBytes + (1 << 20)
	tableID                        = "subscription-group-document-template-table"
	dateLayout                     = "2006-01-02"
	templateBindingPageLimit       = 100
	templateBindingMaxPages        = 100
)

type Deps struct {
	Routes       outcome_summary.Routes
	Labels       outcome_summary.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels
	Options      outcome_summary.Options

	ListPriceSchedules func(context.Context, *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	ListPlans          func(context.Context, *planpb.ListPlansRequest) (*planpb.ListPlansResponse, error)
	ListJobCategories  func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)

	// StoreTemplate writes the generated object key and returns the exact
	// physical container persisted on the artifact row.
	StoreTemplate        func(context.Context, string, []byte, string) (string, error)
	DeleteTemplateObject func(context.Context, string, string) error

	CreateUploadPair     func(context.Context, *documenttemplatepb.DocumentTemplate, *bindingpb.SubscriptionGroupDocumentTemplate) (*documenttemplatepb.DocumentTemplate, *bindingpb.SubscriptionGroupDocumentTemplate, error)
	ListTemplateBindings func(context.Context, *bindingpb.ListSubscriptionGroupDocumentTemplatesRequest) (*bindingpb.ListSubscriptionGroupDocumentTemplatesResponse, error)
	// DeleteDraftPair atomically soft-deletes the DRAFT binding and its exact
	// artifact row. The returned locator is safe to use only after this call
	// returns nil error (the use case owns the commit boundary).
	DeleteDraftPair        func(context.Context, string) (*documenttemplatepb.DocumentTemplate, error)
	PublishTemplateBinding func(context.Context, *bindingpb.PublishSubscriptionGroupDocumentTemplateRequest) (*bindingpb.PublishSubscriptionGroupDocumentTemplateResponse, error)
}

type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

type UploadFormData struct {
	FormAction      string
	WorkspaceID     string
	Labels          outcome_summary.SectionTemplateSettingsLabels
	CommonLabels    any
	ProfileLabel    string
	ScheduleOptions []types.SelectOption
	PlanOptions     []types.SelectOption
	CategoryOptions []types.SelectOption
	AcceptTypes     string
}

func NewListView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can(bindingPermissionEntity, "list") {
			return view.Forbidden(bindingPermissionEntity + ":list")
		}
		l := deps.Labels.SectionTemplateSettings
		table := &types.TableConfig{
			ID:          tableID,
			RefreshURL:  deps.Routes.SectionTemplateSettingsURL,
			Columns:     bindingColumns(l),
			Rows:        bindingRows(ctx, deps, perms),
			ShowSearch:  true,
			ShowActions: true,
			ShowEntries: true,
			Labels:      deps.TableLabels,
			EmptyState:  types.TableEmptyState{Title: l.EmptyTitle, Message: l.EmptyMessage},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.UploadAction,
				ActionURL:       deps.Routes.SectionTemplateUploadURL,
				Icon:            "icon-upload",
				TestID:          "section-template-upload",
				Disabled:        !canCreate(perms),
				DisabledTooltip: l.NotConfigured,
			},
		}
		types.ApplyColumnStyles(table.Columns, table.Rows)
		types.ApplyTableSettings(table)
		return view.OK("section-template-settings", &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion, Title: l.Title,
				CurrentPath: viewCtx.CurrentPath, ActiveNav: deps.Routes.ActiveNav,
				ActiveSubNav: "section-templates", HeaderTitle: l.Title,
				HeaderSubtitle: l.Subtitle, HeaderIcon: "icon-file-text",
				CommonLabels: deps.CommonLabels,
			},
			ContentTemplate: "section-template-settings-content",
			Table:           table,
		})
	})
}

func NewUploadAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		l := deps.Labels.SectionTemplateSettings
		perms := view.GetUserPermissions(ctx)
		if !canCreate(perms) {
			return view.HTMXError(l.NotConfigured)
		}
		categories, err := mappedCategories(ctx, deps)
		if err != nil {
			log.Printf("section template settings: category choices: %v", err)
			return view.HTMXError(l.NotConfigured)
		}
		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("section-template-upload-drawer-form", &UploadFormData{
				FormAction: deps.Routes.SectionTemplateUploadURL, Labels: l,
				CommonLabels:    deps.CommonLabels,
				ProfileLabel:    profileLabel(bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1, l),
				ScheduleOptions: scheduleOptions(ctx, deps, l.ScheduleFallback),
				PlanOptions:     planOptions(ctx, deps, l.PlanFallback),
				CategoryOptions: categoryOptions(categories, l.CategoryFallback),
				AcceptTypes:     docxExt,
			})
		}
		if deps.StoreTemplate == nil || deps.DeleteTemplateObject == nil || deps.CreateUploadPair == nil {
			return view.HTMXError(l.NotConfigured)
		}

		viewCtx.Request.Body = http.MaxBytesReader(nil, viewCtx.Request.Body, maxRequestBytes)
		if err := viewCtx.Request.ParseMultipartForm(maxUploadBytes); err != nil {
			return view.HTMXError(l.UploadFailed)
		}
		name := strings.TrimSpace(viewCtx.Request.FormValue("name"))
		if name == "" {
			return view.HTMXError(l.NameLabel)
		}
		category, profile, ok := selectedCategory(categories, viewCtx.Request.FormValue("job_category_id"), deps.Options.SectionExport)
		if !ok {
			return view.HTMXError(l.CategoryRequiredForProfile)
		}
		priceScheduleID := strings.TrimSpace(viewCtx.Request.FormValue("price_schedule_id"))
		if priceScheduleID != "" && !containsOption(scheduleOptions(ctx, deps, l.ScheduleFallback), priceScheduleID) {
			return view.HTMXError(l.NotConfigured)
		}
		planID := strings.TrimSpace(viewCtx.Request.FormValue("plan_id"))
		if planID != "" && !containsOption(planOptions(ctx, deps, l.PlanFallback), planID) {
			return view.HTMXError(l.NotConfigured)
		}

		file, header, err := viewCtx.Request.FormFile("template_file")
		if err != nil {
			return view.HTMXError(l.InvalidFile)
		}
		defer file.Close()
		if !strings.EqualFold(filepath.Ext(header.Filename), docxExt) || header.Size > maxUploadBytes {
			return view.HTMXError(l.InvalidFile)
		}
		content, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
		if err != nil || len(content) == 0 || int64(len(content)) > maxUploadBytes {
			return view.HTMXError(l.UploadFailed)
		}
		if err := sectiondocument.ValidateTemplate(profile, content); err != nil {
			log.Printf("section template settings: manifest reject: %v", err)
			return view.HTMXError(l.InvalidManifest)
		}

		validityStart, err := optionalDate(viewCtx.Request.FormValue("validity_start"))
		if err != nil {
			return view.HTMXError(l.ValidityStartLabel)
		}
		validityEnd, err := optionalDate(viewCtx.Request.FormValue("validity_end"))
		if err != nil || (validityStart != nil && validityEnd != nil && !validityStart.AsTime().Before(validityEnd.AsTime())) {
			return view.HTMXError(l.ValidityEndLabel)
		}

		documentID := newID()
		bindingID := newID()
		objectKey := fmt.Sprintf("%s/%s%s", storagePrefix, documentID, docxExt)
		container, storeErr := deps.StoreTemplate(ctx, objectKey, content, docxContentType)
		if storeErr != nil {
			if container != "" {
				if cleanupErr := deps.DeleteTemplateObject(ctx, container, objectKey); cleanupErr != nil {
					log.Printf("section template settings: partial upload cleanup: %v", cleanupErr)
					return view.HTMXError(l.CleanupFailed)
				}
			}
			return view.HTMXError(l.UploadFailed)
		}

		originalName := filepath.Base(header.Filename)
		fileSize := int64(len(content))
		artifact := &documenttemplatepb.DocumentTemplate{
			Id: documentID, Name: name, TemplateType: "docx",
			DocumentPurpose: documentPurpose, StorageContainer: &container,
			StorageKey: &objectKey, OriginalFilename: &originalName,
			FileSizeBytes: &fileSize, Status: "active", Active: true,
		}
		if description := strings.TrimSpace(viewCtx.Request.FormValue("description")); description != "" {
			artifact.Description = &description
		}
		binding := &bindingpb.SubscriptionGroupDocumentTemplate{
			Id: bindingID, DocumentTemplateId: documentID, RenderProfile: profile,
			JobCategoryId: &category.Id, ValidityStart: validityStart,
			ValidityEnd: validityEnd, Active: true,
		}
		if priceScheduleID != "" {
			binding.PriceScheduleId = &priceScheduleID
		}
		if planID != "" {
			binding.PlanId = &planID
		}
		createdArtifact, createdBinding, err := deps.CreateUploadPair(ctx, artifact, binding)
		if err != nil || createdArtifact == nil || createdBinding == nil ||
			createdArtifact.GetId() != documentID || createdBinding.GetId() != bindingID {
			if err != nil {
				log.Printf("section template settings: atomic pair create: %v", err)
			}
			if cleanupErr := deps.DeleteTemplateObject(ctx, container, objectKey); cleanupErr != nil {
				log.Printf("section template settings: orphan object cleanup: %v", cleanupErr)
				return view.HTMXError(l.CleanupFailed)
			}
			return view.HTMXError(l.UploadFailed)
		}

		return view.ViewResult{StatusCode: http.StatusOK, Headers: map[string]string{
			"HX-Trigger": `{"formSuccess":true}`, "HX-Redirect": deps.Routes.SectionTemplateSettingsURL,
		}}
	})
}

func NewPublishAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		l := deps.Labels.SectionTemplateSettings
		if !view.GetUserPermissions(ctx).Can(bindingPermissionEntity, "update") || deps.PublishTemplateBinding == nil {
			return view.HTMXError(l.NotConfigured)
		}
		id := bindingID(viewCtx)
		if id == "" {
			return view.HTMXError(l.NotConfigured)
		}
		if _, err := deps.PublishTemplateBinding(ctx, &bindingpb.PublishSubscriptionGroupDocumentTemplateRequest{Id: id}); err != nil {
			log.Printf("section template settings: publish: %v", err)
			return view.HTMXError(err.Error())
		}
		return view.HTMXSuccess(tableID)
	})
}

func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		l := deps.Labels.SectionTemplateSettings
		perms := view.GetUserPermissions(ctx)
		if !canDelete(perms) || deps.DeleteDraftPair == nil || deps.DeleteTemplateObject == nil {
			return view.HTMXError(l.NotConfigured)
		}
		id := bindingID(viewCtx)
		artifact, err := deps.DeleteDraftPair(ctx, id)
		if err != nil || artifact == nil || artifact.GetStorageContainer() == "" || artifact.GetStorageKey() == "" {
			if err != nil {
				log.Printf("section template settings: atomic draft-pair delete: %v", err)
			}
			return view.HTMXError(l.NotConfigured)
		}
		if err := deps.DeleteTemplateObject(ctx, artifact.GetStorageContainer(), artifact.GetStorageKey()); err != nil {
			log.Printf("section template settings: delete object: %v", err)
			return view.HTMXError(l.CleanupFailed)
		}
		return view.HTMXSuccess(tableID)
	})
}

func canCreate(perms *types.UserPermissions) bool {
	return perms.Can(bindingPermissionEntity, "create") && perms.Can(artifactPermissionEntity, "create")
}

func canDelete(perms *types.UserPermissions) bool {
	return perms.Can(bindingPermissionEntity, "delete") && perms.Can(artifactPermissionEntity, "delete")
}

func bindingColumns(l outcome_summary.SectionTemplateSettingsLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.NameColumn},
		{Key: "schedule", Label: l.ScheduleColumn},
		{Key: "plan", Label: l.PlanColumn},
		{Key: "category", Label: l.CategoryColumn},
		{Key: "profile", Label: l.ProfileColumn},
		{Key: "version", Label: l.VersionColumn},
		{Key: "status", Label: l.StatusColumn},
		{Key: "validity", Label: l.ValidityColumn},
	}
}

func bindingRows(ctx context.Context, deps *Deps, perms *types.UserPermissions) []types.TableRow {
	bindings, err := listAllTemplateBindings(ctx, deps)
	if err != nil {
		log.Printf("section template settings: list: %v", err)
		return nil
	}
	if len(bindings) == 0 {
		return nil
	}
	l := deps.Labels.SectionTemplateSettings
	rows := make([]types.TableRow, 0, len(bindings))
	for _, binding := range bindings {
		if binding == nil {
			continue
		}
		name := "—"
		if artifact := binding.GetDocumentTemplate(); artifact != nil && strings.TrimSpace(artifact.GetName()) != "" {
			name = artifact.GetName()
		}
		actions := []types.TableAction{}
		if binding.GetVersionStatus() == enums.VersionStatus_VERSION_STATUS_DRAFT {
			confirm := l.PublishConfirm
			if binding.GetPriceScheduleId() == "" || binding.GetPlanId() == "" {
				confirm = l.BroadScopeConfirm
			}
			actions = append(actions, types.TableAction{Type: "activate", Label: l.PublishAction, Action: "activate", URL: deps.Routes.SectionTemplatePublishURL, ItemName: name, ConfirmTitle: l.PublishAction, ConfirmMessage: confirm, TestID: "section-template-publish-" + short(binding.GetId()), Disabled: !perms.Can(bindingPermissionEntity, "update"), DisabledTooltip: l.NotConfigured})
			actions = append(actions, types.TableAction{Type: "delete", Label: l.DeleteAction, Action: "delete", URL: deps.Routes.SectionTemplateDeleteURL, ItemName: name, ConfirmTitle: l.DeleteAction, ConfirmMessage: l.DeleteConfirm, TestID: "section-template-delete-" + short(binding.GetId()), Disabled: !canDelete(perms), DisabledTooltip: l.NotConfigured})
		}
		status, variant := statusBadge(binding.GetVersionStatus(), l)
		profileKey := ""
		if profile, ok := sectiondocument.LookupProfile(binding.GetRenderProfile()); ok {
			profileKey = profile.Key
		}
		categoryCode := ""
		if category := binding.GetJobCategory(); category != nil {
			categoryCode = strings.TrimSpace(category.GetCode())
		}
		rows = append(rows, types.TableRow{ID: binding.GetId(), Cells: []types.TableCell{
			{Type: "text", Value: name},
			{Type: "text", Value: nestedScheduleName(binding, l.ScheduleFallback)},
			{Type: "text", Value: nestedPlanName(binding, l.PlanFallback)},
			{Type: "text", Value: nestedCategoryName(binding, l.CategoryFallback)},
			{Type: "text", Value: profileLabel(binding.GetRenderProfile(), l)},
			{Type: "text", Value: fmt.Sprintf("v%d", binding.GetVersion())},
			{Type: "badge", Value: status, Variant: variant},
			{Type: "text", Value: validity(binding)},
		}, DataAttrs: map[string]string{
			"testid":        "section-template-row-" + short(binding.GetId()),
			"renderprofile": profileKey,
			"category":      categoryCode,
			"categoryscope": axisScope(binding.GetJobCategoryId()),
			"schedulescope": axisScope(binding.GetPriceScheduleId()),
			"planscope":     axisScope(binding.GetPlanId()),
			"status":        statusKey(binding.GetVersionStatus()),
		}, Actions: actions})
	}
	return rows
}

func listAllTemplateBindings(ctx context.Context, deps *Deps) ([]*bindingpb.SubscriptionGroupDocumentTemplate, error) {
	if deps.ListTemplateBindings == nil {
		return nil, nil
	}
	bindings := make([]*bindingpb.SubscriptionGroupDocumentTemplate, 0, templateBindingPageLimit)
	for page := int32(1); page <= templateBindingMaxPages; page++ {
		resp, err := deps.ListTemplateBindings(ctx, &bindingpb.ListSubscriptionGroupDocumentTemplatesRequest{
			Pagination: templateBindingPagination(page),
		})
		if err != nil || resp == nil || !resp.GetSuccess() {
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		for _, binding := range resp.GetData() {
			if binding != nil {
				bindings = append(bindings, binding)
			}
		}
		if len(resp.GetData()) < templateBindingPageLimit {
			break
		}
		if page == templateBindingMaxPages {
			return nil, fmt.Errorf("section template settings: list template bindings: %d full pages without termination", templateBindingMaxPages)
		}
	}
	return bindings, nil
}

func templateBindingPagination(page int32) *commonpb.PaginationRequest {
	return &commonpb.PaginationRequest{
		Limit: int32(templateBindingPageLimit),
		Method: &commonpb.PaginationRequest_Offset{
			Offset: &commonpb.OffsetPagination{Page: page},
		},
	}
}

func mappedCategories(ctx context.Context, deps *Deps) ([]*jobcategorypb.JobCategory, error) {
	if deps.ListJobCategories == nil {
		return nil, fmt.Errorf("job category list is unavailable")
	}
	resp, err := deps.ListJobCategories(ctx, &jobcategorypb.ListJobCategoriesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list job categories: %w", err)
	}
	if resp == nil || !resp.GetSuccess() {
		return nil, fmt.Errorf("list job categories: unavailable response")
	}
	result := make([]*jobcategorypb.JobCategory, 0, len(resp.GetData()))
	for _, category := range resp.GetData() {
		if category == nil || category.GetId() == "" {
			continue
		}
		profile, ok := deps.Options.SectionExport.ProfileForCategoryCode(category.GetCode())
		if _, registered := sectiondocument.LookupProfile(profile); ok && registered {
			result = append(result, category)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].GetSortOrder() != result[j].GetSortOrder() {
			return result[i].GetSortOrder() < result[j].GetSortOrder()
		}
		in, jn := strings.ToLower(result[i].GetName()), strings.ToLower(result[j].GetName())
		if in != jn {
			return in < jn
		}
		return result[i].GetId() < result[j].GetId()
	})
	return result, nil
}

func selectedCategory(categories []*jobcategorypb.JobCategory, id string, options outcome_summary.SectionExportOptions) (*jobcategorypb.JobCategory, bindingpb.RenderProfile, bool) {
	id = strings.TrimSpace(id)
	for _, category := range categories {
		if category.GetId() != id {
			continue
		}
		profile, ok := options.ProfileForCategoryCode(category.GetCode())
		_, registered := sectiondocument.LookupProfile(profile)
		return category, profile, ok && registered
	}
	return nil, bindingpb.RenderProfile_RENDER_PROFILE_UNSPECIFIED, false
}

func categoryOptions(categories []*jobcategorypb.JobCategory, fallback string) []types.SelectOption {
	options := []types.SelectOption{{Value: "", Label: fallback}}
	for _, category := range categories {
		options = append(options, types.SelectOption{Value: category.GetId(), Label: category.GetName()})
	}
	return options
}

func scheduleOptions(ctx context.Context, deps *Deps, fallback string) []types.SelectOption {
	options := []types.SelectOption{{Value: "", Label: fallback}}
	if deps.ListPriceSchedules == nil {
		return options
	}
	resp, err := deps.ListPriceSchedules(ctx, &priceschedulepb.ListPriceSchedulesRequest{})
	if err != nil || resp == nil || !resp.GetSuccess() {
		return options
	}
	for _, item := range resp.GetData() {
		if item.GetId() != "" {
			options = append(options, types.SelectOption{Value: item.GetId(), Label: item.GetName()})
		}
	}
	sortOptions(options[1:])
	return options
}

func planOptions(ctx context.Context, deps *Deps, fallback string) []types.SelectOption {
	options := []types.SelectOption{{Value: "", Label: fallback}}
	if deps.ListPlans == nil {
		return options
	}
	resp, err := deps.ListPlans(ctx, &planpb.ListPlansRequest{})
	if err != nil || resp == nil || !resp.GetSuccess() {
		return options
	}
	for _, item := range resp.GetData() {
		if item.GetId() != "" {
			options = append(options, types.SelectOption{Value: item.GetId(), Label: item.GetName()})
		}
	}
	sortOptions(options[1:])
	return options
}

func sortOptions(options []types.SelectOption) {
	sort.SliceStable(options, func(i, j int) bool {
		li, lj := strings.ToLower(options[i].Label), strings.ToLower(options[j].Label)
		if li != lj {
			return li < lj
		}
		return options[i].Value < options[j].Value
	})
}

func containsOption(options []types.SelectOption, id string) bool {
	for _, option := range options {
		if option.Value == id {
			return true
		}
	}
	return false
}

func optionalDate(raw string) (*timestamppb.Timestamp, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(dateLayout, raw)
	if err != nil {
		return nil, err
	}
	return timestamppb.New(parsed.UTC()), nil
}

func bindingID(viewCtx *view.ViewContext) string {
	if viewCtx == nil || viewCtx.Request == nil {
		return ""
	}
	if id := strings.TrimSpace(viewCtx.Request.URL.Query().Get("id")); id != "" {
		return id
	}
	return strings.TrimSpace(viewCtx.Request.FormValue("id"))
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(raw[:])
}

func short(id string) string {
	if len(id) > 8 {
		return id[len(id)-8:]
	}
	return id
}

func axisScope(id string) string {
	if strings.TrimSpace(id) == "" {
		return "all"
	}
	return "exact"
}

func statusKey(status enums.VersionStatus) string {
	switch status {
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "published"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "deprecated"
	default:
		return "draft"
	}
}

func profileLabel(profile bindingpb.RenderProfile, l outcome_summary.SectionTemplateSettingsLabels) string {
	if profile == bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1 {
		return l.ProfileSubscriptionGroupOutcomeMatrixSinglePeriod11V1
	}
	return "—"
}

func nestedScheduleName(binding *bindingpb.SubscriptionGroupDocumentTemplate, fallback string) string {
	if binding.GetPriceScheduleId() == "" {
		return fallback
	}
	if item := binding.GetPriceSchedule(); item != nil && item.GetName() != "" {
		return item.GetName()
	}
	return binding.GetPriceScheduleId()
}

func nestedPlanName(binding *bindingpb.SubscriptionGroupDocumentTemplate, fallback string) string {
	if binding.GetPlanId() == "" {
		return fallback
	}
	if item := binding.GetPlan(); item != nil && item.GetName() != "" {
		return item.GetName()
	}
	return binding.GetPlanId()
}

func nestedCategoryName(binding *bindingpb.SubscriptionGroupDocumentTemplate, fallback string) string {
	if binding.GetJobCategoryId() == "" {
		return fallback
	}
	if item := binding.GetJobCategory(); item != nil && item.GetName() != "" {
		return item.GetName()
	}
	return binding.GetJobCategoryId()
}

func statusBadge(status enums.VersionStatus, l outcome_summary.SectionTemplateSettingsLabels) (string, string) {
	switch status {
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return l.StatusPublished, "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return l.StatusDeprecated, "warning"
	default:
		return l.StatusDraft, "default"
	}
}

func validity(binding *bindingpb.SubscriptionGroupDocumentTemplate) string {
	start, end := "", ""
	if binding.GetValidityStart() != nil {
		start = binding.GetValidityStart().AsTime().Format(dateLayout)
	}
	if binding.GetValidityEnd() != nil {
		end = binding.GetValidityEnd().AsTime().Format(dateLayout)
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
