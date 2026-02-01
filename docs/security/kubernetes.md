# Kubernetes Architecture

Architecture overview for Kubernetes clusters deployed by pltf.

## Description
pltf does not enforce a fixed Kubernetes layout. The cluster shape depends on the modules you include (for example, `aws_eks` for the control plane and `helm_chart` or custom modules for workloads and add-ons).

### Third-party integrations
Common add-ons are typically deployed via Helm charts into their own namespaces. Which components you install (ingress, metrics, service mesh, logging, etc.) depends on your module selection and organization standards.

### Services (pltf modules)
Service modules such as `helm_chart` (or custom modules) are responsible for workload deployment. Typical responsibilities include:
- Namespace creation scoped to the service/layer name.
- Deployments, Services, and optional Ingress resources.
- Service accounts and IAM bindings when the module implements them.
- ConfigMaps/Secrets for app config and credentials.

## Security Overview
- Use least-privilege IAM roles via module links where supported.
- Prefer short-lived cloud auth (IRSA/Workload Identity/OIDC) when your modules configure it.
- Store sensitive values in secret managers and inject via Terraform variables.
- Helm v3 is used for Helm chart deployments.
