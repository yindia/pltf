# aws_documentdb

Provision a DocumentDB cluster with subnet groups, encryption, backups, and security groups in the base VPC.

## What it does

- Creates a DocumentDB cluster with configurable engine version and instance count.
- Uses subnet groups and security groups from the VPC; enables encryption.
- Supports deletion protection and exposes host/user/password outputs.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
deletion_protection | A value that indicates whether the DB cluster has deletion protection enabled. The database can't be deleted when deletion protection is enabled. | False | False
engine_version |  | 4.0.0 | False
instance_class |  | db.r5.large | False
instance_count | Number of Instances for aws_docdb_cluster_instance | 1 | False

## Outputs

Name | Description
--- | ---
db_host | Cluster endpoint.
db_password | Master password.
db_user | Master username.

