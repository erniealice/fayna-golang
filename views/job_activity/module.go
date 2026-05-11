package job_activity

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"

	jobactivityaction "github.com/erniealice/fayna-golang/views/job_activity/action"
	jobactivitydetail "github.com/erniealice/fayna-golang/views/job_activity/detail"
	jobactivitylist "github.com/erniealice/fayna-golang/views/job_activity/list"
)

// ModuleDeps holds all dependencies for the job activity module.
type ModuleDeps struct {
	Routes       fayna.JobActivityRoutes
	Labels       fayna.JobActivityLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ActivityLaborRoutes is passed to the detail view so the charge tab can
	// resolve the labor edit URL without importing the activity_labor module.
	ActivityLaborRoutes fayna.ActivityLaborRoutes

	// ActivityMaterialRoutes is passed to the detail view so the charge tab can
	// resolve the material edit URL without importing the activity_material module.
	ActivityMaterialRoutes fayna.ActivityMaterialRoutes

	// ActivityExpenseRoutes is passed to the detail view so the charge tab can
	// resolve the expense edit URL without importing the activity_expense module.
	ActivityExpenseRoutes fayna.ActivityExpenseRoutes

	// GetInUseIDs checks which activity IDs are referenced by other tables
	// (posted activities, revenue line items, subtype rows). When non-nil,
	// matched rows have their delete action disabled and are excluded from
	// bulk-delete selections via data-deletable="false".
	GetInUseIDs func(ctx context.Context, ids []string) (map[string]bool, error)

	// Job activity use case functions
	GetJobActivityListPageData func(ctx context.Context, req *jobactivitypb.GetJobActivityListPageDataRequest) (*jobactivitypb.GetJobActivityListPageDataResponse, error)
	ReadJobActivity            func(ctx context.Context, req *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error)
	CreateJobActivity          func(ctx context.Context, req *jobactivitypb.CreateJobActivityRequest) (*jobactivitypb.CreateJobActivityResponse, error)
	UpdateJobActivity          func(ctx context.Context, req *jobactivitypb.UpdateJobActivityRequest) (*jobactivitypb.UpdateJobActivityResponse, error)
	DeleteJobActivity          func(ctx context.Context, req *jobactivitypb.DeleteJobActivityRequest) (*jobactivitypb.DeleteJobActivityResponse, error)

	// Approval workflow
	SubmitForApproval func(ctx context.Context, req *jobactivitypb.SubmitForApprovalRequest) (*jobactivitypb.SubmitForApprovalResponse, error)
	ApproveActivity   func(ctx context.Context, req *jobactivitypb.ApproveJobActivityRequest) (*jobactivitypb.ApproveJobActivityResponse, error)
	RejectActivity    func(ctx context.Context, req *jobactivitypb.RejectJobActivityRequest) (*jobactivitypb.RejectJobActivityResponse, error)
	PostActivity      func(ctx context.Context, req *jobactivitypb.PostJobActivityRequest) (*jobactivitypb.PostJobActivityResponse, error)
	ReverseActivity   func(ctx context.Context, req *jobactivitypb.ReverseJobActivityRequest) (*jobactivitypb.ReverseJobActivityResponse, error)

	// Activity subtype read functions (for detail page)
	ReadActivityLabor    func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	ReadActivityMaterial func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	ReadActivityExpense  func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)

	NewID func() string

	// GenerateInvoiceFromActivities creates a revenue record from a set of
	// activity IDs. Returns the new revenue ID on success.
	GenerateInvoiceFromActivities func(ctx context.Context, activityIDs []string, clientID, locationID, currency, name string) (string, error)

	// Attachment operations
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
}

// Module holds all constructed job activity views.
type Module struct {
	routes              fayna.JobActivityRoutes
	List                view.View
	Detail              view.View
	TabAction           view.View
	Create              view.View
	Update              view.View
	Delete              view.View
	BulkDelete          view.View
	Submit              view.View
	Approve             view.View
	Reject              view.View
	Post                view.View
	Reverse             view.View
	BulkGenerateInvoice view.View
	AttachmentUpload    view.View
	AttachmentDelete    view.View
}

// NewModule creates the job activity module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	listView := jobactivitylist.NewView(&jobactivitylist.ListViewDeps{
		Routes:                     deps.Routes,
		GetJobActivityListPageData: deps.GetJobActivityListPageData,
		GetInUseIDs:                deps.GetInUseIDs,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
	})

	detailDeps := &jobactivitydetail.DetailViewDeps{
		AttachmentOps: attachment.AttachmentOps{
			UploadFile:       deps.UploadFile,
			ListAttachments:  deps.ListAttachments,
			CreateAttachment: deps.CreateAttachment,
			DeleteAttachment: deps.DeleteAttachment,
			NewAttachmentID:  deps.NewID,
		},
		Routes:                 deps.Routes,
		ActivityLaborRoutes:    deps.ActivityLaborRoutes,
		ActivityMaterialRoutes: deps.ActivityMaterialRoutes,
		ActivityExpenseRoutes:  deps.ActivityExpenseRoutes,
		ReadJobActivity:        deps.ReadJobActivity,
		ReadActivityLabor:      deps.ReadActivityLabor,
		ReadActivityMaterial:   deps.ReadActivityMaterial,
		ReadActivityExpense:    deps.ReadActivityExpense,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
	}

	actionDeps := &jobactivityaction.Deps{
		Routes:                        deps.Routes,
		Labels:                        deps.Labels,
		NewID:                         deps.NewID,
		CreateJobActivity:             deps.CreateJobActivity,
		ReadJobActivity:               deps.ReadJobActivity,
		UpdateJobActivity:             deps.UpdateJobActivity,
		DeleteJobActivity:             deps.DeleteJobActivity,
		SubmitForApproval:             deps.SubmitForApproval,
		ApproveActivity:               deps.ApproveActivity,
		RejectActivity:                deps.RejectActivity,
		PostActivity:                  deps.PostActivity,
		ReverseActivity:               deps.ReverseActivity,
		GenerateInvoiceFromActivities: deps.GenerateInvoiceFromActivities,
	}

	return &Module{
		routes:              deps.Routes,
		List:                listView,
		Detail:              jobactivitydetail.NewView(detailDeps),
		TabAction:           jobactivitydetail.NewTabAction(detailDeps),
		Create:              jobactivityaction.NewAddAction(actionDeps),
		Update:              jobactivityaction.NewEditAction(actionDeps),
		Delete:              jobactivityaction.NewDeleteAction(actionDeps),
		BulkDelete:          jobactivityaction.NewBulkDeleteAction(actionDeps),
		Submit:              jobactivityaction.NewSubmitAction(actionDeps),
		Approve:             jobactivityaction.NewApproveAction(actionDeps),
		Reject:              jobactivityaction.NewRejectAction(actionDeps),
		Post:                jobactivityaction.NewPostAction(actionDeps),
		Reverse:             jobactivityaction.NewReverseAction(actionDeps),
		BulkGenerateInvoice: jobactivityaction.NewBulkGenerateInvoiceAction(actionDeps),
		AttachmentUpload:    jobactivitydetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete:    jobactivitydetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all job activity routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	// List (full page + HTMX partial share the same view)
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.ListURL+"/content", m.List)

	// Detail
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.DetailURL+"/content", m.Detail)
	if m.routes.TabActionURL != "" {
		r.GET(m.routes.TabActionURL, m.TabAction)
	}

	// CRUD actions (GET = drawer form, POST = process submission)
	r.GET(m.routes.AddURL, m.Create)
	r.POST(m.routes.AddURL, m.Create)
	r.GET(m.routes.EditURL, m.Update)
	r.POST(m.routes.EditURL, m.Update)
	r.POST(m.routes.DeleteURL, m.Delete)

	// Approval workflow
	r.POST(m.routes.SubmitURL, m.Submit)
	r.POST(m.routes.ApproveURL, m.Approve)
	r.POST(m.routes.RejectURL, m.Reject)

	// Posting workflow
	r.POST(m.routes.PostURL, m.Post)
	r.POST(m.routes.ReverseURL, m.Reverse)

	// Bulk actions
	if m.routes.BulkDeleteURL != "" {
		r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	}
	r.POST(m.routes.BulkGenerateInvoiceURL, m.BulkGenerateInvoice)

	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
