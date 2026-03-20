package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/api/dto"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/config"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/model"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/repository"
	validator "github.com/AmithSAI007/prj-apex-transcode-submitter/internal/validation"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/constants"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type processContext struct {
	bucket  string
	object  string
	jobName string
}

type SubmitterInterface interface {
	ProcessTranscodeRequest(ctx context.Context, req *dto.TranscodeRequest) error
	validatePayload(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error
	verifyRawFile(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error
	checkStatus(ctx context.Context, videoID string) error
	submitTranscoderJob(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error
	transitionToTranscoding(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error
}

type SubmitterService struct {
	logger         *zap.Logger
	cfg            config.Config
	gcsRepo        repository.GCSRepository
	firestoreRepo  repository.FirestoreRepository
	transcoderRepo repository.TranscoderRepository
}

const (
	VideoID     = "videoID"
	UserID      = "userID"
	RawFilePath = "rawFilePath"
	ContentType = "contentType"
)

func NewSubmitterService(logger *zap.Logger, cfg config.Config, gcsRepo repository.GCSRepository, firestoreRepo repository.FirestoreRepository, transcoderRepo repository.TranscoderRepository) *SubmitterService {
	return &SubmitterService{
		logger:         logger,
		cfg:            cfg,
		gcsRepo:        gcsRepo,
		firestoreRepo:  firestoreRepo,
		transcoderRepo: transcoderRepo,
	}
}

func (s *SubmitterService) ProcessTranscodeRequest(ctx context.Context, req *dto.TranscodeRequest) error {
	ctx, span := utils.Tracer().Start(ctx, "service.ProcessTranscodeRequest",
		trace.WithAttributes(
			attribute.String(constants.LogKeyVideoID, req.VideoID),
			attribute.String(constants.LogKeyUserID, req.UserID),
			attribute.String("event_id", req.EventID),
		),
	)
	defer span.End()

	start := time.Now()
	logFields := append(utils.LogFieldsFromContext(ctx),
		zap.String(constants.LogKeyLayer, "service"),
		zap.String(constants.LogKeyMethod, "ProcessTranscodeRequest"),
		zap.String(constants.LogKeyVideoID, req.VideoID),
		zap.String(constants.LogKeyUserID, req.UserID),
		zap.String("event_id", req.EventID),
	)

	s.logger.Info("processing transcode request", logFields...)

	pc := &processContext{}

	// Step 1: Validate payload
	if err := s.validatePayload(ctx, req, pc); err != nil {
		span.SetStatus(codes.Error, "validation failed")
		span.RecordError(err)
		s.logger.Warn("payload validation failed",
			append(logFields, zap.Error(err))...,
		)
		return err
	}
	span.AddEvent("payload.validated")
	s.logger.Debug("payload validation passed", logFields...)

	// Step 2: Verify raw file in GCS
	if err := s.verifyRawFile(ctx, req, pc); err != nil {
		span.SetStatus(codes.Error, "raw file verification failed")
		span.RecordError(err)
		s.logger.Warn("raw file verification failed",
			append(logFields,
				zap.Error(err),
				zap.String(constants.LogKeyBucket, pc.bucket),
				zap.String(constants.LogKeyObject, pc.object),
			)...,
		)
		return err
	}
	span.AddEvent("raw_file.verified")
	s.logger.Debug("raw file verified in GCS", logFields...)

	// Step 3: Check Firestore document status
	if err := s.checkStatus(ctx, req.VideoID); err != nil {
		span.RecordError(err)
		// Log level depends on error type: idempotency is informational, not an error.
		var idempErr *IdempotencyError
		if errors.As(err, &idempErr) {
			span.SetStatus(codes.Ok, "idempotent duplicate")
			span.AddEvent("idempotency.duplicate_detected")
			s.logger.Info("idempotent duplicate detected",
				append(logFields, zap.Error(err))...,
			)
		} else {
			span.SetStatus(codes.Error, "status check failed")
			s.logger.Warn("status check failed",
				append(logFields, zap.Error(err))...,
			)
		}
		return err
	}
	span.AddEvent("status.verified", trace.WithAttributes(attribute.String(constants.LogKeyStatus, model.StatusQueued)))
	s.logger.Debug("document status verified as QUEUED", logFields...)

	// Step 4: Submit transcoder job
	if err := s.submitTranscoderJob(ctx, req, pc); err != nil {
		span.SetStatus(codes.Error, "transcoder submission failed")
		span.RecordError(err)
		s.logger.Error("transcoder job submission failed",
			append(logFields, zap.Error(err))...,
		)
		return err
	}
	span.AddEvent("transcoder.job_submitted", trace.WithAttributes(attribute.String(constants.LogKeyJobName, pc.jobName)))
	s.logger.Info("transcoder job submitted",
		append(logFields, zap.String(constants.LogKeyJobName, pc.jobName))...,
	)

	// Step 5: Transition Firestore status
	if err := s.transitionToTranscoding(ctx, req, pc); err != nil {
		span.SetStatus(codes.Error, "status transition failed")
		span.RecordError(err)
		s.logger.Error("Firestore status transition failed",
			append(logFields, zap.Error(err))...,
		)
		return err
	}
	span.AddEvent("status.transitioned",
		trace.WithAttributes(
			attribute.String("from", model.StatusQueued),
			attribute.String("to", model.StatusTranscoding),
		),
	)

	elapsed := time.Since(start).Milliseconds()
	s.logger.Info("transcode request processed successfully",
		append(logFields,
			zap.String(constants.LogKeyJobName, pc.jobName),
			zap.Int64(constants.LogKeyDuration, elapsed),
		)...,
	)

	return nil
}

func (s *SubmitterService) validatePayload(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error {
	_, span := utils.Tracer().Start(ctx, "service.validatePayload")
	defer span.End()

	if err := validator.ValidateStringLength(VideoID, req.VideoID, 36); err != nil {
		return NewPermanentError(err)
	}
	if err := validator.ValidateStringLength(UserID, req.UserID, 36); err != nil {
		return NewPermanentError(err)
	}
	if err := validator.ValidateStringLength(RawFilePath, req.RawFilePath, 1024); err != nil {
		return NewPermanentError(err)
	}
	if err := validator.ValidateStringLength(ContentType, req.ContentType, 256); err != nil {
		return NewPermanentError(err)
	}

	if !validator.ValidateUUID(req.VideoID) {
		return NewPermanentError(fmt.Errorf("%s: %w", "videoID", validator.ErrInvalidUUID))
	}
	if !validator.ValidateUUID(req.UserID) {
		return NewPermanentError(fmt.Errorf("%s: %w", "userID", validator.ErrInvalidUUID))
	}

	if err := validator.ValidateNoPathInjection(req.VideoID); err != nil {
		return NewPermanentError(err)
	}

	if err := validator.ValidateNoPathInjection(req.UserID); err != nil {
		return NewPermanentError(err)
	}

	bucket, object, err := validator.ValidateGCSURI(req.RawFilePath)
	if err != nil {
		return NewPermanentError(err)
	}

	if err := validator.ValidateGCSBucket(bucket, s.cfg.GCSBucket); err != nil {
		return NewPermanentError(err)
	}

	if err := validator.ValidatePathConsistency(object, req.UserID, req.VideoID); err != nil {
		return NewPermanentError(err)
	}

	if err := validator.ValidateFileSize(req.FileSize, s.cfg.MinFileSizeBytes, s.cfg.MaxFileSizeBytes); err != nil {
		return NewPermanentError(err)
	}

	allowedTypes := strings.Split(s.cfg.AllowedContentTypes, ",")
	if err := validator.ValidateContentType(req.ContentType, allowedTypes); err != nil {
		return NewPermanentError(err)
	}

	pc.bucket = bucket
	pc.object = object

	return nil
}

func (s *SubmitterService) verifyRawFile(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error {
	metadata, err := s.gcsRepo.HeadObject(ctx, pc.bucket, pc.object)
	if err != nil {
		return classifyError(fmt.Sprintf("gcs head object for %s", req.VideoID), err)
	}
	if metadata.Size != req.FileSize {
		return NewPermanentError(validator.ErrSizeMismatch)
	}
	return nil
}

func (s *SubmitterService) checkStatus(ctx context.Context, videoID string) error {
	currentStatus, err := s.firestoreRepo.GetStatus(ctx, videoID)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			return NewPermanentError(fmt.Errorf("video document not found: %s: %w", videoID, err))
		}

		return classifyError(fmt.Sprintf("firestore read failed for: %s", videoID), err)

	}

	switch currentStatus {
	case model.StatusQueued:
		return nil
	case model.StatusTranscoding, model.StatusCompleted, model.StatusFailed:
		return NewIdempotencyError(fmt.Errorf("video %s already %s", videoID, currentStatus))
	case model.StatusPendingUpload:
		return NewPermanentError(fmt.Errorf("video %s is pending upload", videoID))
	default:
		return NewPermanentError(fmt.Errorf("video %s unexpected status: %s", videoID, currentStatus))

	}
}

func (s *SubmitterService) submitTranscoderJob(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error {

	input := model.TranscodeInput{
		VideoID:     req.VideoID,
		UserID:      req.UserID,
		RawFilePath: req.RawFilePath,
		OutputURI:   fmt.Sprintf("gs://%s/%s/%s/output/", s.cfg.OutputGCSBucket, req.UserID, req.VideoID),
	}

	jobName, err := s.transcoderRepo.SubmitJob(ctx, &input)
	if err != nil {
		return classifyError(fmt.Sprintf("transcoder submit for %s", req.VideoID), err)
	}

	pc.jobName = jobName
	return nil
}

func (s *SubmitterService) transitionToTranscoding(ctx context.Context, req *dto.TranscodeRequest, pc *processContext) error {
	updates := map[string]any{
		repository.TranscoderJobNameField: pc.jobName,
		repository.UpdatedAtField:         firestore.ServerTimestamp,
	}
	err := s.firestoreRepo.TransitionStatus(ctx, req.VideoID, model.StatusQueued, model.StatusTranscoding, updates)
	if err != nil {
		return classifyError(fmt.Sprintf("firestore transition to transcoding for %s", req.VideoID), err)
	}
	return nil
}

var _ SubmitterInterface = (*SubmitterService)(nil)
