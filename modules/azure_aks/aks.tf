data "azurerm_nat_gateway" "nat" {
  name                = "pltf-${var.env_name}-nat-gateway"
  resource_group_name = data.azurerm_resource_group.pltf.name
}

resource "azurerm_user_assigned_identity" "pltf" {
  name                = "pltf-${var.env_name}-aks"
  location            = data.azurerm_resource_group.pltf.location
  resource_group_name = data.azurerm_resource_group.pltf.name
  lifecycle {
    ignore_changes = [
      location
    ]
  }
}

resource "azurerm_user_assigned_identity" "agent_pool" {
  name                = "pltf-${var.env_name}-aks-agent-pool"
  location            = data.azurerm_resource_group.pltf.location
  resource_group_name = data.azurerm_resource_group.pltf.name
  lifecycle {
    ignore_changes = [
      location
    ]
  }
}

resource "azurerm_role_assignment" "pltf" {
  scope                = data.azurerm_resource_group.pltf.id
  role_definition_name = "Network Contributor"
  principal_id         = azurerm_user_assigned_identity.pltf.principal_id
  lifecycle { ignore_changes = [scope] }
}

resource "azurerm_role_assignment" "k8s_assign_identities" {
  scope                = data.azurerm_resource_group.pltf.id
  role_definition_name = "Managed Identity Operator"
  principal_id         = azurerm_user_assigned_identity.pltf.principal_id
  lifecycle { ignore_changes = [scope] }
}

resource "azurerm_role_assignment" "azurerm_container_registry" {
  scope                = data.azurerm_resource_group.pltf.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.pltf.principal_id
  lifecycle { ignore_changes = [scope] }
}

resource "azurerm_role_assignment" "azurerm_container_registry_agent_pool" {
  scope                = data.azurerm_resource_group.pltf.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.agent_pool.principal_id
  lifecycle { ignore_changes = [scope] }
}

# Ignore because we dont know which containers of users need to access API server
# tfsec:ignore:azure-container-limit-authorized-ips
resource "azurerm_kubernetes_cluster" "main" {
  name                = var.cluster_name
  location            = data.azurerm_resource_group.pltf.location
  resource_group_name = data.azurerm_resource_group.pltf.name
  dns_prefix          = "pltf"
  //  disk_encryption_set_id = azurerm_disk_encryption_set.pltf.id
  kubernetes_version                = var.kubernetes_version
  role_based_access_control_enabled = true

  network_profile {
    network_plugin     = "azure"
    network_policy     = "calico"
    service_cidr       = var.service_cidr
    dns_service_ip     = var.dns_service_ip
    docker_bridge_cidr = "172.17.0.1/16"
    outbound_type      = "userAssignedNATGateway"
  }

  azure_active_directory_role_based_access_control {
    managed                = true
    tenant_id              = data.azurerm_client_config.current.tenant_id
    admin_group_object_ids = var.admin_group_object_ids
  }

  oms_agent {
    log_analytics_workspace_id = var.log_analytics_workspace_id
  }

  default_node_pool {
    name                = "default"
    node_count          = var.min_nodes
    vm_size             = var.node_instance_type
    enable_auto_scaling = true
    max_count           = var.max_nodes
    min_count           = var.min_nodes
    vnet_subnet_id      = data.azurerm_subnet.pltf.id
    os_disk_size_gb     = var.node_disk_size
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.pltf.id]
  }

  kubelet_identity {
    client_id                 = azurerm_user_assigned_identity.agent_pool.client_id
    user_assigned_identity_id = azurerm_user_assigned_identity.agent_pool.id
    object_id                 = azurerm_user_assigned_identity.agent_pool.principal_id
  }

  lifecycle {
    ignore_changes = [
      default_node_pool["node_count"],
      location
    ]
  }

  depends_on = [
    azurerm_role_assignment.pltf,
    azurerm_role_assignment.azurerm_container_registry,
    azurerm_role_assignment.k8s_assign_identities,
    azurerm_role_assignment.azurerm_container_registry_agent_pool
  ]
}
