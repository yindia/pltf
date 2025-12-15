# pltf

pltf is a higher-level Infrastructure-as-Code framework. Instead of hand-crafting low-level cloud config, you describe **environments** and **services** in concise YAML. pltf turns those high-level constructs into Terraform so you keep full portability—generate the code, extend it, or take it with you.

![Architecture](images/hero.png) <!-- Replace with your downloaded image -->

## Why pltf
Infrastructure-as-code is essential, but working directly with low-level cloud and Terraform can be complex. pltf bakes in cloud/IaC best practices so you can set up automated, scalable, and secure infrastructure quickly—without being a full-time DevOps engineer. Because pltf emits Terraform, you avoid lock-in and can extend or own the generated code at any time.

## How It Works
With pltf you write configuration files and run the CLI (locally or in CI/CD). The CLI connects to your cloud, renders Terraform (providers/backends/locals/remote state), and can execute Terraform for you.

There are two primary spec types:

- **Environment**: defines cloud/provider, account, region, backend, and shared modules (clusters, networks, IAM, ingress, etc.). You might have one per staging/prod/QA, or per engineer/PR for isolated sandboxes.
- **Service**: defines an application workload and the non-Kubernetes resources it needs, linked to an Environment. Service specs seamlessly connect to environment outputs and modules.

Environment and service specs are linked via `metadata.ref` and `envRef`.

## What You Can Do
- **Generate IaC fast**: turn Environment/Service YAML into Terraform with consistent providers/backends/remote state.
- **Mix modules**: use the embedded catalog or your own (`source: custom`) with the same wiring rules.
- **Choose backends**: store state in `s3|gcs|azurerm` regardless of target cloud; use profiles for cross-account S3.
- **Run Terraform safely**: `pltf terraform plan/apply/destroy/output/unlock` auto-generate before executing TF.
- **Validate & lint**: structural checks plus suggestions (labels, unused vars).
- **Preview**: see provider/backend/labels/modules without running TF.

## Next Steps
- Follow [Getting Started](getting-started/aws.md).
- Explore examples (coming soon).
- Review [Security](security/aws.md).

## Quick Links
- [Installation](installation.md)
- [Getting Started](getting-started/aws.md)
- [CLI Usage](usage.md)
- [Spec Guide](specs.md)
- [Modules & Wiring](modules.md)
- [Features](features.md)
- [References](references/aws.md)
- [Security](security/aws.md)
