resource "aws_route53_record" "service" {
	zone_id = "Z020978430FOM7BMGABHT"
	name    = "[[.ServiceName]].hack.heero.me"
	type    = "A"
	alias {
		name                   = "eks-heero-hack-ext-1712184268.eu-west-1.elb.amazonaws.com"
		zone_id                = "Z32O12XQLNTSW2"
		evaluate_target_health = true
	}
}
