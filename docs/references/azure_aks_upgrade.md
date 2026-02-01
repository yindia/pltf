# Azure AKS Upgrade

How to upgrade an AKS cluster created by pltf.

## Overview
AKS does not auto-upgrade clusters by default. Upgrade one minor version at a time and validate workloads between steps.

## Step 1: Control plane
1. Inspect available versions:
   ```bash
   az aks get-upgrades --name <cluster> --resource-group pltf-<env>
   ```
2. Upgrade the control plane:
   ```bash
   az aks upgrade --name <cluster> --resource-group pltf-<env> --kubernetes-version <version>
   ```

## Step 2: Node pools
Upgrade each node pool after the control plane:
```bash
az aks nodepool upgrade \
  --resource-group pltf-<env> \
  --cluster-name <cluster> \
  --name <node-pool> \
  --kubernetes-version <version>
```

## Step 3: Pin versions in specs
Update your spec so future applies stay consistent:
```yaml
modules:
  - id: aks
    type: azure_aks
    inputs:
      kubernetes_version: "1.28.3"
```
Then:
```bash
pltf terraform plan -f env.yaml -e prod
pltf terraform apply -f env.yaml -e prod
```

## Notes
- Review API deprecations and add-on compatibility before upgrading.
- Upgrade node pools one at a time to reduce impact.

## References
- AKS upgrades: https://learn.microsoft.com/azure/aks/upgrade-cluster
