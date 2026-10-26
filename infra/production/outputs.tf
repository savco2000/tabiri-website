output "application_backend_service" {
  description = "Backend service to attach to the existing URL map after cutover approval."
  value       = google_compute_backend_service.web.self_link
}

output "data_disk" {
  description = "Persistent disk that stores the SQLite database."
  value       = google_compute_disk.data.self_link
}

output "release_bucket" {
  description = "Private bucket for versioned application releases."
  value       = google_storage_bucket.releases.name
}

output "secret_names" {
  description = "Secret Manager containers that require versions before application deployment."
  value = {
    form_hmac_key        = google_secret_manager_secret.form_hmac_key.secret_id
    turnstile_secret_key = google_secret_manager_secret.turnstile_secret_key.secret_id
  }
}

output "existing_routing" {
  description = "Production edge resources managed by Terraform."
  value = {
    backend_bucket = data.google_compute_backend_bucket.existing_static_site.self_link
    https_proxy    = google_compute_target_https_proxy.production.self_link
    url_map        = google_compute_url_map.production.self_link
  }
}
