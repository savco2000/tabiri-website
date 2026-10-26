terraform {
  required_version = ">= 1.10.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }

  backend "gcs" {
    bucket = "tabiri-public-sites-terraform-state"
    prefix = "terraform/production"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}
