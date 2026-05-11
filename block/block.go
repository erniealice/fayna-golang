// Package block implements the Lego pattern for the fayna domain.
//
// Block() returns a pyeza.AppOption that registers all fayna modules
// (operations: jobs, job templates, job activities, outcome criteria,
// task outcomes, outcome summaries, and fulfillment) using AppContext
// as the shared infrastructure carrier.
//
// Usage:
//
//	// Register all fayna modules
//	app.Apply(faynablock.Block())
//
//	// Register only specific modules
//	app.Apply(faynablock.Block(faynablock.WithJob(), faynablock.WithFulfillment()))
package block

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/erniealice/espyna-golang/reference"
	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	fayna "github.com/erniealice/fayna-golang"
	activityexpensemod "github.com/erniealice/fayna-golang/views/activity_expense"
	activitylabormod "github.com/erniealice/fayna-golang/views/activity_labor"
	activitymaterialmod "github.com/erniealice/fayna-golang/views/activity_material"
	fulfillmentmod "github.com/erniealice/fayna-golang/views/fulfillment"
	jobmod "github.com/erniealice/fayna-golang/views/job"
	jobactivitymod "github.com/erniealice/fayna-golang/views/job_activity"
	jobphasemod "github.com/erniealice/fayna-golang/views/job_phase"
	jobtaskmod "github.com/erniealice/fayna-golang/views/job_task"
	jobtemplatemod "github.com/erniealice/fayna-golang/views/job_template"
	jobtemplatePhasemod "github.com/erniealice/fayna-golang/views/job_template_phase"
	jobtemplateTaskmod "github.com/erniealice/fayna-golang/views/job_template_task"
	outcomecriteriaMod "github.com/erniealice/fayna-golang/views/outcome_criteria"
	outcomesummaryMod "github.com/erniealice/fayna-golang/views/outcome_summary"
	taskoutcomeMod "github.com/erniealice/fayna-golang/views/task_outcome"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"
	pyeza "github.com/erniealice/pyeza-golang"
)

// ---------------------------------------------------------------------------
// BlockOption — per-module granular selection
// ---------------------------------------------------------------------------

// BlockOption enables specific fayna sub-modules within Block().
type BlockOption func(*blockConfig)

type blockConfig struct {
	enableAll        bool
	job              bool
	jobTemplate      bool
	jobActivity      bool
	jobPhase         bool
	jobTask          bool
	activityLabor    bool
	activityMaterial bool
	activityExpense  bool
	outcomeCriteria  bool
	taskOutcome      bool
	outcomeSummary   bool
	fulfillment      bool
	jobTemplatePhase bool
	jobTemplateTask  bool
	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 — cross-package
	// URL pattern (e.g. "/app/subscriptions/detail/{id}") supplied by the
	// consuming app via WithSubscriptionDetailURL. Empty = breadcrumb hidden.
	subscriptionDetailURL string
}

// WithJob registers the Job module (list, detail, CRUD, attachment ops).
func WithJob() BlockOption { return func(c *blockConfig) { c.job = true } }

// WithJobTemplate registers the JobTemplate module (list, detail, CRUD, attachment ops).
func WithJobTemplate() BlockOption { return func(c *blockConfig) { c.jobTemplate = true } }

// WithJobActivity registers the JobActivity module (list, detail, CRUD, approval workflow).
func WithJobActivity() BlockOption { return func(c *blockConfig) { c.jobActivity = true } }

// WithJobPhase registers the JobPhase standalone module (list + detail + CRUD + bulk ops).
// The list page is a power-user/debugging surface; phases are normally accessed via the
// Job detail Phases tab deep links.
func WithJobPhase() BlockOption { return func(c *blockConfig) { c.jobPhase = true } }

// WithJobTask registers the JobTask standalone module (list + detail + CRUD + bulk ops).
// The list page is a power-user/debugging surface; tasks are normally accessed via the
// JobPhase detail Tasks tab deep links.
func WithJobTask() BlockOption { return func(c *blockConfig) { c.jobTask = true } }

// WithActivityLabor registers the ActivityLabor sibling module (labor charge detail).
// Not in the sidebar — reached via JobActivity detail charge tab (entry_type=LABOR).
func WithActivityLabor() BlockOption { return func(c *blockConfig) { c.activityLabor = true } }

// WithActivityMaterial registers the ActivityMaterial sibling module (material charge detail).
// Not in the sidebar — reached via JobActivity detail charge tab (entry_type=MATERIAL).
func WithActivityMaterial() BlockOption { return func(c *blockConfig) { c.activityMaterial = true } }

// WithActivityExpense registers the ActivityExpense sibling module (expense charge detail).
// Not in the sidebar — reached via JobActivity detail charge tab (entry_type=EXPENSE).
func WithActivityExpense() BlockOption { return func(c *blockConfig) { c.activityExpense = true } }

// WithOutcomeCriteria registers the OutcomeCriteria module (list, detail, CRUD).
func WithOutcomeCriteria() BlockOption { return func(c *blockConfig) { c.outcomeCriteria = true } }

// WithTaskOutcome registers the TaskOutcome module (list, detail, CRUD).
func WithTaskOutcome() BlockOption { return func(c *blockConfig) { c.taskOutcome = true } }

// WithOutcomeSummary registers the OutcomeSummary module (job + phase summaries).
func WithOutcomeSummary() BlockOption { return func(c *blockConfig) { c.outcomeSummary = true } }

// WithFulfillment registers the Fulfillment module (list, detail, CRUD, status transitions).
func WithFulfillment() BlockOption { return func(c *blockConfig) { c.fulfillment = true } }

// WithJobTemplatePhase registers the JobTemplatePhase drawer-only module (Add/Edit/Delete CTAs).
// Not in the sidebar — reached via the JobTemplate detail Phases tab.
func WithJobTemplatePhase() BlockOption {
	return func(c *blockConfig) { c.jobTemplatePhase = true }
}

// WithJobTemplateTask registers the JobTemplateTask drawer-only module (Add/Edit/Delete CTAs).
// Not in the sidebar — reached via the JobTemplate detail Tasks tab.
func WithJobTemplateTask() BlockOption {
	return func(c *blockConfig) { c.jobTemplateTask = true }
}

// WithSubscriptionDetailURL supplies the centymo subscription-detail path
// template (e.g. "/app/subscriptions/detail/{id}") so the Job detail page
// can render a "Spawned from Subscription" breadcrumb when
// Job.origin_type = SUBSCRIPTION. Optional — when unset the breadcrumb is
// hidden.
// 2026-04-29 auto-spawn-jobs-from-subscription Phase D.
func WithSubscriptionDetailURL(url string) BlockOption {
	return func(c *blockConfig) { c.subscriptionDetailURL = url }
}

func (c *blockConfig) wantJob() bool              { return c.enableAll || c.job }
func (c *blockConfig) wantJobTemplate() bool      { return c.enableAll || c.jobTemplate }
func (c *blockConfig) wantJobActivity() bool      { return c.enableAll || c.jobActivity }
func (c *blockConfig) wantJobPhase() bool         { return c.enableAll || c.jobPhase }
func (c *blockConfig) wantJobTask() bool          { return c.enableAll || c.jobTask }
func (c *blockConfig) wantActivityLabor() bool    { return c.enableAll || c.activityLabor }
func (c *blockConfig) wantActivityMaterial() bool { return c.enableAll || c.activityMaterial }
func (c *blockConfig) wantActivityExpense() bool  { return c.enableAll || c.activityExpense }
func (c *blockConfig) wantOutcomeCriteria() bool  { return c.enableAll || c.outcomeCriteria }
func (c *blockConfig) wantTaskOutcome() bool      { return c.enableAll || c.taskOutcome }
func (c *blockConfig) wantOutcomeSummary() bool   { return c.enableAll || c.outcomeSummary }
func (c *blockConfig) wantFulfillment() bool      { return c.enableAll || c.fulfillment }
func (c *blockConfig) wantJobTemplatePhase() bool { return c.enableAll || c.jobTemplatePhase }
func (c *blockConfig) wantJobTemplateTask() bool  { return c.enableAll || c.jobTemplateTask }

// ---------------------------------------------------------------------------
// routeRegistrarFull — optional HandleFunc extension for raw HTTP handlers.
// ---------------------------------------------------------------------------

// routeRegistrarFull extends RouteRegistrar with HandleFunc for JSON endpoints.
type routeRegistrarFull interface {
	pyeza.RouteRegistrar
	HandleFunc(method, path string, handler http.HandlerFunc, middlewares ...string)
}

// handleFunc registers an http.HandlerFunc route if the registrar supports it.
// Silently skips if the registrar does not implement HandleFunc.
func handleFunc(r pyeza.RouteRegistrar, method, path string, handler http.HandlerFunc) {
	if path == "" || handler == nil {
		return
	}
	if full, ok := r.(routeRegistrarFull); ok {
		full.HandleFunc(method, path, handler)
		return
	}
	log.Printf("fayna.Block: RouteRegistrar does not support HandleFunc — skipping %s %s", method, path)
}

// ---------------------------------------------------------------------------
// listSimpler — minimal DB interface for location search (avoids centymo import).
// ---------------------------------------------------------------------------

// listSimpler is satisfied by centymo.DataSource and espyna's DatabaseAdapter.
type listSimpler interface {
	ListSimple(ctx context.Context, collection string) ([]map[string]any, error)
}

// ---------------------------------------------------------------------------
// Search handler builders for job drawer auto-complete pickers.
// ---------------------------------------------------------------------------

// searchOption is the JSON shape for auto-complete responses.
type searchOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

const jobSearchResultLimit = 20

// writeSearchJSON writes a JSON response for the auto-complete component.
func writeSearchJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("fayna.Block: failed to encode search JSON: %v", err)
	}
}

// newJobClientSearchHandler builds a JSON search handler for the client picker.
// Uses SearchClientsByName use case when available (SQL ILIKE); falls back to
// ListClients with in-process filtering. Returns nil when no client use cases
// are wired.
func newJobClientSearchHandler(
	searchByName func(ctx context.Context, req *clientpb.SearchClientsByNameRequest) (*clientpb.SearchClientsByNameResponse, error),
	listClients func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error),
) http.HandlerFunc {
	if searchByName == nil && listClients == nil {
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := strings.TrimSpace(r.URL.Query().Get("q"))

		if searchByName != nil {
			resp, err := searchByName(ctx, &clientpb.SearchClientsByNameRequest{Query: query})
			if err != nil {
				log.Printf("fayna.Block: client search failed: %v", err)
				writeSearchJSON(w, []searchOption{})
				return
			}
			results := make([]searchOption, 0, len(resp.GetResults()))
			for _, item := range resp.GetResults() {
				results = append(results, searchOption{Value: item.GetId(), Label: item.GetLabel()})
			}
			writeSearchJSON(w, results)
			return
		}

		// Fallback: full list + in-process filter.
		resp, err := listClients(ctx, &clientpb.ListClientsRequest{})
		if err != nil {
			log.Printf("fayna.Block: list clients for search failed: %v", err)
			writeSearchJSON(w, []searchOption{})
			return
		}
		queryLower := strings.ToLower(query)
		var results []searchOption
		for _, c := range resp.GetData() {
			if !c.GetActive() {
				continue
			}
			label := c.GetName()
			if label == "" {
				if u := c.GetUser(); u != nil {
					label = strings.TrimSpace(u.GetFirstName() + " " + u.GetLastName())
				}
			}
			if label == "" {
				label = c.GetId()
			}
			if queryLower != "" && !strings.Contains(strings.ToLower(label), queryLower) {
				continue
			}
			results = append(results, searchOption{Value: c.GetId(), Label: label})
			if len(results) >= jobSearchResultLimit {
				break
			}
		}
		if results == nil {
			results = []searchOption{}
		}
		writeSearchJSON(w, results)
	}
}

// newActivityLaborStaffSearchHandler builds a JSON search handler for the staff picker
// used by the activity labor drawer form.
// Uses ListStaffs use case with in-process name filter.
// Returns nil when listStaffs is nil.
//
// TODO(P5 wave 3): add SearchStaffByName use case to espyna entity/staff for SQL ILIKE
// (mirrors the client search pattern). For now we fall back to ListStaffs + in-process filter.
// When neither use case is available, StaffSearchURL is left empty and the auto-complete
// falls back to flat filter mode (operator types staff_id directly).
func newActivityLaborStaffSearchHandler(
	listStaffs func(ctx context.Context, req *staffpb.ListStaffsRequest) (*staffpb.ListStaffsResponse, error),
) http.HandlerFunc {
	if listStaffs == nil {
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		queryLower := strings.ToLower(query)

		resp, err := listStaffs(ctx, &staffpb.ListStaffsRequest{})
		if err != nil {
			log.Printf("fayna.Block: staff search failed: %v", err)
			writeSearchJSON(w, []searchOption{})
			return
		}

		var results []searchOption
		for _, s := range resp.GetData() {
			if !s.GetActive() {
				continue
			}
			id := s.GetId()
			if id == "" {
				continue
			}
			// Build display label: first + last name from embedded user, fallback to id.
			label := ""
			if u := s.GetUser(); u != nil {
				label = strings.TrimSpace(u.GetFirstName() + " " + u.GetLastName())
			}
			if label == "" {
				label = id
			}
			if queryLower != "" && !strings.Contains(strings.ToLower(label), queryLower) {
				continue
			}
			results = append(results, searchOption{Value: id, Label: label})
			if len(results) >= jobSearchResultLimit {
				break
			}
		}
		if results == nil {
			results = []searchOption{}
		}
		writeSearchJSON(w, results)
	}
}

// newJobLocationSearchHandler builds a JSON search handler for the location picker.
// Uses db.ListSimple("location") with in-process active+name filter.
// Returns nil when db is nil.
func newJobLocationSearchHandler(db listSimpler) http.HandlerFunc {
	if db == nil {
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		queryLower := strings.ToLower(query)

		records, err := db.ListSimple(ctx, "location")
		if err != nil {
			log.Printf("fayna.Block: location search failed: %v", err)
			writeSearchJSON(w, []searchOption{})
			return
		}
		var results []searchOption
		for _, rec := range records {
			active, _ := rec["active"].(bool)
			if !active {
				continue
			}
			id, _ := rec["id"].(string)
			name, _ := rec["name"].(string)
			if id == "" {
				continue
			}
			if queryLower != "" && !strings.Contains(strings.ToLower(name), queryLower) {
				continue
			}
			results = append(results, searchOption{Value: id, Label: name})
			if len(results) >= jobSearchResultLimit {
				break
			}
		}
		if results == nil {
			results = []searchOption{}
		}
		writeSearchJSON(w, results)
	}
}

// newActivityMaterialProductSearchHandler builds a JSON search handler for the product picker
// used by the activity material drawer form.
// Uses db.ListSimple("product") with in-process active+name filter.
// Returns nil when db is nil.
//
// TODO(P5): add SearchProductsByName use case to espyna for SQL ILIKE (mirrors client pattern).
func newActivityMaterialProductSearchHandler(db listSimpler) http.HandlerFunc {
	if db == nil {
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		queryLower := strings.ToLower(query)

		records, err := db.ListSimple(ctx, "product")
		if err != nil {
			log.Printf("fayna.Block: product search failed: %v", err)
			writeSearchJSON(w, []searchOption{})
			return
		}
		var results []searchOption
		for _, rec := range records {
			active, _ := rec["active"].(bool)
			if !active {
				continue
			}
			id, _ := rec["id"].(string)
			name, _ := rec["name"].(string)
			if id == "" {
				continue
			}
			if queryLower != "" && !strings.Contains(strings.ToLower(name), queryLower) {
				continue
			}
			results = append(results, searchOption{Value: id, Label: name})
			if len(results) >= jobSearchResultLimit {
				break
			}
		}
		if results == nil {
			results = []searchOption{}
		}
		writeSearchJSON(w, results)
	}
}

// newActivityExpenseExpenseCategorySearchHandler builds a JSON search handler for the
// expense category picker used by the activity expense drawer form.
// Uses db.ListSimple("expense_category") with in-process active+name filter.
// Returns nil when db is nil.
//
// TODO(P5): add SearchExpenseCategoriesByName use case to espyna for SQL ILIKE.
func newActivityExpenseExpenseCategorySearchHandler(db listSimpler) http.HandlerFunc {
	if db == nil {
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		queryLower := strings.ToLower(query)

		records, err := db.ListSimple(ctx, "expense_category")
		if err != nil {
			log.Printf("fayna.Block: expense category search failed: %v", err)
			writeSearchJSON(w, []searchOption{})
			return
		}
		var results []searchOption
		for _, rec := range records {
			active, _ := rec["active"].(bool)
			if !active {
				continue
			}
			id, _ := rec["id"].(string)
			name, _ := rec["name"].(string)
			if id == "" {
				continue
			}
			if queryLower != "" && !strings.Contains(strings.ToLower(name), queryLower) {
				continue
			}
			results = append(results, searchOption{Value: id, Label: name})
			if len(results) >= jobSearchResultLimit {
				break
			}
		}
		if results == nil {
			results = []searchOption{}
		}
		writeSearchJSON(w, results)
	}
}

// ---------------------------------------------------------------------------
// Block — the main Lego entry point
// ---------------------------------------------------------------------------

// Block registers fayna domain modules (operations: jobs, templates, activities,
// outcomes, fulfillment). Call with no options to register ALL modules. Call with
// specific With*() options to register a subset.
func Block(opts ...BlockOption) pyeza.AppOption {
	cfg := &blockConfig{enableAll: len(opts) == 0}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx *pyeza.AppContext) error {
		// --- Type-assert translations ---
		translations, ok := ctx.Translations.(*lynguaV1.TranslationProvider)
		if !ok || translations == nil {
			return fmt.Errorf("fayna.Block: ctx.Translations must be *lynguaV1.TranslationProvider")
		}

		// --- Type-assert reference checker (optional; nil = no in-use gating) ---
		var refChecker reference.Checker
		if ctx.RefChecker != nil {
			refChecker, _ = ctx.RefChecker.(reference.Checker)
		}

		// --- Type-assert attachment operations ---
		uploadFile, _ := ctx.UploadFile.(func(context.Context, string, string, []byte, string) error)
		listAttachments, _ := ctx.ListAttachments.(func(context.Context, string, string) (*attachmentpb.ListAttachmentsResponse, error))
		createAttachment, _ := ctx.CreateAttachment.(func(context.Context, *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error))
		deleteAttachment, _ := ctx.DeleteAttachment.(func(context.Context, *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error))
		newAttachmentID, _ := ctx.NewAttachmentID.(func() string)

		// --- Load routes (defaults + optional lyngua overrides) ---
		jobRoutes := fayna.DefaultJobRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "job", &jobRoutes)

		jtRoutes := fayna.DefaultJobTemplateRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "job_template", &jtRoutes)

		jaRoutes := fayna.DefaultJobActivityRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "job_activity", &jaRoutes)

		jpRoutes := fayna.DefaultJobPhaseRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "job_phase", &jpRoutes)

		jkRoutes := fayna.DefaultJobTaskRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "job_task", &jkRoutes)

		alRoutes := fayna.DefaultActivityLaborRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "activity_labor", &alRoutes)

		amRoutes := fayna.DefaultActivityMaterialRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "activity_material", &amRoutes)

		aeRoutes := fayna.DefaultActivityExpenseRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "activity_expense", &aeRoutes)

		ocRoutes := fayna.DefaultOutcomeCriteriaRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "outcome_criteria", &ocRoutes)

		toRoutes := fayna.DefaultTaskOutcomeRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "task_outcome", &toRoutes)

		osRoutes := fayna.DefaultOutcomeSummaryRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "outcome_summary", &osRoutes)

		ffRoutes := fayna.DefaultFulfillmentRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "fulfillment", &ffRoutes)

		jtpRoutes := fayna.DefaultJobTemplatePhaseRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "job_template_phase", &jtpRoutes)

		jttRoutes := fayna.DefaultJobTemplateTaskRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "job_template_task", &jttRoutes)

		// --- Load labels (defaults + optional lyngua overrides) ---
		jobLabels := fayna.DefaultJobLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job.json", "job", &jobLabels)

		jtLabels := fayna.DefaultJobTemplateLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_template.json", "job_template", &jtLabels)

		jaLabels := fayna.DefaultJobActivityLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_activity.json", "job_activity", &jaLabels)

		jpLabels := fayna.DefaultJobPhaseLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_phase.json", "job_phase", &jpLabels)

		jkLabels := fayna.DefaultJobTaskLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_task.json", "job_task", &jkLabels)

		alLabels := fayna.DefaultActivityLaborLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "activity_labor.json", "activity_labor", &alLabels)

		amLabels := fayna.DefaultActivityMaterialLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "activity_material.json", "activity_material", &amLabels)

		aeLabels := fayna.DefaultActivityExpenseLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "activity_expense.json", "activity_expense", &aeLabels)

		ocLabels := fayna.DefaultOutcomeCriteriaLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "outcome_criteria.json", "outcome_criteria", &ocLabels)

		toLabels := fayna.DefaultTaskOutcomeLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "task_outcome.json", "task_outcome", &toLabels)

		osLabels := fayna.DefaultOutcomeSummaryLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "outcome_summary.json", "outcome_summary", &osLabels)

		ffLabels := defaultFulfillmentLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "fulfillment.json", "fulfillment", &ffLabels)

		jtpLabels := fayna.DefaultJobTemplatePhaseLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_template_phase.json", "job_template_phase", &jtpLabels)

		jttLabels := fayna.DefaultJobTemplateTaskLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_template_task.json", "job_template_task", &jttLabels)

		// --- Reflect into use cases aggregate (espyna-free wiring) ---
		uc := assertUseCases(ctx.UseCases)

		// --- Build job drawer search handlers (client + location pickers) ---
		// These are registered as HandleFunc routes alongside the job module routes.
		// Client search uses entity use cases via reflection (SQL ILIKE when available,
		// ListClients fallback). Location search uses ctx.DB (ListSimple on "location").
		var jobClientSearchFn, jobLocationSearchFn http.HandlerFunc
		// activityLaborStaffSearchFn is the staff picker for the activity labor drawer.
		var activityLaborStaffSearchFn http.HandlerFunc
		// activityMaterialProductSearchFn is the product picker for the activity material drawer.
		var activityMaterialProductSearchFn http.HandlerFunc
		// activityMaterialLocationSearchFn is the location picker for the activity material drawer.
		var activityMaterialLocationSearchFn http.HandlerFunc
		// activityExpenseExpenseCategorySearchFn is the expense category picker for the activity expense drawer.
		var activityExpenseExpenseCategorySearchFn http.HandlerFunc
		if uc != nil {
			// Extract SearchClientsByName and ListClients from UseCases.Entity.Client
			// via the same reflection pattern used by wireJobDeps.
			entity := ptrField(uc.v, "Entity")
			if entity.IsValid() {
				clientUC := ptrField(entity, "Client")
				if clientUC.IsValid() {
					searchByNameRaw := execFn(clientUC, "SearchClientsByName")
					listClientsRaw := execFn(clientUC, "ListClients")
					searchByName, _ := searchByNameRaw.(func(context.Context, *clientpb.SearchClientsByNameRequest) (*clientpb.SearchClientsByNameResponse, error))
					listClients, _ := listClientsRaw.(func(context.Context, *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error))
					jobClientSearchFn = newJobClientSearchHandler(searchByName, listClients)
				}

				// Extract ListStaffs from UseCases.Entity.Staff for the activity labor staff picker.
				// TODO(P5 wave 3): add SearchStaffByName to espyna for SQL ILIKE (mirrors client pattern).
				staffUC := ptrField(entity, "Staff")
				if staffUC.IsValid() {
					listStaffsRaw := execFn(staffUC, "ListStaffs")
					listStaffs, _ := listStaffsRaw.(func(context.Context, *staffpb.ListStaffsRequest) (*staffpb.ListStaffsResponse, error))
					activityLaborStaffSearchFn = newActivityLaborStaffSearchHandler(listStaffs)
				}
			}
		}
		if db, ok := ctx.DB.(listSimpler); ok {
			jobLocationSearchFn = newJobLocationSearchHandler(db)
			activityMaterialProductSearchFn = newActivityMaterialProductSearchHandler(db)
			activityMaterialLocationSearchFn = newJobLocationSearchHandler(db)
			activityExpenseExpenseCategorySearchFn = newActivityExpenseExpenseCategorySearchHandler(db)
		}

		// Wire the staff search URL into activityLabor routes only when the handler is available.
		// When nil the URL is left empty and the auto-complete falls back to flat filter mode.
		if activityLaborStaffSearchFn == nil {
			alRoutes.StaffSearchURL = ""
			log.Printf("fayna.Block: staff search handler not available — ActivityLabor drawer will use flat filter mode for staff picker")
		}
		// Wire product/location search URLs into activityMaterial routes.
		if activityMaterialProductSearchFn == nil {
			amRoutes.ProductSearchURL = ""
		}
		if activityMaterialLocationSearchFn == nil {
			amRoutes.LocationSearchURL = ""
		}
		// Wire expense category search URL into activityExpense routes.
		if activityExpenseExpenseCategorySearchFn == nil {
			aeRoutes.ExpenseCategorySearchURL = ""
		}

		// --- Register Job module ---
		if cfg.wantJob() {
			jobDeps := &jobmod.ModuleDeps{
				Routes:           jobRoutes,
				Labels:           jobLabels,
				CommonLabels:     ctx.Common,
				TableLabels:      ctx.Table,
				UploadFile:       uploadFile,
				ListAttachments:  listAttachments,
				CreateAttachment: createAttachment,
				DeleteAttachment: deleteAttachment,
				NewID:            newAttachmentID,
				// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4.
				SubscriptionDetailURL: cfg.subscriptionDetailURL,
				// 2026-04-29 milestone-billing plan §5/§6 — Activities tab on
				// Job detail uses the JobActivity routes/labels to render the
				// "+ Add Activity" CTA + per-row Edit CTA.
				JobActivityRoutes: jaRoutes,
				JobActivityLabels: jaLabels,
				// Search endpoints for job drawer client + location pickers.
				// Served by handlers registered below via handleFunc.
				ClientSearchURL:   jobRoutes.ClientSearchURL,
				LocationSearchURL: jobRoutes.LocationSearchURL,
			}
			if refChecker != nil {
				jobDeps.GetInUseIDs = refChecker.GetJobInUseIDs
			}
			if uc != nil {
				wireJobDeps(jobDeps, uc)
				wireJobDashboard(jobDeps, uc)
			}
			jobmod.NewModule(jobDeps).RegisterRoutes(ctx.Routes)
			// Register the client and location search endpoints for the job drawer.
			handleFunc(ctx.Routes, "GET", jobRoutes.ClientSearchURL, jobClientSearchFn)
			handleFunc(ctx.Routes, "GET", jobRoutes.LocationSearchURL, jobLocationSearchFn)
		}

		// --- Register JobTemplate module ---
		if cfg.wantJobTemplate() {
			jtDeps := &jobtemplatemod.ModuleDeps{
				Routes:           jtRoutes,
				Labels:           jtLabels,
				CommonLabels:     ctx.Common,
				TableLabels:      ctx.Table,
				UploadFile:       uploadFile,
				ListAttachments:  listAttachments,
				CreateAttachment: createAttachment,
				DeleteAttachment: deleteAttachment,
				NewID:            newAttachmentID,
				// Pass sibling routes so the Phases/Tasks tabs can render Add/Edit/Delete CTAs.
				PhaseRoutes: jtpRoutes,
				TaskRoutes:  jttRoutes,
			}
			if refChecker != nil {
				jtDeps.GetInUseIDs = refChecker.GetJobTemplateInUseIDs
			}
			if uc != nil {
				wireJobTemplateDeps(jtDeps, uc)
			}
			jobtemplatemod.NewModule(jtDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register JobTemplatePhase module (drawer-only) ---
		// Not in the sidebar — reached via JobTemplate detail Phases tab Add/Edit/Delete CTAs.
		if cfg.wantJobTemplatePhase() {
			jtpDeps := &jobtemplatePhasemod.ModuleDeps{
				Routes:       jtpRoutes,
				Labels:       jtpLabels,
				CommonLabels: ctx.Common,
			}
			if uc != nil {
				wireJobTemplatePhaseDeps(jtpDeps, uc)
			}
			jobtemplatePhasemod.NewModule(jtpDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register JobTemplateTask module (drawer-only) ---
		// Not in the sidebar — reached via JobTemplate detail Tasks tab Add/Edit/Delete CTAs.
		if cfg.wantJobTemplateTask() {
			jttDeps := &jobtemplateTaskmod.ModuleDeps{
				Routes:       jttRoutes,
				Labels:       jttLabels,
				CommonLabels: ctx.Common,
			}
			if uc != nil {
				wireJobTemplateTaskDeps(jttDeps, uc)
			}
			jobtemplateTaskmod.NewModule(jttDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register JobActivity module ---
		// Note: ReadActivityLabor / Material / Expense are not yet wired in espyna
		// use cases — these are nil until activity subtype use cases are added.
		if cfg.wantJobActivity() {
			jaDeps := &jobactivitymod.ModuleDeps{
				Routes:           jaRoutes,
				Labels:           jaLabels,
				CommonLabels:     ctx.Common,
				TableLabels:      ctx.Table,
				UploadFile:       uploadFile,
				ListAttachments:  listAttachments,
				CreateAttachment: createAttachment,
				DeleteAttachment: deleteAttachment,
				NewID:            newAttachmentID,
				// Pass activity charge routes so the charge tab can resolve edit URLs.
				ActivityLaborRoutes:    alRoutes,
				ActivityMaterialRoutes: amRoutes,
				ActivityExpenseRoutes:  aeRoutes,
			}
			if refChecker != nil {
				jaDeps.GetInUseIDs = refChecker.GetJobActivityInUseIDs
			}
			if uc != nil {
				wireJobActivityDeps(jaDeps, uc)
			}
			jobactivitymod.NewModule(jaDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register JobPhase module ---
		// JobPhase is the standalone phase entity module (Shape 3 — list + detail + CRUD).
		// The list page (/app/job-phases/list/{status}) is a power-user/debugging surface
		// with no sidebar entry. Phases are normally reached via Job detail Phases tab
		// deep links (/app/job-phase/{id}).
		if cfg.wantJobPhase() {
			jpDeps := &jobphasemod.ModuleDeps{
				Routes:       jpRoutes,
				Labels:       jpLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
			}
			jpDeps.AttachmentOps.UploadFile = uploadFile
			jpDeps.AttachmentOps.ListAttachments = listAttachments
			jpDeps.AttachmentOps.CreateAttachment = createAttachment
			jpDeps.AttachmentOps.DeleteAttachment = deleteAttachment
			jpDeps.AttachmentOps.NewAttachmentID = newAttachmentID
			if refChecker != nil {
				jpDeps.GetInUseIDs = refChecker.GetJobPhaseInUseIDs
			}
			if uc != nil {
				wireJobPhaseDeps(jpDeps, uc)
			}
			jobphasemod.NewModule(jpDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register JobTask module ---
		// JobTask is the standalone task entity module (Shape 3 — list + detail + CRUD).
		// The list page (/app/job-tasks/list/{status}) is a power-user/debugging surface
		// with no sidebar entry. Tasks are normally reached via JobPhase detail Tasks tab
		// deep links (/app/job-task/{id}).
		if cfg.wantJobTask() {
			jkDeps := &jobtaskmod.ModuleDeps{
				Routes:       jkRoutes,
				Labels:       jkLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
			}
			jkDeps.AttachmentOps.UploadFile = uploadFile
			jkDeps.AttachmentOps.ListAttachments = listAttachments
			jkDeps.AttachmentOps.CreateAttachment = createAttachment
			jkDeps.AttachmentOps.DeleteAttachment = deleteAttachment
			jkDeps.AttachmentOps.NewAttachmentID = newAttachmentID
			if refChecker != nil {
				jkDeps.GetInUseIDs = refChecker.GetJobTaskInUseIDs
			}
			if uc != nil {
				wireJobTaskDeps(jkDeps, uc)
			}
			// Staff search — reuse the same activityLaborStaffSearchFn (same data source).
			// Resource and template-task search endpoints fall back to flat filter mode
			// when the handler is nil (URLs cleared below).
			if activityLaborStaffSearchFn == nil {
				jkRoutes.StaffSearchURL = ""
			}
			jkRoutes.ResourceSearchURL = ""     // resource search not yet implemented
			jkRoutes.TemplateTaskSearchURL = "" // template-task search not yet implemented
			jobtaskmod.NewModule(jkDeps).RegisterRoutes(ctx.Routes)
			handleFunc(ctx.Routes, "GET", jkRoutes.StaffSearchURL, activityLaborStaffSearchFn)
		}

		// --- Register ActivityLabor module ---
		// ActivityLabor is the charge-detail sibling for ENTRY_TYPE_LABOR job activities.
		// Not in the sidebar — reached via JobActivity detail charge tab.
		// Use cases (Create/Read/Update/Delete/List) are stubbed with TODO comments until
		// ActivityLabor is added to espyna OperationUseCases (see wiring.go).
		if cfg.wantActivityLabor() {
			alDeps := &activitylabormod.ModuleDeps{
				Routes:       alRoutes,
				Labels:       alLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
				// TODO(P5 wave 3): wire from espyna OperationUseCases.ActivityLabor
				// via wireActivityLaborDeps() in wiring.go once the use case is added.
				// Until then all CRUD handlers return clear gap error messages.
			}
			if uc != nil {
				wireActivityLaborDeps(alDeps, uc)
			}
			activitylabormod.NewModule(alDeps).RegisterRoutes(ctx.Routes)
			// Register the staff search endpoint (nil-safe — skipped when handler is nil).
			handleFunc(ctx.Routes, "GET", alRoutes.StaffSearchURL, activityLaborStaffSearchFn)
		}

		// --- Register ActivityMaterial module ---
		// ActivityMaterial is the charge-detail sibling for ENTRY_TYPE_MATERIAL job activities.
		// Not in the sidebar — reached via JobActivity detail charge tab.
		if cfg.wantActivityMaterial() {
			amDeps := &activitymaterialmod.ModuleDeps{
				Routes:       amRoutes,
				Labels:       amLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
			}
			if uc != nil {
				wireActivityMaterialDeps(amDeps, uc)
			}
			activitymaterialmod.NewModule(amDeps).RegisterRoutes(ctx.Routes)
			// Register search endpoints (nil-safe — skipped when handlers are nil).
			handleFunc(ctx.Routes, "GET", amRoutes.ProductSearchURL, activityMaterialProductSearchFn)
			handleFunc(ctx.Routes, "GET", amRoutes.LocationSearchURL, activityMaterialLocationSearchFn)
		}

		// --- Register ActivityExpense module ---
		// ActivityExpense is the charge-detail sibling for ENTRY_TYPE_EXPENSE job activities.
		// Not in the sidebar — reached via JobActivity detail charge tab.
		if cfg.wantActivityExpense() {
			aeDeps := &activityexpensemod.ModuleDeps{
				Routes:       aeRoutes,
				Labels:       aeLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
			}
			if uc != nil {
				wireActivityExpenseDeps(aeDeps, uc)
			}
			activityexpensemod.NewModule(aeDeps).RegisterRoutes(ctx.Routes)
			// Register search endpoint (nil-safe — skipped when handler is nil).
			handleFunc(ctx.Routes, "GET", aeRoutes.ExpenseCategorySearchURL, activityExpenseExpenseCategorySearchFn)
		}

		// --- Register OutcomeCriteria module ---
		if cfg.wantOutcomeCriteria() {
			ocDeps := &outcomecriteriaMod.ModuleDeps{
				Routes:           ocRoutes,
				Labels:           ocLabels,
				CommonLabels:     ctx.Common,
				TableLabels:      ctx.Table,
				UploadFile:       uploadFile,
				ListAttachments:  listAttachments,
				CreateAttachment: createAttachment,
				DeleteAttachment: deleteAttachment,
				NewID:            newAttachmentID,
			}
			if uc != nil {
				wireOutcomeCriteriaDeps(ocDeps, uc)
			}
			outcomecriteriaMod.NewModule(ocDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register TaskOutcome module ---
		if cfg.wantTaskOutcome() {
			toDeps := &taskoutcomeMod.ModuleDeps{
				Routes:           toRoutes,
				Labels:           toLabels,
				CommonLabels:     ctx.Common,
				TableLabels:      ctx.Table,
				UploadFile:       uploadFile,
				ListAttachments:  listAttachments,
				CreateAttachment: createAttachment,
				DeleteAttachment: deleteAttachment,
				NewID:            newAttachmentID,
			}
			if uc != nil {
				wireTaskOutcomeDeps(toDeps, uc)
			}
			taskoutcomeMod.NewModule(toDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register OutcomeSummary module ---
		if cfg.wantOutcomeSummary() {
			osDeps := &outcomesummaryMod.ModuleDeps{
				Routes:       osRoutes,
				Labels:       osLabels,
				CommonLabels: ctx.Common,
			}
			if uc != nil {
				wireOutcomeSummaryDeps(osDeps, uc)
			}
			outcomesummaryMod.NewModule(osDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register Fulfillment module ---
		if cfg.wantFulfillment() {
			ffDeps := &fulfillmentmod.ModuleDeps{
				Routes:           ffRoutes,
				Labels:           ffLabels,
				CommonLabels:     ctx.Common,
				TableLabels:      ctx.Table,
				UploadFile:       uploadFile,
				ListAttachments:  listAttachments,
				CreateAttachment: createAttachment,
				DeleteAttachment: deleteAttachment,
				NewID:            newAttachmentID,
			}
			if uc != nil {
				wireFulfillmentDeps(ffDeps, uc)
				wireFulfillmentDashboard(ffDeps, uc)
			}
			fulfillmentmod.NewModule(ffDeps).RegisterRoutes(ctx.Routes)
		}

		log.Println("  ✓ Operations domain initialized (fayna)")
		return nil
	}
}

// defaultFulfillmentLabels returns FulfillmentLabels with sensible English defaults.
// Mirrors the service-admin composition helper so the block is self-contained.
func defaultFulfillmentLabels() fayna.FulfillmentLabels {
	return fayna.FulfillmentLabels{
		PageTitle: "Fulfillment",
		AppLabel:  "Fulfillment",
		Title:     "Fulfillments",
		Status: fayna.FulfillmentStatusLabels{
			Pending:            "Pending",
			Ready:              "Ready",
			InTransit:          "In Transit",
			Delivered:          "Delivered",
			PartiallyDelivered: "Partially Delivered",
			Failed:             "Failed",
			Cancelled:          "Cancelled",
		},
		Type: fayna.DeliveryModeLabels{
			Instant:      "Instant",
			Scheduled:    "Scheduled",
			Shipped:      "Shipped",
			Digital:      "Digital",
			Project:      "Project",
			Subscription: "Subscription",
		},
		Columns: fayna.FulfillmentColumnLabels{
			DeliveryMode: "Method",
			Status:       "Status",
			SupplierName: "Supplier",
			ScheduledAt:  "Scheduled",
			ItemCount:    "Items",
			Notes:        "Notes",
		},
		Tabs: fayna.FulfillmentTabLabels{
			Info:        "Information",
			Items:       "Items",
			History:     "History",
			Returns:     "Returns",
			Attachments: "Attachments",
		},
		Actions: fayna.FulfillmentActionLabels{
			MarkReady:      "Mark Ready",
			Dispatch:       "Dispatch",
			Deliver:        "Deliver",
			DeliverPartial: "Partial Delivery",
			MarkFailed:     "Mark Failed",
			Cancel:         "Cancel",
			Retry:          "Retry",
		},
		Buttons: fayna.FulfillmentButtonLabels{
			AddFulfillment: "Add Fulfillment",
			Edit:           "Edit",
			Delete:         "Delete",
			Transition:     "Update Status",
			Return:         "Create Return",
		},
		Empty: fayna.FulfillmentEmptyLabels{
			Title:   "No fulfillments found",
			Message: "No fulfillments to display.",
		},
		Errors: fayna.FulfillmentErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			LoadFailed:       "Failed to load fulfillment data",
			TransitionFailed: "Failed to update fulfillment status",
		},
	}
}
