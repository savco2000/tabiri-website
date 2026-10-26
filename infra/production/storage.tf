resource "google_storage_bucket" "releases" {
  name                        = var.release_bucket_name
  location                    = upper(var.region)
  force_destroy               = false
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  versioning {
    enabled = true
  }
}

resource "google_storage_bucket_iam_member" "release_reader" {
  bucket = google_storage_bucket.releases.name
  role   = "roles/storage.objectViewer"
  member = google_service_account.web.member
}

resource "google_compute_disk" "data" {
  name = "tabiri-web-data"
  type = "pd-balanced"
  zone = var.zone
  size = 10

  lifecycle {
    prevent_destroy = true
  }

  depends_on = [google_project_service.required]
}

resource "google_compute_resource_policy" "daily_snapshots" {
  name   = "tabiri-web-daily-snapshots"
  region = var.region

  snapshot_schedule_policy {
    schedule {
      daily_schedule {
        days_in_cycle = 1
        start_time    = "03:00"
      }
    }

    retention_policy {
      max_retention_days    = 14
      on_source_disk_delete = "KEEP_AUTO_SNAPSHOTS"
    }

    snapshot_properties {
      storage_locations = ["us"]
    }
  }
}

resource "google_compute_disk_resource_policy_attachment" "daily_snapshots" {
  name = google_compute_resource_policy.daily_snapshots.name
  disk = google_compute_disk.data.name
  zone = var.zone
}
