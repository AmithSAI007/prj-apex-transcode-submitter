package platform

import (
	"context"

	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	"go.uber.org/zap"
)

type TranscoderClient struct {
	logger *zap.Logger
	client *transcoder.Client
}

func NewTranscoderClient(ctx context.Context, logger *zap.Logger) (*TranscoderClient, error) {
	client, err := transcoder.NewClient(ctx)
	if err != nil {
		logger.Error("Failed to create Transcoder client", zap.Error(err))
		return nil, err
	}

	return &TranscoderClient{
		logger: logger,
		client: client,
	}, nil
}

func (tc *TranscoderClient) Client() *transcoder.Client {
	return tc.client
}

func (tc *TranscoderClient) Close() error {
	if tc == nil || tc.client == nil {
		return nil
	}
	return tc.client.Close()
}
