package outcome_summary

import (
	"context"
	"errors"
	"testing"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// listFn builds a ListJobCategories closure returning the given categories.
func listFn(cats ...*jobcategorypb.JobCategory) func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
	return func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
		return &jobcategorypb.ListJobCategoriesResponse{Data: cats, Success: true}, nil
	}
}

func errListFn() func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
	return func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
		return nil, errors.New("boom")
	}
}

// H2 fail-open hardening (B4 codex finding #2): a configured filter that cannot
// be resolved MUST fail closed (ok=false ⇒ caller drops every job), never
// regress to unfiltered data.
func strptr(s string) *string { return &s }

func TestResolveCategoryID_FailClosedSemantics(t *testing.T) {
	academic := &jobcategorypb.JobCategory{Id: "cat-academic", Code: strptr("academic")}

	t.Run("no code configured — no filter, keep all", func(t *testing.T) {
		id, ok := ResolveCategoryID(context.Background(), listFn(academic), "")
		if id != "" || !ok {
			t.Fatalf("blank code: want (\"\", true), got (%q, %v)", id, ok)
		}
	})

	t.Run("nil closure — no filter, keep all", func(t *testing.T) {
		id, ok := ResolveCategoryID(context.Background(), nil, "academic")
		if id != "" || !ok {
			t.Fatalf("nil list: want (\"\", true), got (%q, %v)", id, ok)
		}
	})

	t.Run("code resolves — filter active", func(t *testing.T) {
		id, ok := ResolveCategoryID(context.Background(), listFn(academic), "academic")
		if id != "cat-academic" || !ok {
			t.Fatalf("resolved: want (\"cat-academic\", true), got (%q, %v)", id, ok)
		}
	})

	t.Run("list read errors — FAIL CLOSED", func(t *testing.T) {
		id, ok := ResolveCategoryID(context.Background(), errListFn(), "academic")
		if id != "" || ok {
			t.Fatalf("lookup error must fail closed: want (\"\", false), got (%q, %v)", id, ok)
		}
	})

	t.Run("code does not resolve — FAIL CLOSED", func(t *testing.T) {
		id, ok := ResolveCategoryID(context.Background(), listFn(academic), "nonexistent")
		if id != "" || ok {
			t.Fatalf("unresolved code must fail closed: want (\"\", false), got (%q, %v)", id, ok)
		}
	})
}

func TestKeepJobInCategory(t *testing.T) {
	// No filter (catID "") keeps everything.
	if !KeepJobInCategory("", "anything") {
		t.Error("no filter must keep every job")
	}
	// Active filter keeps matching category + the NULL legacy-academic allowance.
	if !KeepJobInCategory("cat-academic", "cat-academic") {
		t.Error("matching category must be kept")
	}
	if !KeepJobInCategory("cat-academic", "") {
		t.Error("NULL job_category_id must be kept (legacy-academic allowance)")
	}
	// Active filter drops out-of-category (e.g. deportment) jobs.
	if KeepJobInCategory("cat-academic", "cat-deportment") {
		t.Error("out-of-category job must be dropped")
	}
}
