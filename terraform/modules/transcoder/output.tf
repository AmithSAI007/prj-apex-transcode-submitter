output "transcoder_job_template_name" {
  description = "The name of the Transcoder API job template to use for video ingestion."
  value       = data.google_transcoder_job_template.apex_ingestion_job_template.name
}
