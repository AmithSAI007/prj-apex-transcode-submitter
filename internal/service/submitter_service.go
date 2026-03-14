package service

import (
	"go.uber.org/zap"
)

type SubmitterService struct {
	logger *zap.Logger
}

func NewSubmitterService(logger *zap.Logger) *SubmitterService {
	return &SubmitterService{
		logger: logger,
	}
}
