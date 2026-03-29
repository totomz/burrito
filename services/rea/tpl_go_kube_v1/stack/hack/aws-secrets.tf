resource "aws_secretsmanager_secret" "service_secret" {
  name        = "heero/${var.env}/service/${var.service_name}"
  description = "Secrets for ${var.service_name}"
}

# Placeholder - populate it later
resource "aws_secretsmanager_secret_version" "service_secret_version" {
  secret_id     = aws_secretsmanager_secret.service_secret.id
  secret_string = jsonencode({
    placeholder = "change-me"
  })

  # ignore changes to the secret
  lifecycle {
    ignore_changes = all
  }
}