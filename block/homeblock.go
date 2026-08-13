package block

// homeblock.go — the SELF-CONTAINED persona-aware home dashboard block
// (docs/plan/20260801-persona-home-dashboard §4.2/§4.4; Q1–Q7 locked
// 2026-08-01). Deliberately NOT part of engineblock.go / usecases.go /
// wiring.go (coordination-hazard files, plan §9): this file owns its whole
// chain — option grammar entry, use-case adapter closure, unit Mount, and
// assembly — and touches zero shared block files.
//
// Inert-zero contract (Q6-B): with no declared variants the block registers
// NOTHING — no routes, no templates, no route-map keys — so a consumer that
// does not configure it (service-admin) is provably byte-identical.

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/erniealice/espyna-golang/consumer"
	consumerapp "github.com/erniealice/espyna-golang/consumer/app"
	compose "github.com/erniealice/espyna-golang/consumer/compose"
	"github.com/erniealice/espyna-golang/ports"

	home "github.com/erniealice/fayna-golang/domain/home"
	homedashboard "github.com/erniealice/fayna-golang/domain/home/dashboard"

	ocpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/outcome_completion"
)

// HomeOption configures the home block from the consuming app.
type HomeOption func(*homeConfig)

type homeConfig struct {
	options              home.Options
	resolvePrincipalKind func(ctx context.Context) int32
	resolveWorkspaceName func(ctx context.Context) string
	resolveOverviewURL   func(ctx context.Context) string
}

// WithHomeOptions sets the deployment-declared home-surface config (persona
// slot → widget roster, tab axis, quick links). Zero value = inert block.
func WithHomeOptions(o home.Options) HomeOption {
	return func(c *homeConfig) { c.options = o }
}

// WithHomePrincipalKindResolver injects the composition-owned persona
// closure (the getDashboardData precedent): the app wraps identity.Require
// so fayna never imports espyna/shared/identity. Unresolved sessions return
// 0 → home.Options.DefaultSlot → the fail-closed lowest-capability variant.
func WithHomePrincipalKindResolver(fn func(ctx context.Context) int32) HomeOption {
	return func(c *homeConfig) { c.resolvePrincipalKind = fn }
}

// WithHomeWorkspaceNameResolver injects the workspace display-name resolver
// (replacing the retired hardcoded demo workspace name). Nil-safe: absent →
// the workspace line is omitted.
func WithHomeWorkspaceNameResolver(fn func(ctx context.Context) string) HomeOption {
	return func(c *homeConfig) { c.resolveWorkspaceName = fn }
}

// WithHomeOverviewURLResolver injects the WORKSPACE-QUALIFIED overview URL the
// bare /home route redirects to (20260809 D2/M5). The app builds it with its
// workspace-slug helper so fayna never imports espyna's slug package. Nil-safe:
// absent ⇒ the redirect falls back to the bare Routes.OverviewURL.
func WithHomeOverviewURLResolver(fn func(ctx context.Context) string) HomeOption {
	return func(c *homeConfig) { c.resolveOverviewURL = fn }
}

// HomeBlock returns the consumerapp.AppOption that mounts the persona-aware
// home dashboard. It must run AFTER the domain engine blocks whose route
// keys the quick links reference and BEFORE the app-owned shell block (which
// reads the fully-merged compose result and skips its app-local home module
// when this block is configured).
func HomeBlock(opts ...HomeOption) consumerapp.AppOption {
	return func(ctx *consumerapp.AppContext) error {
		cfg := homeConfig{}
		for _, o := range opts {
			o(&cfg)
		}
		// Inert-zero contract: no declared variants → register nothing.
		if !cfg.options.Enabled() {
			log.Printf("  compose: fayna home block — no variants declared, inert (zero-value contract)")
			return nil
		}

		uc, err := consumerapp.RequireUseCases(ctx, "faynaHomeBlock")
		if err != nil {
			return err
		}

		// The ONE aggregate read (Q5 lock). Identity + row scope resolve
		// from ctx inside the use case/adapter (Q-EIB-BRIDGE): the request
		// carries no principal/workspace fields from here. Nil-safe: an
		// unwired dashboard aggregate leaves the closure nil and the data
		// widgets render their designed empty states.
		//
		// Error classification (Q4/T-9, skeptic F2): the use case types its
		// fail-closed refusals as ports.AuthorizationError; those are wrapped
		// with homedashboard.ErrSummaryDenied so the view renders the
		// designed DENIED cards + AUTHZ_RBAC_DENY log line, while data/
		// infrastructure errors pass through and render the labelled
		// UNAVAILABLE state — a deny never masquerades as "nothing assigned".
		var getSummary func(reqCtx context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error)
		if uc.Service != nil && uc.Service.Dashboard != nil &&
			uc.Service.Dashboard.OutcomeCompletion != nil &&
			uc.Service.Dashboard.OutcomeCompletion.GetOutcomeCompletionSummary != nil {
			exec := uc.Service.Dashboard.OutcomeCompletion.GetOutcomeCompletionSummary.Execute
			getSummary = func(reqCtx context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error) {
				resp, err := exec(reqCtx, &ocpb.GetOutcomeCompletionSummaryRequest{})
				if err != nil {
					var authzErr *ports.AuthorizationError
					if errors.As(err, &authzErr) {
						return nil, fmt.Errorf("%w: %v", homedashboard.ErrSummaryDenied, err)
					}
					return nil, err
				}
				return resp, nil
			}
		}

		// View-level deny log (the T-9 observability half): mirrors the RBAC
		// authorizer's AUTHZ_RBAC_DENY line so log-grep assertions catch
		// view-level widget denies too. Identity fields come from session
		// context accessors — never the wire.
		logDeny := func(reqCtx context.Context, code, widget string) {
			log.Printf("AUTHZ_RBAC_DENY | mode=VIEW | user=%s | workspace=%s | code=%s | widget=%s",
				consumer.GetUserIDFromContext(reqCtx),
				consumer.GetWorkspaceIDFromContext(reqCtx),
				code, widget)
		}

		u := home.Describe()
		u.Mount = func(mc *compose.MountContext) error {
			r := u.Routes.(*home.Routes)
			homedashboard.NewModule(&homedashboard.Deps{
				Routes:               *r,
				Options:              cfg.options,
				CommonLabels:         mc.Common,
				GetCompletionSummary: getSummary,
				ResolvePrincipalKind: cfg.resolvePrincipalKind,
				ResolveWorkspaceName: cfg.resolveWorkspaceName,
				ResolveOverviewURL:   cfg.resolveOverviewURL,
				LogDeny:              logDeny,
			}).RegisterRoutes(mc.Routes)
			return nil
		}

		return consumerapp.AssembleEngineBlock("fayna-home", []compose.Unit{u}, ctx)
	}
}
