package action

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func testLabels() fayna.FulfillmentLabels {
	return fayna.FulfillmentLabels{
		Errors: fayna.FulfillmentErrorLabels{
			PermissionDenied: "permission denied",
			LoadFailed:       "load failed",
			TransitionFailed: "transition failed",
		},
	}
}

func testRoutes() fayna.FulfillmentRoutes {
	return fayna.DefaultFulfillmentRoutes()
}

func ctxWithPerms(perms ...string) context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(perms))
}

func ctxNoPerm() context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(nil))
}

// ---------------------------------------------------------------------------
// NewAddAction tests
// ---------------------------------------------------------------------------

func TestNewAddAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewAddAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/add", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "permission denied" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "permission denied")
	}
}

func TestNewAddAction_GET_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewAddAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/fulfillments/add", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewAddAction_POST_Success(t *testing.T) {
	t.Parallel()

	createCalled := false
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillment: func(_ context.Context, req *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error) {
			createCalled = true
			return &fulfillmentpb.CreateFulfillmentResponse{
				Data: &fulfillmentpb.Fulfillment{Id: "ff-123"},
			}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("revenue_id", "rev-1")
	form.Set("supplier_id", "sup-1")
	form.Set("fulfillment_method", "Physical")
	form.Set("notes", "Test note")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:create"), viewCtx)

	if !createCalled {
		t.Fatal("CreateFulfillment was not called")
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if redirect := result.Headers["HX-Redirect"]; redirect == "" {
		t.Fatal("expected HX-Redirect for successful create with returned ID")
	}
}

func TestNewAddAction_POST_CreateError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillment: func(_ context.Context, _ *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error) {
			return nil, errors.New("db connection error")
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("revenue_id", "rev-1")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:create"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "db connection error" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "db connection error")
	}
}

func TestNewAddAction_POST_EmptyFields(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.CreateFulfillmentRequest
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillment: func(_ context.Context, req *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error) {
			captured = req
			return &fulfillmentpb.CreateFulfillmentResponse{
				Data: &fulfillmentpb.Fulfillment{Id: "ff-empty"},
			}, nil
		},
	}

	v := NewAddAction(deps)

	// POST with empty form body — all fields empty.
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:create"), viewCtx)

	if captured == nil {
		t.Fatal("CreateFulfillment was not called")
	}
	if captured.Data.RevenueId != "" {
		t.Fatalf("RevenueId = %q, want empty", captured.Data.RevenueId)
	}
	// Handler does not validate required fields — passes through.
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestNewAddAction_POST_NoResponseData(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillment: func(_ context.Context, _ *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error) {
			return &fulfillmentpb.CreateFulfillmentResponse{Data: nil}, nil
		},
	}

	v := NewAddAction(deps)

	form := url.Values{}
	form.Set("revenue_id", "rev-1")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:create"), viewCtx)

	// When no data is returned, falls through to HTMXSuccess.
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if redirect := result.Headers["HX-Redirect"]; redirect != "" {
		t.Fatalf("expected no HX-Redirect when response has nil data, got %q", redirect)
	}
}

func TestNewAddAction_POST_SpecialCharsInFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		notes string
	}{
		{"html_script", `<script>alert('xss')</script>`},
		{"sql_injection", `'; DROP TABLE fulfillments; --`},
		{"null_byte", "note\x00with\x00nulls"},
		{"unicode", "\u202Egnirts\u200B"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var capturedNotes string
			deps := &Deps{
				Routes: testRoutes(),
				Labels: testLabels(),
				CreateFulfillment: func(_ context.Context, req *fulfillmentpb.CreateFulfillmentRequest) (*fulfillmentpb.CreateFulfillmentResponse, error) {
					capturedNotes = req.Data.Notes
					return &fulfillmentpb.CreateFulfillmentResponse{
						Data: &fulfillmentpb.Fulfillment{Id: "ff-special"},
					}, nil
				},
			}

			v := NewAddAction(deps)

			form := url.Values{}
			form.Set("notes", tt.notes)

			req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/add", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			viewCtx := &view.ViewContext{Request: req}

			result := v.Handle(ctxWithPerms("fulfillment:create"), viewCtx)

			if result.StatusCode != http.StatusOK {
				t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
			}
			if capturedNotes != tt.notes {
				t.Fatalf("Notes = %q, want %q", capturedNotes, tt.notes)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewEditAction tests
// ---------------------------------------------------------------------------

func TestNewEditAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewEditAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/fulfillments/edit/ff-1", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewEditAction_GET_LoadError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		GetFulfillmentItemPageData: func(_ context.Context, _ *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error) {
			return nil, errors.New("not found")
		},
	}

	v := NewEditAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/fulfillments/edit/ff-1", nil)
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "load failed" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "load failed")
	}
}

func TestNewEditAction_GET_NilFulfillmentInResponse(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		GetFulfillmentItemPageData: func(_ context.Context, _ *fulfillmentpb.GetFulfillmentItemPageDataRequest) (*fulfillmentpb.GetFulfillmentItemPageDataResponse, error) {
			return &fulfillmentpb.GetFulfillmentItemPageDataResponse{
				Fulfillment: nil, // response with nil fulfillment
			}, nil
		},
	}

	v := NewEditAction(deps)

	req := httptest.NewRequest(http.MethodGet, "/action/fulfillments/edit/ff-1", nil)
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "load failed" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "load failed")
	}
}

func TestNewEditAction_POST_UpdateError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateFulfillment: func(_ context.Context, _ *fulfillmentpb.UpdateFulfillmentRequest) (*fulfillmentpb.UpdateFulfillmentResponse, error) {
			return nil, errors.New("conflict: version mismatch")
		},
	}

	v := NewEditAction(deps)

	form := url.Values{}
	form.Set("fulfillment_method", "Digital")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/edit/ff-1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "conflict: version mismatch" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "conflict: version mismatch")
	}
}

func TestNewEditAction_POST_EmptyID(t *testing.T) {
	t.Parallel()

	var capturedID string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		UpdateFulfillment: func(_ context.Context, req *fulfillmentpb.UpdateFulfillmentRequest) (*fulfillmentpb.UpdateFulfillmentResponse, error) {
			capturedID = req.Data.Id
			return &fulfillmentpb.UpdateFulfillmentResponse{}, nil
		},
	}

	v := NewEditAction(deps)

	form := url.Values{}
	form.Set("fulfillment_method", "Service")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/edit/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// No SetPathValue — id is empty.
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	// Documents passthrough behavior: empty ID is sent to UpdateFulfillment.
	if capturedID != "" {
		t.Fatalf("expected empty ID passthrough, got %q", capturedID)
	}
}

// ---------------------------------------------------------------------------
// NewDeleteAction tests
// ---------------------------------------------------------------------------

func TestNewDeleteAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/delete?id=ff-1", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewDeleteAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/delete", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "ID is required")
	}
}

func TestNewDeleteAction_EmptyIDInQueryAndForm(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewDeleteAction(deps)

	form := url.Values{}
	form.Set("id", "")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/delete?id=", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewDeleteAction_Success(t *testing.T) {
	t.Parallel()

	var deletedID string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteFulfillment: func(_ context.Context, req *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error) {
			deletedID = req.Id
			return &fulfillmentpb.DeleteFulfillmentResponse{}, nil
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/delete?id=ff-42", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "ff-42" {
		t.Fatalf("DeleteFulfillment called with ID %q, want %q", deletedID, "ff-42")
	}
}

func TestNewDeleteAction_DeleteError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteFulfillment: func(_ context.Context, _ *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error) {
			return nil, errors.New("has active shipments")
		},
	}

	v := NewDeleteAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/delete?id=ff-1", nil)
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:delete"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "has active shipments" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "has active shipments")
	}
}

func TestNewDeleteAction_IDFromFormBody(t *testing.T) {
	t.Parallel()

	var deletedID string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		DeleteFulfillment: func(_ context.Context, req *fulfillmentpb.DeleteFulfillmentRequest) (*fulfillmentpb.DeleteFulfillmentResponse, error) {
			deletedID = req.Id
			return &fulfillmentpb.DeleteFulfillmentResponse{}, nil
		},
	}

	v := NewDeleteAction(deps)

	form := url.Values{}
	form.Set("id", "ff-form-id")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:delete"), viewCtx)

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if deletedID != "ff-form-id" {
		t.Fatalf("DeleteFulfillment called with ID %q, want %q", deletedID, "ff-form-id")
	}
}

// ---------------------------------------------------------------------------
// NewTransitionAction tests
// ---------------------------------------------------------------------------

func TestNewTransitionAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewTransitionAction(deps)

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/transition", nil)
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewTransitionAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewTransitionAction(deps)

	form := url.Values{}
	form.Set("event", "mark_ready")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/transition", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// No SetPathValue — empty ID.
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Fulfillment ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Fulfillment ID is required")
	}
}

func TestNewTransitionAction_MissingEvent(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewTransitionAction(deps)

	form := url.Values{}
	// event omitted

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/transition", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Event is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Event is required")
	}
}

func TestNewTransitionAction_InvalidEvent(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		TransitionStatus: func(_ context.Context, req *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error) {
			return nil, errors.New("invalid event: BOGUS_EVENT")
		},
	}

	v := NewTransitionAction(deps)

	form := url.Values{}
	form.Set("event", "BOGUS_EVENT")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/transition", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	// Handler maps TransitionStatus errors to the generic TransitionFailed label.
	if msg := result.Headers["HX-Error-Message"]; msg != "transition failed" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "transition failed")
	}
}

func TestNewTransitionAction_Success(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.TransitionStatusRequest
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		TransitionStatus: func(_ context.Context, req *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error) {
			captured = req
			return &fulfillmentpb.TransitionStatusResponse{}, nil
		},
	}

	v := NewTransitionAction(deps)

	form := url.Values{}
	form.Set("event", "mark_ready")
	form.Set("reason", "All items packed")
	form.Set("provider_status", "ready")
	form.Set("provider_reference", "REF-123")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/transition", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if captured == nil {
		t.Fatal("TransitionStatus was not called")
	}
	if captured.FulfillmentId != "ff-1" {
		t.Fatalf("FulfillmentId = %q, want %q", captured.FulfillmentId, "ff-1")
	}
	if captured.Event != "mark_ready" {
		t.Fatalf("Event = %q, want %q", captured.Event, "mark_ready")
	}
	if captured.Reason != "All items packed" {
		t.Fatalf("Reason = %q, want %q", captured.Reason, "All items packed")
	}
	if redirect := result.Headers["HX-Redirect"]; redirect == "" {
		t.Fatal("expected HX-Redirect for successful transition")
	}
}

func TestNewTransitionAction_SpecialCharsInEvent(t *testing.T) {
	t.Parallel()

	var capturedEvent string
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		TransitionStatus: func(_ context.Context, req *fulfillmentpb.TransitionStatusRequest) (*fulfillmentpb.TransitionStatusResponse, error) {
			capturedEvent = req.Event
			return nil, errors.New("invalid event")
		},
	}

	v := NewTransitionAction(deps)

	form := url.Values{}
	form.Set("event", `<script>alert('xss')</script>`)

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/transition", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	// Handler passes through the event string without sanitization.
	if capturedEvent != `<script>alert('xss')</script>` {
		t.Fatalf("Event = %q, want the script tag string", capturedEvent)
	}
}

// ---------------------------------------------------------------------------
// NewReturnAction tests
// ---------------------------------------------------------------------------

func TestNewReturnAction_PermissionDenied(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Defective")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxNoPerm(), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestNewReturnAction_MissingID(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Defective")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// No SetPathValue — empty ID.
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Fulfillment ID is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Fulfillment ID is required")
	}
}

func TestNewReturnAction_MissingReason(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	// reason omitted

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "Reason is required" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "Reason is required")
	}
}

func TestNewReturnAction_NegativeRefundAmount(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.FulfillmentReturn
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillmentReturn: func(_ context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error) {
			captured = req
			return req, nil
		},
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Defective item")
	form.Set("refund_amount", "-100.50")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	// Handler does not validate negative refund amounts — passes through.
	if captured == nil {
		t.Fatal("CreateFulfillmentReturn was not called")
	}
	if captured.RefundAmount == nil {
		t.Fatal("RefundAmount should be set")
	}
	if *captured.RefundAmount != -10050 {
		t.Fatalf("RefundAmount = %d, want -10050", *captured.RefundAmount)
	}
}

func TestNewReturnAction_InvalidRefundAmount(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.FulfillmentReturn
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillmentReturn: func(_ context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error) {
			captured = req
			return req, nil
		},
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Defective item")
	form.Set("refund_amount", "not-a-number")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	// Invalid float string is silently ignored — RefundAmount stays nil.
	if captured == nil {
		t.Fatal("CreateFulfillmentReturn was not called")
	}
	if captured.RefundAmount != nil {
		t.Fatalf("RefundAmount = %d, want nil (invalid float should be skipped)", *captured.RefundAmount)
	}
}

func TestNewReturnAction_ZeroRefundAmount(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.FulfillmentReturn
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillmentReturn: func(_ context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error) {
			captured = req
			return req, nil
		},
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Wrong item")
	form.Set("refund_amount", "0")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if captured == nil {
		t.Fatal("CreateFulfillmentReturn was not called")
	}
	if captured.RefundAmount == nil {
		t.Fatal("RefundAmount should be set for 0")
	}
	if *captured.RefundAmount != 0 {
		t.Fatalf("RefundAmount = %d, want 0", *captured.RefundAmount)
	}
}

func TestNewReturnAction_VeryLargeRefundAmount(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.FulfillmentReturn
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillmentReturn: func(_ context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error) {
			captured = req
			return req, nil
		},
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Full refund")
	form.Set("refund_amount", "999999999999.99")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if captured == nil {
		t.Fatal("CreateFulfillmentReturn was not called")
	}
	if captured.RefundAmount == nil {
		t.Fatal("RefundAmount should be set")
	}
	if *captured.RefundAmount != 99999999999999 {
		t.Fatalf("RefundAmount = %d, want 99999999999999", *captured.RefundAmount)
	}
}

func TestNewReturnAction_CreateReturnError(t *testing.T) {
	t.Parallel()

	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillmentReturn: func(_ context.Context, _ *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error) {
			return nil, errors.New("return policy violation")
		},
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Defective")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}
	if msg := result.Headers["HX-Error-Message"]; msg != "return policy violation" {
		t.Fatalf("HX-Error-Message = %q, want %q", msg, "return policy violation")
	}
}

func TestNewReturnAction_Success(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.FulfillmentReturn
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillmentReturn: func(_ context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error) {
			captured = req
			return req, nil
		},
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "Damaged in transit")
	form.Set("notes", "Box was crushed")
	form.Set("refund_amount", "50.00")
	form.Set("currency", "USD")

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	result := v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if captured == nil {
		t.Fatal("CreateFulfillmentReturn was not called")
	}
	if captured.FulfillmentId != "ff-1" {
		t.Fatalf("FulfillmentId = %q, want %q", captured.FulfillmentId, "ff-1")
	}
	if captured.Reason != "Damaged in transit" {
		t.Fatalf("Reason = %q, want %q", captured.Reason, "Damaged in transit")
	}
	if captured.Notes != "Box was crushed" {
		t.Fatalf("Notes = %q, want %q", captured.Notes, "Box was crushed")
	}
	if captured.Currency != "USD" {
		t.Fatalf("Currency = %q, want %q", captured.Currency, "USD")
	}
	if captured.Status != "PENDING" {
		t.Fatalf("Status = %q, want %q", captured.Status, "PENDING")
	}
	if redirect := result.Headers["HX-Redirect"]; redirect == "" {
		t.Fatal("expected HX-Redirect for successful return")
	}
	if redirect := result.Headers["HX-Redirect"]; !strings.Contains(redirect, "tab=returns") {
		t.Fatalf("HX-Redirect = %q, expected to contain tab=returns", redirect)
	}
}

func TestNewReturnAction_EmptyRefundAmountString(t *testing.T) {
	t.Parallel()

	var captured *fulfillmentpb.FulfillmentReturn
	deps := &Deps{
		Routes: testRoutes(),
		Labels: testLabels(),
		CreateFulfillmentReturn: func(_ context.Context, req *fulfillmentpb.FulfillmentReturn) (*fulfillmentpb.FulfillmentReturn, error) {
			captured = req
			return req, nil
		},
	}

	v := NewReturnAction(deps)

	form := url.Values{}
	form.Set("reason", "No longer needed")
	form.Set("refund_amount", "") // empty string — should not set RefundAmount

	req := httptest.NewRequest(http.MethodPost, "/action/fulfillments/ff-1/return", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "ff-1")
	viewCtx := &view.ViewContext{Request: req}

	_ = v.Handle(ctxWithPerms("fulfillment:update"), viewCtx)

	if captured == nil {
		t.Fatal("CreateFulfillmentReturn was not called")
	}
	if captured.RefundAmount != nil {
		t.Fatalf("RefundAmount = %d, want nil (empty string should be skipped)", *captured.RefundAmount)
	}
}
