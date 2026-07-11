package list

// template_summary.go — the education-tier ("Classes") template-grain
// delivery summary that replaces the per-job table on the job List view.
//
// docs/plan/20260710-staff-class-list/problem-statement.md (LOCKED design):
// one row per job_template that has >=1 (resolver-scoped) job for the URL
// {status} segment. Columns: template name, delivery group name, deliverer
// (staff of record), item count (DISTINCT scoped jobs), schedule name. Row
// link -> outcome_matrix.matrix ("/outcome-matrix/{id}", id=job_template_id).
//
// Row scoping is entirely resolver-level (espyna principalscope — STAFF
// principals get seat-scoped ListJobs + own-seats-only
// GetSubscriptionSeatListPageData automatically; non-staff see everything).
// This file adds NO staff_id request filter anywhere.
//
// NAMING: every identifier here is generic (JobTemplate / SubscriptionGroup /
// SubscriptionSeat / Staff — the real espyna entity names). Education
// vocabulary ("Classes"/"Section"/"Teacher"/"Students"/"Academic Year")
// enters ONLY via lyngua (packages/lyngua/translations/en/education/job.json)
// — never as a Go identifier, filename, or default label here.

import (
	"context"
	"log"
	"sort"
	"strconv"
	"strings"

	job "github.com/erniealice/fayna-golang/domain/operation/job"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	productplanpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	subscriptionseatpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_seat"
)

// templateSummaryPageLimit is the page size used by every page-looped fetch
// below (fetchSubscriptionsForPricePlan shape — centymo price_schedule/
// detail/plan/subscriptions.go), capped at the adapter's per-call maximum.
const templateSummaryPageLimit = 100

// templateSummaryRow is one job_template's aggregated delivery-summary row.
type templateSummaryRow struct {
	TemplateID    string
	TemplateName  string
	GroupName     string
	DelivererName string
	ItemCount     int
	ScheduleName  string
}

// buildDeliverySummaryTable builds the template-grain TableConfig for the
// education tier: fetch the resolver-scoped jobs for {status}, aggregate by
// job_template_id, then resolve each template's delivery group + deliverer +
// schedule. Templates with no output_product_id (job_template.OutputProductId
// unset) or no scoped jobs produce no row.
func buildDeliverySummaryTable(ctx context.Context, deps *ListViewDeps, status string, perms *types.UserPermissions) (*types.TableConfig, error) {
	jobs, err := fetchScopedJobs(ctx, deps, status)
	if err != nil {
		return nil, err
	}

	rows := buildTemplateSummaryRows(ctx, deps, jobs)

	l := deps.Labels
	columns := templateSummaryColumns(l)
	tableRows := make([]types.TableRow, 0, len(rows))
	for _, r := range rows {
		matrixURL := route.ResolveURL(deps.MatrixDetailURL, "id", r.TemplateID)
		tableRows = append(tableRows, types.TableRow{
			ID:   r.TemplateID,
			Href: matrixURL,
			Cells: []types.TableCell{
				{Type: "text", Value: r.TemplateName},
				{Type: "text", Value: r.GroupName},
				{Type: "text", Value: r.DelivererName},
				{Type: "number", Value: strconv.Itoa(r.ItemCount)},
				{Type: "text", Value: r.ScheduleName},
			},
			DataAttrs: map[string]string{
				"name":      r.TemplateName,
				"group":     r.GroupName,
				"deliverer": r.DelivererName,
				"schedule":  r.ScheduleName,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: matrixURL},
			},
		})
	}
	types.ApplyColumnStyles(columns, tableRows)

	tableConfig := &types.TableConfig{
		ID:                   "job-template-summary-table",
		Columns:              columns,
		Rows:                 tableRows,
		ShowSearch:           true,
		ShowActions:          true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "group",
		DefaultSortDirection: "asc",
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}
	types.ApplyTableSettings(tableConfig)
	return tableConfig, nil
}

// templateSummaryColumns declares the template-grain columns. Go defaults
// (job.Labels.Columns.{Group,Deliverer,Items,Schedule}) stay generic
// ("Group"/"Deliverer"/"Items"/"Schedule"); education overrides them to
// "Section"/"Teacher"/"Students"/"Academic Year" via lyngua.
func templateSummaryColumns(l job.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "group", Label: l.Columns.Group},
		{Key: "deliverer", Label: l.Columns.Deliverer},
		{Key: "items", Label: l.Columns.Items},
		{Key: "schedule", Label: l.Columns.Schedule},
	}
}

// buildTemplateSummaryRows aggregates jobs by job_template_id (DISTINCT
// scoped-job count per template), then resolves each template's name +
// output product, delivery group (via the SAMPLE job's own origin
// subscription's subscription_group_member row — never via sgpps(plan,staff),
// which over-counts to cohort grain per the S2 build-time note), deliverer
// (the seat on that subscription whose product_plan.product_id matches the
// template's output_product_id), and schedule (the group's nested
// price_schedule name). Rows are sorted by group name then template name
// (LOCKED order).
func buildTemplateSummaryRows(ctx context.Context, deps *ListViewDeps, jobs []*jobpb.Job) []templateSummaryRow {
	type templateAgg struct {
		itemCount    int
		sampleOrigin string
	}
	aggs := map[string]*templateAgg{}
	var order []string
	for _, j := range jobs {
		tid := j.GetJobTemplateId()
		if tid == "" {
			continue
		}
		a, ok := aggs[tid]
		if !ok {
			a = &templateAgg{}
			aggs[tid] = a
			order = append(order, tid)
		}
		a.itemCount++
		if a.sampleOrigin == "" && j.GetOriginType() == enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION {
			if oid := j.GetOriginId(); oid != "" {
				a.sampleOrigin = oid
			}
		}
	}
	if len(order) == 0 {
		return nil
	}

	templates := fetchJobTemplatesByID(ctx, deps, order)
	productPlans := fetchAllProductPlans(ctx, deps)
	seatDeliverers := fetchSeatDeliverers(ctx, deps, productPlans)
	staff := fetchAllStaff(ctx, deps)

	var sampleOrigins []string
	seenOrigin := map[string]bool{}
	for _, tid := range order {
		if o := aggs[tid].sampleOrigin; o != "" && !seenOrigin[o] {
			seenOrigin[o] = true
			sampleOrigins = append(sampleOrigins, o)
		}
	}
	subToGroup := fetchSubscriptionGroupIDs(ctx, deps, sampleOrigins)
	groups := fetchAllSubscriptionGroups(ctx, deps)

	rows := make([]templateSummaryRow, 0, len(order))
	for _, tid := range order {
		a := aggs[tid]
		tpl := templates[tid]
		if tpl == nil || tpl.GetOutputProductId() == "" {
			// No output product => not a deliverable template; no row
			// (problem-statement.md §3 — "Templates without output_product_id
			// or without scoped jobs simply don't produce rows").
			continue
		}

		var groupName, scheduleName string
		if groupID := subToGroup[a.sampleOrigin]; groupID != "" {
			if g := groups[groupID]; g != nil {
				groupName = g.GetName()
				scheduleName = g.GetPriceSchedule().GetName()
			}
		}

		var delivererName string
		if staffID := seatDeliverers[a.sampleOrigin+"|"+tpl.GetOutputProductId()]; staffID != "" {
			if s := staff[staffID]; s != nil {
				delivererName = staffDisplayName(s)
			}
		}

		rows = append(rows, templateSummaryRow{
			TemplateID:    tid,
			TemplateName:  tpl.GetName(),
			GroupName:     groupName,
			DelivererName: delivererName,
			ItemCount:     a.itemCount,
			ScheduleName:  scheduleName,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].GroupName != rows[j].GroupName {
			return rows[i].GroupName < rows[j].GroupName
		}
		return rows[i].TemplateName < rows[j].TemplateName
	})
	return rows
}

// staffDisplayName mirrors the established fayna staff-name pattern
// (block.go newActivityLaborStaffSearchHandler / newJobClientSearchHandler):
// first + last name from the embedded User, falling back to the staff id.
func staffDisplayName(s *staffpb.Staff) string {
	if u := s.GetUser(); u != nil {
		if name := strings.TrimSpace(u.GetFirstName() + " " + u.GetLastName()); name != "" {
			return name
		}
	}
	return s.GetId()
}

// fetchJobTemplatesByID pages through ListJobTemplates (Pagination IS
// forwarded by this adapter — see block/usecases.go JobTemplateUseCases)
// and returns every requested id found, keyed by id. Unlike the other
// lookups below, this one is bounded by real page-looping rather than a
// single-call-under-cap assumption, since job_template count (130 in
// education1) already exceeds the adapter's 100-row default.
func fetchJobTemplatesByID(ctx context.Context, deps *ListViewDeps, ids []string) map[string]*jobtemplatepb.JobTemplate {
	out := map[string]*jobtemplatepb.JobTemplate{}
	if deps.ListJobTemplates == nil || len(ids) == 0 {
		return out
	}
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	for page := int32(1); ; page++ {
		resp, err := deps.ListJobTemplates(ctx, &jobtemplatepb.ListJobTemplatesRequest{
			Pagination: &commonpb.PaginationRequest{
				Limit:  templateSummaryPageLimit,
				Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
			},
		})
		if err != nil {
			log.Printf("Failed to list job templates: %v", err)
			return out
		}
		batch := resp.GetData()
		for _, t := range batch {
			if id := t.GetId(); id != "" && want[id] {
				out[id] = t
			}
		}
		if int32(len(batch)) < templateSummaryPageLimit {
			return out
		}
	}
}

// fetchAllProductPlans returns every product_plan in one call, keyed by id.
// Single call — ground truth: 43 rows (education1), under the adapter's
// 100-row default; ListProductPlans' listAll does not forward Pagination
// (verified against the postgres adapter), so this is complete only while
// the workspace's product_plan count stays under 100.
func fetchAllProductPlans(ctx context.Context, deps *ListViewDeps) map[string]*productplanpb.ProductPlan {
	out := map[string]*productplanpb.ProductPlan{}
	if deps.ListProductPlans == nil {
		return out
	}
	resp, err := deps.ListProductPlans(ctx, &productplanpb.ListProductPlansRequest{})
	if err != nil {
		log.Printf("Failed to list product plans: %v", err)
		return out
	}
	for _, p := range resp.GetData() {
		if id := p.GetId(); id != "" {
			out[id] = p
		}
	}
	return out
}

// fetchAllSubscriptionGroups returns every subscription_group in one call,
// keyed by id. Single call — ground truth: 16 rows (education1, both AY
// 2024-25 and 2025-26 sections). The adapter
// hydrates each group's nested PriceSchedule (STATUS-AGNOSTIC join), so no
// separate price_schedule fetch is needed for the schedule-name column.
func fetchAllSubscriptionGroups(ctx context.Context, deps *ListViewDeps) map[string]*subscriptiongrouppb.SubscriptionGroup {
	out := map[string]*subscriptiongrouppb.SubscriptionGroup{}
	if deps.ListSubscriptionGroups == nil {
		return out
	}
	resp, err := deps.ListSubscriptionGroups(ctx, &subscriptiongrouppb.ListSubscriptionGroupsRequest{})
	if err != nil {
		log.Printf("Failed to list subscription groups: %v", err)
		return out
	}
	for _, g := range resp.GetData() {
		if id := g.GetId(); id != "" {
			out[id] = g
		}
	}
	return out
}

// fetchAllStaff pages through GetStaffListPageData (the User-hydrating read
// — the bare ListStaffs RPC never populates Staff.User, confirmed against
// the postgres adapter — a plain table scan, no join to "user" — which is
// why the deliverer column showed a raw staff id before this fetch existed)
// and returns every row keyed by id. Ground truth: 47 active staff rows
// (education1; the 21-22 "seat-holding" figure elsewhere in this file/plan
// is the task-assigned SUBSET, not the workspace's total staff headcount),
// so this is one page in practice, but the loop is real (this adapter DOES
// honor Pagination, unlike ListSubscriptionGroups/ListSubscriptionGroupMembers/
// ListProductPlans above).
func fetchAllStaff(ctx context.Context, deps *ListViewDeps) map[string]*staffpb.Staff {
	out := map[string]*staffpb.Staff{}
	if deps.GetStaffListPageData == nil {
		return out
	}
	for page := int32(1); ; page++ {
		resp, err := deps.GetStaffListPageData(ctx, &staffpb.GetStaffListPageDataRequest{
			Pagination: &commonpb.PaginationRequest{
				Limit:  templateSummaryPageLimit,
				Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
			},
		})
		if err != nil {
			log.Printf("Failed to list staff: %v", err)
			return out
		}
		batch := resp.GetStaffList()
		for _, s := range batch {
			if id := s.GetId(); id != "" {
				out[id] = s
			}
		}
		if int32(len(batch)) < templateSummaryPageLimit {
			return out
		}
	}
}

// fetchSubscriptionGroupIDs resolves subscription_id -> subscription_group_id
// for exactly the given (sample, deduplicated) subscription ids, chunked into
// batches of templateSummaryPageLimit so each ListFilter(IN) call's result
// set stays under the adapter's 100-row default (ListSubscriptionGroupMembers'
// listAll does not forward Pagination — verified against the postgres
// adapter — so a single unchunked call over >100 ids could silently drop
// matches; chunking the INPUT side keeps every call's OUTPUT bounded).
func fetchSubscriptionGroupIDs(ctx context.Context, deps *ListViewDeps, subscriptionIDs []string) map[string]string {
	out := map[string]string{}
	if deps.ListSubscriptionGroupMembers == nil || len(subscriptionIDs) == 0 {
		return out
	}
	for start := 0; start < len(subscriptionIDs); start += templateSummaryPageLimit {
		end := start + templateSummaryPageLimit
		if end > len(subscriptionIDs) {
			end = len(subscriptionIDs)
		}
		chunk := subscriptionIDs[start:end]
		resp, err := deps.ListSubscriptionGroupMembers(ctx, &subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{{
					Field: "subscription_id",
					FilterType: &commonpb.TypedFilter_ListFilter{
						ListFilter: &commonpb.ListFilter{Values: chunk, Operator: commonpb.ListOperator_LIST_IN},
					},
				}},
			},
		})
		if err != nil {
			log.Printf("Failed to list subscription group members: %v", err)
			continue
		}
		for _, m := range resp.GetData() {
			sid := m.GetSubscriptionId()
			if sid == "" {
				continue
			}
			if _, ok := out[sid]; !ok {
				out[sid] = m.GetSubscriptionGroupId()
			}
		}
	}
	return out
}

// fetchSeatDeliverers pages through GetSubscriptionSeatListPageData (staff-
// scoped + fail-closed at the espyna adapter — principalscope.StaffScopeClause
// — so a STAFF principal only ever sees its own seats here) and builds
// "subscriptionID|productID" -> staffID, translating each seat's
// product_plan_id to a product_id via the productPlans lookup.
func fetchSeatDeliverers(ctx context.Context, deps *ListViewDeps, productPlans map[string]*productplanpb.ProductPlan) map[string]string {
	out := map[string]string{}
	if deps.GetSubscriptionSeatListPageData == nil {
		return out
	}
	for page := int32(1); ; page++ {
		resp, err := deps.GetSubscriptionSeatListPageData(ctx, &subscriptionseatpb.GetSubscriptionSeatListPageDataRequest{
			Pagination: &commonpb.PaginationRequest{
				Limit:  templateSummaryPageLimit,
				Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
			},
		})
		if err != nil {
			log.Printf("Failed to list subscription seats: %v", err)
			return out
		}
		batch := resp.GetSubscriptionSeatList()
		for _, seat := range batch {
			plan := productPlans[seat.GetProductPlanId()]
			if plan == nil {
				continue
			}
			productID := plan.GetProductId()
			subID := seat.GetSubscriptionId()
			staffID := seat.GetStaffId()
			if productID == "" || subID == "" || staffID == "" {
				continue
			}
			key := subID + "|" + productID
			if _, ok := out[key]; !ok {
				out[key] = staffID
			}
		}
		if int32(len(batch)) < templateSummaryPageLimit {
			return out
		}
	}
}
