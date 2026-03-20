data "google_cloud_tasks_queue" "transcode_task_queue" {
  name = var.transcoder_queue_name
}
