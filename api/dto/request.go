package dto

// TranscodeRequest is the payload delivered by Google Cloud Tasks when a new
// video transcode job needs to be processed. Every field is mandatory; the
// task will be rejected with HTTP 400 if any field is missing or malformed.
type TranscodeRequest struct {
	// EventID is the Cloud Tasks event identifier for deduplication and tracing.
	EventID string `json:"eventId" binding:"required" example:"evt_a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	// FileSize is the expected size of the raw media file in bytes, used for integrity verification against GCS metadata.
	FileSize int64 `json:"fileSize" binding:"required" example:"104857600"`
	// ContentType is the MIME type of the raw media file (e.g. video/mp4, video/mkv).
	ContentType string `json:"contentType" binding:"required" example:"video/mp4"`
	// TraceID is a client-supplied correlation identifier propagated through logs and downstream services.
	TraceID string `json:"traceId" binding:"required" example:"trace_7f3a9c2b1d4e8f0a"`
	// VideoID is the UUID of the video resource in Firestore. Must be a valid UUIDv4.
	VideoID string `json:"videoId" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// UserID is the UUID of the owning user. Must be a valid UUIDv4.
	UserID string `json:"userId" binding:"required" example:"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
	// RawFilePath is the fully-qualified GCS URI of the raw media file (gs://bucket/path/to/file).
	RawFilePath string `json:"rawFilePath" binding:"required" example:"gs://apex-raw-uploads/6ba7b810-9dad-11d1-80b4-00c04fd430c8/550e8400-e29b-41d4-a716-446655440000/video.mp4"`
}

// TranscodeResponse is returned on successful task acceptance (HTTP 200).
type TranscodeResponse struct {
	// Message is a human-readable confirmation that the task was accepted.
	Message string `json:"message" example:"Transcode task received successfully"`
}

// HealthResponse is returned by the health check endpoint.
type HealthResponse struct {
	// Status indicates the service health. Always "ok" when the service is running.
	Status string `json:"status" example:"ok"`
}
