resource "azurerm_public_ip" "pltf" {
  name                = "pltf-${var.env_name}-nat-public-ip"
  location            = data.azurerm_resource_group.pltf.location
  resource_group_name = data.azurerm_resource_group.pltf.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_public_ip_prefix" "pltf" {
  name                = "pltf-${var.env_name}-nat-public-ip-prefix"
  location            = data.azurerm_resource_group.pltf.location
  resource_group_name = data.azurerm_resource_group.pltf.name
  prefix_length       = 30
}

resource "azurerm_nat_gateway" "pltf" {
  name                    = "pltf-${var.env_name}-nat-gateway"
  location                = data.azurerm_resource_group.pltf.location
  resource_group_name     = data.azurerm_resource_group.pltf.name
  sku_name                = "Standard"
  idle_timeout_in_minutes = 10
}

resource "azurerm_nat_gateway_public_ip_prefix_association" "pltf" {
  nat_gateway_id      = azurerm_nat_gateway.pltf.id
  public_ip_prefix_id = azurerm_public_ip_prefix.pltf.id
}

resource "azurerm_nat_gateway_public_ip_association" "pltf" {
  nat_gateway_id       = azurerm_nat_gateway.pltf.id
  public_ip_address_id = azurerm_public_ip.pltf.id
}

resource "azurerm_subnet_nat_gateway_association" "pltf" {
  subnet_id      = azurerm_subnet.pltf.id
  nat_gateway_id = azurerm_nat_gateway.pltf.id
}
