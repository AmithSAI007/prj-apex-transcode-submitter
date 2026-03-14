package dto

type TranscodeRequest struct {
	EventID     string `json:"eventId"`
	TraceID     string `json:"traceId"`
	VideoID     string `json:"videoId" example:"vid_12345"`
	UserID      string `json:"userId" example:"user_67890"`
	RawFilePath string `json:"rawFilePath" example:"gs://bucket_name/path/to/video.mp4"`
}

type TranscodeResponse struct {
	Message string `json:"message" example:"Transcode task received successfully"`
}
