# pltf

The next generation of Infrastructure-as-Code. Describe high-level constructs; ship Terraform you can own.

pltf turns concise YAML into ready-to-run Terraform for AWS, GCP, and Azure. You model three things:

- **Stack**: reusable module bundles (EKS + nodegroups + observability).
- **Environment**: cloud, account/project, region, shared modules (VPC, DNS, EKS/GKE/AKS, IAM).
- **Service**: an app’s resources (databases, queues, buckets, roles, charts) wired into an Environment.

The CLI validates your specs, renders providers/backends/locals/remote state, and can run `terraform plan/apply` end-to-end. Because the output is plain Terraform, you keep portability—extend it or take it with you.

> Status: active development; review generated Terraform before applying to production.

## Why teams use pltf
- **Faster IaC**: generate consistent Terraform from YAML instead of hand-rolling cloud glue.
- **Cloud agnostic outputs**: state backends `s3|gcs|azurerm`, provider wiring, and locals are emitted for you.
- **Module catalog + custom modules**: ship with AWS modules and accept your own `module.yaml` definitions.
- **Safe automation**: `pltf terraform plan/apply/destroy/output/graph` re-renders code every run to keep drift in check.
- **Validation**: structural checks catch missing refs, bad wiring, and secrets placement early.

## Grounded example
Use the repo samples as a blueprint:

- `example/stack.yaml` (and `example/stacks/*`) define reusable cluster building blocks.
- `example/env.yaml` defines an AWS environment (`example-aws`) and references stacks for EKS and observability.
- `example/service.yaml` defines a service (`payments-api`) that binds to the environment and shows how to pass variables/secrets and link modules.

## Typical workflow
1. Create or edit your stack, environment, and service specs.
2. Validate and preview wiring:
   - `pltf validate -f example/env.yaml`
   - `pltf preview -f example/service.yaml --env prod`
3. Generate or run Terraform:
   - `pltf terraform plan -f example/service.yaml --env prod`
   - `pltf terraform apply -f example/service.yaml --env prod`
4. Inspect outputs and graphs:
   - `pltf terraform output -f example/service.yaml --env prod`
   - `pltf terraform graph -f example/service.yaml --env prod | dot -Tpng > graph.png`

## Quick links
- [Installation](installation.md)
- [Getting Started](getting-started/aws.md)
- [Platform Usage](platform.md)
- [CLI Reference](usage.md)
- [Spec Guide](specs.md)
- [Modules & Wiring](modules.md)
- [Features](features.md)
- [Security](security/aws.md)
