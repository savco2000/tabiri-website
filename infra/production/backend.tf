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

data "google_compute_backend_bucket" "existing_static_site" {
  name = "tabirianalytics-backend"
}
