output "vpc_id" {
  value = azurerm_virtual_network.pltf.id
}

output "vpc_name" {
  value = azurerm_virtual_network.pltf.name
}

output "private_subnet_id" {
  value = azurerm_subnet.pltf.id
}

output "private_subnet_name" {
  value = azurerm_subnet.pltf.name
}

output "acr_id" {
  value = azurerm_container_registry.acr.id
}

output "acr_name" {
  value = azurerm_container_registry.acr.name
}

output "acr_login_url" {
  value = azurerm_container_registry.acr.login_server
}

output "log_analytics_workspace_id" {
  value = azurerm_log_analytics_workspace.watcher.id
}

output "public_nat_ips" {
  value = azurerm_public_ip.pltf.ip_address
}
