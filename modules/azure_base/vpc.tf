resource "azurerm_virtual_network" "pltf" {
  name                = "pltf-${var.env_name}"
  location            = data.azurerm_resource_group.pltf.location
  resource_group_name = data.azurerm_resource_group.pltf.name
  address_space       = [var.private_ipv4_cidr_block]
}

resource "azurerm_subnet" "pltf" {
  name                                           = "pltf-${var.env_name}-subnet"
  resource_group_name                            = data.azurerm_resource_group.pltf.name
  virtual_network_name                           = azurerm_virtual_network.pltf.name
  address_prefixes                               = [var.subnet_ipv4_cidr_block]
  enforce_private_link_endpoint_network_policies = true
}