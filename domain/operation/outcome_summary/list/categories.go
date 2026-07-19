package list

// categories.go — the landing's dynamic per-job_category count columns (R9
// W-A2). Structural mirror of job/list/tabs.go (the "/classes" ?jc= tabstrip):
// the SAME single-statement tab-support UNION read supplies the ACTIVE
// category corpus (plan §3.0 — the ONE shared corpus, sort_order-ordered) and
// the ACTIVE template→category map the historical fallback buckets with.
//
// NAMING: every identifier here is generic. The category display strings
// ("Academic" / "Subject Deportment" / …) are the job_category rows' own name
// column (per-workspace DATA) — vertical vocabulary never enters a Go
// identifier, and the column HEADERS are data, not labels (lyngua.md).

import (
	"context"
	"log"
	"sort"
	"strings"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// loadLandingCategories issues the ONE tab-support read (R9 W-A2) and shapes it
// for the landing: the ACTIVE job_category rows sorted by sort_order (NULLS
// LAST → name ASC → id — the landing's own explicit sort contract, plan §3.3),
// the ACTIVE template→category map (the historical fallback's authoritative
// current-FK lookup, §3.0), and whether any ACTIVE template carries a NULL
// category (the Uncategorized bucket trigger).
//
// Config-gated: with Options.List.CategoryColumns unset (service-admin's
// zero-valued options) NO read is issued and the static column set renders
// byte-identical. Data-gated: a nil closure, a read error, or a denied kind
// (the espyna use case returns an empty slice per denied permission) all
// degrade to (nil, nil, false) → the SAME static columns, never an error table.
func loadLandingCategories(ctx context.Context, deps *ListViewDeps) (cats []*jobcategorypb.JobCategory, templateToCat map[string]string, hasNullTemplate bool) {
	if !deps.Options.List.CategoryColumns() || deps.ListJobListTabSupport == nil {
		return nil, nil, false
	}
	allCats, templates, err := deps.ListJobListTabSupport(ctx)
	if err != nil {
		log.Printf("report cards landing: list category tab support: %v", err)
		return nil, nil, false // data-gated degrade → static columns
	}
	for _, c := range allCats {
		if c.GetActive() {
			cats = append(cats, c)
		}
	}
	sortLandingCategories(cats)
	templateToCat = make(map[string]string, len(templates))
	for _, t := range templates {
		templateToCat[t.GetId()] = t.GetJobCategoryId()
		if t.GetJobCategoryId() == "" {
			hasNullTemplate = true
		}
	}
	return cats, templateToCat, hasNullTemplate
}

// sortLandingCategories orders the landing's category columns by
// job_category.sort_order ASC with NULLs LAST, then name ASC, then id. This is
// the category primitive's OWN sort contract (plan §3.3) — TabOptions only
// governs the price_schedule tabstrip and never this axis.
func sortLandingCategories(cats []*jobcategorypb.JobCategory) {
	sort.SliceStable(cats, func(i, j int) bool {
		a, b := cats[i], cats[j]
		ai, aok := categoryOrderOf(a)
		bi, bok := categoryOrderOf(b)
		if aok != bok {
			return aok // NULLS LAST
		}
		if aok && bok && ai != bi {
			return ai < bi
		}
		an := strings.ToLower(a.GetName())
		bn := strings.ToLower(b.GetName())
		if an == bn {
			return a.GetId() < b.GetId()
		}
		return an < bn
	})
}

// categoryOrderOf returns the category's sort_order and whether it is set
// (NULL = not set → sorts last).
func categoryOrderOf(c *jobcategorypb.JobCategory) (int32, bool) {
	if c != nil && c.SortOrder != nil {
		return c.GetSortOrder(), true
	}
	return 0, false
}

// categoryIDSet returns the corpus membership set for the sorted active
// category slice.
func categoryIDSet(cats []*jobcategorypb.JobCategory) map[string]bool {
	set := make(map[string]bool, len(cats))
	for _, c := range cats {
		set[c.GetId()] = true
	}
	return set
}

// uncategorizedCount folds every bucket OUTSIDE the active corpus into the
// single Uncategorized count for one section: the "" NULL bucket PLUS any
// count keyed by a non-corpus category id (a stale/inactive/foreign effective
// category — §3.0's corpus/authority mismatch). Folding keeps the pinned
// "never dropped" invariant without adding inactive columns or a dangling eye
// (the bucket cell renders eye-less).
func uncategorizedCount(byCat map[string]int, corpus map[string]bool) int {
	n := 0
	for cat, count := range byCat {
		if !corpus[cat] {
			n += count
		}
	}
	return n
}

// anyUncategorizedCount reports whether any section holds ≥1 subject outside
// the active corpus (the Uncategorized bucket-column trigger).
func anyUncategorizedCount(subjectsByCat map[string]map[string]int, corpus map[string]bool) bool {
	for _, byCat := range subjectsByCat {
		if uncategorizedCount(byCat, corpus) > 0 {
			return true
		}
	}
	return false
}

// cellAccessibleName substitutes both DATA nouns into the lyngua aria frame
// for a category cell's eye link. A frame missing either placeholder returns
// "" so BuildCompositeCell composes its default name (which always carries
// both) — the accessible name never loses a dimension to a bad translation.
func cellAccessibleName(frame, category, section string) string {
	if !strings.Contains(frame, "{category}") || !strings.Contains(frame, "{section}") {
		return ""
	}
	return strings.NewReplacer("{category}", category, "{section}", section).Replace(frame)
}

// landingColumnsFor returns the landing column set: the static 3-column set
// when the category corpus is empty (the degrade contract — byte-identical to
// landingColumns), else section + students + one column per ACTIVE category
// (headers = job_category.name DATA, sort_order-ordered) + the optional
// Uncategorized bucket column. Column keys embed the FULL category id
// (collision-proof, mirroring the composite cell's test-id contract).
func landingColumnsFor(l outcome_summary.Labels, cats []*jobcategorypb.JobCategory, showUncategorized bool) []types.TableColumn {
	if len(cats) == 0 {
		return landingColumns(l)
	}
	cols := []types.TableColumn{
		{Key: "section", Label: l.Landing.GroupColumn, MinWidth: "12.5rem"},
		{Key: "students", Label: l.Landing.MembersColumn, MinWidth: "6.25rem", Align: "right"},
	}
	for _, c := range cats {
		cols = append(cols, types.TableColumn{Key: "jc-" + c.GetId(), Label: c.GetName(), MinWidth: "6.25rem", Align: "center"})
	}
	if showUncategorized {
		cols = append(cols, types.TableColumn{Key: "jc-uncategorized", Label: l.Landing.UncategorizedColumn, MinWidth: "6.25rem", Align: "center"})
	}
	return cols
}
