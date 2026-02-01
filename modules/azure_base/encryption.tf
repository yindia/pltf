resource "random_id" "key_suffix" {
  byte_length = 8
}

resource "azurerm_key_vault" "pltf" {
  name                            = "pltf-${random_id.key_suffix.hex}"
  location                        = data.azurerm_resource_group.pltf.location
  resource_group_name             = data.azurerm_resource_group.pltf.name
  tenant_id                       = data.azurerm_subscription.current.tenant_id
  enable_rbac_authorization       = true
  sku_name                        = "premium"
  enabled_for_disk_encryption     = true
  enabled_for_deployment          = true
  enabled_for_template_deployment = true
  purge_protection_enabled        = true
  soft_delete_retention_days      = var.key_vault_soft_delete_retention_days

  network_acls {
    bypass                     = "AzureServices"
    default_action             = "Deny"
    ip_rules                   = var.key_vault_ip_rules
    virtual_network_subnet_ids = var.key_vault_subnet_ids
  }
  lifecycle {
    ignore_changes = [location]
  }
}

resource "azurerm_key_vault_key" "acr" {
  name            = "pltf-${var.env_name}-acr"
  key_vault_id    = azurerm_key_vault.pltf.id
  key_type        = "RSA"
  key_size        = 2048
  expiration_date = var.key_vault_key_expiration_date

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey",
  ]
}
