package document

// ratings.go — the block-layout item-rating enrichment: the rotation-pair
// merge (G1) and the per-item / group rating (deportment) tables.
//
// Rotation pairs: a canonical item (e.g. a school subject "Arts") is graded on
// ONE job whose phase 1 marks come from one strand variant ("Arts: Visual
// Arts") and phase 2 from the other ("Arts: Music"). The strand identities
// survive only on the rating-category jobs, whose names keep the "Prefix:
// Variant" form. The merged display title lists the phase-1 strand first —
// exactly the operator's printed card ("Arts: Visual Arts / Arts: Music").
//
// Which strand is phase 1 is decided data-first:
//  1. the strand rating job with a phase-1 phase summary is the phase-1
//     strand (the rating semester import writes a summary only for the
//     strand's active half);
//  2. else the strand whose name matches an INACTIVE academic job is phase 2
//     (the canonicalization deactivated the merged-in phase-2 strand);
//  3. else alphabetical (deterministic fallback).

import (
	"context"
	"log"
	"sort"
	"strings"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// ratingContext is the one-shot fetch bundle for the rating-category jobs:
// strand names, per-phase phase summaries, frozen year averages, and the
// group-category job's per-phase summaries.
type ratingContext struct {
	strandJobs []*jobpb.Job
	nameOf     map[string]string           // strand jobID → cleaned display name
	pos        map[string]map[int32]string // strand jobID → phase_order → scaled label
	avg        map[string]string           // strand jobID → frozen year average label
	groupPos   map[int32]string            // group job phase_order → scaled label
}

// fetchItemRatings loads the item-rating enrichment sources for the card's
// non-academic jobs. Fully nil-safe: missing closures leave the affected maps
// empty and every rating field renders blank.
func fetchItemRatings(ctx context.Context, d *Deps, deportJobs []*jobpb.Job, groupJob *jobpb.Job, historical bool) *ratingContext {
	c := &ratingContext{
		nameOf:   map[string]string{},
		pos:      map[string]map[int32]string{},
		avg:      map[string]string{},
		groupPos: map[int32]string{},
	}
	groupID := ""
	if groupJob != nil {
		groupID = groupJob.GetId()
	}

	ids := []string{}
	templateIDs := []string{}
	tmplSeen := map[string]bool{}
	jobTemplate := map[string]string{}
	for _, j := range deportJobs {
		jid, tid := j.GetId(), j.GetJobTemplateId()
		if jid == "" || tid == "" {
			continue
		}
		if jid != groupID {
			c.strandJobs = append(c.strandJobs, j)
		}
		ids = append(ids, jid)
		jobTemplate[jid] = tid
		if !tmplSeen[tid] {
			tmplSeen[tid] = true
			templateIDs = append(templateIDs, tid)
		}
	}
	if len(ids) == 0 {
		return c
	}

	tmplNames := fetchTemplateNames(ctx, d, templateIDs, historical)
	for _, j := range c.strandJobs {
		jid := j.GetId()
		c.nameOf[jid] = cleanSubject(colName(tmplNames, jobTemplate[jid]))
	}

	order, _ := fetchPhaseOrders(ctx, d, ids)
	pos := fetchSemesterLabels(ctx, d, ids, order)
	for jid, byOrder := range pos {
		if jid == groupID {
			c.groupPos = byOrder
			continue
		}
		c.pos[jid] = byOrder
	}

	strandIDs := make([]string, 0, len(c.strandJobs))
	for _, j := range c.strandJobs {
		strandIDs = append(strandIDs, j.GetId())
	}
	c.avg = fetchYearLabels(ctx, d, strandIDs)
	return c
}

// rotationPair is one merged strand pair, period-1 strand first.
type rotationPair struct {
	sem1Name, sem2Name string
	sem1Job, sem2Job   string
}

// mergedPairs indexes the rotation pairs by their canonical academic subject
// name (lower-cased).
type mergedPairs struct {
	byCanonical map[string]rotationPair
}

// titleFor returns the merged pair title for a canonical subject display name,
// or "" when the subject has no rotation pair.
func (m mergedPairs) titleFor(name string) string {
	p, ok := m.byCanonical[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return ""
	}
	return p.sem1Name + " / " + p.sem2Name
}

// mergeRotationPairs pairs the strand conduct jobs ("Prefix: Variant") under
// their canonical academic subject ("Prefix") and decides the period order.
// Only exact two-strand pairs whose prefix IS an academic subject merge;
// anything else renders unmerged (fail-soft).
func mergeRotationPairs(c *ratingContext, academicNames map[string]bool, inactiveNames map[string]bool) mergedPairs {
	m := mergedPairs{byCanonical: map[string]rotationPair{}}
	if c == nil || len(c.strandJobs) == 0 {
		return m
	}
	byPrefix := map[string][]string{}
	for _, j := range c.strandJobs {
		jid := j.GetId()
		name := c.nameOf[jid]
		idx := strings.Index(name, ":")
		if idx <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(name[:idx]))
		if key == "" || !academicNames[key] {
			continue
		}
		byPrefix[key] = append(byPrefix[key], jid)
	}
	for key, ids := range byPrefix {
		if len(ids) != 2 {
			continue
		}
		a, b := ids[0], ids[1]
		if c.nameOf[a] > c.nameOf[b] {
			a, b = b, a
		}
		sem1, sem2 := a, b
		aPos, bPos := c.pos[a], c.pos[b]
		_, a1 := aPos[1]
		_, a2 := aPos[2]
		_, b1 := bPos[1]
		_, b2 := bPos[2]
		aInactive := inactiveNames[strings.ToLower(c.nameOf[a])]
		bInactive := inactiveNames[strings.ToLower(c.nameOf[b])]
		switch {
		case a1 && !b1:
			sem1, sem2 = a, b
		case b1 && !a1:
			sem1, sem2 = b, a
		case b2 && !a2:
			sem1, sem2 = a, b
		case a2 && !b2:
			sem1, sem2 = b, a
		case aInactive && !bInactive:
			sem1, sem2 = b, a
		case bInactive && !aInactive:
			sem1, sem2 = a, b
		}
		m.byCanonical[key] = rotationPair{
			sem1Name: c.nameOf[sem1], sem2Name: c.nameOf[sem2],
			sem1Job: sem1, sem2Job: sem2,
		}
	}
	return m
}

// buildItemRatings assembles the per-item rating table rows (rotation pairs
// merged into one row, phase-1 strand's value in the phase-1 column) plus the
// group rating per-phase values. Non-enrolled strands (frozen average at the
// transmute-of-zero floor) are suppressed exactly like the academic transcript.
func buildItemRatings(c *ratingContext, merged mergedPairs) (rows []ratingRow, groupPhase1, groupPhase2 string) {
	if c == nil {
		return nil, "", ""
	}
	enrolled := func(jid string) bool {
		ev := outcome_summary.EnrollmentEvidence{HasMarks: true}
		return !outcome_summary.IsNonEnrolledCell(ev, strings.TrimSpace(c.avg[jid]))
	}
	phaseValue := func(jid string, phase int32) string {
		if v, ok := c.pos[jid][phase]; ok {
			return strings.TrimSpace(v)
		}
		return ""
	}

	inPair := map[string]bool{}
	for _, p := range merged.byCanonical {
		inPair[p.sem1Job] = true
		inPair[p.sem2Job] = true
		e1, e2 := enrolled(p.sem1Job), enrolled(p.sem2Job)
		if !e1 && !e2 {
			continue
		}
		row := ratingRow{Title: p.sem1Name + " / " + p.sem2Name}
		if e1 {
			row.Phase1 = firstNonEmpty(phaseValue(p.sem1Job, 1), strings.TrimSpace(c.avg[p.sem1Job]))
		}
		if e2 {
			row.Phase2 = firstNonEmpty(phaseValue(p.sem2Job, 2), strings.TrimSpace(c.avg[p.sem2Job]))
		}
		rows = append(rows, row)
	}
	for _, j := range c.strandJobs {
		jid := j.GetId()
		if inPair[jid] || !enrolled(jid) {
			continue
		}
		rows = append(rows, ratingRow{
			Title:  c.nameOf[jid],
			Phase1: phaseValue(jid, 1),
			Phase2: phaseValue(jid, 2),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].Title) < strings.ToLower(rows[j].Title)
	})
	return rows, strings.TrimSpace(c.groupPos[1]), strings.TrimSpace(c.groupPos[2])
}

// fetchInactiveSubjectNames lists the INACTIVE jobs on the enrollment
// subscription and returns their cleaned template names (lower-cased set) —
// the rotation period-2 strand fallback signal (the canonicalization
// deactivated the merged-in strand job). Nil-safe.
func fetchInactiveSubjectNames(ctx context.Context, d *Deps, subID string) map[string]bool {
	out := map[string]bool{}
	if d.ListJobs == nil || subID == "" {
		return out
	}
	templateIDs := []string{}
	tmplSeen := map[string]bool{}
	jobTemplates := []string{}
	for page := int32(1); page <= maxPages; page++ {
		resp, err := d.ListJobs(ctx, &jobpb.ListJobsRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
				stringEq("origin_id", subID), boolEq("active", false),
			}},
			Pagination: &commonpb.PaginationRequest{
				Limit:  int32(pageLimit),
				Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
			},
		})
		if err != nil {
			log.Printf("report card doc: list inactive jobs: %v", err)
			return out
		}
		data := resp.GetData()
		for _, j := range data {
			if tid := j.GetJobTemplateId(); tid != "" {
				jobTemplates = append(jobTemplates, tid)
				if !tmplSeen[tid] {
					tmplSeen[tid] = true
					templateIDs = append(templateIDs, tid)
				}
			}
		}
		if len(data) < pageLimit {
			break
		}
	}
	if len(templateIDs) == 0 {
		return out
	}
	names := fetchTemplateNames(ctx, d, templateIDs, true)
	for _, tid := range jobTemplates {
		if n := cleanSubject(strings.TrimSpace(names[tid])); n != "" {
			out[strings.ToLower(n)] = true
		}
	}
	return out
}
