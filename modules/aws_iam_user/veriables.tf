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

variable "iam_policy" {
  description = "iam policy"
  type        = any
  default     = null
}

variable "extra_iam_policies" {
  type    = list(string)
  default = []
}

# tflint-ignore: terraform_unused_declarations
variable "links" {
  description = "Links for module"
  type        = any
  default     = []
}
