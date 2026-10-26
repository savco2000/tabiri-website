resource "google_compute_instance" "web" {
  count = var.deploy_application ? 1 : 0

  name                      = "tabiri-web-1"
  zone                      = var.zone
  machine_type              = var.machine_type
  allow_stopping_for_update = true
  deletion_protection       = true
  tags                      = [local.network_tag]

  boot_disk {
    initialize_params {
      image = "projects/ubuntu-os-cloud/global/images/family/ubuntu-2404-lts-amd64"
      size  = 20
      type  = "pd-balanced"
    }
  }

  attached_disk {
    source      = google_compute_disk.data.id
    device_name = google_compute_disk.data.name
    mode        = "READ_WRITE"
  }

  network_interface {
    subnetwork = google_compute_subnetwork.web.id

    access_config {
      network_tier = "PREMIUM"
    }
  }

  service_account {
    email  = google_service_account.web.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    block-project-ssh-keys = "TRUE"
    enable-oslogin         = "TRUE"
  }

  metadata_startup_script = templatefile("${path.module}/templates/startup.sh.tftpl", {
    project_id                  = var.project_id
    release_bucket_name         = google_storage_bucket.releases.name
    release_object              = var.release_object
    release_sha256              = var.release_sha256
    turnstile_site_key          = var.turnstile_site_key
    turnstile_allowed_hostnames = var.turnstile_allowed_hostnames
  })

  shielded_instance_config {
    enable_integrity_monitoring = true
    enable_secure_boot          = true
    enable_vtpm                 = true
  }

  lifecycle {
    precondition {
      condition     = trimspace(var.turnstile_site_key) != ""
      error_message = "turnstile_site_key must be set when deploy_application is true."
    }

    precondition {
      condition     = can(regex("^[0-9a-fA-F]{64}$", var.release_sha256))
      error_message = "release_sha256 must be a 64-character SHA-256 checksum when deploy_application is true."
    }
  }

  depends_on = [
    google_secret_manager_secret_iam_member.form_hmac_reader,
    google_secret_manager_secret_iam_member.turnstile_reader,
    google_storage_bucket_iam_member.release_reader,
  ]
}

resource "google_compute_instance_group" "web" {
  count = var.deploy_application ? 1 : 0

  name      = "tabiri-web-instance-group"
  zone      = var.zone
  instances = [google_compute_instance.web[0].id]

  named_port {
    name = "http"
    port = 4000
  }
}
