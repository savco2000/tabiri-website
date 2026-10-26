variable "project_id" {
  description = "Google Cloud project that hosts the website."
  type        = string
  default     = "tabiri-public-sites"
}

variable "region" {
  description = "Region for regional application resources."
  type        = string
  default     = "us-east1"
}

variable "zone" {
  description = "Zone for zonal application resources."
  type        = string
  default     = "us-east1-b"
}

variable "machine_type" {
  description = "Compute Engine machine type for the single application VM."
  type        = string
  default     = "e2-small"
}

variable "release_bucket_name" {
  description = "Globally unique bucket containing versioned application releases."
  type        = string
  default     = "tabiri-public-sites-releases"
}

variable "release_object" {
  description = "Object path of the release archive in the release bucket."
  type        = string
  default     = "releases/tabiri-web.tar.gz"
}

variable "release_sha256" {
  description = "Optional SHA-256 checksum for the release archive."
  type        = string
  default     = ""
}

variable "turnstile_site_key" {
  description = "Public Cloudflare Turnstile site key."
  type        = string
  default     = ""
}

variable "turnstile_allowed_hostnames" {
  description = "Comma-separated hostnames accepted in Turnstile responses."
  type        = string
  default     = "www.tabirianalytics.com,tabirianalytics.com"
}

variable "deploy_application" {
  description = "Create the VM and attach it to the new backend service."
  type        = bool
  default     = false
}
