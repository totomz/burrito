resource "kubernetes_namespace_v1" "service" {
  metadata {
    name = var.service_name
  }
  lifecycle {
    ignore_changes = [
      metadata[0].annotations
    ]
  }
}