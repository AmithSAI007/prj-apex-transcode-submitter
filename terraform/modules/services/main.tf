resource "google_cloud_run_v2_service" "apex_ingestion_service" {
  name     = var.service_name
  location = var.project_region

  scaling {
    min_instance_count = var.min_instance_count
    max_instance_count = var.max_instance_count
  }



  template {

    service_account = var.service_account_name
    containers {
      image = var.container_image

      resources {
        limits = {
          memory = var.memory_limit
          cpu    = var.cpu_limit
        }
      }

      env {
        name  = "APP_ENV"
        value = var.app_env
      }
      env {
        name  = "GCP_PROJECT_REGION"
        value = var.project_region
      }
      env {
        name  = "HTTP_PORT"
        value = var.http_port
      }
      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }
      env {
        name  = "GCS_BUCKET"
        value = var.gcs_bucket
      }
      env {
        name  = "OUTPUT_GCS_BUCKET"
        value = var.output_gcs_bucket
      }
      env {
        name  = "OTEL_SERVICE_NAME"
        value = var.otel_service_name
      }
      env {
        name  = "OTEL_EXPORTER_OTLP_ENDPOINT"
        value = var.otel_exporter_otlp_endpoint
      }
      env {
        name  = "OTEL_EXPORTER_OLTP_HEADERS"
        value = var.otel_exporter_otlp_headers
      }
      env {
        name  = "MAX_FILE_SIZE_BYTES"
        value = var.max_file_size_bytes
      }
      env {
        name  = "MIN_FILE_SIZE_BYTES"
        value = var.min_file_size_bytes
      }
      env {
        name  = "ALLOWED_CONTENT_TYPES"
        value = var.allowed_content_types
      }
      env {
        name  = "TRANSCODE_TASK_QUEUE"
        value = var.transcode_task_queue
      }
      env {
        name  = "FIRESTORE_COLLECTION"
        value = var.firestore_collection
      }
      env {
        name  = "TRANSCODE_TEMPLATE_ID"
        value = var.transcode_template_id
      }
      env {
        name  = "FIRESTORE_DATABASE_ID"
        value = var.firestore_database_id
      }
      env {
        name  = "OTEL_RESOURCE_ATTRIBUTES"
        value = var.otel_resource_attributes
      }
    }
  }

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }
}
