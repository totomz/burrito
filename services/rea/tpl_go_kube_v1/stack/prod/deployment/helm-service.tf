
data "kubernetes_namespace_v1" "service" {
  metadata {
    name = var.service_name
  }
}

data "aws_ecr_repository" "service" {
  name = var.service_name
}

data "aws_iam_role" "service" {
  name = "${var.service_name}-${var.env}"
}

data "google_service_account" "service" {
  account_id   = "${local.service_name}-${var.env}"
}

resource "google_service_account_key" "service" {
  service_account_id = data.google_service_account.service.id
}

resource "helm_release" "service" {
  name       = "[[.ServiceName]]"
  namespace  = data.kubernetes_namespace_v1.service.id
  repository = "oci://${data.aws_caller_identity.current.account_id}.dkr.ecr.${local.region}.amazonaws.com"
  chart      = data.aws_ecr_repository.service.name
  version    = data.external.chart_version.result["version"]
  timeout    = 180

  values = [
    templatefile("${path.module}/helm-service-values.yaml", {
      googleCredentials = google_service_account_key.service.private_key
      env               = var.env
      awsRoleArn        = data.aws_iam_role.service.arn
      imageRepository   = data.aws_ecr_repository.service.repository_url
      configText = indent(4, templatefile("${path.module}/../../../envs/config-${var.env}.yaml", {} ))
    })
  ]

  depends_on = [
  ]
  
}

data "external" "chart_version" {
  program = [
    "bash", "-c",
    "version=$(aws ecr describe-images --repository-name ${data.aws_ecr_repository.service.name} --region ${local.region} --output text --no-cli-pager --query 'sort_by(imageDetails[?artifactMediaType==`application/vnd.cncf.helm.config.v1+json`], &imagePushedAt)[-1:].imageTags[0]'); echo \"{\\\"version\\\": \\\"$version\\\"}\""
  ]
}