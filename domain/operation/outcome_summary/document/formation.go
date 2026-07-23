package document

import (
	"context"
	"log"
	"sort"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
)

// The Formation page (report-card DOCX v2) mirrors the MMIS report card's
// "Student Formation" page: one table per non-academic outcome category
// (subject-deportment, homeroom-deportment), each strand shown with its frozen
// authoritative average. It is fully generic — a "formation group" is any
// job_category OTHER than the configured academic filter, and its table title is
// the category's own display NAME (data, never a code literal). The MMIS-specific
// static wording (the "STUDENT FORMATION" heading, the grade-descriptor legend,
// the certificate of transfer) lives ONLY in the .docx template artifact, exactly
// like the existing "MYP Report Card" cover text.
//
// Data faithfulness: education1's per-phase deportment task_outcome ("Conduct") is
// loader-default scaffold (dominated by 100/0), so the Formation renders the FROZEN
// job_outcome_summary.scaled_label — the 0-100 average that the MMIS deportment
// import proved 100% faithful — NOT the scaffold per-semester cells. See
// docs/plan/20260716-formation-page/plan.md §"one deliberate deviation".

// formationRow is one row of a formation-page category table: a subject strand (or
// the homeroom) and its frozen authoritative average
// (job_outcome_summary.scaled_label), pre-formatted as a display string.
type formationRow struct {
	Subject string
	Average string
}

// formationGroup is one category block on the Formation page — the analog of a
// MMIS "Subject Deportment" / "Homeroom Deportment" table. Title is the
// job_category display NAME (data); Rows are the enrolled strands in that category.
type formationGroup struct {
	Title string
	Rows  []formationRow
}

// catInfo is the display metadata for a job_category (name + sort order +
// code, the app-config match key for DocumentOptions.GroupCategoryFilter).
type catInfo struct {
	name  string
	order int32
	code  string
}

// collectFormationGroups builds the Formation-page category blocks from the
// student's NON-academic categorized jobs (the deportment complement of the
// academic subject set, partitioned by the caller in collectCard). It renders one
// table per job_category, titled by the category's display name, ordered by the
// category's sort_order (academic=10 < subject_deportment=20 < homeroom_deportment
// =30 on education1 → Subject before Homeroom, the MMIS order). Fully generic: no
// education vocabulary in the code, only the shared category axis.
//
// deportJobs are the caller's already-fetched jobs whose category is present and is
// NOT the academic category. When no category filter is configured (service-admin),
// the caller passes an empty slice, so the Formation page never fabricates a group
// from an unrelated category. Fail-soft: a missing closure or lookup error yields
// no groups, and the page renders its static sections only.
func collectFormationGroups(ctx context.Context, d *Deps, deportJobs []*jobpb.Job, historical bool) []formationGroup {
	if len(deportJobs) == 0 {
		return nil
	}

	cats := fetchCategories(ctx, d)

	jobIDs := make([]string, 0, len(deportJobs))
	jobTemplate := map[string]string{}
	jobCat := map[string]string{}
	templateIDs := []string{}
	tmplSeen := map[string]bool{}
	for _, j := range deportJobs {
		jid, tid := j.GetId(), j.GetJobTemplateId()
		if jid == "" || tid == "" {
			continue
		}
		jobIDs = append(jobIDs, jid)
		jobTemplate[jid] = tid
		jobCat[jid] = j.GetJobCategoryId()
		if !tmplSeen[tid] {
			tmplSeen[tid] = true
			templateIDs = append(templateIDs, tid)
		}
	}

	tmplNames := fetchTemplateNames(ctx, d, templateIDs, historical)
	avgByJob := fetchYearLabels(ctx, d, jobIDs)

	// Group rows by category id, suppressing non-enrolled strands (the untaken
	// parallel track — e.g. a Korean strand — rides in as a frozen average of "0";
	// reuse the ONE shared enrollment predicate: it keeps the row iff the frozen
	// average parses > 1, dropping the "0"/"1" transmute-of-zero floor).
	rowsByCat := map[string][]formationRow{}
	for _, jid := range jobIDs {
		avg := strings.TrimSpace(avgByJob[jid])
		// A deportment job always carries the Conduct scaffold, so HasMarks=true;
		// the frozen average is the enrollment signal.
		ev := outcome_summary.EnrollmentEvidence{HasMarks: true}
		if outcome_summary.IsNonEnrolledCell(ev, avg) {
			continue
		}
		cid := jobCat[jid]
		rowsByCat[cid] = append(rowsByCat[cid], formationRow{
			Subject: cleanSubject(colName(tmplNames, jobTemplate[jid])),
			Average: avg,
		})
	}
	if len(rowsByCat) == 0 {
		return nil
	}

	// Order categories by sort_order (then name) so the tables render in the MMIS
	// order regardless of map iteration.
	catIDs := make([]string, 0, len(rowsByCat))
	for cid := range rowsByCat {
		catIDs = append(catIDs, cid)
	}
	sort.SliceStable(catIDs, func(i, j int) bool {
		a, b := cats[catIDs[i]], cats[catIDs[j]]
		if a.order != b.order {
			return a.order < b.order
		}
		if a.name != b.name {
			return a.name < b.name
		}
		return catIDs[i] < catIDs[j]
	})

	groups := make([]formationGroup, 0, len(catIDs))
	for _, cid := range catIDs {
		rows := rowsByCat[cid]
		sort.SliceStable(rows, func(i, j int) bool {
			return strings.ToLower(rows[i].Subject) < strings.ToLower(rows[j].Subject)
		})
		title := strings.TrimSpace(cats[cid].name)
		if title == "" {
			// Defensive: a category with no display name still renders a titled
			// table rather than a raw id.
			title = "Formation"
		}
		groups = append(groups, formationGroup{Title: title, Rows: rows})
	}
	return groups
}

// fetchCategories loads all job_category display metadata (id → name + sort_order)
// via the per-request ListJobCategories closure. Nil-safe: a missing closure or a
// lookup error yields an empty map (formation titles fall back to a generic label).
func fetchCategories(ctx context.Context, d *Deps) map[string]catInfo {
	out := map[string]catInfo{}
	if d.ListJobCategories == nil {
		return out
	}
	resp, err := d.ListJobCategories(ctx, &jobcategorypb.ListJobCategoriesRequest{})
	if err != nil {
		log.Printf("report card doc: list job categories: %v", err)
		return out
	}
	for _, c := range resp.GetData() {
		if id := c.GetId(); id != "" {
			out[id] = catInfo{name: c.GetName(), order: c.GetSortOrder(), code: c.GetCode()}
		}
	}
	return out
}

// formationData flattens the assembled formation groups into the doctemplate data
// shape: a body-level {{#formation_groups}} loop, each item a category table with a
// nested {{#rows}} table-loop. Every key is emitted as a pre-formatted string so no
// raw {{..}} leaks; an empty group list renders no formation tables (the engine
// removes the loop block).
func formationData(groups []formationGroup) []any {
	out := make([]any, 0, len(groups))
	for _, g := range groups {
		rows := make([]any, 0, len(g.Rows))
		for _, r := range g.Rows {
			rows = append(rows, map[string]any{
				"row_subject": orBlank(r.Subject),
				"row_average": orDash(r.Average),
			})
		}
		out = append(out, map[string]any{
			"category_title": orBlank(g.Title),
			"rows":           rows,
		})
	}
	return out
}
