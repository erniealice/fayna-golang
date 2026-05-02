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
	"fmt"
	"log"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	fayna "github.com/erniealice/fayna-golang"
	fulfillmentmod "github.com/erniealice/fayna-golang/views/fulfillment"
	jobmod "github.com/erniealice/fayna-golang/views/job"
	jobactivitymod "github.com/erniealice/fayna-golang/views/job_activity"
	jobtemplatemod "github.com/erniealice/fayna-golang/views/job_template"
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
	enableAll       bool
	job             bool
	jobTemplate     bool
	jobActivity     bool
	outcomeCriteria bool
	taskOutcome     bool
	outcomeSummary  bool
	fulfillment     bool
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

// WithOutcomeCriteria registers the OutcomeCriteria module (list, detail, CRUD).
func WithOutcomeCriteria() BlockOption { return func(c *blockConfig) { c.outcomeCriteria = true } }

// WithTaskOutcome registers the TaskOutcome module (list, detail, CRUD).
func WithTaskOutcome() BlockOption { return func(c *blockConfig) { c.taskOutcome = true } }

// WithOutcomeSummary registers the OutcomeSummary module (job + phase summaries).
func WithOutcomeSummary() BlockOption { return func(c *blockConfig) { c.outcomeSummary = true } }

// WithFulfillment registers the Fulfillment module (list, detail, CRUD, status transitions).
func WithFulfillment() BlockOption { return func(c *blockConfig) { c.fulfillment = true } }

// WithSubscriptionDetailURL supplies the centymo subscription-detail path
// template (e.g. "/app/subscriptions/detail/{id}") so the Job detail page
// can render a "Spawned from Subscription" breadcrumb when
// Job.origin_type = SUBSCRIPTION. Optional — when unset the breadcrumb is
// hidden.
// 2026-04-29 auto-spawn-jobs-from-subscription Phase D.
func WithSubscriptionDetailURL(url string) BlockOption {
	return func(c *blockConfig) { c.subscriptionDetailURL = url }
}

func (c *blockConfig) wantJob() bool             { return c.enableAll || c.job }
func (c *blockConfig) wantJobTemplate() bool     { return c.enableAll || c.jobTemplate }
func (c *blockConfig) wantJobActivity() bool     { return c.enableAll || c.jobActivity }
func (c *blockConfig) wantOutcomeCriteria() bool { return c.enableAll || c.outcomeCriteria }
func (c *blockConfig) wantTaskOutcome() bool     { return c.enableAll || c.taskOutcome }
func (c *blockConfig) wantOutcomeSummary() bool  { return c.enableAll || c.outcomeSummary }
func (c *blockConfig) wantFulfillment() bool     { return c.enableAll || c.fulfillment }

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

		ocRoutes := fayna.DefaultOutcomeCriteriaRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "outcome_criteria", &ocRoutes)

		toRoutes := fayna.DefaultTaskOutcomeRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "task_outcome", &toRoutes)

		osRoutes := fayna.DefaultOutcomeSummaryRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "outcome_summary", &osRoutes)

		ffRoutes := fayna.DefaultFulfillmentRoutes()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "fulfillment", &ffRoutes)

		// --- Load labels (defaults + optional lyngua overrides) ---
		jobLabels := fayna.DefaultJobLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job.json", "job", &jobLabels)

		jtLabels := fayna.DefaultJobTemplateLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_template.json", "job_template", &jtLabels)

		jaLabels := fayna.DefaultJobActivityLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "job_activity.json", "job_activity", &jaLabels)

		ocLabels := fayna.DefaultOutcomeCriteriaLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "outcome_criteria.json", "outcome_criteria", &ocLabels)

		toLabels := fayna.DefaultTaskOutcomeLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "task_outcome.json", "task_outcome", &toLabels)

		osLabels := fayna.DefaultOutcomeSummaryLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "outcome_summary.json", "outcome_summary", &osLabels)

		ffLabels := defaultFulfillmentLabels()
		_ = translations.LoadPathIfExists("en", ctx.BusinessType, "fulfillment.json", "fulfillment", &ffLabels)

		// --- Reflect into use cases aggregate (espyna-free wiring) ---
		uc := assertUseCases(ctx.UseCases)

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
			}
			if uc != nil {
				wireJobDeps(jobDeps, uc)
				wireJobDashboard(jobDeps, uc)
			}
			jobmod.NewModule(jobDeps).RegisterRoutes(ctx.Routes)
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
			}
			if uc != nil {
				wireJobTemplateDeps(jtDeps, uc)
			}
			jobtemplatemod.NewModule(jtDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register JobActivity module ---
		// Note: ReadActivityLabor / Material / Expense are not yet wired in espyna
		// use cases — these are nil until activity subtype use cases are added.
		if cfg.wantJobActivity() {
			jaDeps := &jobactivitymod.ModuleDeps{
				Routes:       jaRoutes,
				Labels:       jaLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
			}
			if uc != nil {
				wireJobActivityDeps(jaDeps, uc)
			}
			jobactivitymod.NewModule(jaDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register OutcomeCriteria module ---
		if cfg.wantOutcomeCriteria() {
			ocDeps := &outcomecriteriaMod.ModuleDeps{
				Routes:       ocRoutes,
				Labels:       ocLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
			}
			if uc != nil {
				wireOutcomeCriteriaDeps(ocDeps, uc)
			}
			outcomecriteriaMod.NewModule(ocDeps).RegisterRoutes(ctx.Routes)
		}

		// --- Register TaskOutcome module ---
		if cfg.wantTaskOutcome() {
			toDeps := &taskoutcomeMod.ModuleDeps{
				Routes:       toRoutes,
				Labels:       toLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
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
				Routes:       ffRoutes,
				Labels:       ffLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
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
			Status:            "Status",
			SupplierName:      "Supplier",
			ScheduledAt:       "Scheduled",
			ItemCount:         "Items",
			Notes:             "Notes",
		},
		Tabs: fayna.FulfillmentTabLabels{
			Info:    "Information",
			Items:   "Items",
			History: "History",
			Returns: "Returns",
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
