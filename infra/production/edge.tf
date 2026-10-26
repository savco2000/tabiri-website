resource "google_compute_managed_ssl_certificate" "production" {
  name = "tabirianalytics-cert-v2"

  managed {
    domains = [
      "tabirianalytics.com",
      "www.tabirianalytics.com",
    ]
  }
}

resource "google_compute_url_map" "production" {
  name            = "tabirianalytics-map"
  default_service = google_compute_backend_service.web.id

  host_rule {
    hosts        = ["tabirianalytics.com"]
    path_matcher = "apex-application"
  }

  host_rule {
    hosts        = ["www.tabirianalytics.com"]
    path_matcher = "www-redirect"
  }

  path_matcher {
    name            = "apex-application"
    default_service = google_compute_backend_service.web.id
  }

  path_matcher {
    name = "www-redirect"

    default_url_redirect {
      host_redirect          = "tabirianalytics.com"
      https_redirect         = true
      redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
      strip_query            = false
    }
  }
}

resource "google_compute_target_https_proxy" "production" {
  name    = "tabirianalytics-proxy"
  url_map = google_compute_url_map.production.id
  ssl_certificates = [
    google_compute_managed_ssl_certificate.production.id,
  ]
  quic_override  = "NONE"
  tls_early_data = "DISABLED"
}
