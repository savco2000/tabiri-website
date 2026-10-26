resource "google_secret_manager_secret" "turnstile_secret_key" {
  secret_id = "tabiri-turnstile-secret"

  replication {
    auto {}
  }

  depends_on = [google_project_service.required]
}

resource "google_secret_manager_secret" "form_hmac_key" {
  secret_id = "tabiri-form-hmac-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.required]
}

resource "google_secret_manager_secret_iam_member" "turnstile_reader" {
  secret_id = google_secret_manager_secret.turnstile_secret_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.web.member
}

resource "google_secret_manager_secret_iam_member" "form_hmac_reader" {
  secret_id = google_secret_manager_secret.form_hmac_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.web.member
}
