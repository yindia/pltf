# aws_mysql

Provisions an Aurora MySQL cluster with subnet group, encryption, backups, and optional multi-AZ.

## What it does

- Provisions an Aurora MySQL cluster in private subnets with encryption.
- Supports multi-AZ, backups, retention, and public accessibility toggle.
- Exposes writer/reader endpoints and security/subnet group metadata.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
backup_retention_days | How many days to keep the backup retention |  | True
db_name |  | app | False
engine_version |  | 5.7.mysql_aurora.2.04.2 | False
instance_class |  | db.t3.medium | False
multi_az |  | False | False
safety |  | False | False

## Outputs

Name | Description
--- | ---
db_host | 
db_name | 
db_password | 
db_user | 

