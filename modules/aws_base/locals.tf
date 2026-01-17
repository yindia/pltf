locals {
  create_vpc         = var.vpc_id == null
  vpc_id             = local.create_vpc ? aws_vpc.vpc[0].id : data.aws_vpc.vpc[0].id
  vpc_cidr_blocks    = local.create_vpc ? [aws_vpc.vpc[0].cidr_block] : data.aws_vpc.vpc[0].cidr_block_associations[*].cidr_block
  private_subnet_ids = local.create_vpc ? aws_subnet.private_subnets[*].id : values(data.aws_subnet.private_subnets)[*].id
  public_subnet_ids  = local.create_vpc ? aws_subnet.public_subnets[*].id : values(data.aws_subnet.public_subnets)[*].id
  public_nat_ips     = local.create_vpc ? aws_eip.nat_eips[*].public_ip : []
  cluster_name       = trimspace(coalesce(var.cluster_name, "")) != "" ? var.cluster_name : "pltf-${var.layer_name}"
  cluster_tags = merge(
    {
      "kubernetes.io/cluster/${local.cluster_name}" = "shared"
    },
    var.enable_karpenter ? { "karpenter.sh/discovery" = local.cluster_name } : {}
  )
}
