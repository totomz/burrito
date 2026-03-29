resource "grafana_folder" "service" {
  title = var.service_name
  uid   = lower(var.service_name)
}