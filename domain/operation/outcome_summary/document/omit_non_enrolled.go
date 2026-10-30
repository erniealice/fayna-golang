package document

import (
	"context"
	"log"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/erniealice/espyna-golang/consumer"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// omit_non_enrolled.go — the "classmates graded" omission rule for explicit
// client documents. Every client of a group can carry a job for each parallel
// offering of the plan (only some of which the client actually takes), and no
// per-client choice is recorded. A job is therefore left off the document only
// when BOTH hold for the selected period:
//
//  1. the client's own job carries no grade evidence (the shared
//     outcome_summary.IsNonEnrolledCell predicate over the UNREDACTED card), and
//  2. at least one other member of the same subscription group is graded for
//     the same job template in that period (one group-scoped export read per
//     category, cached briefly so a batch download reads it once).
//
// A job nobody in the group is graded for yet stays on the document, so an
// ungraded sheet remains visible instead of silently disappearing. Every
// failure (no export port, a read error, an unknown category) omits nothing.
// Per-item jobs of other categories whose name pairs with an omitted job (the
// same trailing-qualifier pairing the period summaries already display) are
// omitted too, but only when they carry no evidence of their own.

// OmittedItemsHeader names the response header listing the omitted job names
// (URL-encoded, "; "-joined) so a batch download can record them per client.
const OmittedItemsHeader = "X-Omitted-Items"

// classmateEvidenceTTL bounds how long one group read is reused. A batch of a
// few dozen documents finishes well inside it; a teacher's later save is seen
// on the next document after expiry.
const classmateEvidenceTTL = 30 * time.Second

type omittedJob struct {
	jobID string
	name  string
}

// classmateGraded is template id → set of client ids graded in the period.
type classmateGraded map[string]map[string]bool

func (g classmateGraded) othersGraded(templateID, clientID string) bool {
	for id := range g[templateID] {
		if id != clientID {
			return true
		}
	}
	return false
}

type classmateEntry struct {
	done   chan struct{}
	graded classmateGraded
	err    error
	at     time.Time
}

var classmateCache = struct {
	sync.Mutex
	entries map[string]*classmateEntry
}{entries: map[string]*classmateEntry{}}

// fetchClassmateGraded reads one group/category/period matrix, sharing an
// in-flight or recent read for the same caller (the cache key includes the
// user id, so one principal's scope never answers for another's).
func fetchClassmateGraded(ctx context.Context, d *Deps, groupID, categoryID, period string) (classmateGraded, error) {
	key := strings.Join([]string{consumer.GetUserIDFromContext(ctx), groupID, categoryID, period}, "\x00")
	now := time.Now()
	classmateCache.Lock()
	for k, e := range classmateCache.entries {
		select {
		case <-e.done:
			if now.Sub(e.at) > classmateEvidenceTTL {
				delete(classmateCache.entries, k)
			}
		default:
		}
	}
	if e := classmateCache.entries[key]; e != nil {
		classmateCache.Unlock()
		<-e.done
		return e.graded, e.err
	}
	e := &classmateEntry{done: make(chan struct{})}
	classmateCache.entries[key] = e
	classmateCache.Unlock()

	e.graded, e.err = readClassmateGraded(ctx, d, groupID, categoryID, period)
	e.at = time.Now()
	if e.err != nil {
		classmateCache.Lock()
		delete(classmateCache.entries, key)
		classmateCache.Unlock()
	}
	close(e.done)
	return e.graded, e.err
}

func readClassmateGraded(ctx context.Context, d *Deps, groupID, categoryID, period string) (classmateGraded, error) {
	req := &exportpb.GetSubscriptionGroupOutcomeExportRequest{SubscriptionGroupId: groupID}
	req.JobCategoryId = &categoryID
	if period == clientDocumentPeriodYearFinal {
		req.OutcomeSelector = &exportpb.GetSubscriptionGroupOutcomeExportRequest_FinalOutcome{FinalOutcome: true}
	} else {
		req.OutcomeSelector = &exportpb.GetSubscriptionGroupOutcomeExportRequest_JobTemplatePhaseCode{JobTemplatePhaseCode: period}
	}
	resp, err := d.GetSubscriptionGroupOutcomeExport(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil || !resp.GetSuccess() {
		return nil, errIncompleteClassmateEvidence
	}
	graded := classmateGraded{}
	for _, row := range resp.GetClientRows() {
		clientID := strings.TrimSpace(row.GetClientId())
		if clientID == "" {
			continue
		}
		for _, cell := range row.GetCells() {
			if !cellGraded(cell) {
				continue
			}
			templateID := strings.TrimSpace(cell.GetJobTemplateId())
			if graded[templateID] == nil {
				graded[templateID] = map[string]bool{}
			}
			graded[templateID][clientID] = true
		}
	}
	return graded, nil
}

type classmateEvidenceError string

func (e classmateEvidenceError) Error() string { return string(e) }

const errIncompleteClassmateEvidence = classmateEvidenceError("classmate evidence: incomplete export response")

// cellGraded is the positive side of the shared placeholder rule: a present job
// with a positive mark, or a displayable (non-placeholder) stored value.
func cellGraded(cell *exportpb.SubscriptionGroupOutcomeCell) bool {
	if cell == nil || !cell.GetJobPresent() || strings.TrimSpace(cell.GetJobTemplateId()) == "" {
		return false
	}
	if cell.GetEnrollmentEvidence().GetHasPositiveMark() {
		return true
	}
	return outcome_summary.ExportCellValue(cell, "") != ""
}

// omitNonEnrolledJobs returns the jobs to leave off this client's document for
// the period, in document order. The card must be the UNREDACTED projection.
func omitNonEnrolledJobs(ctx context.Context, d *Deps, groupID string, card *exportpb.ClientReportCardProjection, period string) []omittedJob {
	if d == nil || d.GetSubscriptionGroupOutcomeExport == nil || card == nil {
		return nil
	}
	historical := card.GetContext().GetHistorical()
	clientID := strings.TrimSpace(card.GetClient().GetClientId())
	if clientID == "" {
		return nil
	}
	var codes []string
	if period == clientDocumentPeriodYearFinal {
		codes = []string{d.CategoryFilter}
	} else {
		codes = d.DocOptions.PhaseJobCategoryCodes
	}
	categoryByCode := map[string]string{}
	for _, category := range card.GetJobCategories() {
		if category != nil {
			categoryByCode[strings.ToLower(strings.TrimSpace(category.GetCode()))] = strings.TrimSpace(category.GetId())
		}
	}
	targets := map[string]bool{}
	for _, code := range codes {
		if id := categoryByCode[strings.ToLower(strings.TrimSpace(code))]; id != "" {
			targets[id] = true
		}
	}
	if len(targets) == 0 {
		return nil
	}

	ev := newCardEvidence(card, period, historical)
	jobs := append(card.GetJobs()[:0:0], card.GetJobs()...)
	sort.SliceStable(jobs, func(i, j int) bool { return jobs[i].GetId() < jobs[j].GetId() })

	var omitted []omittedJob
	omittedNames := map[string]bool{}
	graded := map[string]classmateGraded{}
	ownEvidence := false
	for _, job := range jobs {
		if job == nil || (!historical && !job.GetActive()) {
			continue
		}
		categoryID := ev.categoryOf(job)
		if !targets[categoryID] {
			continue
		}
		if !ev.placeholder(job.GetId()) {
			ownEvidence = true
			continue
		}
		g, ok := graded[categoryID]
		if !ok {
			var err error
			g, err = fetchClassmateGraded(ctx, d, groupID, categoryID, period)
			if err != nil {
				log.Printf("client document: classmate evidence unavailable, omitting nothing: %v", err)
				return nil
			}
			graded[categoryID] = g
		}
		if !g.othersGraded(strings.TrimSpace(job.GetJobTemplateId()), clientID) {
			continue
		}
		name := ev.nameOf(job)
		omitted = append(omitted, omittedJob{jobID: job.GetId(), name: name})
		omittedNames[pairKey(trimTrailingQualifier(name))] = true
	}
	// A client without evidence in any job is ungraded (not yet assessed or
	// newly placed), not choosing among offerings: keep the whole document.
	if len(omitted) == 0 || !ownEvidence {
		return nil
	}
	for _, job := range jobs {
		if job == nil || (!historical && !job.GetActive()) || targets[ev.categoryOf(job)] {
			continue
		}
		name := ev.nameOf(job)
		if omittedNames[pairKey(trimTrailingQualifier(name))] && ev.placeholder(job.GetId()) {
			omitted = append(omitted, omittedJob{jobID: job.GetId(), name: name})
		}
	}
	sort.SliceStable(omitted, func(i, j int) bool { return strings.ToLower(omitted[i].name) < strings.ToLower(omitted[j].name) })
	return omitted
}

func pairKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// withoutJobs returns a copy of the card without the omitted jobs; every
// builder walks card.Jobs, so their phases, tasks and summaries fall away.
func withoutJobs(card *exportpb.ClientReportCardProjection, omitted []omittedJob) *exportpb.ClientReportCardProjection {
	if card == nil || len(omitted) == 0 {
		return card
	}
	drop := make(map[string]bool, len(omitted))
	for _, o := range omitted {
		drop[o.jobID] = true
	}
	out := proto.Clone(card).(*exportpb.ClientReportCardProjection)
	kept := out.Jobs[:0]
	for _, job := range out.Jobs {
		if job != nil && drop[job.GetId()] {
			continue
		}
		kept = append(kept, job)
	}
	out.Jobs = kept
	return out
}

// omittedItemsHeaderValue encodes the omitted names for OmittedItemsHeader.
func omittedItemsHeaderValue(omitted []omittedJob) string {
	parts := make([]string, 0, len(omitted))
	for _, o := range omitted {
		parts = append(parts, url.PathEscape(o.name))
	}
	return strings.Join(parts, "; ")
}

// cardEvidence indexes one client's own grade evidence per job for a period.
type cardEvidence struct {
	period     string
	historical bool
	card       *exportpb.ClientReportCardProjection
	category   map[string]string // template id -> category id
	template   map[string]string // template id -> name
	evidence   map[string]outcome_summary.EnrollmentEvidence
	bands      map[string][]string
}

func newCardEvidence(card *exportpb.ClientReportCardProjection, period string, historical bool) *cardEvidence {
	ev := &cardEvidence{
		period:     period,
		historical: historical,
		card:       card,
		category:   map[string]string{},
		template:   map[string]string{},
		evidence:   map[string]outcome_summary.EnrollmentEvidence{},
		bands:      map[string][]string{},
	}
	for _, t := range card.GetJobTemplates() {
		if t != nil {
			ev.category[t.GetId()] = strings.TrimSpace(t.GetJobCategoryId())
			ev.template[t.GetId()] = strings.TrimSpace(t.GetName())
		}
	}
	codeByTemplatePhase := map[string]string{}
	for _, tp := range card.GetJobTemplatePhases() {
		if tp != nil && (historical || tp.GetActive()) {
			codeByTemplatePhase[tp.GetId()] = strings.ToLower(strings.TrimSpace(tp.GetCode()))
		}
	}
	jobByPhase := map[string]string{}
	for _, phase := range card.GetJobPhases() {
		if phase == nil || (!historical && !phase.GetActive()) {
			continue
		}
		if period != clientDocumentPeriodYearFinal && codeByTemplatePhase[phase.GetTemplatePhaseId()] != strings.ToLower(period) {
			continue
		}
		jobByPhase[phase.GetId()] = phase.GetJobId()
	}
	jobByTask := map[string]string{}
	for _, task := range card.GetJobTasks() {
		if task == nil || (!historical && !task.GetActive()) {
			continue
		}
		if jobID := jobByPhase[task.GetJobPhaseId()]; jobID != "" {
			jobByTask[task.GetId()] = jobID
		}
	}
	for _, outcome := range card.GetTaskOutcomes() {
		jobID := jobByTask[outcome.GetJobTaskId()]
		if jobID == "" || outcome.NumericValue == nil {
			continue
		}
		e := ev.evidence[jobID]
		e.HasMarks = true
		if outcome.GetNumericValue() > 0 {
			e.HasPositiveMark = true
		}
		ev.evidence[jobID] = e
	}
	for phaseID, summary := range latestPhaseSummaryByID(card.GetPhaseOutcomeSummaries(), historical) {
		if jobID := jobByPhase[phaseID]; jobID != "" {
			ev.bands[jobID] = append(ev.bands[jobID], summaryBand(summary.GetScaledLabel(), summary.ScaledScore))
			if summary.SummaryScore != nil {
				// The per-period summary rows print this score; it is evidence too.
				ev.bands[jobID] = append(ev.bands[jobID], formatClientReportMaximum(summary.GetSummaryScore()))
			}
		}
	}
	if period == clientDocumentPeriodYearFinal {
		for jobID, summary := range latestFinalSummaryByJob(card.GetJobOutcomeSummaries(), historical) {
			ev.bands[jobID] = append(ev.bands[jobID], summaryBand(summary.GetScaledLabel(), summary.ScaledScore))
			if summary.SummaryScore != nil {
				ev.bands[jobID] = append(ev.bands[jobID], formatClientReportMaximum(summary.GetSummaryScore()))
			}
		}
	}
	return ev
}

func summaryBand(label string, score *float64) string {
	if label = strings.TrimSpace(label); label != "" {
		return label
	}
	if score != nil {
		return formatClientReportMaximum(*score)
	}
	return ""
}

func (ev *cardEvidence) placeholder(jobID string) bool {
	return outcome_summary.IsNonEnrolledCell(ev.evidence[jobID], ev.bands[jobID]...)
}

func (ev *cardEvidence) categoryOf(job interface {
	GetJobTemplateId() string
	GetJobCategoryId() string
}) string {
	if id := ev.category[job.GetJobTemplateId()]; id != "" {
		return id
	}
	return strings.TrimSpace(job.GetJobCategoryId())
}

func (ev *cardEvidence) nameOf(job interface {
	GetName() string
	GetJobTemplateId() string
}) string {
	return firstNonEmpty(ev.template[job.GetJobTemplateId()], strings.TrimSpace(job.GetName()))
}
