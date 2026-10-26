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

    route_rules {
      priority = 10

      match_rules {
        full_path_match = "/blog-post.html"
      }

      url_redirect {
        path_redirect          = "/blog/why-secure-ai-teams-outperform-generic-automation-tools"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 20

      match_rules {
        full_path_match = "/blog-post-2.html"
      }

      url_redirect {
        path_redirect          = "/blog/inside-a-regulated-fintechs-agentic-secops-penetration-test"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 30

      match_rules {
        full_path_match = "/blog-post-3.html"
      }

      url_redirect {
        path_redirect          = "/blog/from-alert-queues-to-agentic-defense-protecting-utility-infrastructure-mapping-systems"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 40

      match_rules {
        full_path_match = "/blog-post-4.html"
      }

      url_redirect {
        path_redirect          = "/blog/how-tabiris-ai-agent-triaged-php2shell-safely-in-less-than-20-minutes"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 50

      match_rules {
        full_path_match = "/blog-post-5.html"
      }

      url_redirect {
        path_redirect          = "/blog/one-brotli-error-five-leads-how-tabirids-ai-agent-traced-node2shell-risk"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 60

      match_rules {
        full_path_match = "/digital-employees.html"
      }

      url_redirect {
        path_redirect          = "/digital-employees"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 70

      match_rules {
        full_path_match = "/operational-assessment.html"
      }

      url_redirect {
        path_redirect          = "/operational-assessment"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 80

      match_rules {
        full_path_match = "/penetration-testing.html"
      }

      url_redirect {
        path_redirect          = "/penetration-testing"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 90

      match_rules {
        full_path_match = "/managed-cybersecurity-services.html"
      }

      url_redirect {
        path_redirect          = "/managed-cybersecurity-services"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 100

      match_rules {
        full_path_match = "/cybersecurity-insurance.html"
      }

      url_redirect {
        path_redirect          = "/cybersecurity-insurance"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 110

      match_rules {
        full_path_match = "/data-protection.html"
      }

      url_redirect {
        path_redirect          = "/data-protection"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 120

      match_rules {
        full_path_match = "/case-studies.html"
      }

      url_redirect {
        path_redirect          = "/case-studies"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 130

      match_rules {
        full_path_match = "/channel-partners.html"
      }

      url_redirect {
        path_redirect          = "/channel-partners"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 140

      match_rules {
        full_path_match = "/capability-statement.html"
      }

      url_redirect {
        path_redirect          = "/capability-statement"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 150

      match_rules {
        full_path_match = "/blog.html"
      }

      url_redirect {
        path_redirect          = "/blog"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }

    route_rules {
      priority = 160

      match_rules {
        full_path_match = "/about.html"
      }

      url_redirect {
        path_redirect          = "/about"
        redirect_response_code = "FOUND"
        strip_query            = false
      }
    }
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
