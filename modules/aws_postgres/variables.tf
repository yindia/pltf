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


variable "backup_retention_days" {
  description = "How many days to keep the backup retention"
  type        = number
  default     = 7
}


variable "restore_from_snapshot" {
  type    = string
  default = false
}

variable "engine_version" {
  type    = string
  default = "11.9"
}

variable "instance_class" {
  type    = string
  default = "db.t3.medium"
}

variable "safety" {
  type    = bool
  default = false
}

variable "multi_az" {
  type    = bool
  default = false
}

variable "extra_security_groups_ids" {
  type    = list(string)
  default = []
}

variable "create_global_database" {
  type    = bool
  default = false
}

variable "existing_global_database_id" {
  type    = string
  default = false
}

variable "database_name" {
  type = string
}

variable "db_aws_security_group" {
  type = string

}

variable "kms_key_alias" {
  type = string

}