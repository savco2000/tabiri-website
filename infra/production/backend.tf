resource "google_compute_health_check" "web" {
  name                = "tabiri-web-health"
  check_interval_sec  = 10
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 3

  http_health_check {
    port         = 4000
    request_path = "/healthz"
  }

  depends_on = [google_project_service.required]
}

resource "google_compute_backend_service" "web" {
  name                  = "tabiri-web-backend"
  protocol              = "HTTP"
  port_name             = "http"
  timeout_sec           = 30
  load_balancing_scheme = "EXTERNAL"
  health_checks         = [google_compute_health_check.web.id]

  dynamic "backend" {
    for_each = var.deploy_application ? [google_compute_instance_group.web[0].id] : []

    content {
      group           = backend.value
      balancing_mode  = "UTILIZATION"
      max_utilization = 0.8
    }
  }
}

resource "google_compute_global_network_endpoint_group" "google_tag_gateway" {
  name                  = "tabiri-google-tag-gateway"
  network_endpoint_type = "INTERNET_FQDN_PORT"
  default_port          = 443
}

resource "google_compute_global_network_endpoint" "google_tag_gateway" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.google_tag_gateway.name
  fqdn                          = "gtm-wh3cmbpj.fps.goog"
  port                          = 443
}

resource "google_compute_backend_service" "google_tag_gateway" {
  name                  = "tabiri-google-tag-gateway"
  protocol              = "HTTPS"
  timeout_sec           = 30
  load_balancing_scheme = "EXTERNAL"
  custom_request_headers = [
    "Host: gtm-wh3cmbpj.fps.goog",
    "X-Forwarded-CountryRegion: {client_region_subdivision}",
    "X-Forwarded-Geolocation: latlong={client_city_lat_long};city={client_city}",
  ]

  backend {
    group = google_compute_global_network_endpoint_group.google_tag_gateway.id
  }

  depends_on = [google_compute_global_network_endpoint.google_tag_gateway]
}

data "google_compute_backend_bucket" "existing_static_site" {
  name = "tabirianalytics-backend"
}
