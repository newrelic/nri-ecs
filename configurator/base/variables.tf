variable "ssh_key_name" {
  description = "Which SSH key to use"
  default     = "coreint-ssh-key"
}

variable "vpc_id" {
  description = "Which VPC to use"
  type        = string
  default     = "vpc-0bbd46e8821f6a9a2"
}

variable "worker_nodes" {
  description = "How many worker nodes to spawn"
  type        = number
  default     = 1
}

variable "instance_type" { 
  description = "Worker node instance type (AWS)"
  type = string
  default = "t3.medium"
}