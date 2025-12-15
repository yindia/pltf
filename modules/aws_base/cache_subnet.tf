resource "aws_elasticache_subnet_group" "main" {
  name       = "pltf-${var.layer_name}"
  subnet_ids = local.private_subnet_ids
}
