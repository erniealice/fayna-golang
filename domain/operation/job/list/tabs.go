package list

// tabs.go — the job_category tabstrip for the "/classes" (jobs) list. Mirrors
// the report-cards price_schedule tabstrip (outcome_summary/list/page.go) 1:1:
// one pyeza.TabItem per job_category row (active + inactive), ordered by
// job_category.sort_order (NULLS LAST → name ASC), with ?jc=<id> selecting the
// active tab. The tabstrip mounts ONLY when the app configures
// Options.Tab.GroupByField = "job_category" (school-admin); with zero-valued
// options (service-admin) the list renders its flat table unchanged.
//
// NAMING: every identifier here is generic. The category display strings
// ("Academic" / "Subject Deportment" / "Homeroom Deportment") are the
// job_category rows' own name column (per-workspace data) — vertical vocabulary
// never enters a Go identifier.

import (
	"context"
	"sort"
	"strings"

	job "github.com/erniealice/fayna-golang/domain/operation/job"
	pyeza "github.com/erniealice/pyeza-golang"

	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// loadJobListTabSupport issues the ONE tab-support read per page load (20260718
// courses-list-perf Rank-1) and splits it into the tabstrip's two shapes: every
// job_category (active + inactive — the tab rows) and every ACTIVE job_template
// stub (id/name/job_category_id). This replaces the former listAllJobCategories
// two-call concat + the paged activeTemplatesByCategory loop (12 statements → 1).
//
// Per-kind authorization (job_category:list, job_template:list) is enforced in the
// espyna use case: a denied kind returns an empty slice, so the tabstrip degrades
// exactly as before (no tabs / no fallback for the denied kind). Unlike the old
// helpers, a DB/query error is NOT swallowed — it propagates so renderTabbed can
// render a real error state instead of silently-partial tabs.
//
// Nil-safe: no closure wired → (nil, nil, nil) → the list degrades to no tabs.
func loadJobListTabSupport(ctx context.Context, deps *ListViewDeps) ([]*jobcategorypb.JobCategory, []*jobtemplatepb.JobTemplate, error) {
	if deps.ListJobListTabSupport == nil {
		return nil, nil, nil
	}
	return deps.ListJobListTabSupport(ctx)
}

// sortJobCategories orders categories per the Tab options: when SortByOrder is
// set, sort_order ASC with NULLs LAST then name ASC; otherwise name ASC. A
// "desc" direction reverses the primary key. Stable. (outcome_summary.sortSchedules.)
func sortJobCategories(cats []*jobcategorypb.JobCategory, opts job.TabOptions) {
	desc := opts.Direction() == "desc"
	byOrder := opts.SortByOrder()
	sort.SliceStable(cats, func(i, j int) bool {
		a, b := cats[i], cats[j]
		if byOrder {
			ai, aok := categoryOrder(a)
			bi, bok := categoryOrder(b)
			if aok != bok {
				return aok // NULLS LAST regardless of direction.
			}
			if aok && bok && ai != bi {
				if desc {
					return ai > bi
				}
				return ai < bi
			}
		}
		an := strings.ToLower(a.GetName())
		bn := strings.ToLower(b.GetName())
		if an == bn {
			return a.GetId() < b.GetId()
		}
		if desc && byOrder {
			return an < bn // name stays the ascending tiebreak under desc sort_order.
		}
		if desc {
			return an > bn
		}
		return an < bn
	})
}

// categoryOrder returns the category's sort_order and whether it is set (NULL =
// not set → sorts last).
func categoryOrder(c *jobcategorypb.JobCategory) (int32, bool) {
	if c != nil && c.SortOrder != nil {
		return c.GetSortOrder(), true
	}
	return 0, false
}

// categoryExists reports whether id matches a fetched job_category — the guard
// that keeps a stale/foreign/tampered ?jc= from selecting a non-existent tab.
func categoryExists(cats []*jobcategorypb.JobCategory, id string) bool {
	for _, c := range cats {
		if c.GetId() == id {
			return true
		}
	}
	return false
}

// defaultCategory returns the id of the first ACTIVE category in the (already
// sorted) slice, falling back to the first category, or "" when empty.
func defaultCategory(cats []*jobcategorypb.JobCategory) string {
	for _, c := range cats {
		if c.GetActive() {
			return c.GetId()
		}
	}
	if len(cats) > 0 {
		return cats[0].GetId()
	}
	return ""
}

// buildJobCategoryTabs builds one TabItem per category; Count is the number of
// list rows in that category (subjects with jobs on the education path, jobs on
// the flat path); the tab whose id == selected is marked active. Href is
// listURL + "?jc=<id>", where listURL is the CURRENT status-resolved list URL
// (e.g. "/courses/list/active") so a tab click preserves the {status} segment.
func buildJobCategoryTabs(
	cats []*jobcategorypb.JobCategory,
	counts map[string]int,
	selected string,
	listURL string,
) []pyeza.TabItem {
	tabs := make([]pyeza.TabItem, 0, len(cats))
	for _, c := range cats {
		tabs = append(tabs, pyeza.TabItem{
			Key:   tabKey(c.GetId()),
			Label: c.GetName(),
			Href:  listURL + "?jc=" + c.GetId(),
			Count: counts[c.GetId()],
		})
	}
	return tabs
}

// activeTemplatesByCategory groups an already-fetched set of ACTIVE job_templates
// (from loadJobListTabSupport) by job_category_id. It returns two views of the
// same set:
//   - catToTemplates: category id → its active templates, used to render (and
//     count) a job_category the delivery-summary aggregate never reaches —
//     deportment conduct records are JOB_STATUS_COMPLETED jobs with no active
//     subscription_seat / product_plan match, so they fall out of the aggregate's
//     inner joins and their tab would otherwise be empty.
//   - templateToCat: template id → category id, used to map each aggregate
//     summary row back to its category. The tab-support read returns only ACTIVE
//     templates (the aggregate joins `jt.active`), so an active-only map is
//     complete for that purpose.
//
// Pure (no ctx/DB): the single tab-support statement already fetched the rows, so
// this is an in-memory regroup. Nil/empty input → empty maps (no tabs claim any
// category, so the list degrades to the unfiltered aggregate).
func activeTemplatesByCategory(templates []*jobtemplatepb.JobTemplate) (catToTemplates map[string][]*jobtemplatepb.JobTemplate, templateToCat map[string]string) {
	catToTemplates = map[string][]*jobtemplatepb.JobTemplate{}
	templateToCat = map[string]string{}
	for _, t := range templates {
		cat := t.GetJobCategoryId()
		catToTemplates[cat] = append(catToTemplates[cat], t)
		templateToCat[t.GetId()] = cat
	}
	return catToTemplates, templateToCat
}

// templateGrainRows renders job_templates directly at template grain, for a
// job_category the delivery-summary aggregate doesn't cover (deportment conduct
// records — see activeTemplatesByCategory). Only the template name and id are
// known at this grain, so the delivery columns (group / deliverer / schedule)
// and the item count render blank rather than a misleading zero; the row still
// links to the same outcome_matrix grid as an aggregate row. Sorted by template
// name for a stable order (the aggregate rows sort by group-then-name; these
// have no group).
func templateGrainRows(tmpls []*jobtemplatepb.JobTemplate) []templateSummaryRow {
	rows := make([]templateSummaryRow, 0, len(tmpls))
	for _, t := range tmpls {
		rows = append(rows, templateSummaryRow{
			TemplateID:    t.GetId(),
			TemplateName:  t.GetName(),
			hideItemCount: true,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].TemplateName < rows[j].TemplateName })
	return rows
}

// tabKey builds a stable, greppable tab key from a job_category id.
func tabKey(id string) string {
	if id == "" {
		return ""
	}
	return "jc-tab-" + short(id)
}

// short returns a stable, collision-resistant slug for keys/testids. It takes
// the LAST 8 chars (the uuidv7 random tail), NOT the first 8 — education1
// entities share the ~13-char timestamp PREFIX, so a first-8 slug collides
// across rows and silently marks every tab active (the report-cards lesson).
func short(id string) string {
	if len(id) > 8 {
		return id[len(id)-8:]
	}
	return id
}
