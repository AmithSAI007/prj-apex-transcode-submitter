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

variable "app_env" {
  description = "The application environment (e.g., development, staging, production)."
  type        = string
  default     = "development"
}

variable "container_image" {
  description = "The container image to be used for the Cloud Run service."
  type        = string
}

variable "service_account_name" {
  description = "The service account name to be used by the Cloud Run service."
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

variable "otel_exporter_otlp_endpoint" {
  description = "The OTLP endpoint for OpenTelemetry exporter."
  type        = string
}

variable "max_file_size_bytes" {
  description = "The maximum file size in bytes for uploaded videos."
  type        = number
}

variable "min_file_size_bytes" {
  description = "The minimum file size in bytes for uploaded videos."
  type        = number
}

variable "allowed_content_types" {
  description = "A list of allowed content types for uploaded videos."
  type        = string
}

variable "firestore_database_id" {
  description = "The ID of the Firestore database to use for storing transcoding job metadata."
  type        = string
}

variable "otel_resource_attributes" {
  description = "A comma-separated list of key=value pairs to be added as resource attributes in OpenTelemetry telemetry data."
  type        = string
}

variable "otel_exporter_otlp_headers" {
  description = "A comma-separated list of key=value pairs to be included as headers in OpenTelemetry OTLP exporter requests."
  type        = string
}
