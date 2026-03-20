module "storage" {
  source = "../../modules/storage"
}

module "service" {
  source                      = "../../modules/services"
  project_id                  = var.project_id
  project_region              = var.project_region
  container_image             = var.container_image
  min_instance_count          = var.min_instance_count
  max_instance_count          = var.max_instance_count
  memory_limit                = var.memory_limit
  cpu_limit                   = var.cpu_limit
  gcs_bucket                  = module.storage.raw_videos_bucket_name
  output_gcs_bucket           = module.storage.transcoded_videos_bucket_name
  app_env                     = var.app_env
  otel_exporter_otlp_endpoint = var.otel_exporter_otlp_endpoint
  max_file_size_bytes         = var.max_file_size_bytes
  min_file_size_bytes         = var.min_file_size_bytes
  allowed_content_types       = var.allowed_content_types
  transcode_task_queue        = var.transcode_task_queue
  transcode_template_id       = var.transcode_template_id
  firestore_database_id       = var.firestore_database_id
  otel_resource_attributes    = var.otel_resource_attributes
  service_account_name        = var.service_account_name
  otel_exporter_otlp_headers  = var.otel_exporter_otlp_headers
}
