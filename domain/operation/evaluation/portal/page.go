package portal

import (
	"context"
	"fmt"
	"log"

	evaluation "github.com/erniealice/fayna-golang/domain/operation/evaluation"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
)

// portal/page.go — Surface 5: client-portal "Rate My Team" (§H).
//
// Q-EVAL-IDOR-1 (the SHIP-GATE invariant): the view supplies NO client_id and
// CANNOT re-introduce one — the GetListPageData request type carries no
// client_id field. ALL three IDOR gates are enforced INSIDE the injected
// closure (use-case/adapter QUERY PREDICATE), which the espyna backend already
// implements:
//   1. row-scope: WHERE client_id = session.acting_as_client_id (session-derived)
//   2. fail-closed: nil/empty acting_as_client_id → empty result, deny-before-SQL
//   3. visibility: visibility_type != INTERNAL_ONLY in the predicate
// The empty-state (client-eval-list-empty) asserts the IDOR boundary (TEST-1).

// PortalViewDeps holds the client-portal dependencies. GetListPageData is the
// CLIENT-scoped projection (the closure carries the IDOR predicate server-side).
type PortalViewDeps struct {
	Routes          evaluation.Routes
	Labels          evaluation.Labels
	CommonLabels    pyeza.CommonLabels
	TableLabels     types.TableLabels
	GetListPageData func(ctx context.Context, req *evalpb.GetEvaluationListPageDataRequest) (*evalpb.GetEvaluationListPageDataResponse, error)
}

// PageData holds the data for the "Rate My Team" portal page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Cards           []Card
	Labels          evaluation.Labels
	StartReviewURL  string
	IsEmpty         bool
}

// Card is one associate card (subject_staff_id-scoped roster row).
type Card struct {
	EvaluationID  string
	StaffID       string
	Period        string
	StatusLabel   string
	StatusVariant string
	Rating        string
	RatingTestID  string // associate-rating-{staff_id}
	CardTestID    string // seat-card / associate card
	StartTestID   string // start-review-{...}
	StartURL      string
}

// NewView creates the client-portal "Rate My Team" view (Surface 5).
func NewView(deps *PortalViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "read") {
			return view.Forbidden("evaluation:read")
		}

		cards, res, ok := loadCards(ctx, deps)
		if !ok {
			return res
		}

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Portal.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				HeaderTitle:    l.Portal.Heading,
				HeaderSubtitle: l.Portal.Caption,
				HeaderIcon:     "icon-users",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "evaluation-portal-content",
			Cards:           cards,
			Labels:          l,
			StartReviewURL:  deps.Routes.AddURL,
			IsEmpty:         len(cards) == 0,
		}
		return view.OK("evaluation-portal", pageData)
	})
}

// NewTableView creates the HTMX roster-refresh partial (scoped, §H.1).
func NewTableView(deps *PortalViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation", "read") {
			return view.Forbidden("evaluation:read")
		}
		cards, res, ok := loadCards(ctx, deps)
		if !ok {
			return res
		}
		l := deps.Labels
		return view.OK("evaluation-portal-roster", &PageData{
			Cards:          cards,
			Labels:         l,
			StartReviewURL: deps.Routes.AddURL,
			IsEmpty:        len(cards) == 0,
		})
	})
}

// loadCards runs the CLIENT-scoped page-data closure (IDOR predicate is inside
// the closure — the view passes no client_id) and maps the result to cards.
func loadCards(ctx context.Context, deps *PortalViewDeps) ([]Card, view.ViewResult, bool) {
	resp, err := deps.GetListPageData(ctx, &evalpb.GetEvaluationListPageDataRequest{})
	if err != nil {
		log.Printf("Failed to load portal page data: %v", err)
		return nil, view.Error(fmt.Errorf("failed to load reviews: %w", err)), false
	}
	cards := buildCards(resp.GetEvaluationList(), deps)
	return cards, view.ViewResult{}, true
}

func buildCards(items []*evalpb.Evaluation, deps *PortalViewDeps) []Card {
	cards := make([]Card, 0, len(items))
	for _, e := range items {
		staffID := e.GetSubjectStaffId()
		cards = append(cards, Card{
			EvaluationID:  e.GetId(),
			StaffID:       staffID,
			Period:        formatPeriod(e.GetPeriodStart(), e.GetPeriodEnd()),
			StatusLabel:   statusLabel(e.GetStatus(), deps.Labels),
			StatusVariant: statusVariant(e.GetStatus()),
			Rating:        rating(e),
			RatingTestID:  "associate-rating-" + staffID,
			CardTestID:    "associate-card-" + staffID,
			StartTestID:   "start-review-" + staffID,
			StartURL:      route.ResolveURL(deps.Routes.AddURL),
		})
	}
	return cards
}

func rating(e *evalpb.Evaluation) string {
	if e.GetStatus() == evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT || e.GetOverallScore() == 0 {
		return "—"
	}
	return fmt.Sprintf("%.2f", e.GetOverallScore())
}

func statusLabel(s evalpb.EvaluationStatus, l evaluation.Labels) string {
	switch s {
	case evalpb.EvaluationStatus_EVALUATION_STATUS_DRAFT:
		return l.Status.Draft
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED:
		return l.Status.Submitted
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF:
		return l.Status.SignedOff
	case evalpb.EvaluationStatus_EVALUATION_STATUS_ARCHIVED:
		return l.Status.Archived
	default:
		return l.Status.Draft
	}
}

func statusVariant(s evalpb.EvaluationStatus) string {
	switch s {
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SIGNED_OFF:
		return "success"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_SUBMITTED:
		return "info"
	case evalpb.EvaluationStatus_EVALUATION_STATUS_ARCHIVED:
		return "secondary"
	default:
		return "default"
	}
}

func formatPeriod(start, end string) string {
	if start == "" && end == "" {
		return "—"
	}
	if end == "" {
		return start
	}
	return start + " – " + end
}
