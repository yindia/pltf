output "cache_auth_token" {
  value     = azurerm_redis_cache.pltf.primary_access_key
  sensitive = true
}

output "cache_host" {
  value = azurerm_private_endpoint.pltf.private_service_connection[0].private_ip_address
}
