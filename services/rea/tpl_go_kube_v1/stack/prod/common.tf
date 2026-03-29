module "common_config" {
  source = "../../../../environments/v2/common/config"
  env = var.env
}

data "aws_eks_cluster" "heero" {
  name = "heero-${var.env}"
}

data "aws_iam_openid_connect_provider" "heero_eks" {
  url = data.aws_eks_cluster.heero.identity[0].oidc[0].issuer
}