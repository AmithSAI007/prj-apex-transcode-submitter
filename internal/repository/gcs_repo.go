package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/model"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/constants"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type GCSRepositoryService struct {
	logger *zap.Logger
	client *storage.Client
}

func NewGCSRepositoryService(logger *zap.Logger, client *storage.Client) *GCSRepositoryService {
	return &GCSRepositoryService{
		logger: logger,
		client: client,
	}
}

var ErrNotFound = errors.New("object not found in GCS")

func (r *GCSRepositoryService) HeadObject(ctx context.Context, bucket, object string) (*model.ObjectMetadata, error) {
	ctx, span := utils.Tracer().Start(ctx, "gcs.HeadObject",
		trace.WithAttributes(
			attribute.String(constants.LogKeyBucket, bucket),
			attribute.String(constants.LogKeyObject, object),
		),
	)
	defer span.End()

	start := time.Now()
	logFields := append(utils.LogFieldsFromContext(ctx),
		zap.String(constants.LogKeyLayer, "repository"),
		zap.String(constants.LogKeyMethod, "HeadObject"),
		zap.String(constants.LogKeyBucket, bucket),
		zap.String(constants.LogKeyObject, object),
	)

	r.logger.Debug("checking GCS object existence", logFields...)

	storageObject := r.client.Bucket(bucket).Object(object)

	attrs, err := storageObject.Attrs(ctx)
	if err != nil {
		elapsed := time.Since(start).Milliseconds()
		if err == storage.ErrObjectNotExist {
			span.SetStatus(codes.Error, "object not found")
			span.RecordError(ErrNotFound)
			r.logger.Warn("GCS object does not exist",
				append(logFields, zap.Int64(constants.LogKeyDuration, elapsed))...,
			)
			return nil, ErrNotFound
		}
		span.SetStatus(codes.Error, "gcs attrs failed")
		span.RecordError(err)
		r.logger.Error("failed to retrieve GCS object attributes",
			append(logFields,
				zap.Error(err),
				zap.Int64(constants.LogKeyDuration, elapsed),
			)...,
		)
		return nil, fmt.Errorf("gcs head object error: %w", err)
	}

	elapsed := time.Since(start).Milliseconds()
	span.SetAttributes(
		attribute.Int64("gcs.object.size", attrs.Size),
		attribute.String("gcs.object.content_type", attrs.ContentType),
	)
	span.AddEvent("gcs.object.resolved")

	r.logger.Info("GCS object verified",
		append(logFields,
			zap.Int64("file_size", attrs.Size),
			zap.String("content_type", attrs.ContentType),
			zap.Int64(constants.LogKeyDuration, elapsed),
		)...,
	)

	metadata := &model.ObjectMetadata{
		ContentType: attrs.ContentType,
		Size:        attrs.Size,
	}
	return metadata, nil
}

var _ GCSRepository = (*GCSRepositoryService)(nil)
