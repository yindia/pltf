# aws_dns

Creates a Route53 hosted zone and ACM certificate with DNS validation, wiring records for ingress/load balancers.

## What it does

- Creates Route53 hosted zone and ACM cert with DNS validation or import.
- Exposes NS records and cert ARN for downstream modules.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
cert_chain_included |  | False | False
delegated |  | False | False
domain |  |  | True
external_cert_arn |  |  | False
force_update |  | False | False
upload_cert |  | False | False

## Outputs

Name | Description
--- | ---
cert_arn | ACM certificate ARN (created/imported/external).
domain | Domain name of the hosted zone.
name_servers | Delegated name servers.
zone_id | Route53 hosted zone ID.

