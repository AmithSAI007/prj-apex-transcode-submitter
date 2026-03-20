package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/constants"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrDocumentNotFound  = errors.New("document not found")
	ErrInvalidTransition = errors.New("invalid status transition")
)

const (
	StatusField            = "status"
	UpdatedAtField         = "updatedAt"
	TranscoderJobNameField = "transcoderJobName"
)

type FirestoreRepo struct {
	logger     *zap.Logger
	client     *firestore.Client
	collection string
}

func NewFirestoreRepo(logger *zap.Logger, client *firestore.Client, collection string) *FirestoreRepo {
	return &FirestoreRepo{
		logger:     logger,
		client:     client,
		collection: collection,
	}
}

func (r *FirestoreRepo) GetStatus(ctx context.Context, videoID string) (string, error) {
	ctx, span := utils.Tracer().Start(ctx, "firestore.GetStatus",
		trace.WithAttributes(
			attribute.String(constants.LogKeyVideoID, videoID),
			attribute.String("firestore.collection", r.collection),
		),
	)
	defer span.End()

	start := time.Now()
	logFields := append(utils.LogFieldsFromContext(ctx),
		zap.String(constants.LogKeyLayer, "repository"),
		zap.String(constants.LogKeyMethod, "GetStatus"),
		zap.String(constants.LogKeyVideoID, videoID),
	)

	r.logger.Debug("reading document status from Firestore", logFields...)

	docRef := r.client.Collection(r.collection).Doc(videoID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		elapsed := time.Since(start).Milliseconds()
		if status.Code(err) == grpccodes.NotFound {
			span.SetStatus(codes.Error, "document not found")
			span.RecordError(ErrDocumentNotFound)
			r.logger.Warn("Firestore document not found",
				append(logFields, zap.Int64(constants.LogKeyDuration, elapsed))...,
			)
			return "", ErrDocumentNotFound
		}
		span.SetStatus(codes.Error, "firestore get failed")
		span.RecordError(err)
		r.logger.Error("failed to read Firestore document",
			append(logFields,
				zap.Error(err),
				zap.Int64(constants.LogKeyDuration, elapsed),
			)...,
		)
		return "", fmt.Errorf("firestore get document %s: %w", videoID, err)
	}

	rawStatus, err := docSnap.DataAt(StatusField)
	if err != nil {
		span.SetStatus(codes.Error, "missing status field")
		span.RecordError(err)
		r.logger.Error("Firestore document missing status field",
			append(logFields, zap.Error(err))...,
		)
		return "", fmt.Errorf("firestore get status field for %s: %w", videoID, err)
	}

	statusStr, ok := rawStatus.(string)
	if !ok {
		err := fmt.Errorf("firestore status field for %s: expected string, got: %T", videoID, rawStatus)
		span.SetStatus(codes.Error, "invalid status type")
		span.RecordError(err)
		r.logger.Error("Firestore status field has unexpected type",
			append(logFields, zap.Any("raw_status", rawStatus))...,
		)
		return "", err
	}

	elapsed := time.Since(start).Milliseconds()
	span.SetAttributes(attribute.String(constants.LogKeyStatus, statusStr))
	span.AddEvent("firestore.status.read")

	r.logger.Info("Firestore document status retrieved",
		append(logFields,
			zap.String(constants.LogKeyStatus, statusStr),
			zap.Int64(constants.LogKeyDuration, elapsed),
		)...,
	)

	return statusStr, nil
}

func (r *FirestoreRepo) TransitionStatus(ctx context.Context, videoID, fromStatus, toStatus string, updates map[string]any) error {
	ctx, span := utils.Tracer().Start(ctx, "firestore.TransitionStatus",
		trace.WithAttributes(
			attribute.String(constants.LogKeyVideoID, videoID),
			attribute.String("firestore.collection", r.collection),
			attribute.String("firestore.from_status", fromStatus),
			attribute.String("firestore.to_status", toStatus),
		),
	)
	defer span.End()

	start := time.Now()
	logFields := append(utils.LogFieldsFromContext(ctx),
		zap.String(constants.LogKeyLayer, "repository"),
		zap.String(constants.LogKeyMethod, "TransitionStatus"),
		zap.String(constants.LogKeyVideoID, videoID),
		zap.String("from_status", fromStatus),
		zap.String("to_status", toStatus),
	)

	r.logger.Debug("beginning Firestore status transition", logFields...)

	docRef := r.client.Collection(r.collection).Doc(videoID)
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docSnap, err := tx.Get(docRef)
		if err != nil {
			if status.Code(err) == grpccodes.NotFound {
				return ErrDocumentNotFound
			}
			return fmt.Errorf("firestore transaction read %s: %w", videoID, err)
		}

		rawStatus, err := docSnap.DataAt(StatusField)
		if err != nil {
			return fmt.Errorf("firestore get status field for %s: %w", videoID, err)
		}

		currentStatus, ok := rawStatus.(string)
		if !ok {
			return fmt.Errorf("firestore status field for %s: expected string, got: %T", videoID, rawStatus)
		}

		if currentStatus != fromStatus {
			return ErrInvalidTransition
		}

		firestoreUpdates := []firestore.Update{
			{Path: StatusField, Value: toStatus},
			{Path: UpdatedAtField, Value: updates[UpdatedAtField]},
			{Path: TranscoderJobNameField, Value: updates[TranscoderJobNameField]},
		}

		return tx.Update(docRef, firestoreUpdates)
	})

	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		span.SetStatus(codes.Error, "transition failed")
		span.RecordError(err)
		r.logger.Error("Firestore status transition failed",
			append(logFields,
				zap.Error(err),
				zap.Int64(constants.LogKeyDuration, elapsed),
			)...,
		)
		return err
	}

	span.AddEvent("firestore.status.transitioned",
		trace.WithAttributes(
			attribute.String("from", fromStatus),
			attribute.String("to", toStatus),
		),
	)

	r.logger.Info("Firestore status transitioned successfully",
		append(logFields, zap.Int64(constants.LogKeyDuration, elapsed))...,
	)

	return nil
}

var _ FirestoreRepository = (*FirestoreRepo)(nil)
