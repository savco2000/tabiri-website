resource "google_compute_network" "web" {
  name                    = "tabiri-web-network"
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"

  depends_on = [google_project_service.required]
}

resource "google_compute_subnetwork" "web" {
  name                     = "tabiri-web-subnet"
  region                   = var.region
  network                  = google_compute_network.web.id
  ip_cidr_range            = "10.20.0.0/24"
  private_ip_google_access = true
}

resource "google_compute_firewall" "load_balancer" {
  name      = "tabiri-web-allow-load-balancer"
  network   = google_compute_network.web.name
  direction = "INGRESS"

  source_ranges = [
    "35.191.0.0/16",
    "130.211.0.0/22",
  ]
  target_tags = [local.network_tag]

  allow {
    protocol = "tcp"
    ports    = ["4000"]
  }
}

resource "google_compute_firewall" "iap_ssh" {
  name      = "tabiri-web-allow-iap-ssh"
  network   = google_compute_network.web.name
  direction = "INGRESS"

  source_ranges = ["35.235.240.0/20"]
  target_tags   = [local.network_tag]

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}
