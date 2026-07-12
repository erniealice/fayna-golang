// Package section renders view-2 of the report-cards surface: a per-section
// (subscription_group) grid of students × subjects, each cell = the student's
// year-final rating (job_outcome_summary.scaled_label) for that subject,
// linking out to the existing per-job summary. It is a READ-ONLY reporting
// table (types.TableConfig + Groups gender bands) — the editing surface is the
// grade sheet (outcome_matrix).
//
// Security (Q-SEC-7): the section id is EXISTS-gated against the session
// workspace (the workspace-aware ListSubscriptionGroups adapter returns a
// foreign-workspace group as no-rows → fail-closed). Every read is
// workspace-bound at the espyna adapter; the row/column set is derived from the
// section's JOBS (ListJobs), which the adapter narrows to the acting STAFF
// principal's reachable jobs — so a teacher sees only their students' cards.
package section

import (
	"context"
	"html"
	"log"
	"sort"
	"strconv"
	"strings"

	texttemplate "html/template"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

// pageLimit chunks ListFilter(IN) id sets so each call's result set stays under
// the adapter's default cap (the fetchClientNames pattern).
const pageLimit = 100

// Deps holds the section-grid view dependencies. All list closures are
// workspace-bound at the espyna adapter; the view composes no client_id/
// workspace filter of its own.
type Deps struct {
	Routes       outcome_summary.Routes
	Labels       outcome_summary.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Options — app-configured row presentation (bands + sort). Zero value →
	// flat rows (no bands), name sort.
	Options outcome_summary.Options

	ListSubscriptionGroups       func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListSubscriptionGroupMembers func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListJobs                     func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListJobTemplates             func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ListClients                  func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
	ListJobOutcomeSummarys       func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)
	ListClientAttributes         func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ResolveAttributeIDByCode     func(ctx context.Context, code string) (string, error)
}

// PageData is the section-grid page data.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	NotComputed     bool
	Banner          string
}

// student is one row's resolved identity.
type student struct {
	clientID string
	name     string
	lastName string
}

// NewView creates the per-section report-card grid view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "list") {
			return view.Forbidden("job_outcome_summary:list")
		}

		sectionID := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		if sectionID == "" {
			return view.Forbidden("job_outcome_summary:list")
		}

		// EXISTS gate: the group must belong to the session workspace. The
		// ListSubscriptionGroups adapter is workspace-scoped, so a foreign or
		// missing id returns no rows → fail-closed (no leak).
		group := fetchSection(ctx, deps, sectionID)
		if group == nil {
			return view.Forbidden("job_outcome_summary:list")
		}

		l := deps.Labels

		// members(section) → subscription_id → client_id.
		subToClient := fetchMembers(ctx, deps, sectionID)

		// jobs(origin_id IN subs, active, SUBSCRIPTION) — staff-narrowed at the
		// adapter. Rows + columns are derived from THIS set (Q-SEC-7).
		jobs := fetchSectionJobs(ctx, deps, keysOf(subToClient))

		// Build the (client, template) → job map + distinct sets.
		type cell struct{ jobID string }
		cellJob := map[string]string{} // clientID+"\x00"+templateID -> jobID
		clientIDs := []string{}
		clientSeen := map[string]bool{}
		templateIDs := []string{}
		tmplSeen := map[string]bool{}
		jobIDs := []string{}
		for _, j := range jobs {
			jobID := j.GetId()
			tid := j.GetJobTemplateId()
			clientID := subToClient[j.GetOriginId()]
			if clientID == "" {
				clientID = j.GetClientId()
			}
			if clientID == "" || tid == "" || jobID == "" {
				continue
			}
			if !clientSeen[clientID] {
				clientSeen[clientID] = true
				clientIDs = append(clientIDs, clientID)
			}
			if !tmplSeen[tid] {
				tmplSeen[tid] = true
				templateIDs = append(templateIDs, tid)
			}
			cellJob[clientID+"\x00"+tid] = jobID
			jobIDs = append(jobIDs, jobID)
		}

		// summaries(job_id IN jobs) → jobID → scaled label.
		labelByJob := fetchSummaryLabels(ctx, deps, jobIDs, l)

		// Empty-state: no computed summaries for this section → banner, not a
		// blank grid (D.4 do-not-ship-blank).
		if len(labelByJob) == 0 {
			return okPage(viewCtx, deps, group, nil, l.Section.NotComputedBanner)
		}

		// client display names + last_name (for the row sort).
		students := fetchStudents(ctx, deps, clientIDs)

		// gender (or configured) attribute values for bands.
		attrValues := fetchAttributeValues(ctx, deps, clientIDs)

		// template names (columns), name ASC.
		tmplNames := fetchTemplateNames(ctx, deps, templateIDs)
		columns := buildColumns(templateIDs, tmplNames, l)

		table := &types.TableConfig{
			ID:          "report-cards-grid",
			Columns:     columns,
			ShowSearch:  true,
			ShowColumns: true,
			ShowDensity: true,
			Labels:      deps.TableLabels,
			Caption:     l.Section.Title,
			FixedLayout: false,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Section.NotComputedBanner,
			},
		}

		rows := buildRows(students, orderedColumnIDs(columns), cellJob, labelByJob, deps.Routes, l)
		applyRowPresentation(table, rows, students, attrValues, deps.Options)
		types.ApplyColumnStyles(table.Columns, allRows(table))

		return okPage(viewCtx, deps, group, table, "")
	})
}

// okPage assembles the PageData (grid or empty-state banner).
func okPage(viewCtx *view.ViewContext, deps *Deps, group *subscriptiongrouppb.SubscriptionGroup, table *types.TableConfig, banner string) view.ViewResult {
	l := deps.Labels
	pd := &PageData{
		PageData: types.PageData{
			CacheVersion:   viewCtx.CacheVersion,
			Title:          l.Section.Title,
			CurrentPath:    viewCtx.CurrentPath,
			ActiveNav:      deps.Routes.ActiveNav,
			ActiveSubNav:   "report-cards",
			HeaderTitle:    l.Section.Title,
			HeaderSubtitle: group.GetName(),
			HeaderIcon:     "icon-award",
			CommonLabels:   deps.CommonLabels,
		},
		ContentTemplate: "outcome-summary-section-content",
		Table:           table,
		NotComputed:     table == nil,
		Banner:          banner,
	}
	return view.OK("outcome-summary-section", pd)
}

// fetchSection returns the section (workspace-scoped EXISTS gate) or nil.
func fetchSection(ctx context.Context, deps *Deps, sectionID string) *subscriptiongrouppb.SubscriptionGroup {
	if deps.ListSubscriptionGroups == nil {
		return nil
	}
	resp, err := deps.ListSubscriptionGroups(ctx, &subscriptiongrouppb.ListSubscriptionGroupsRequest{
		Filters: &commonpb.FilterRequest{
			Filters: []*commonpb.TypedFilter{stringEq("id", sectionID)},
		},
	})
	if err != nil {
		log.Printf("report cards section: list subscription group by id: %v", err)
		return nil
	}
	for _, g := range resp.GetData() {
		if g.GetId() == sectionID {
			return g
		}
	}
	return nil
}

// fetchMembers returns subscription_id → client_id for the active members.
func fetchMembers(ctx context.Context, deps *Deps, sectionID string) map[string]string {
	out := map[string]string{}
	if deps.ListSubscriptionGroupMembers == nil {
		return out
	}
	resp, err := deps.ListSubscriptionGroupMembers(ctx, &subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
		Filters: &commonpb.FilterRequest{
			Filters: []*commonpb.TypedFilter{stringEq("subscription_group_id", sectionID)},
		},
	})
	if err != nil {
		log.Printf("report cards section: list members: %v", err)
		return out
	}
	for _, m := range resp.GetData() {
		if !m.GetActive() {
			continue
		}
		if sid, cid := m.GetSubscriptionId(), m.GetClientId(); sid != "" && cid != "" {
			out[sid] = cid
		}
	}
	return out
}

// fetchSectionJobs returns the section's active subscription-origin jobs
// (staff-narrowed at the adapter), chunked by origin_id.
func fetchSectionJobs(ctx context.Context, deps *Deps, subIDs []string) []*jobpb.Job {
	var out []*jobpb.Job
	if deps.ListJobs == nil || len(subIDs) == 0 {
		return out
	}
	for start := 0; start < len(subIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(subIDs) {
			end = len(subIDs)
		}
		resp, err := deps.ListJobs(ctx, &jobpb.ListJobsRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{listIn("origin_id", subIDs[start:end])},
			},
		})
		if err != nil {
			log.Printf("report cards section: list jobs: %v", err)
			continue
		}
		for _, j := range resp.GetData() {
			if j.GetActive() && j.GetOriginType() == enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION {
				out = append(out, j)
			}
		}
	}
	return out
}

// fetchSummaryLabels returns jobID → year-final label (scaled_label, falling
// back to a formatted scaled_score), chunked by job_id.
func fetchSummaryLabels(ctx context.Context, deps *Deps, jobIDs []string, l outcome_summary.Labels) map[string]string {
	out := map[string]string{}
	if deps.ListJobOutcomeSummarys == nil || len(jobIDs) == 0 {
		return out
	}
	for start := 0; start < len(jobIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(jobIDs) {
			end = len(jobIDs)
		}
		resp, err := deps.ListJobOutcomeSummarys(ctx, &jobsumpb.ListJobOutcomeSummarysRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{listIn("job_id", jobIDs[start:end])},
			},
		})
		if err != nil {
			log.Printf("report cards section: list job outcome summaries: %v", err)
			continue
		}
		for _, s := range resp.GetData() {
			if !s.GetActive() {
				continue
			}
			jid := s.GetJobId()
			if jid == "" {
				continue
			}
			if lbl := strings.TrimSpace(s.GetScaledLabel()); lbl != "" {
				out[jid] = lbl
			} else if s.ScaledScore != nil {
				out[jid] = strconv.FormatFloat(s.GetScaledScore(), 'f', -1, 64)
			}
		}
	}
	return out
}

// fetchStudents resolves client_id → display name + last_name, chunked.
func fetchStudents(ctx context.Context, deps *Deps, clientIDs []string) map[string]student {
	out := map[string]student{}
	if deps.ListClients == nil || len(clientIDs) == 0 {
		return out
	}
	for start := 0; start < len(clientIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(clientIDs) {
			end = len(clientIDs)
		}
		resp, err := deps.ListClients(ctx, &clientpb.ListClientsRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{listIn("id", clientIDs[start:end])},
			},
		})
		if err != nil {
			log.Printf("report cards section: list clients: %v", err)
			continue
		}
		for _, c := range resp.GetData() {
			id := c.GetId()
			if id == "" {
				continue
			}
			out[id] = student{clientID: id, name: clientDisplayName(c), lastName: clientLastName(c)}
		}
	}
	return out
}

// fetchAttributeValues resolves each Options-referenced attribute code to its id
// then loads the roster clients' values — code → (client_id → value). Mirrors
// the outcome_matrix hydration. Nil-safe.
func fetchAttributeValues(ctx context.Context, deps *Deps, clientIDs []string) map[string]map[string]string {
	codes := deps.Options.AttributeCodes()
	if len(codes) == 0 || deps.ListClientAttributes == nil || deps.ResolveAttributeIDByCode == nil || len(clientIDs) == 0 {
		return nil
	}
	out := make(map[string]map[string]string, len(codes))
	for _, code := range codes {
		attrID, err := deps.ResolveAttributeIDByCode(ctx, code)
		if err != nil || attrID == "" {
			log.Printf("report cards section: attribute code %q did not resolve (bands ignored for it): %v", code, err)
			continue
		}
		vals := map[string]string{}
		for start := 0; start < len(clientIDs); start += pageLimit {
			end := start + pageLimit
			if end > len(clientIDs) {
				end = len(clientIDs)
			}
			resp, err := deps.ListClientAttributes(ctx, &clientattributepb.ListClientAttributesRequest{
				Filters: &commonpb.FilterRequest{
					Filters: []*commonpb.TypedFilter{
						stringEq("attribute_id", attrID),
						listIn("client_id", clientIDs[start:end]),
					},
				},
			})
			if err != nil {
				log.Printf("report cards section: list client attributes for %q: %v", code, err)
				continue
			}
			for _, ca := range resp.GetData() {
				if cid, v := ca.GetClientId(), strings.TrimSpace(ca.GetValue()); cid != "" && v != "" {
					vals[cid] = v
				}
			}
		}
		out[code] = vals
	}
	return out
}

// fetchTemplateNames resolves job_template_id → name, chunked.
func fetchTemplateNames(ctx context.Context, deps *Deps, templateIDs []string) map[string]string {
	out := map[string]string{}
	if deps.ListJobTemplates == nil || len(templateIDs) == 0 {
		return out
	}
	for start := 0; start < len(templateIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(templateIDs) {
			end = len(templateIDs)
		}
		resp, err := deps.ListJobTemplates(ctx, &jobtemplatepb.ListJobTemplatesRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{listIn("id", templateIDs[start:end])},
			},
		})
		if err != nil {
			log.Printf("report cards section: list job templates: %v", err)
			continue
		}
		for _, t := range resp.GetData() {
			if id := t.GetId(); id != "" {
				out[id] = t.GetName()
			}
		}
	}
	return out
}

// buildColumns builds the first (student) column + one column per template,
// ordered by template name ASC (prod's canonical subject order).
func buildColumns(templateIDs []string, names map[string]string, l outcome_summary.Labels) []types.TableColumn {
	ordered := append([]string(nil), templateIDs...)
	sort.SliceStable(ordered, func(i, j int) bool {
		a := strings.ToLower(colName(names, ordered[i]))
		b := strings.ToLower(colName(names, ordered[j]))
		if a == b {
			return ordered[i] < ordered[j]
		}
		return a < b
	})
	cols := make([]types.TableColumn, 0, len(ordered)+1)
	cols = append(cols, types.TableColumn{Key: "student", Label: l.Section.ClientColumn, MinWidth: "11.25rem"})
	for _, tid := range ordered {
		cols = append(cols, types.TableColumn{Key: "tmpl-" + tid, Label: colName(names, tid), MinWidth: "6.25rem", Align: "center", NoSort: true})
	}
	return cols
}

// orderedColumnIDs returns the template ids in the same order as buildColumns
// laid out the subject columns (skipping the leading student column).
func orderedColumnIDs(cols []types.TableColumn) []string {
	var ids []string
	for _, c := range cols {
		if tid, ok := strings.CutPrefix(c.Key, "tmpl-"); ok {
			ids = append(ids, tid)
		}
	}
	return ids
}

// buildRows builds one row per student: student-name cell + a rating cell per
// subject column (linking to the per-job summary; "—" when no summary).
func buildRows(
	students map[string]student,
	templateIDs []string,
	cellJob, labelByJob map[string]string,
	routes outcome_summary.Routes,
	l outcome_summary.Labels,
) []types.TableRow {
	empty := l.Section.RatingEmpty
	if empty == "" {
		empty = "—"
	}
	rows := make([]types.TableRow, 0, len(students))
	for clientID, st := range students {
		cells := make([]types.TableCell, 0, len(templateIDs)+1)
		cells = append(cells, types.TableCell{Value: st.name})
		for _, tid := range templateIDs {
			testid := "rc-cell-" + short(clientID) + "-" + short(tid)
			jobID := cellJob[clientID+"\x00"+tid]
			label := labelByJob[jobID]
			if jobID != "" && label != "" {
				url := route.ResolveURL(routes.JobSummaryURL, "id", jobID)
				anchor := `<a href="` + html.EscapeString(url) + `" class="table-link" data-testid="` + testid + `" hx-push-url="true">` + html.EscapeString(label) + `</a>`
				cells = append(cells, types.TableCell{Type: "html", HTML: texttemplate.HTML(anchor)})
			} else {
				span := `<span class="rc-cell-empty" data-testid="` + testid + `">` + html.EscapeString(empty) + `</span>`
				cells = append(cells, types.TableCell{Type: "html", HTML: texttemplate.HTML(span)})
			}
		}
		rows = append(rows, types.TableRow{
			ID:        clientID,
			DataAttrs: map[string]string{"testid": "rc-row-" + short(clientID)},
			Cells:     cells,
		})
	}
	return rows
}

// applyRowPresentation applies the configured sort + gender bands. When a
// group-by attribute is configured the rows are partitioned into
// TableRowGroup bands (value-ascending, no-value band last — the documented
// fallback ordering); otherwise the flat rows are sorted in place.
func applyRowPresentation(
	table *types.TableConfig,
	rows []types.TableRow,
	students map[string]student,
	attrValues map[string]map[string]string,
	opts outcome_summary.Options,
) {
	sortRows(rows, students, opts.Row)

	groupCode, ok := outcome_summary.ClientAttributeCode(opts.Row.GroupByField)
	var vals map[string]string
	if ok {
		vals = attrValues[groupCode]
	}
	if !ok || len(vals) == 0 {
		table.Rows = rows
		return
	}

	// Partition into value bands (rows keep their sorted order within a band).
	order := []string{}
	seen := map[string]bool{}
	buckets := map[string][]types.TableRow{}
	for _, r := range rows {
		v := vals[r.ID]
		if !seen[v] {
			seen[v] = true
			order = append(order, v)
		}
		buckets[v] = append(buckets[v], r)
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if (a == "") != (b == "") {
			return b == "" // no-value band last
		}
		return strings.ToLower(a) < strings.ToLower(b)
	})
	groups := make([]types.TableRowGroup, 0, len(order))
	for _, v := range order {
		title := v
		if title == "" {
			title = "—"
		}
		groups = append(groups, types.TableRowGroup{
			ID:        "rc-band-" + slug(v),
			Title:     title,
			Rows:      buckets[v],
			DataAttrs: map[string]string{"testid": "rc-band-" + slug(v)},
		})
	}
	table.Groups = groups
}

// sortRows orders rows by the configured client-column SortField (only
// "last_name" is implemented today), direction-aware, stable; ties fall back to
// display name then id.
func sortRows(rows []types.TableRow, students map[string]student, opts outcome_summary.RowOptions) {
	desc := opts.Direction() == "desc"
	field := strings.TrimSpace(opts.SortField)
	sort.SliceStable(rows, func(i, j int) bool {
		si, sj := students[rows[i].ID], students[rows[j].ID]
		var a, b string
		switch field {
		case "last_name":
			a, b = strings.ToLower(si.lastName), strings.ToLower(sj.lastName)
		default:
			a, b = strings.ToLower(si.name), strings.ToLower(sj.name)
		}
		if a == b {
			an, bn := strings.ToLower(si.name), strings.ToLower(sj.name)
			if an == bn {
				return rows[i].ID < rows[j].ID
			}
			return an < bn
		}
		// values present sort before empties regardless of direction.
		if (a == "") != (b == "") {
			return a != ""
		}
		if desc {
			return a > b
		}
		return a < b
	})
}

// allRows flattens the table's rows/groups for ApplyColumnStyles.
func allRows(table *types.TableConfig) []types.TableRow {
	if table == nil {
		return nil
	}
	if len(table.Groups) == 0 {
		return table.Rows
	}
	var out []types.TableRow
	for _, g := range table.Groups {
		out = append(out, g.Rows...)
	}
	return out
}

// --- small helpers -------------------------------------------------------

func stringEq(field, value string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_StringFilter{
			StringFilter: &commonpb.StringFilter{Value: value, Operator: commonpb.StringOperator_STRING_EQUALS},
		},
	}
}

func listIn(field string, values []string) *commonpb.TypedFilter {
	return &commonpb.TypedFilter{
		Field: field,
		FilterType: &commonpb.TypedFilter_ListFilter{
			ListFilter: &commonpb.ListFilter{Values: values, Operator: commonpb.ListOperator_LIST_IN},
		},
	}
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func colName(names map[string]string, id string) string {
	if n := strings.TrimSpace(names[id]); n != "" {
		return n
	}
	return id
}

// clientDisplayName mirrors the fayna client-name pattern (client's own name
// column first, then embedded User first+last, then id).
func clientDisplayName(c *clientpb.Client) string {
	if name := strings.TrimSpace(c.GetName()); name != "" {
		return name
	}
	if fn := strings.TrimSpace(c.GetFirstName() + " " + c.GetLastName()); fn != "" {
		return fn
	}
	if u := c.GetUser(); u != nil {
		if name := strings.TrimSpace(u.GetFirstName() + " " + u.GetLastName()); name != "" {
			return name
		}
	}
	return c.GetId()
}

// clientLastName prefers the client's own last_name column, then the embedded
// User's last name.
func clientLastName(c *clientpb.Client) string {
	if ln := strings.TrimSpace(c.GetLastName()); ln != "" {
		return ln
	}
	if u := c.GetUser(); u != nil {
		if ln := strings.TrimSpace(u.GetLastName()); ln != "" {
			return ln
		}
	}
	return ""
}

// short returns a stable, collision-resistant slug (the uuidv7 random TAIL,
// not the shared timestamp prefix — see the list view's short() note). Keeps
// rc-row / rc-cell testids unique across a section's students and subjects.
func short(id string) string {
	if len(id) > 8 {
		return id[len(id)-8:]
	}
	return id
}

func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "none"
	}
	return b.String()
}
