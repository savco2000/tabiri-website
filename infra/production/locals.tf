locals {
  application_name = "tabiri-web"
  network_tag      = "tabiri-web"

  required_services = toset([
    "compute.googleapis.com",
    "iam.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "secretmanager.googleapis.com",
    "storage.googleapis.com",
  ])
}
