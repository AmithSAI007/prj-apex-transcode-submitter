package model

import "time"

const (
	StatusPendingUpload = "PENDING_UPLOAD"
	StatusQueued        = "QUEUED"
	StatusTranscoding   = "TRANSCODING"
	StatusCompleted     = "COMPLETED"
	StatusFailed        = "FAILED"
)

type VideoDocument struct {
	VideoID           string    `firestore:"videoId"`
	UserID            string    `firestore:"userId"`
	Status            string    `firestore:"status"`
	RawFilePath       string    `firestore:"rawFilePath"`
	FileSize          int64     `firestore:"fileSize"`
	ContentType       string    `firestore:"contentType"`
	FileHash          string    `firestore:"fileHash"`
	TranscoderJobName string    `firestore:"transcoderJobName,omitempty"`
	CreatedAt         time.Time `firestore:"createdAt"`
	UpdatedAt         time.Time `firestore:"updatedAt"`
}

type ObjectMetadata struct {
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

type TranscodeInput struct {
	VideoID     string `json:"videoId" binding:"required"`
	UserID      string `json:"userId" binding:"required"`
	RawFilePath string `json:"rawFilePath" binding:"required"`
	OutputURI   string `json:"outputUri" binding:"required"`
}
