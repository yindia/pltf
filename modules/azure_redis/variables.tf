data "azurerm_resource_group" "main" {
  name = "pltf-${var.env_name}"
}

data "azurerm_subnet" "pltf" {
  name                 = "pltf-${var.env_name}-subnet"
  resource_group_name  = data.azurerm_resource_group.main.name
  virtual_network_name = "pltf-${var.env_name}"
}

variable "sku_name" {
  type    = string
  default = "Standard"
}

variable "family" {
  type    = string
  default = "C"
}

variable "capacity" {
  type    = number
  default = 2
}

variable "env_name" {
  description = "Env name"
  type        = string
}

variable "layer_name" {
  description = "Layer name"
  type        = string
}

variable "module_name" {
  description = "Module name"
  type        = string
}
