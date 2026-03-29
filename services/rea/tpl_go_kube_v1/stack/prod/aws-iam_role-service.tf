data "aws_iam_policy_document" "service_assumerole_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"
    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_iam_openid_connect_provider.heero_eks.url, "https://", "")}:sub"
      values   = ["system:serviceaccount:${kubernetes_namespace_v1.service.id}:${var.service_name}"]
    }
    principals {
      type        = "Federated"
      identifiers = [data.aws_iam_openid_connect_provider.heero_eks.arn]
    }
  }
  statement {
    actions = ["sts:AssumeRole"]
    effect = "Allow"
    principals {
      type = "AWS"
      identifiers = ["arn:aws:iam::654654538697:role/aws-reserved/sso.amazonaws.com/eu-west-1/AWSReservedSSO_AdministratorAccess_e3a8930eb4c9cfc3"]
    }
  }
  depends_on = [data.aws_iam_openid_connect_provider.heero_eks]
}

resource "aws_iam_role" "service_role" {
  name               = "${var.service_name}-${var.env}"
  assume_role_policy = data.aws_iam_policy_document.service_assumerole_policy.json
}

resource "aws_iam_policy" "service_policy" {
  name        = "${var.service_name}-default-policy"
  description = "Default policy for service ${var.service_name}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:Describe*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource =  "${aws_secretsmanager_secret.service_secret.arn}-*"
        # Resource =  "arn:aws:secretsmanager:eu-west-1:767398121280:secret:heero/${var.env}/service/kfc/temporal-*"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "service_role_policy" {
  role       = aws_iam_role.service_role.name
  policy_arn = aws_iam_policy.service_policy.arn
}

