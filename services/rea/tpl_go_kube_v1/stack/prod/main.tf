terraform {
  required_providers {
    grafana = {
      source = "grafana/grafana"
      version = "~> 4.10.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.15.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.21"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.38"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~>3.0"
    }

  }
  backend "s3" {
    bucket = "heero-tfstates-prod"
    key = "heero/apps/[[.ServiceName]]"
    region = "eu-west-1"
  }
}

locals {
  region = "eu-west-1"
  gcloud_project_id = "heero-${var.env}-412816"
  service_name = var.service_name
}


provider "aws" {
  region = "eu-west-1"
  default_tags {
    tags = {
      project     = "heero"
      environment = var.env
      source = "services/${local.service_name}"
    }
  }
}

provider "google" {
  project = local.gcloud_project_id
}

provider "grafana" {
  url  = module.common_config.grafana[var.env]
  auth = module.common_config.infra_token["apitoken-grafana"]
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.heero.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.heero.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.default.token
}

data "aws_eks_cluster_auth" "default" {
  name = "heero-${var.env}"
}

provider "helm" {
  kubernetes = {
    host                   = data.aws_eks_cluster.heero.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.heero.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.default.token
  }
}

data "aws_caller_identity" "current" {}
