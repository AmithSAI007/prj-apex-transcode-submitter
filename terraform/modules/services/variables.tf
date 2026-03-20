variable "project_id" {
  type        = string
  description = "The unique identifier for the GCP project for resource organization and billing."
  validation {
    condition     = length(var.project_id) > 0
    error_message = "The project_id must not be empty."
  }
}

variable "project_region" {
  type        = string
  description = "The GCP region where the resources will be deployed, impacting latency and compliance."
  validation {
    condition     = length(var.project_region) > 0
    error_message = "The project_region must be specified."
  }
}

variable "service_name" {
  description = "The name of the Cloud Run service."
  type        = string
  default     = "prj-apex-transcode-submitter"
}

variable "container_image" {
  description = "The container image to be used for the Cloud Run service."
  type        = string
}

variable "min_instance_count" {
  description = "The minimum number of instances for the Cloud Run service."
  type        = number
}

variable "max_instance_count" {
  description = "The maximum number of instances for the Cloud Run service."
  type        = number
}

variable "memory_limit" {
  description = "The memory limit for the Cloud Run service."
  type        = string
}

variable "cpu_limit" {
  description = "The CPU limit for the Cloud Run service."
  type        = string
}

variable "app_env" {
  description = "The application environment (e.g., development, staging, production)."
  type        = string
}

variable "http_port" {
  description = "The HTTP port for the Cloud Run service."
  type        = string
  default     = ":8080"
}

variable "gcs_bucket" {
  description = "The GCS bucket where uploaded objects are stored."
  type        = string
}

variable "output_gcs_bucket" {
  description = "The GCS bucket where transcoded outputs are stored."
  type        = string
}

variable "otel_service_name" {
  description = "The logical service name reported to the OpenTelemetry collector."
  type        = string
  default     = "prj-apex-transcode-submitter"
}

variable "otel_exporter_otlp_endpoint" {
  description = "The endpoint URL for the OpenTelemetry Protocol (OTLP) exporter to send telemetry data."
  type        = string
}

variable "max_file_size_bytes" {
  description = "The maximum file size in bytes that the service will accept for processing."
  type        = number
}

variable "min_file_size_bytes" {
  description = "The minimum file size in bytes that the service will accept for processing."
  type        = number
}

variable "allowed_content_types" {
  description = "A comma-separated list of allowed MIME content types for uploaded files."
  type        = string
}

variable "transcode_task_queue" {
  description = "The name of the Pub/Sub topic used as the task queue for transcoding jobs."
  type        = string
}

variable "firestore_collection" {
  description = "The name of the Firestore collection where transcoding job metadata will be stored."
  type        = string
  default     = "videos"
}

variable "transcode_template_id" {
  description = "The ID of the Transcoder API job template to use for transcoding tasks."
  type        = string
}

variable "firestore_database_id" {
  description = "The ID of the Firestore database to use for storing transcoding job metadata."
  type        = string
}

variable "otel_resource_attributes" {
  description = "A comma-separated list of key=value pairs to be included as resource attributes in OpenTelemetry telemetry data."
  type        = string
}

variable "service_account_name" {
  description = "The name of the service account to be used by the Cloud Run service."
  type        = string
}
