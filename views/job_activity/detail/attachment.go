package detail

import (
	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/pyeza-golang/view"
)

func attachmentConfig(deps *DetailViewDeps) *attachment.Config {
	return &attachment.Config{
		EntityType:       "job_activity",
		BucketName:       "attachments",
		UploadURL:        deps.Routes.AttachmentUploadURL,
		DeleteURL:        deps.Routes.AttachmentDeleteURL,
		Labels:           attachment.DefaultLabels(),
		CommonLabels:     deps.CommonLabels,
		NewID:            deps.NewAttachmentID,
		UploadFile:       deps.UploadFile,
		ListAttachments:  deps.ListAttachments,
		CreateAttachment: deps.CreateAttachment,
		DeleteAttachment: deps.DeleteAttachment,
	}
}

// NewAttachmentUploadAction creates the upload handler for job activity attachments.
func NewAttachmentUploadAction(deps *DetailViewDeps) view.View {
	return attachment.NewUploadAction(attachmentConfig(deps))
}

// NewAttachmentDeleteAction creates the delete handler for job activity attachments.
func NewAttachmentDeleteAction(deps *DetailViewDeps) view.View {
	return attachment.NewDeleteAction(attachmentConfig(deps))
}
