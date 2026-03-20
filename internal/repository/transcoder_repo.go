package repository

import (
	"context"
	"fmt"
	"time"

	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	"cloud.google.com/go/video/transcoder/apiv1/transcoderpb"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/internal/model"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/constants"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TranscoderService struct {
	logger     *zap.Logger
	client     *transcoder.Client
	projectID  string
	location   string
	templateID string
}

func NewTranscoderService(logger *zap.Logger, client *transcoder.Client, projectID, location, templateID string) *TranscoderService {
	return &TranscoderService{
		logger:     logger,
		client:     client,
		projectID:  projectID,
		location:   location,
		templateID: templateID,
	}
}

func (s *TranscoderService) SubmitJob(ctx context.Context, input *model.TranscodeInput) (string, error) {
	ctx, span := utils.Tracer().Start(ctx, "transcoder.SubmitJob",
		trace.WithAttributes(
			attribute.String(constants.LogKeyVideoID, input.VideoID),
			attribute.String(constants.LogKeyUserID, input.UserID),
			attribute.String("transcoder.template_id", s.templateID),
		),
	)
	defer span.End()

	start := time.Now()
	logFields := append(utils.LogFieldsFromContext(ctx),
		zap.String(constants.LogKeyLayer, "repository"),
		zap.String(constants.LogKeyMethod, "SubmitJob"),
		zap.String(constants.LogKeyVideoID, input.VideoID),
		zap.String(constants.LogKeyUserID, input.UserID),
	)

	jobName := "projects/" + s.projectID + "/locations/" + s.location + "/jobs/" + input.VideoID
	span.SetAttributes(attribute.String(constants.LogKeyJobName, jobName))

	s.logger.Info("submitting transcoder job", append(logFields, zap.String(constants.LogKeyJobName, jobName))...)

	req := &transcoderpb.CreateJobRequest{
		Parent: "projects/" + s.projectID + "/locations/" + s.location,
		Job: &transcoderpb.Job{
			Name:      jobName,
			InputUri:  input.RawFilePath,
			OutputUri: input.OutputURI,
			JobConfig: &transcoderpb.Job_TemplateId{
				TemplateId: s.templateID,
			},
		},
	}

	job, err := s.client.CreateJob(ctx, req)
	if err != nil {
		elapsed := time.Since(start).Milliseconds()
		if status.Code(err) == grpccodes.AlreadyExists {
			span.AddEvent("transcoder.job.already_exists")
			s.logger.Info("transcoder job already exists (idempotent)",
				append(logFields,
					zap.String(constants.LogKeyJobName, jobName),
					zap.Int64(constants.LogKeyDuration, elapsed),
				)...,
			)
			return jobName, nil
		}
		span.SetStatus(codes.Error, "transcoder submit failed")
		span.RecordError(err)
		s.logger.Error("failed to submit transcoder job",
			append(logFields,
				zap.Error(err),
				zap.String(constants.LogKeyJobName, jobName),
				zap.Int64(constants.LogKeyDuration, elapsed),
			)...,
		)
		return "", fmt.Errorf("transcoder submit job error: %w", err)
	}

	elapsed := time.Since(start).Milliseconds()
	span.SetAttributes(attribute.String(constants.LogKeyJobName, job.Name))
	span.AddEvent("transcoder.job.created")

	s.logger.Info("transcoder job submitted successfully",
		append(logFields,
			zap.String(constants.LogKeyJobName, job.Name),
			zap.Int64(constants.LogKeyDuration, elapsed),
		)...,
	)

	return job.Name, nil
}

var _ TranscoderRepository = (*TranscoderService)(nil)
