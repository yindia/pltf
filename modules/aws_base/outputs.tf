output "kms_account_key_arn" {
  value = aws_kms_key.key.arn
}

output "s3_log_bucket_name" {
  value = aws_s3_bucket.log_bucket.id
}

output "kms_account_key_id" {
  value = aws_kms_key.key.id
}

output "vpc_id" {
  value = local.vpc_id
}

output "private_subnet_ids" {
  value = local.private_subnet_ids
}

output "public_subnets_ids" {
  value = local.public_subnet_ids
}

output "public_nat_ips" {
  value = local.public_nat_ips
}

output "db_aws_security_group" {
  value = "pltf-${var.env_name}-db-sg"
}

output "documentdb_aws_security_group" {
  value = "pltf-${var.env_name}-documentdb-sg"
}

output "elasticache_aws_security_group" {
  value = "pltf-${var.env_name}-elasticache-sg"
}

output "kms_key_alias" {
  value = "alias/pltf-${var.env_name}"
}

