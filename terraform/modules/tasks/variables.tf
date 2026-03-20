variable "transcoder_queue_name" {
  description = "The name of the Cloud Tasks queue for transcoding tasks."
  type        = string
  default     = "apex-transcoder-service-queue"
}
