module "storage" {
  source = "../../modules/storage"
}


module "service" {
  source             = "../../modules/services"
  project_id         = var.project_id
  project_region     = var.project_region
  container_image    = var.container_image
  min_instance_count = var.min_instance_count
  max_instance_count = var.max_instance_count
  memory_limit       = var.memory_limit
  cpu_limit          = var.cpu_limit
  gcs_bucket         = module.storage.raw_videos_bucket_name
}
