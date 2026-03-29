# resource "aws_route53_record" "service" {
#   zone_id = "Z06584323AIP4M5Z2LMTK"
#   name    = "[[.ServiceName]].app.heero.me"
#   type    = "A"
#   alias {
#     name                   = "eks-heero-prod-ext-361581886.eu-west-1.elb.amazonaws.com"
#     zone_id                = "Z32O12XQLNTSW2"
#     evaluate_target_health = true
#   }
# }
