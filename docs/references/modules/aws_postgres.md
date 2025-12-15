# aws_postgres

Provisions an Aurora Postgres cluster with subnet group, encryption, backups, and optional multi-AZ.

## What it does

- Provisions an Aurora Postgres cluster in private subnets with encryption.
- Supports multi-AZ, backups, retention, and public accessibility toggle.
- Exposes writer/reader endpoints and security/subnet group metadata.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
backup_retention_days | How many days to keep the backup retention |  | True
create_global_database |  |  | True
database_name |  |  | True
engine_version |  | 11.9 | False
existing_global_database_id |  |  | True
extra_security_groups_ids |  |  | True
instance_class |  | db.t3.medium | False
multi_az |  | False | False
restore_from_snapshot |  |  | True
safety |  | False | False

## Outputs

Name | Description
--- | ---
db_host | 
db_name | 
db_password | 
db_user | 
global_database_id | 

