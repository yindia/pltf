# Modules & Wiring

Modules are discovered from a modules root where each module type directory contains a `module.yaml`. The CLI scans custom roots (when provided) and the embedded catalog. Modules marked `source: custom` must be found in your custom root; others fall back to embedded.

## Wiring rules
- Inputs auto-wire to outputs with the same name (current scope, or parent env for services).
- Required inputs without a value or matching output fail validation.
- Optional/default inputs can stay unwired if nothing matches.
- Templates `${module.*}`, `${var.*}`, `${parent.*}` are supported and converted to Terraform traversals.

## Module metadata (module.yaml)
Example fields:
```yaml
name: aws_eks
type: aws_eks
provider: aws
version: 1.0.0
description: EKS cluster
inputs:
  - name: cluster_name
    type: string
    required: true
  - name: enable_metrics
    type: bool
    required: true
  - name: env_name
    type: string
    required: true
outputs:
  - name: k8s_cluster_name
    type: string
```
Notes:
- Inputs may include `description`, `default`, `capability` (optional).
- Outputs may include `description`, `capability`.
- Capabilities can declare `provides`/`accepts` to describe contracts.

## Embedded modules (AWS)
- `aws_base`, `aws_dns`, `aws_eks`, `aws_k8s_base`, `aws_k8s_service`, `aws_nodegroup`
- `aws_postgres`, `aws_mysql`, `aws_redis`, `aws_dynamodb`, `aws_s3`, `aws_ses`, `aws_sns`, `aws_sqs`, `aws_documentdb`
- `aws_iam_role`, `aws_iam_policy`, `aws_iam_user`
- `cloudfront_distribution`

GCP/Azure: no bundled modules yet; use custom modules or your own registry. You can target GCP/Azure providers with custom modules and backends.

## Custom modules
- Mark spec entries with `source: custom` to force lookup in your custom modules root (`--modules` or profile `modules_root`).
- Generate `module.yaml` for your module with `pltf module init --path <module_dir> [--force]`.
- Inventory commands: `pltf module list|get [-m ./modules] -o table|json|yaml`.

Treat modules as black boxes: configure via `inputs`, consume declared `outputs`, and let wiring handle references.

## Module init helper
Use `pltf module init --path <module_dir> [--force]` to generate or refresh `module.yaml` from an existing Terraform module. This inspects variables/outputs and writes a fresh descriptor (backing up or overwriting if `--force`).
