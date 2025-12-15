# aws_redis

Provisions ElastiCache Redis with subnet group, encryption in-transit/at-rest, and parameter options.

## What it does

- Deploys ElastiCache Redis cluster/subnet group with in-transit/at-rest encryption.
- Configurable engine version, node class, cluster size, and parameter family.
- Outputs cache endpoints and security group details.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
node_type |  | cache.m4.large | False
redis_version |  | 6.x | False
snapshot_retention_limit | Days for which the Snapshot should be retained. | 0 | False
snapshot_window | When should the Snapshot for redis cache be done. UTC Time. Snapshot Retention Limit should be set to more than 0. | 04:00-05:00 | False

## Outputs

Name | Description
--- | ---
cache_auth_token | 
cache_host | Redis host.

