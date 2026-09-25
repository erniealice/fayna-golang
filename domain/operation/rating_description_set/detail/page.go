package detail

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/rating_description_set"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	entrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	scalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// entryPageLimit/entryMaxPages bound the offset-pagination loop that fetches
// a set's entries (fetchSetEntries below). The adapter's query budget caps a
// single page at 100 rows (espyna-golang query_budget.go maxQueryPageSize) —
// codex-review-impl2.out.md finding #7 found the clone already has 128 active
// entries in one set, silently truncated to 100 by an unfiltered, unpaginated
// single call. entryMaxPages bounds the loop independently of the adapter's
// own short-final-page termination, mirroring the established pattern in
// fayna-golang/domain/operation/outcome_summary/enrollment.go
// (markEvidencePageLimit/markEvidenceMaxPages).
const (
	entryPageLimit = 100
	entryMaxPages  = 50 // 5,000 entries/set ceiling — far beyond any real rubric
)

// ratingDescriptionSetIDFilter builds a STRING_EQUALS filter so the entries
// fetch is scoped to one set AT THE DATABASE BOUNDARY (finding #7's
// "filter by parent at the database boundary", not a client-side filter over
// a workspace-wide, cap-truncated result set).
func ratingDescriptionSetIDFilter(setID string) *commonpb.FilterRequest {
	return &commonpb.FilterRequest{
		Filters: []*commonpb.TypedFilter{{
			Field: "rating_description_set_id",
			FilterType: &commonpb.TypedFilter_StringFilter{
				StringFilter: &commonpb.StringFilter{Value: setID, Operator: commonpb.StringOperator_STRING_EQUALS},
			},
		}},
	}
}

// idSort orders each page by the primary key so OFFSET pagination is
// deterministic across pages (without an explicit unique sort key, ties in
// the adapter's default `date_created DESC` order can drop or duplicate rows
// across page boundaries — see enrollment.go markEvidenceSortByID).
func idSort() *commonpb.SortRequest {
	return &commonpb.SortRequest{Fields: []*commonpb.SortField{{Field: "id"}}}
}

func entryPage(page int32) *commonpb.PaginationRequest {
	return &commonpb.PaginationRequest{
		Limit:  entryPageLimit,
		Method: &commonpb.PaginationRequest_Offset{Offset: &commonpb.OffsetPagination{Page: page}},
	}
}

// fetchSetEntries loops offset pages (filtered to setID, sorted by id) until
// a short page signals the end, aggregating every active entry. Errors are
// returned, not swallowed (finding #6/#7's "surface load failures").
func fetchSetEntries(
	ctx context.Context,
	listFn func(ctx context.Context, req *entrypb.ListRatingDescriptionSetEntriesRequest) (*entrypb.ListRatingDescriptionSetEntriesResponse, error),
	setID string,
) ([]*entrypb.RatingDescriptionSetEntry, error) {
	if listFn == nil || setID == "" {
		return nil, nil
	}
	var all []*entrypb.RatingDescriptionSetEntry
	for page := int32(1); page <= entryMaxPages; page++ {
		resp, err := listFn(ctx, &entrypb.ListRatingDescriptionSetEntriesRequest{
			Filters:    ratingDescriptionSetIDFilter(setID),
			Sort:       idSort(),
			Pagination: entryPage(page),
		})
		if err != nil {
			return nil, fmt.Errorf("list rating description set entries (set %s, page %d): %w", setID, page, err)
		}
		data := resp.GetData()
		all = append(all, data...)
		if len(data) < entryPageLimit {
			break
		}
	}
	return all, nil
}

// MatrixColumn is one criterion column of the descriptors matrix. Columns
// are the criteria that already have >= 1 entry in this set (Q25 rule — an
// empty column is never stored).
type MatrixColumn struct {
	CriterionID string
	Name        string
}

// MatrixCell is one (band, criterion) cell.
type MatrixCell struct {
	CriterionID string
	Description string
	Empty       bool
	TestID      string
}

// MatrixRow is one band (level) row — cells are positionally aligned with
// PageData.MatrixColumns.
type MatrixRow struct {
	BandID  string
	Level   string
	IsZero  bool // band_role == "no_description" (Q20) — never editable
	Cells   []MatrixCell
}

// PageData holds the data for the rating description set detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Set             map[string]any
	Labels          rating_description_set.Labels
	Routes          rating_description_set.Routes

	MatrixColumns []MatrixColumn
	MatrixRows    []MatrixRow
	FilledCount   int
	TotalCount    int

	CanPublish   bool
	CanDeprecate bool
	// CanAddEntry gates every "+ Add entry" control (toolbar, empty-cell
	// links, empty-state CTA): DRAFT status alone is not authorization
	// (codex-review-impl2.out.md finding #11 — the entry create use case
	// requires rating_description_set:update; the drawer/matrix controls
	// must match, not just check status).
	CanAddEntry bool
	AddEntryURL string
}

func setToMap(s *setpb.RatingDescriptionSet) map[string]any {
	return map[string]any{
		"id":             s.GetId(),
		"name":           s.GetName(),
		"code":           s.GetCode(),
		"version":        s.GetVersion(),
		"version_status": versionStatusString(s.GetVersionStatus()),
		"status_variant": versionStatusVariant(s.GetVersionStatus()),
		"score_scale_id": s.GetScoreScaleId(),
		"is_draft":       s.GetVersionStatus() == enums.VersionStatus_VERSION_STATUS_DRAFT,
	}
}

func versionStatusString(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "draft"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "published"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "deprecated"
	default:
		return "draft"
	}
}

func versionStatusVariant(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "default"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "warning"
	default:
		return "default"
	}
}

// NewView creates the rating description set detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set", "read") {
			return view.Forbidden("rating_description_set:read")
		}

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadRatingDescriptionSet(ctx, &setpb.ReadRatingDescriptionSetRequest{
			Data: &setpb.RatingDescriptionSet{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read rating description set %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load rating description set: %w", err))
		}
		rows := resp.GetData()
		if len(rows) == 0 {
			return view.Error(fmt.Errorf("rating description set not found"))
		}
		record := rows[0]
		l := deps.Labels

		matrixCols, matrixRows, filled, total, err := buildMatrix(ctx, deps, record)
		if err != nil {
			log.Printf("Failed to build rating description set matrix for set %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load rating description set entries: %w", err))
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          record.GetName(),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    record.GetName(),
				HeaderSubtitle: l.Page.Heading,
				HeaderIcon:     "icon-file-text",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "rating-description-set-detail-content",
			Set:             setToMap(record),
			Labels:          l,
			Routes:          deps.Routes,
			MatrixColumns:   matrixCols,
			MatrixRows:      matrixRows,
			FilledCount:     filled,
			TotalCount:      total,
			CanPublish:      record.GetVersionStatus() == enums.VersionStatus_VERSION_STATUS_DRAFT && perms.Can("rating_description_set", "publish"),
			CanDeprecate:    record.GetVersionStatus() == enums.VersionStatus_VERSION_STATUS_PUBLISHED && perms.Can("rating_description_set", "deprecate"),
			CanAddEntry:     record.GetVersionStatus() == enums.VersionStatus_VERSION_STATUS_DRAFT && perms.Can("rating_description_set", "update"),
			AddEntryURL:     route.ResolveURL(deps.EntryRoutes.AddURL) + "?rating_description_set_id=" + record.GetId(),
		}

		return view.OK("rating-description-set-detail", pageData)
	})
}

// buildMatrix loads the set's scale bands (rows, sorted by sequence_order)
// and its active entries (cells), then the criteria referenced by those
// entries (columns, sorted by name). Bands and criteria are still small,
// workspace-wide lookup fetches (score_scale/template_task_criteria pattern);
// entries are fetched via fetchSetEntries, which filters by
// rating_description_set_id AT THE DATABASE BOUNDARY and pages to
// completeness (codex-review-impl2.out.md finding #7 — the clone already has
// 128 active entries in one set, more than the adapter's single-page cap).
func buildMatrix(ctx context.Context, deps *DetailViewDeps, set *setpb.RatingDescriptionSet) ([]MatrixColumn, []MatrixRow, int, int, error) {
	var bandRows []MatrixRow
	var criterionNames map[string]string
	var pageDataLoaded bool
	if deps.GetEntryFormPageData != nil {
		// Preferred: page data authorized under rating_description_set:read,
		// no separate score_scale_band:list / outcome_criteria:list grant
		// (codex-review-impl2.out.md finding #6 — live-confirmed: "the
		// rating_description_set detail matrix's band ROWS never render").
		pageData, err := deps.GetEntryFormPageData(ctx)
		if err != nil {
			return nil, nil, 0, 0, fmt.Errorf("get rating description set entry form page data: %w", err)
		}
		pageDataLoaded = true
		type band struct {
			id, level string
			isZero    bool
			seq       int32
		}
		var bands []band
		for _, b := range pageData.Bands {
			if b.ScoreScaleID != set.GetScoreScaleId() {
				continue
			}
			bands = append(bands, band{
				id:     b.ID,
				level:  b.Name,
				isZero: strings.EqualFold(b.BandRole, "no_description"),
				seq:    b.SequenceOrder,
			})
		}
		sort.SliceStable(bands, func(i, j int) bool { return bands[i].seq < bands[j].seq })
		for _, b := range bands {
			bandRows = append(bandRows, MatrixRow{BandID: b.id, Level: b.level, IsZero: b.isZero})
		}
		criterionNames = make(map[string]string, len(pageData.Criteria))
		for _, c := range pageData.Criteria {
			criterionNames[c.ID] = c.Name
		}
	} else if deps.ListScoreScaleBands != nil {
		if bandResp, err := deps.ListScoreScaleBands(ctx, &scalebandpb.ListScoreScaleBandsRequest{}); err == nil && bandResp != nil {
			type band struct {
				id, level string
				isZero    bool
				seq       int32
			}
			var bands []band
			for _, b := range bandResp.GetData() {
				if b == nil || !b.GetActive() || b.GetScoreScaleId() != set.GetScoreScaleId() {
					continue
				}
				bands = append(bands, band{
					id:     b.GetId(),
					level:  b.GetOutputLabel(),
					isZero: b.GetBandRole() == "no_description",
					seq:    b.GetSequenceOrder(),
				})
			}
			sort.SliceStable(bands, func(i, j int) bool { return bands[i].seq < bands[j].seq })
			for _, b := range bands {
				bandRows = append(bandRows, MatrixRow{BandID: b.id, Level: b.level, IsZero: b.isZero})
			}
		}
	}

	entries, err := fetchSetEntries(ctx, deps.ListRatingDescriptionSetEntries, set.GetId())
	if err != nil {
		return nil, nil, 0, 0, err
	}
	entriesByBandCriterion := map[string]map[string]string{} // bandID -> criterionID -> description
	criterionIDSet := map[string]bool{}
	for _, e := range entries {
		// fetchSetEntries already filters by rating_description_set_id at
		// the DB boundary; GetActive() still needs a client check because
		// the request has no active=true/false filter of its own (the
		// adapter's List defaults active=true only when the caller omits an
		// explicit "active" BooleanFilter, which fetchSetEntries does).
		if e == nil || !e.GetActive() {
			continue
		}
		criterionIDSet[e.GetOutcomeCriteriaId()] = true
		if entriesByBandCriterion[e.GetScoreScaleBandId()] == nil {
			entriesByBandCriterion[e.GetScoreScaleBandId()] = map[string]string{}
		}
		entriesByBandCriterion[e.GetScoreScaleBandId()][e.GetOutcomeCriteriaId()] = e.GetDescription()
	}

	if !pageDataLoaded {
		criterionNames = map[string]string{}
		if deps.ListOutcomeCriterias != nil && len(criterionIDSet) > 0 {
			if cResp, err := deps.ListOutcomeCriterias(ctx, &criteriapb.ListOutcomeCriteriasRequest{}); err == nil && cResp != nil {
				for _, c := range cResp.GetData() {
					if c == nil {
						continue
					}
					criterionNames[c.GetId()] = c.GetName()
				}
			}
		}
	}

	var columns []MatrixColumn
	for cid := range criterionIDSet {
		name := criterionNames[cid]
		if name == "" {
			name = cid
		}
		columns = append(columns, MatrixColumn{CriterionID: cid, Name: name})
	}
	sort.SliceStable(columns, func(i, j int) bool { return columns[i].Name < columns[j].Name })

	filled, total := 0, 0
	for i := range bandRows {
		row := &bandRows[i]
		for _, col := range columns {
			desc := entriesByBandCriterion[row.BandID][col.CriterionID]
			empty := desc == ""
			if !row.IsZero {
				total++
				if !empty {
					filled++
				}
			}
			row.Cells = append(row.Cells, MatrixCell{
				CriterionID: col.CriterionID,
				Description: desc,
				Empty:       empty,
				TestID:      fmt.Sprintf("rds-cell-%s-%s", col.CriterionID, row.Level),
			})
		}
	}

	return columns, bandRows, filled, total, nil
}
