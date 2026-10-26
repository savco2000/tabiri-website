resource "google_storage_bucket" "terraform_state" {
  name                        = var.state_bucket_name
  location                    = "US"
  force_destroy               = false
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  versioning {
    enabled = true
  }

  lifecycle {
    prevent_destroy = true
  }
}
