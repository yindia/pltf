# GCP GKE Upgrade

How to upgrade a GKE cluster created by pltf.

## Overview
pltf configures the GKE release channel via `gke_channel` (RAPID, REGULAR, STABLE). Upgrades happen according to Google’s channel policy. You can also trigger manual upgrades with `gcloud`.

## Option 1: Change release channel (recommended)
Update your spec:
```yaml
modules:
  - id: gke
    type: gcp_gke
    inputs:
      gke_channel: "REGULAR"
```
Then:
```bash
pltf terraform plan -f env.yaml -e prod
pltf terraform apply -f env.yaml -e prod
```

## Option 2: Manual upgrade with gcloud
1. Inspect available versions:
   ```bash
   gcloud container get-server-config --region <region>
   ```
2. Upgrade control plane:
   ```bash
   gcloud container clusters upgrade <cluster> \
     --region <region> \
     --project <project-id>
   ```
3. Upgrade node pools:
   ```bash
   gcloud container node-pools upgrade <node-pool> \
     --cluster <cluster> \
     --region <region> \
     --project <project-id>
   ```

## Notes
- Upgrade one minor version at a time.
- Check add-on compatibility (CNI, ingress, metrics).
- For private clusters, ensure your admin network can reach the control plane before upgrading.

## References
- GKE upgrades: https://cloud.google.com/kubernetes-engine/docs/how-to/upgrading-a-cluster
- Release channels: https://cloud.google.com/kubernetes-engine/docs/concepts/release-channels
