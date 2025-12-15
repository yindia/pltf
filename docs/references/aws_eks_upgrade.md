# AWS EKS Upgrade

How to upgrade the version of your EKS cluster created by pltf.

## Overview
EKS does not auto-upgrade clusters. Upgrade one minor version at a time (e.g., 1.24 → 1.25). Steps below use the AWS console; CLI works too.

## Step 1: Control Plane
1. Open the EKS cluster (correct region).
2. Click **Update now** on the control plane.
3. Select the next Kubernetes version and start the update.

**Important:** During control plane upgrade (~20 min) avoid new deploys or kubectl changes. Running workloads keep serving traffic.

## Step 2: Node Groups
1. Go to **Configuration → Compute**.
2. For each managed node group, click **Update now**.
3. Use **Rolling update** strategy and start the upgrade.

**Important:** Nodes are replaced. If ingress is not HA, expect brief downtime while pods reschedule. Upgrade node groups one at a time.

## Step 3: Pin versions in specs
After upgrading, pin the new versions so future applies stay consistent:
```yaml
modules:
  - type: aws_eks
    k8s_version: "1.25"
  - type: aws_nodegroup
    name: default
    k8s_version: "1.25"
```
Then:
```bash
pltf terraform plan -f env.yaml -e prod
pltf terraform apply -f env.yaml -e prod
```

## Breaking Changes
- Review Kubernetes API deprecations for your target version.
- Ensure add-ons (CNI, metrics, ingress) are compatible.
- For multi-hop upgrades (e.g., 1.22 → 1.24), step through each minor version.

## References
- AWS: https://docs.aws.amazon.com/eks/latest/userguide/update-cluster.html
- Versions: https://docs.aws.amazon.com/eks/latest/userguide/kubernetes-versions.html
