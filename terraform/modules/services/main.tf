resource "google_cloud_run_v2_service" "apex_ingestion_service" {
  name     = var.service_name
  location = var.project_region

  scaling {
    min_instance_count = var.min_instance_count
    max_instance_count = var.max_instance_count
  }

  template {
    containers {
      image = var.container_image

      resources {
        limits = {
          memory = var.memory_limit
          cpu    = var.cpu_limit
        }
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
        name  = "GCP_PROJECT_REGION"
        value = var.project_region
      }
      env {
        name  = "GCS_BUCKET"
        value = var.gcs_bucket
      }
      env {
        name  = "OTEL_SERVICE_NAME"
        value = var.otel_service_name
      }
    }
  }

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }
}
