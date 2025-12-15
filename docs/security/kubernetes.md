# Kubernetes Architecture

Architecture overview for Kubernetes clusters deployed by pltf.

## Description
pltf divides the cluster into namespaces for third-party integrations and for your services. Third-party components are deployed via Helm charts (v3) into their own namespaces; your services are deployed into namespaces derived from the service/layer name.

### Third-party integrations (common set)
- **Linkerd** (service mesh) — mTLS, traffic control, golden metrics; chosen for simplicity and security.
- **Metrics Server** — HPA metrics (built-in on GKE/Azure; installed on EKS).
- **Cluster Autoscaler** — scales nodes (built-in on GKE/Azure; installed on EKS).
- **Ingress NGINX** — ingress controller routing LB traffic into the cluster.
- **External DNS** — manages DNS records for LBs (not needed on GKE/Azure by default).
- **Datadog (optional)** — metrics/logs/APM via the Datadog K8s integration module.

### Services (pltf modules)
Each service (`aws_k8s_service`/`gcp_k8s_service`) creates:
- Namespace named from the service (layer) name.
- Deployment + pods, Service, optional Ingress.
- Horizontal Pod Autoscaler (CPU/memory driven).
- Service Account wired to cloud IAM via IRSA/Workload Identity; least privilege via links.
- ConfigMap/Secrets for app config and credentials (secrets encrypted at rest by the cloud).
- Internal DNS of the form `<module_name>.<layer_name>` for service-to-service calls.

## Security Overview
- Linkerd mTLS secures cross-service traffic.
- Official/Bitnami Helm charts, version-locked; IAM roles scoped to least privilege.
- Service accounts per service; no extra cluster roles granted by default.
- IRSA/Workload Identity/OIDC for cloud access; no long-lived credentials in pods.
- Secrets stored in K8s are encrypted at rest; cloud KMS used by the control plane.
- plft does not modify `aws-auth` beyond optional `admin_arns` configuration.
- Helm v3 used for all chart deployments.
