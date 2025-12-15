resource "aws_route53_zone" "public" {
  name = var.domain
  tags = {
    Name               = "pltf-${var.env_name}"
    "pltf-environment" = var.env_name
  }
  force_destroy = true
}