variable "raw_videos_bucket_name" {
  description = "The name of the Cloud Storage bucket to monitor for new video uploads."
  type        = string
  default     = "apex-dev-gcs-raw-videos"
}

variable "transcoded_videos_bucket_name" {
  description = "The name of the Cloud Storage bucket where transcoded videos will be stored."
  type        = string
  default     = "apex-dev-gcs-processed-videos"
}
