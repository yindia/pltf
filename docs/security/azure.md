# Azure Architecture

Security highlights for Azure environments managed by pltf.

## Identity and access
- Use AAD identities and RBAC for AKS (`azure_aks` enables RBAC by default).
- Prefer managed identities over long-lived secrets.

## Network controls
- `azure_base` provisions VNets/subnets and applies restrictive network rules.
- Key Vault is configured with network ACLs and purge protection.

## State and secrets
- Use the `azurerm` backend with a dedicated storage account and container.
- Keep secrets in Key Vault or CI secret stores.

## Logging
- AKS connects to Log Analytics via OMS agent.
- Enable diagnostic settings for critical resources (Key Vault, networking).
