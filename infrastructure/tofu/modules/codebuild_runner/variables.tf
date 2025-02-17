variable "repository_url" {
  type = string
}

variable "github_pat" {
  type      = string
  sensitive = true
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "security_group_ids" {
  type = list(string)
}

variable "use_fleet" {
  type    = bool
  default = false
}
