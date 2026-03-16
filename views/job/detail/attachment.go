package detail

import (
	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/pyeza-golang/view"
)

func attachmentConfig(deps *Deps) *attachment.Config {
	return &attachment.Config{
		EntityType:       "job",
		BucketName:       "attachments",
		UploadURL:        deps.Routes.AttachmentUploadURL,
		DeleteURL:        deps.Routes.AttachmentDeleteURL,
		Labels:           attachment.DefaultLabels(),
		CommonLabels:     deps.CommonLabels,
		TableLabels:      deps.TableLabels,
		NewID:            deps.NewID,
		UploadFile:       deps.UploadFile,
		ListAttachments:  deps.ListAttachments,
		CreateAttachment: deps.CreateAttachment,
		DeleteAttachment: deps.DeleteAttachment,
	}
}

// NewAttachmentUploadAction creates the upload handler for job attachments.
func NewAttachmentUploadAction(deps *Deps) view.View {
	return attachment.NewUploadAction(attachmentConfig(deps))
}

// NewAttachmentDeleteAction creates the delete handler for job attachments.
func NewAttachmentDeleteAction(deps *Deps) view.View {
	return attachment.NewDeleteAction(attachmentConfig(deps))
}
