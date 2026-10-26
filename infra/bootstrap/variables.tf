variable "project_id" {
  description = "Google Cloud project that hosts the website."
  type        = string
  default     = "tabiri-public-sites"
}

variable "state_bucket_name" {
  description = "Globally unique Cloud Storage bucket for Terraform state."
  type        = string
  default     = "tabiri-public-sites-terraform-state"
}
