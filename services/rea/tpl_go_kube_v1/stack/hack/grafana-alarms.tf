locals {
  labels = {
    team = var.team
    app = var.service_name
    env = var.env
  }
}
module "pod_alarms" {
  source = "../../../../environments/v2/common/tf_grafana_alarms"
  grafana_folder_uid = grafana_folder.service.uid
  labels = local.labels
  namespace          = kubernetes_namespace_v1.service.id
  rule_group_name    = "Pods"
}

resource "grafana_rule_group" "service" {
  name       = "Service"
  folder_uid = grafana_folder.service.uid
  org_id     = ""

  interval_seconds = 60

  rule {
    name = "generic errors"
    annotations = {
      "summary"     = "Generic error in the log"
      "description" = ""
      "runbook_url" = ""
    }
    labels         = local.labels
    no_data_state  = "OK"
    exec_err_state = "Alerting"
    
    condition      = "C"

    data {
      ref_id         = "A"
      query_type = "instant"
      relative_time_range {
        from = 300
        to   = 0
      }
      datasource_uid = "loki-heero-grafanacloud"
      model = jsonencode({
        queryType = "instant"
        range     = false
        instant   = true
        intervalMs = 1000
        refId = "A"
        expr      = <<EOF
sum(
    count_over_time(
        {cluster="${module.common_config.eks[var.env]}",namespace="${kubernetes_namespace_v1.service.id}"}
        |= "level=ERROR" 
        [$__auto]
    )
)
EOF
        "intervalMs": 1000,
				"maxDataPoints": 43200,
				"queryType": "instant",
      })
    }

    data {
      datasource_uid = "__expr__"
      ref_id         = "C"

      relative_time_range {
        from = 300
        to   = 0
      }
      model = jsonencode({
        expression = "$A > 0"
        type       = "math"
        intervalMs = 1000
      })
    }
  }
}

