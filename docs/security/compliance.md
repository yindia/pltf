# Compliance

SOC 2 and PCI considerations for infrastructure deployed by pltf.

## Overview
pltf aims to make SOC 2 and PCI alignment the default for cloud resources it provisions. Compliance is broader than infrastructure; this covers only the cloud layer. Engage a compliance partner for full org-level readiness.

## Methodology
- We scan representative environments with Fugue/Regula for SOC2/PCI controls before releases.
- Findings are fixed when possible; otherwise documented below. Backward-incompatible changes are avoided; new defaults apply to newly created resources.

## AWS
AWS infrastructure can meet SOC2/PCI with the following settings:
- S3 buckets: deny non-SSL traffic; enable `same_region_replication` for backups.
- Postgres (Aurora): enable `multi_az`.

Example:
```yaml
modules:
  - name: db
    type: aws_postgres
    multi_az: true
  - name: s3
    type: aws_s3
    same_region_replication: true
    bucket_policy:
      Version: "2012-10-17"
      Statement:
        - Sid: denyInsecureTransport
          Effect: Deny
          Principal: "*"
          Action: "s3:*"
          Resource:
            - "arn:aws:s3:::${parent_name}-${layer_name}/*"
            - "arn:aws:s3:::${parent_name}-${layer_name}"
          Condition:
            Bool:
              aws:SecureTransport: "false"
```
Notes auditors may raise:
- Terraform lock DynamoDB table is unencrypted (no customer data, only hashes).
- Terraform state bucket logging is not enabled by default (bootstrap ordering). You may manually add logging to the log bucket.
- The log bucket does not log itself.

## GCP
Current gaps to full SOC2/PCI:
- GKE nodegroup VMs cannot disable `block-project-ssh-keys` easily.
- GKE node disks via KMS key encryption are still limited (beta in TF); will adopt when GA.
- Defaults without uniform bucket-level access (GCS state bucket, GCR-backed bucket) to avoid tedious per-user grants; can be manually enabled if desired.

## Azure
Azure can meet SOC2/PCI with an extra user step:
- Enable flow logs for the agent pool security group.

We continue to monitor provider capabilities and will tighten defaults as features mature.
