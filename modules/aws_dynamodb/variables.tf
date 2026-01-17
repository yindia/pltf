# tflint-ignore: terraform_unused_declarations
data "aws_caller_identity" "current" {}
# tflint-ignore: terraform_unused_declarations
data "aws_region" "current" {}

# tflint-ignore: terraform_unused_declarations
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

variable "read_capacity" {
  type    = number
  default = 20
}

variable "write_capacity" {
  type    = number
  default = 20
}

# tflint-ignore: terraform_unused_declarations
variable "billing_mode" {
  type    = string
  default = "PROVISIONED"
}

variable "hash_key" {
  type    = string
  default = ""
}

variable "range_key" {
  type    = string
  default = null
}

variable "attributes" {
  type = list(map(string))
}
