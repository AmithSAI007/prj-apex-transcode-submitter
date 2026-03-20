data "google_storage_bucket" "apex_dev_gcs_raw_videos" {
  name = var.raw_videos_bucket_name
}

data "google_storage_bucket" "apex_dev_gcs_transcoded_videos" {
  name = var.transcoded_videos_bucket_name
}
