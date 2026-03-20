package repository

import (
	"context"

	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/model"
)

type GCSRepository interface {
	HeadObject(ctx context.Context, bucket, object string) (*model.ObjectMetadata, error)
}

type FirestoreRepository interface {
	GetStatus(ctx context.Context, videoID string) (string, error)
	TransitionStatus(ctx context.Context, videoID, fromStatus, toStatus string, updates map[string]any) error
}

type TranscoderRepository interface {
	SubmitJob(ctx context.Context, input *model.TranscodeInput) (string, error)
}
