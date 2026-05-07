package detail

import (
	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/pyeza-golang/view"
)

func attachmentConfig(deps *DetailViewDeps) *attachment.Config {
	return &attachment.Config{
		EntityType:       "task_outcome",
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

// NewAttachmentUploadAction creates the upload handler for task outcome attachments.
func NewAttachmentUploadAction(deps *DetailViewDeps) view.View {
	return attachment.NewUploadAction(attachmentConfig(deps))
}

// NewAttachmentDeleteAction creates the delete handler for task outcome attachments.
func NewAttachmentDeleteAction(deps *DetailViewDeps) view.View {
	return attachment.NewDeleteAction(attachmentConfig(deps))
}
