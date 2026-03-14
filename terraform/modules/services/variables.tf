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

variable "http_port" {
  description = "The HTTP port for the Cloud Run service."
  type        = string
  default     = ":8080"
}

variable "gcs_bucket" {
  description = "The GCS bucket where uploaded objects are stored."
  type        = string
}

variable "otel_service_name" {
  description = "The logical service name reported to the OpenTelemetry collector."
  type        = string
  default     = "prj-apex-transcode-submitter"
}
