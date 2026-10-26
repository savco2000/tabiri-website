output "state_bucket_name" {
  description = "Bucket to use for the production Terraform backend."
  value       = google_storage_bucket.terraform_state.name
}
