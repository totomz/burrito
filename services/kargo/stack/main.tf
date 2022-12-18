terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.7.0"
    }
    
  }

  backend "s3" {
    profile = "croccocode"
    region  = "eu-west-1"
    bucket  = "infra-tf-states-fjow94"
    key     = "infra/v1"
  }
}


provider "aws" {
  profile = var.aws_profile
  region  = var.aws_region
  
  default_tags {
    tags = {
      managedBy = "terraform/ff_infra" 
    }
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = var.kubectl_context
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
    config_context = var.kubectl_context
  }
}

provider "kubectl" {
  config_path = "~/.kube/config"
  config_context = var.kubectl_context
}